'use client'; // クライアント（ブラウザ）側で動作

import ProductForm from '@/components/ProductForm';
import { useState } from 'react';
import { useRouter } from 'next/navigation';

// 商品登録ページ
export default function ProductRegisterPage() {
  const router = useRouter();
  const [errorMessage, setErrorMessage] = useState(''); // エラーメッセージ

  // フォーム送信時のイベントハンドラ
  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault(); // デフォルトの送信動作をキャンセル
    setErrorMessage(''); // 送信前にエラーをクリア

    // フォームデータを生成
    const formData = new FormData(e.currentTarget);

    // 各入力データを取得
    const name = formData.get('name') as string;
    const imageFile = formData.get('imageFile') as File;
    const price = Number(formData.get('price') as string);
    const stock = Number(formData.get('stock') as string);

    // 入力データのバリデーション
    if (!name.trim() || !imageFile) {
      setErrorMessage('すべての必須項目を入力してください。');
      return;
    }
    if (price < 0 || stock < 0) {
      setErrorMessage('価格と在庫数は0以上の数値を入力してください。');
      return;
    }

    try { // 商品登録APIにPOSTリクエストを送信
      const res = await fetch('/api/products', {
        method: 'POST',
        body: formData
      });

      if (res.ok) { // 登録成功時は管理者向け商品一覧ページへ
        router.push('/admin/products?registered=1'); // 登録成功をクエリパラメータで通知
      } else { // 登録失敗時はエラー情報を表示
        const data = await res.json();
        setErrorMessage(data.error || '登録に失敗しました。');
      }
    } catch {
      setErrorMessage('通信エラーが発生しました。');
    }
  };

  return (
    <main className="max-w-xl mx-auto py-10">
      <h1 className="text-center mb-6">商品登録</h1>
      {errorMessage && (
        <p className="text-red-600 text-center mb-4">{errorMessage}</p>
      )}
      <ProductForm
        onSubmit={handleSubmit}
        submitLabel="登録"
      />
    </main>
  );
}