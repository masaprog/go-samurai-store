'use client'; // クライアント（ブラウザ）側で動作

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import ProductForm from '@/components/ProductForm';
import { type ProductData } from '@/types/product';

// 商品データの型定義
type Product = ProductData; // 基本型から変更なし

// 商品編集ページ
export default function ProductEditPage() {
  const router = useRouter();
  const { id: productId } = useParams();

  // 商品データ、エラーメッセージ、商品読み込み状態の管理
  const [productData, setProductData] = useState<Product | null>(null);
  const [errorMessage, setErrorMessage] = useState('');
  const [loading, setLoading] = useState(true);

  // 商品IDが変更されたときに商品データを取得
  useEffect(() => {
    const getProduct = async () => {
      try {
        const res = await fetch(`/api/products/${productId}`);
        const data = await res.json();
        if (!res.ok) {
          setErrorMessage(data.error || '商品情報の取得に失敗しました。');
          setLoading(false); // 読み込み終了
          return;
        }
        setProductData(data); // 商品データを更新
      } catch {
        setErrorMessage('通信エラーが発生しました。');
      } finally {
        setLoading(false); // 読み込み終了
      }
    };

    // 現在のIDに該当する商品データを取得
    if (productId) {
      getProduct();
    }
  }, [productId]);

  // フォーム送信
  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault(); // デフォルトの送信動作をキャンセル
    setErrorMessage(''); // 送信前にエラーをクリア

    // フォームデータを生成
    const formData = new FormData(e.currentTarget);

    // 各入力データを取得
    const name = formData.get('name') as string;
    const price = Number(formData.get('price') as string);
    const stock = Number(formData.get('stock') as string);

    // 入力値のバリデーション（画像ファイルなしは許容）
    if (!name.trim()) {
      setErrorMessage('商品名は必須です。');
      return;
    }
    if (price < 0 || stock < 0) {
      setErrorMessage('価格と在庫数は0以上の数値を入力してください。');
      return;
    }

    try { // 商品編集APIにPUTリクエストを送信
      const res = await fetch(`/api/products/${productId}`, {
        method: 'PUT',
        body: formData
      });

      if (res.ok) { // 更新成功時は管理者向け商品一覧ページへ
        router.push('/admin/products?edited=1'); // 更新成功をクエリパラメータで通知
      } else {
        const data = await res.json();
        setErrorMessage(data.error || '更新に失敗しました。');
      }
    } catch {
      setErrorMessage('通信エラーが発生しました。');
    }
  };

  // 商品データ読み込み中の表示
  if (loading) {
    return <div className="text-center py-12 text-gray-600 text-lg">商品データを読み込み中です...</div>;
  }

  return (
    <main className="max-w-xl mx-auto py-10">
      <h1 className="text-center mb-6">商品編集</h1>
      {errorMessage && (
        <p className="text-red-600 text-center mb-4">{errorMessage}</p>
      )}
      {productData && (
        <ProductForm
          onSubmit={handleSubmit}
          initialValues={productData}
          submitLabel="更新"
        />
      )}
    </main>
  );
}