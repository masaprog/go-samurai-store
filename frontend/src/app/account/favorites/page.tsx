'use client'; // クライアント（ブラウザ）側で動作

import Link from 'next/link';
import Image from 'next/image';
import { useState, useEffect } from 'react';
import { useCart, CartItem } from '@/hooks/useCart';
import { type ProductData } from '@/types/product';

// 商品データの型定義
type Product = Pick<ProductData, 'id' | 'name' | 'price' | 'image_url'>;

// お気に入り一覧ページ
export default function FavoritesPage() {
  const [favorites, setFavorites] = useState<Product[]>([]);
  const { addItem, isInCart } = useCart();

  // お気に入り一覧を取得
  useEffect(() => {
    const getFavorites = async () => {
      try {
        const res = await fetch('/api/favorites', { cache: 'no-store' });
        if (!res.ok) throw new Error('お気に入り一覧の取得に失敗しました。');
        const data: Product[] = await res.json();
        setFavorites(data);
      } catch (err) {
        console.error(err);
      }
    };
    // お気に入り一覧データを取得
    getFavorites();
  }, []);

  // カートボタン押下時のイベントハンドラ
  const handleCart = (item: Product) => {
    const cartItem: CartItem = {
      id: item.id.toString(),
      title: item.name,
      price: item.price,
      imageUrl: item.image_url ?? '',
      quantity: 1
    };
    addItem(cartItem);
  };

  // お気に入り解除ボタン押下時のイベントハンドラ
  const handleRemoveFavorite = async (productId: number) => {
    if (!confirm('本当にお気に入りから削除しますか？')) return;

    try { // お気に入り削除APIにDELETEリクエストを送信
      const res = await fetch(`/api/favorites/${productId}`, {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' }
      });

      if (res.ok) {
        // 成功したらstateから該当商品を除外
        setFavorites((prev) => prev.filter((item) => item.id !== productId));
      } else {
        const data = await res.json();
        alert(data.error || 'お気に入りの削除に失敗しました。');
      }
    } catch (err) {
      console.error('お気に入り削除エラー：', err);
      alert('通信エラーが発生しました。');
    }
  };

  return (
    <main className="container mx-auto px-4 py-8">
      <div className="my-4">
        <Link href="/account" className="text-indigo-600 hover:underline">
          ← マイページに戻る
        </Link>
      </div>
      <h1 className="text-center mb-6">お気に入り一覧</h1>
      {favorites.length === 0 ? (
        <div className="text-center py-16">
          <p className="text-gray-600">お気に入り商品がありません。</p>
          <Link href="/products" className="text-indigo-600 hover:underline">← 商品一覧でお気に入りを見つける</Link>
        </div>
      ) : (
        <div className="flex flex-col space-y-6">
          {favorites.map((item) => (
            <div key={item.id} className="flex items-center gap-6 border border-gray-200 rounded-lg p-6 shadow-sm bg-white">
              <Link href={`/products/${item.id}`} className="flex-shrink-0">
                <Image
                  src={item.image_url ? `/uploads/${item.image_url}` : '/images/no-image.jpg'}
                  alt={item.name}
                  width={120}
                  height={120}
                  className="object-contain mb-4"
                />
              </Link>

              <div className="flex-1 flex flex-col justify-between gap-4">
                <h2 className="text-xl">{item.name}</h2>
                <p className="text-indigo-600 font-bold text-lg">
                  ¥{item.price.toLocaleString()}
                  <span className="text-base font-normal text-gray-500">（税込）</span>
                </p>
              </div>

              <div className="flex flex-col items-end gap-8">
                <button
                  onClick={!isInCart(item.id.toString()) ? () => handleCart(item) : undefined}
                  disabled={isInCart(item.id.toString())}
                  className={`py-2 px-4 rounded-sm min-w-[150px] ${isInCart(item.id.toString())
                      ? 'bg-gray-400 text-white cursor-not-allowed'
                      : 'bg-indigo-500 hover:bg-indigo-600 text-white'
                    }`}
                >
                  {isInCart(item.id.toString()) ? '追加済み' : 'カートに追加'}
                </button>
                <button
                  onClick={() => handleRemoveFavorite(item.id)}
                  className="py-2 px-4 rounded-sm min-w-[150px] bg-rose-500 hover:bg-rose-600 text-white"
                >
                  お気に入り解除
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </main>
  );
}