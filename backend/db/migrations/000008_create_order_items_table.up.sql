CREATE TABLE order_items (
 id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
 order_id INT NOT NULL,
 product_id INT NOT NULL,
 product_name VARCHAR(255) NOT NULL,
 quantity INT NOT NULL,
 unit_price INT NOT NULL,
 CONSTRAINT fk_order_items_order FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE, -- 注文（orders）削除時に明細（order_items）も削除する
 CONSTRAINT fk_order_items_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT -- 商品（products）が削除されたらエラーにする
);