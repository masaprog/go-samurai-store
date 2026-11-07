'use client'; // クライアント（ブラウザ）側で動作

import { useRouter, usePathname } from 'next/navigation';
import { useState } from 'react';

// レビュー制御用コンポーネントに渡すデータ（props）の型定義
type ReviewControlsProps = {
  productId: number;
  loggedIn: boolean;
};

// レビュー制御用コンポーネント（商品詳細ページ専用）
export default function ReviewControls({ productId, loggedIn }: ReviewControlsProps) {
  const router = useRouter();
  const pathname = usePathname(); // 現在のパスを取得

  // レビュー投稿フォームの状態管理
  const [rating, setRating] = useState(0); // 評価（星の数）
  const [clickedRating, setClickedRating] = useState(0); // クリックして確定した星
  const [content, setContent] = useState(''); // レビュー内容
  const [submitting, setSubmitting] = useState(false); // 送信中フラグ
  const [errorMessage, setErrorMessage] = useState(''); // エラーメッセージ
  const [successMessage, setSuccessMessage] = useState<string | null>(null); // 成功メッセージ

  // 評価の星をクリックした時のイベントハンドラ
  const handleScoreClick = (selectedScore: number) => {
    setRating(selectedScore);
    setClickedRating(selectedScore);
  };

  // レビュー投稿ボタン押下時のイベントハンドラ
  const handleSubmitReview = async (e: React.FormEvent) => {
    e.preventDefault(); // デフォルトの送信動作をキャンセル
    setErrorMessage('');
    setSuccessMessage('');

    // 入力データのバリデーション
    if (rating === 0 || !content.trim()) {
      setErrorMessage('すべての項目を入力してください。');
      return;
    }
    // ログイン必須
    if (!loggedIn) {
      setErrorMessage('レビューの投稿にはログインが必要です。');
      return;
    }

    // 投稿中に表示更新
    setSubmitting(true);

    try { // レビュー登録APIにPOSTリクエストを送信
      const res = await fetch(`/api/products/${productId}/reviews`, {
        method: 'POST',
        body: JSON.stringify({ rating, content }),
        headers: { 'Content-Type': 'application/json' }
      });

      if (res.ok) { // 投稿成功時は再更新
        setSuccessMessage('レビューが正常に投稿されました！');
        // フォームをリセット
        setRating(0);
        setContent('');
        // 再表示して最新のレビューを反映
        router.refresh();
      } else { // 投稿失敗時はエラー情報を表示
        const data = await res.json();
        setErrorMessage(data.error || 'レビューの投稿に失敗しました。');
      }
    } catch {
      setErrorMessage('通信エラーが発生しました。');
    } finally {
      setSubmitting(false);
    }
  };

  // 未ログイン時のボタンクリックハンドラ
  const handleLoginRedirect = () => {
    // 現在のページURLをリダイレクト先に設定
    const redirectUrl = encodeURIComponent(pathname);
    router.push(`/login?redirect=${redirectUrl}`); // ログインページへ遷移
  };

  return (
    <div>
      <h2 className="mt-2">レビューを投稿する</h2>
      {!loggedIn ? (
        <div className="text-center py-4">
          <p className="text-gray-600 mb-4">レビューの投稿にはログインが必要です。</p>
          <button
            onClick={handleLoginRedirect}
            className="bg-indigo-500 hover:bg-indigo-600 text-white py-2 px-6 rounded-md shadow-md"
          >
            ログインしてレビューを投稿
          </button>
        </div>
      ) : (
        <form onSubmit={handleSubmitReview} className="space-y-4">
          <div>
            <label htmlFor="score" className="block text-gray-700 font-semibold mb-2">評価</label>
            <div className="flex text-2xl text-yellow-500">
              {[1, 2, 3, 4, 5].map((s) => (
                <span
                  key={s} className="cursor-pointer"
                  onClick={() => handleScoreClick(s)}
                  onMouseEnter={() => setRating(s)} // カーソルが合わさった星に暫定更新
                  onMouseLeave={() => setRating(clickedRating)} // 最後にクリックされた星に戻す
                >
                  {s <= rating ? '★' : '☆'}
                </span>
              ))}
            </div>
          </div>

          <div>
            <label htmlFor="content" className="block text-gray-700 font-semibold mb-2">コメント</label>
            <textarea
              id="content" name="content" rows={4} value={content} disabled={submitting}
              className="w-full p-3 border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500 resize-y"
              placeholder="商品の感想や体験を具体的にご記入ください。"
              onChange={(e) => setContent(e.target.value)}
            ></textarea>
            {errorMessage && <p className="text-red-500">{errorMessage}</p>}
            {successMessage && <p className="text-green-600">{successMessage}</p>}
          </div>

          <button
            type="submit" disabled={submitting}
            className={`w-full py-3 px-4 rounded-md
              ${submitting ? 'bg-gray-400 cursor-not-allowed' : 'bg-indigo-600 hover:bg-indigo-700 text-white'}
            `}
          >
            {submitting ? '投稿中...' : 'レビューを投稿'}
          </button>
        </form>
      )}
    </div>
  );
}