'use client'; // クライアント（ブラウザ）側で動作

import Image from 'next/image';
import { CartItem } from '@/hooks/useCart';

interface CartItemCardProps {
  item: CartItem; // カートに入った個々の商品データ
  isEditable?: boolean; // カート情報の編集を可能にするか
  // 数量変更時のイベントハンドラ
  onUpdateQuantity?: (id: string, quantity: number) => void;
  // 削除時のイベントハンドラ
  onRemove?: (id: string) => void;
}

// カート商品カードコンポーネント
export default function CartItemCard({
  item,
  isEditable,
  onUpdateQuantity,
  onRemove
}: CartItemCardProps) {
  // 数量選択オプション（1～10）
  const quantityOptions = Array.from({ length: 10 }, (_, i) => (
    <option key={i + 1} value={i + 1}>{i + 1}</option>
  ));

  return (
    <div className="flex items-center gap-12 border border-gray-200 rounded p-8">
      <Image
        src={item.imageUrl ? `/uploads/${item.imageUrl}` : '/images/no-image.jpg'}
        alt={item.title}
        width={120}
        height={120}
        className="object-contain"
      />
      <div className="flex-grow">
        <h2 className="text-xl">{item.title}</h2>
        <p className="text-indigo-600 font-bold text-xl">
          ¥{item.price.toLocaleString()}
          <span className="text-base font-normal text-gray-500">（税込）</span>
        </p>
        {isEditable ? (
          <div className="flex items-center mt-2 gap-4">
            <label htmlFor={`quantity-${item.id}`} className="text-sm text-gray-700">
                数量
            </label>
            <select
                id={`quantity-${item.id}`}
                value={item.quantity}
                onChange={(e) => onUpdateQuantity && onUpdateQuantity(item.id, Number(e.target.value))}
                className="border border-gray-300 rounded-md px-3 py-1 focus:ring-2 focus:ring-indigo-500"
            >
                {quantityOptions}
            </select>
            <button
                onClick={() => onRemove && onRemove(item.id)}
                className="text-red-600 hover:underline cursor-pointer"
            >
                削除
            </button>
          </div>
        ) : (
          <p className="text-lg font-semibold">数量：{item.quantity}</p>
        )}
      </div>
      <p className="text-right font-semibold text-lg w-32">
        小計：¥{(item.price * item.quantity).toLocaleString()}
      </p>
    </div>
  );
}