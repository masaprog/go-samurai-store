CREATE TABLE orders (
 id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
 user_id INT NOT NULL,
 total_price INT NOT NULL,
 status ENUM('未処理', '処理中', '出荷済み', '完了', 'キャンセル', '返金済み') NOT NULL DEFAULT '未処理',
 payment_status ENUM('未決済', '決済処理中', '決済成功', '決済失敗', '返金処理中', '返金済み') NOT NULL DEFAULT '未決済',
 shipping_address TEXT NOT NULL,
 created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
 updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
 CONSTRAINT fk_orders_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE -- ユーザー（users）削除時に注文（orders）も削除する
);