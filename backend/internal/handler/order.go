package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/checkout/session"
	"github.com/stripe/stripe-go/v83/webhook"

	"github.com/masaprog/go-samurai-store/backend/internal/database"
)

// --- 1. 型定義（struct） ---

// カート内商品の構造体
type CartItem struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Price    int    `json:"price"`
	ImageURL string `json:"imageUrl"`
	Quantity int    `json:"quantity"`
}

// Stripe Checkoutセッション作成リクエストの構造体
type CheckoutRequest struct {
	Items   []CartItem `json:"items"`
	Address string     `json:"address" binding:"required"`
}

// 商品情報レスポンスの構造体（在庫チェック、価格計算用）
type productForOrder struct {
	ID    int
	Name  string
	Price int
	Stock int
}

// 注文明細の構造体
type OrderItem struct {
	ProductName string `json:"productName"`
	Quantity    int    `json:"quantity"`
	UnitPrice   int    `json:"unitPrice"`
}

// 注文情報レスポンスの構造体
type OrderData struct {
	ID            int         `json:"id"`
	TotalPrice    int         `json:"totalPrice"`
	Status        string      `json:"status"`        // ENUMはstringとして扱う
	PaymentStatus string      `json:"paymentStatus"` // ENUMはstringとして扱う
	CreatedAt     time.Time   `json:"createdAt"`
	Items         []OrderItem `json:"items"` // 注文明細のスライス
}

// 注文と注文明細の結合レコードの構造体
type orderJoinRecord struct {
	ID            int
	TotalPrice    int
	Status        string
	PaymentStatus string
	CreatedAt     time.Time
	ProductName   string
	Quantity      int
	UnitPrice     int
}

// 注文ステータス（ordersテーブルのENUMに対応）
const (
	OrderStatusPending    = "未処理"
	OrderStatusProcessing = "処理中"
	// 他のステータスも必要になったら定義する
)

// 決済ステータス（ordersテーブルのENUMに対応）
const (
	PaymentStatusUnpaid = "未決済"
	PaymentStatusPaid   = "決済成功"
	// 他のステータスも必要になったら定義する
)

// 送料（円）
const shippingCost = 500

// Stripeを初期化する
func init() {
	// 環境変数からStripeのシークレットキーを取得して設定する
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	if stripe.Key == "" {
		log.Println("警告: 環境変数 STRIPE_SECRET_KEY が設定されていません")
	}
}

// --- 2. ハンドラ定義 ---

