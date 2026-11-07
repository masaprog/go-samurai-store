CREATE TABLE reviews (
 id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
 product_id INT NOT NULL,
 user_id INT NOT NULL,
 score INT NOT NULL,
 content TEXT NOT NULL,
 created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
 FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE, -- 商品（products）削除時にレビュー（reviews）も削除する
 FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE        -- ユーザー（users）削除時にレビュー（reviews）も削除する
);