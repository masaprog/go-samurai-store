CREATE TABLE favorites (
 id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
 product_id INT NOT NULL,
 user_id INT NOT NULL,
 created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
 UNIQUE KEY unique_user_product (user_id, product_id), -- 同じユーザーが同じ商品を複数お気に入り登録できないようにUNIQUE制約を追加する
 FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE, -- 商品（products）削除時にお気に入り（favorites）も削除する
 FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE        -- ユーザー（users）削除時にお気に入り（favorites）も削除する
);