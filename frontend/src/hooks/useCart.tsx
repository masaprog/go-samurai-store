'use client'; // クライアント（ブラウザ）側で動作

import { useState, useEffect, useCallback, useMemo, useContext, createContext } from 'react';

// カート内の商品アイテムの型定義
export interface CartItem {
  id: string;
  title: string;
  price: number;
  imageUrl?: string;
  quantity: number;
}

// コンテクストの型定義
interface CartContextType {
  cartItems: CartItem[];
  addItem: (product: Omit<CartItem, 'quantity'>, addQuantity?: number) => void;
  removeItem: (id: string) => void;
  clearCart: () => void;
  updateQuantity: (id: string, newQuantity: number) => void;
  isInCart: (id: string) => boolean;
  totalPrice: number;
  totalQuantity: number;
}

// コンテクストを生成
const CartContext = createContext<CartContextType | undefined>(undefined);

// ローカルストレージのキー名
const STORAGE_KEY = 'cartItems';

// カート管理用のカスタムフック
export function useCart() {
  // カートのコンテクストを取得
  const context = useContext(CartContext);
  if (context === undefined) {
    throw new Error('useCartはCartProviderの内部でのみ使用してください。');
  }
  return context;
}

// コンテクストを提供するコンポーネント
export function CartProvider({ children }: { children: React.ReactNode }) {
  // カート内の状態を管理する配列
  const [cartItems, setCartItems] = useState<CartItem[]>([]);
  const [isClient, setIsClient] = useState(false);

  // 初回マウント時にクライアントであることを記録し、ローカルストレージから読み込む
  useEffect(() => {
    setIsClient(true);
    try {
      const storedCart = localStorage.getItem(STORAGE_KEY);
      if (storedCart) {
        setCartItems(JSON.parse(storedCart));
      }
    } catch (error) {
      console.error('カート読み込みエラー：', error);
    }
  }, []); // 初回マウント時のみ実行する

  // ローカルストレージへの保存もクライアントサイドでのみ行う
  useEffect(() => {
    // 状態変数isClientがtrue（マウント後）の場合のみローカルストレージに保存する
    if (isClient) {
      try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(cartItems));
      } catch (error) {
        console.error('カート保存エラー：', error);
      }
    }
  }, [cartItems, isClient]); // 状態変数cartItemsまたはisClientが変更されたら実行する

  // カートに商品を追加する関数
  const addItem = useCallback((targetProduct: Omit<CartItem, 'quantity'>, addQuantity = 1) => {
    setCartItems((prevCart) => {
      // すでにカート内に存在する商品かチェック
      const isInCart = prevCart.find((item) => item.id === targetProduct.id);

      if (isInCart) { // 既存商品なら数量を増やす
        return prevCart.map((item) =>
          item.id === targetProduct.id ? { ...item, quantity: item.quantity + addQuantity } : item
        );
      } else { // 新規商品なら追加
        return [...prevCart, { ...targetProduct, quantity: addQuantity }];
      }
    });
  }, []);

  // カートから商品を削除する関数
  const removeItem = useCallback((targetId: string) => {
    setCartItems(prevCart => prevCart.filter(p => p.id !== targetId));
  }, []);

  // カートを空にする関数
  const clearCart = useCallback(() => {
    setCartItems([]);

    if (typeof window !== 'undefined') {
      try {
        localStorage.removeItem(STORAGE_KEY);
      } catch (error) {
        console.error('カートのローカルストレージクリアエラー:', error);
      }
    }
  }, []);

  // 商品の数量を変更する関数
  const updateQuantity = useCallback((targetId: string, newQuantity: number) => {
    // カート配列
    setCartItems((prevCart) => {
      if (newQuantity > 0) { // 数量が1以上の場合のみ更新
        return prevCart.map((item) =>
          item.id === targetId ? { ...item, quantity: newQuantity } : item
        );
      } else { // 数量が0以下なら商品をカートから削除
        return prevCart.filter(item => item.id !== targetId);
      }
    });
  }, []);

  // 指定IDの商品がカートに含まれているか
  const isInCart = useCallback((targetId: string) => {
    return cartItems.some(p => p.id === targetId);
  }, [cartItems]);

  // カート内商品の合計金額
  const totalPrice = useMemo(
    () => cartItems.reduce((sum, item) => sum + item.price * item.quantity, 0),
    [cartItems]
  );

  // カート内商品の総数量
  const totalQuantity = useMemo(
    () => cartItems.reduce((sum, item) => sum + item.quantity, 0),
    [cartItems]
  );

  // コンテクストに渡す値
  const value = useMemo(() => ({
    cartItems: isClient ? cartItems : [],
    addItem,
    removeItem,
    clearCart,
    updateQuantity,
    isInCart,
    totalPrice: isClient ? totalPrice : 0,
    totalQuantity: isClient ? totalQuantity : 0,
  }), [
    isClient, cartItems, addItem, removeItem, clearCart, updateQuantity, isInCart, totalPrice, totalQuantity
  ]);

  return (
    <CartContext.Provider value={value}>
      {children}
    </CartContext.Provider>
  );
}