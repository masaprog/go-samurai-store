'use client'; // クライアント側で動作

import { useState } from 'react';

// お気に入り制御用コンポーネントに渡すデータ（props）の型定義
interface FavoriteControlsProps {
  productId: number;
  initialFavorite: boolean;
}

// お気に入り制御用コンポーネント
export default function FavoriteControls({ productId, initialFavorite }: FavoriteControlsProps) {
  // お気に入り状態
  const [isFavorite, setIsFavorite] = useState(initialFavorite);
  // 処理状態
  const [loading, setLoading] = useState(false);

  // お気に入りリンク押下時のイベントハンドラ
  const handleToggleFavorite = async () => {
    setLoading(true); // 処理中はリンクを無効化
    try {
      const method = isFavorite ? 'DELETE' : 'POST';
      const body = method === 'POST' ? JSON.stringify({ productId }) : null;
      const url = method === 'DELETE' ? `/api/favorites/${productId}` : `/api/favorites`;

      // お気に入り登録／削除APIにPOSTまたはDELETEリクエストを送信
      const res = await fetch(url, {
        method: method,
        body: body,
        headers: { 'Content-Type': 'application/json' }
      });

      if (res.ok) {
        // 成功したら状態を反転
        setIsFavorite(!isFavorite);
      } else {
        const data = await res.json();
        alert(data.error || '操作に失敗しました。');
      }
    } catch (err) {
      console.error('お気に入り操作エラー：', err);
      alert('通信エラーが発生しました。');
    } finally {
      setLoading(false);
    }
  };

  return (
    <button
      onClick={handleToggleFavorite} disabled={loading}
      className="text-teal-800 hover:underline cursor-pointer"
      style={{ fontFamily: 'sans-serif' }}
    >
      {isFavorite ? '♥ お気に入り解除' : '♡ お気に入り追加'}
    </button>
  );
}