'use client'; // クライアント（ブラウザ）側で動作

import Link from 'next/link';
import { useCart, CartItem } from '@/hooks/useCart';
import CartItemCard from '@/components/CartItemCard';

// カートページ
export default function CartPage() {
  // カート情報と制御関数を取得
  const { cartItems, removeItem, updateQuantity, totalPrice } = useCart();

  return (
    <main className="container mx-auto px-4 py-8">
      <h1 className="text-center mb-6">ショッピングカート</h1>

      {cartItems.length === 0 ? (
        <div className="text-center py-16">
          <p className="text-gray-600">カートに商品がありません。</p>
          <Link href="/products" className="text-indigo-600 hover:underline">← 商品一覧に戻る</Link>
        </div>
      ) : (
        <>
          <div className="flex flex-col space-y-6">
            {cartItems.map((item: CartItem) => (
              <CartItemCard
                key={item.id}
                item={item}
                isEditable
                onUpdateQuantity={updateQuantity}
                onRemove={removeItem}
              />
            ))}
          </div>

          <div className="mt-8 flex justify-between items-center border-t border-gray-300 pt-6">
            <div className="flex flex-col">
              <p className="text-2xl font-bold">合計金額：¥{totalPrice.toLocaleString()}</p>
              <p className="text-gray-500">※価格はすべて税込表記です。</p>
            </div>
            <Link
              href="/order-confirm"
              className="bg-indigo-500 hover:bg-indigo-600 text-white py-2 px-6 rounded"
            >
              購入手続きへ
            </Link>
          </div>
        </>
      )}
    </main>
  );
}