// Stripe Checkoutセッションを作成する関数（POST /api/orders/checkout）
func CreateCheckoutSessionHandler(c *gin.Context) {
	// HTTPリクエストボディをCheckoutRequest構造体にバインドする
	var req CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Checkoutリクエストのバインドエラー: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "リクエスト内容が正しくありません"})
		return
	}
	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "カートが空です"})
		return
	}

	// HTTPリクエストのコンテキストからユーザー情報を取得する
	userClaims, exists := c.Get("user")
	if !exists {
		log.Println("CreateCheckoutSessionHandler: ユーザー情報がコンテキストにありません")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
		return
	}
	claims, ok := userClaims.(*JWTCustomClaims)
	if !ok {
		log.Println("CreateCheckoutSessionHandler: ユーザー情報の型アサーションに失敗しました")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}
	userID := claims.UserID
	userEmail := claims.Email

	// データベース接続を取得する
	db := database.GetDB()

	productIDs := []any{} // カート内の商品IDリストを格納するスライス
	quantityMap := map[int]int{}  // 商品IDごとの数量を保持するマップ
	for _, item := range req.Items {
		id, err := strconv.Atoi(item.ID)
		if err != nil {
			log.Printf("不正な商品ID形式です: %s", item.ID)
			c.JSON(http.StatusBadRequest, gin.H{"error": "カート内の商品IDが不正です"})
			return
		}
		productIDs = append(productIDs, id)
		quantityMap[id] = item.Quantity
	}

	// データベースから対象商品の情報を取得する（ID IN ...）
	// 商品IDの数に合わせてプレースホルダを動的に生成する
	placeholders := strings.Repeat("?,", len(productIDs)-1) + "?"
	query := fmt.Sprintf(`SELECT id, name, price, stock FROM products WHERE id IN (%s)`, placeholders)

	rows, err := db.Query(query, productIDs...)
	if err != nil {
		log.Printf("在庫チェック時の商品取得エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}
	defer rows.Close()

	dbProducts := make(map[int]productForOrder) // カート内の商品IDをキーにしたマップ
	shortageItems := []string{}                 // 在庫不足の商品名リスト
	for rows.Next() {
		var p productForOrder
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock); err != nil {
			log.Printf("在庫チェック時の商品スキャンエラー: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
			return
		}
		dbProducts[p.ID] = p
		// 在庫チェックを行う
		if p.Stock < quantityMap[p.ID] {
			shortageItems = append(shortageItems, p.Name)
		}
	}
	if err = rows.Err(); err != nil {
		log.Printf("在庫チェック時の行エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}

	// カート内の商品IDがデータベースに存在しない場合はエラーを返す
	if len(dbProducts) != len(productIDs) {
		log.Printf("カート内の商品が見つかりません: %v vs %v", len(dbProducts), len(productIDs))
		c.JSON(http.StatusBadRequest, gin.H{"error": "カート内の一部商品が見つかりませんでした"})
		return
	}
	// 在庫不足の商品があればエラーを返す
	if len(shortageItems) > 0 {
		errMsg := fmt.Sprintf("在庫不足の商品があります: %s", strings.Join(shortageItems, "、"))
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	// 注文情報をデータベースに登録するトランザクションを開始する
	tx, err := db.Begin()
	if err != nil {
		log.Printf("トランザクション開始エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}
	defer tx.Rollback() // 関数終了時にロールバックする（コミットされなかった場合）

	// 合計金額を計算する
	totalPrice := 0
	for _, item := range req.Items {
		id, _ := strconv.Atoi(item.ID)
		product := dbProducts[id]
		totalPrice += product.Price * item.Quantity
	}
	totalPrice += shippingCost // 送料を加算する

	// ordersテーブルに登録する
	orderQuery := `
		INSERT INTO orders (user_id, total_price, status, payment_status, shipping_address)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := tx.Exec(orderQuery, userID, totalPrice, OrderStatusPending, PaymentStatusUnpaid, req.Address)
	if err != nil {
		log.Printf("注文登録エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注文の登録に失敗しました"})
		return
	}
	orderID, err := result.LastInsertId() // 登録した注文IDを取得する
	if err != nil {
		log.Printf("注文ID取得エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注文の登録に失敗しました"})
		return
	}

	// order_itemsテーブルに登録する
	itemQuery := `
		INSERT INTO order_items (order_id, product_id, product_name, quantity, unit_price)
		VALUES (?, ?, ?, ?, ?)
	`
	stmt, err := tx.Prepare(itemQuery) // プリペアードステートメントを使う
	if err != nil {
		log.Printf("注文明細 INSERT準備エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注文の登録に失敗しました"})
		return
	}
	defer stmt.Close()

	for _, item := range req.Items {
		id, _ := strconv.Atoi(item.ID)
		product := dbProducts[id]
		_, err := stmt.Exec(orderID, product.ID, product.Name, item.Quantity, product.Price)
		if err != nil {
			log.Printf("注文明細 INSERT実行エラー: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "注文の登録に失敗しました"})
			return
		}
	}

	// Stripe Checkoutセッションを作成する
	lineItems := []*stripe.CheckoutSessionLineItemParams{}
	for _, item := range req.Items {
		id, _ := strconv.Atoi(item.ID)
		product := dbProducts[id]
		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String(string(stripe.CurrencyJPY)),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe.String(product.Name),
				},
				UnitAmount: stripe.Int64(int64(product.Price)),
			},
			Quantity: stripe.Int64(int64(item.Quantity)),
		})
	}
	// 送料を追加する
	lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
		PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
			Currency: stripe.String(string(stripe.CurrencyJPY)),
			ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
				Name: stripe.String("送料"),
			},
			UnitAmount: stripe.Int64(int64(shippingCost)),
		},
		Quantity: stripe.Int64(1),
	})

	// フロントエンドのベースURL（リダイレクト先を構築するため）
	// 本番環境では.envファイルなどから取得する
	frontendBaseURL := "http://localhost:3000"

	params := &stripe.CheckoutSessionParams{
		LineItems:     lineItems,
		Mode:          stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:    stripe.String(fmt.Sprintf("%s/account?session_id={CHECKOUT_SESSION_ID}", frontendBaseURL)), // 成功時のリダイレクト先
		CancelURL:     stripe.String(fmt.Sprintf("%s/order-confirm", frontendBaseURL)),                            // キャンセル時のリダイレクト先
		CustomerEmail: stripe.String(userEmail),                                                                   // 注文者のメールアドレス
		Metadata: map[string]string{ // Stripe Webhookで使う情報
			"orderId": strconv.FormatInt(orderID, 10),
			"userId":  strconv.Itoa(userID),
		},
	}

	s, err := session.New(params)
	if err != nil {
		log.Printf("Stripe Checkout セッション作成エラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "決済ページの生成に失敗しました"})
		return
	}

	// トランザクションをコミットする
	if err := tx.Commit(); err != nil {
		log.Printf("トランザクションコミットエラー: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注文の確定に失敗しました"})
		return
	}

	// Stripe CheckoutセッションのURLをレスポンスとして返す
	c.JSON(http.StatusOK, gin.H{"url": s.URL})
}

// Stripe Webhookイベントを処理する関数（POST /api/orders/webhook）
func StripeWebhookHandler(c *gin.Context) {
	// Stripe Webhookの署名シークレットを取得する
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if webhookSecret == "" {
		log.Println("警告: 環境変数 STRIPE_WEBHOOK_SECRET が設定されていません")
		c.Status(http.StatusInternalServerError)
		return
	}

	// HTTPリクエストボディを読み込む
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Webhook リクエストボディ読み込みエラー: %v", err)
		c.Status(http.StatusBadRequest)
		return
	}

	// 署名シークレットを検証し、イベントを構築する
	signature := c.GetHeader("Stripe-Signature")
	event, err := webhook.ConstructEventWithOptions(
		payload,
		signature,
		webhookSecret,
		webhook.ConstructEventOptions{
			IgnoreAPIVersionMismatch: true, // APIバージョンの違いによるエラーは無視する
		},
	)
	if err != nil {
		log.Printf("Webhook 署名検証エラー: %v", err)
		c.Status(http.StatusBadRequest) // 署名が無効
		return
	}

	// checkout.session.completed（決済成功）イベントの処理を行う
	if event.Type == "checkout.session.completed" {
		var session stripe.CheckoutSession
		// イベントデータを CheckoutSession オブジェクトにアンマーシャル
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			log.Printf("Webhook イベントデータ解析エラー: %v", err)
			c.Status(http.StatusBadRequest)
			return
		}

		// メタデータからorderIdとuserIdを取得する
		orderIDStr, okOrderId := session.Metadata["orderId"]
		userIDStr, okUserId := session.Metadata["userId"]
		if !okOrderId || !okUserId {
			log.Printf("Webhook メタデータ不足: orderId=%s, userId=%s", orderIDStr, userIDStr)
			c.Status(http.StatusBadRequest)
			return
		}
		orderID, errOrderId := strconv.ParseInt(orderIDStr, 10, 64)
		userID, errUserId := strconv.Atoi(userIDStr)
		if errOrderId != nil || errUserId != nil {
			log.Printf("Webhook メタデータ形式エラー: orderId=%s, userId=%s", orderIDStr, userIDStr)
			c.Status(http.StatusBadRequest)
			return
		}

		// データベースを更新するトランザクションを開始する
		log.Printf("Webhook 受信 (checkout.session.completed): OrderID=%d, UserID=%d", orderID, userID)
		db := database.GetDB()
		tx, err := db.Begin()
		if err != nil {
			log.Printf("Webhook: トランザクション開始エラー: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// 注文ステータスを更新する
		// （冪等性を考慮し、すでに決済ステータスが「決済成功」なら何もしない）
		updateOrderQuery := `
			UPDATE orders
			SET status = ?, payment_status = ?
			WHERE id = ? AND user_id = ? AND payment_status != ?
		`
		result, err := tx.Exec(updateOrderQuery, OrderStatusProcessing, PaymentStatusPaid, orderID, userID, PaymentStatusPaid)
		if err != nil {
			log.Printf("Webhook: 注文ステータス更新エラー (OrderID=%d): %v", orderID, err)
			c.Status(http.StatusInternalServerError)
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			log.Printf("Webhook: 注文ステータス更新対象なし (OrderID=%d, UserID=%d)", orderID, userID)
			// Stripe側には成功を返す（イベントは受け取ったため）
			c.Status(http.StatusOK)
			return
		}

		// 注文された商品と数量を取得する
		itemsQuery := "SELECT product_id, quantity FROM order_items WHERE order_id = ?"
		rows, err := tx.Query(itemsQuery, orderID)
		if err != nil {
			log.Printf("Webhook: 注文明細取得エラー (OrderID=%d): %v", orderID, err)
			c.Status(http.StatusInternalServerError)
			return
		}
		type orderItem struct{ ProductID, Quantity int }
		itemsToUpdate := []orderItem{}
		for rows.Next() {
			var item orderItem
			if err := rows.Scan(&item.ProductID, &item.Quantity); err != nil {
				log.Printf("Webhook: 注文明細スキャンエラー (OrderID=%d): %v", orderID, err)
				c.Status(http.StatusInternalServerError)
				rows.Close()
				return
			}
			itemsToUpdate = append(itemsToUpdate, item)
		}
		rows.Close()
		if err = rows.Err(); err != nil {
			log.Printf("Webhook: 注文明細行エラー (OrderID=%d): %v", orderID, err)
			c.Status(http.StatusInternalServerError)
			return
		}

		// 在庫数を更新するSQL文を準備する
		updateStockQuery := `
			UPDATE products
			SET stock = stock - ?, sales_count = sales_count + ?
			WHERE id = ? AND stock >= ? -- 在庫が十分あることを再確認 (念のため)
		`
		stmt, err := tx.Prepare(updateStockQuery)
		if err != nil {
			log.Printf("Webhook: 在庫更新準備エラー: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		// 各商品の在庫数を更新する
		for _, item := range itemsToUpdate {
			res, err := stmt.Exec(item.Quantity, item.Quantity, item.ProductID, item.Quantity)
			if err != nil {
				log.Printf("Webhook: 在庫更新実行エラー (ProductID=%d): %v", item.ProductID, err)
				c.Status(http.StatusInternalServerError)
				return
			}
			affected, _ := res.RowsAffected()
			if affected == 0 {
				log.Printf("Webhook: 在庫不足または商品が見つかりません (ProductID=%d)", item.ProductID)
				c.Status(http.StatusInternalServerError)
				return
			}
		}

		// トランザクションをコミットする
		if err := tx.Commit(); err != nil {
			log.Printf("Webhook: トランザクションコミットエラー (OrderID=%d): %v", orderID, err)
			c.Status(http.StatusInternalServerError)
			return
		}

		log.Printf("Webhook 処理成功 (OrderID=%d)", orderID)
	} else {
		// checkout.session.completed以外のイベントは無視する（ログには残す）
		log.Printf("Webhook受信（無視するイベント）: %s", event.Type)
	}

	// Stripe Webhookイベントを正常に受信したことをStripe側に伝える
	c.Status(http.StatusOK)
}

// 注文履歴を取得する関数
func GetOrdersHandler(c *gin.Context) {
	// HTTPリクエストのコンテキストからユーザー情報を取得する
	userClaims, exists := c.Get("user")
	if !exists {
		log.Println("GetOrdersHandler: ユーザー情報がコンテキストにありません")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
		return
	}
	claims, ok := userClaims.(*JWTCustomClaims) // auth.goファイルで定義した型
	if !ok {
		log.Println("GetOrdersHandler: ユーザー情報の型アサーションに失敗しました")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}
	userID := claims.UserID

	// データベースから注文データを取得する
	// ordersテーブルとorder_itemsテーブルを結合し、ユーザーIDで絞り込み、作成日時・注文明細ID順で並べ替える
	db := database.GetDB()
	query := `
		SELECT
			o.id, o.total_price, o.status, o.payment_status, o.created_at,
			oi.product_name, oi.quantity, oi.unit_price
		FROM orders AS o
		JOIN order_items AS oi ON o.id = oi.order_id
		WHERE o.user_id = ?
		ORDER BY o.created_at DESC, oi.id ASC
	`
	rows, err := db.Query(query, userID)
	if err != nil {
		log.Printf("注文履歴取得エラー (UserID=%d): %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}
	defer rows.Close()

	// 注文データを格納するマップ
	ordersMap := make(map[int]*OrderData)
	// スキャン結果を一時的に格納する変数
	var record orderJoinRecord

	// SQL文の実行結果をスキャンする
	for rows.Next() {
		err := rows.Scan(
			&record.ID, &record.TotalPrice, &record.Status, &record.PaymentStatus, &record.CreatedAt,
			&record.ProductName, &record.Quantity, &record.UnitPrice,
		)
		if err != nil {
			log.Printf("注文履歴スキャンエラー (UserID=%d): %v", userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
			return
		}

		// マップを使って注文IDをキーにデータをまとめる（キー: order.id, 値: OrderData構造体へのポインタ）
		orderData, exists := ordersMap[record.ID]
		if !exists {
			// マップにこの注文IDがまだなければ、新しくOrderData構造体へのポインタを作成して追加する
			orderData = &OrderData{
				ID:            record.ID,
				TotalPrice:    record.TotalPrice,
				Status:        record.Status,
				PaymentStatus: record.PaymentStatus,
				CreatedAt:     record.CreatedAt,
				Items:         []OrderItem{}, // 空のスライスで初期化する
			}
			ordersMap[record.ID] = orderData
		}

		// 現在のレコードの注文明細（OrderItem構造体）を作成する
		item := OrderItem{
			ProductName: record.ProductName,
			Quantity:    record.Quantity,
			UnitPrice:   record.UnitPrice,
		}
		// Itemsフィールドに注文明細を追加する
		orderData.Items = append(orderData.Items, item)
	}
	if err = rows.Err(); err != nil {
		log.Printf("注文履歴行エラー (UserID=%d): %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "サーバーエラーが発生しました"})
		return
	}

	// マップから最終的なレスポンス用のスライスを作成する
	orders := make([]OrderData, 0, len(ordersMap))
	for _, order := range ordersMap {
		orders = append(orders, *order) // ポインタの中身をスライスに追加する
	}

	// 作成日時が新しい順で並べ替える
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].CreatedAt.After(orders[j].CreatedAt)
	})

	// JSONとしてレスポンスを返す
	c.JSON(http.StatusOK, gin.H{"orders": orders})
}