'use client'; // クライアント（ブラウザ）側で動作

import { useRouter } from 'next/navigation';
import { useState } from 'react';
import { useCart, CartItem } from '@/hooks/useCart';

// カート制御用コンポーネントに渡すデータ（props）の型定義
type CartControlsProps = {
  cartItem: Omit<CartItem, 'quantity'>;
  stock: number;
  loggedIn: boolean;
};

// カート制御用コンポーネント（商品詳細ページ専用）
export default function CartControls({ cartItem, stock, loggedIn }: CartControlsProps) {
  const router = useRouter();
  const { addItem, isInCart } = useCart();
  const inCart = isInCart(cartItem.id); // カートに追加済みか

  // 数量の選択状態
  const [quantity, setQuantity] = useState(1);

  // 数量セレクトボックスのオプションを生成
  const quantityOptions = [];
  for (let i = 1; i <= Math.min(stock, 10); i++) { // 最大10個まで
    quantityOptions.push(<option key={i} value={i}>{i}</option>);
  }

  // カートボタン押下時のイベントハンドラ
  const handleCart = () => {
    // 選択されている数量をカートに追加
    addItem({...cartItem}, quantity);
  };

  // 購入手続きボタン押下時のイベントハンドラ
  const handleOrder = () => {
    // カートに商品が追加されていない場合のみ追加
    if (!inCart) {
      addItem({ ...cartItem }, quantity);
    }

    // 注文確認ページへ遷移
    router.push('/order-confirm');
  };

  return (
    <div className="space-y-6 mt-8">
      {stock > 0 && (
        <div className="flex items-end gap-4">
          <div className="flex flex-col">
            <label htmlFor="quantity" className="block text-sm text-gray-700">
              数量
            </label>
            <select
              id="quantity" name="quantity" value={quantity}
              onChange={(e) => setQuantity(Number(e.target.value))}
              className="border border-gray-300 rounded-md px-4 py-2 w-24 focus:ring-2 focus:ring-indigo-500"
            >
              {quantityOptions}
            </select>
          </div>

          <button
            onClick={!inCart ? handleCart : undefined}
            disabled={inCart}
            className={`py-2 px-4 rounded-sm min-w-[130px] ${
              inCart
                ? 'bg-gray-400 text-white cursor-not-allowed'
                : 'bg-indigo-500 hover:bg-indigo-600 text-white'
            }`}
          >
            {inCart ? '追加済み' : 'カートに追加'}
          </button>

          {loggedIn && (
            <button
              onClick={handleOrder}
              className="border border-indigo-500 text-indigo-500 py-2 px-4 rounded-sm hover:bg-indigo-50"
            >
              購入手続きへ
            </button>
          )}
        </div>
      )}
    </div>
  );
}