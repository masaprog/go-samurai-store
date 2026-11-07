'use client'; // クライアント（ブラウザ）側で動作

import ProductCard, { type ProductCardProps } from './ProductCard';

// 商品リストコンポーネントに渡すデータ（props）の型定義
interface ProductListProps {
  products: ProductCardProps[]; // ProductCardの配列
}

// 商品リストの共通コンポーネント
export default function ProductList( { products }: ProductListProps ) {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4">
      {products.length === 0 ? (
        <p className="col-span-full text-center text-gray-500 text-lg py-10">
          表示する商品がありません。
        </p>
      ) : (
        // 商品データを1つずつProductCardで表示
        products.map((product) => (
          <ProductCard key={product.id} {...product} imageSize={300} showCartButton />
        ))
      )}
    </div>
  );
}