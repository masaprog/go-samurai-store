'use client'; // クライアント（ブラウザ）側で動作

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';

// 商品フォームコンポーネントに渡すデータ（props）の型定義
interface ProductFormProps {
  // フォーム送信時に実行されるイベントハンドラ
  onSubmit: (e: React.FormEvent<HTMLFormElement>) => void;
  // 入力項目に設定する初期値
  initialValues?: {
    name: string;
    image_url?: string | null | undefined;
    description?: string | null | undefined;
    price: number;
    stock?: number;
    is_featured?: boolean;
  };
  submitLabel: string; // フォーム送信ボタンの表示名
}

// 入力欄の共通スタイル
const inputStyle = 'w-full border border-gray-300 px-3 py-2 rounded-sm focus:ring-2 focus:ring-indigo-500';
// ラベルの共通スタイル
const labelStyle = "block font-bold mb-1";
// バッジの共通スタイル (必須項目を示す)
const badgeStyle = "ml-2 px-2 py-0.5 bg-red-500 text-white text-xs font-semibold rounded-md";

// 商品フォームの共通コンポーネント
export default function ProductForm({
  onSubmit,
  initialValues = { name: '', image_url: '', description: '', price: 0, stock: 0, is_featured: false },
  submitLabel,
}: ProductFormProps) {
  const router = useRouter();
  const [previewUrl, setPreviewUrl] = useState('');

  // 登録済みの画像があれば初期表示
  useEffect(() => {
    if (initialValues.image_url) {
      setPreviewUrl(`/uploads/${initialValues.image_url}`);
    }
  }, [initialValues.image_url]);

  // URL.createObjectURL()で生成された一時URLを解放する
  useEffect(() => {
    return () => {
      if (previewUrl && previewUrl.startsWith('blob:')) {
        URL.revokeObjectURL(previewUrl);
      }
    };
  }, [previewUrl]);

  const handleImageChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) { // 画像情報をクリア
      setPreviewUrl('');
      return;
    }

    // ブラウザで画像を表示するための一時的なURLを生成
    const tempUrl = URL.createObjectURL(file);
    setPreviewUrl(tempUrl);
  };

  // キャンセル時のイベントハンドラ
  const handleCancel = () => {
    // 確認画面を表示
    if (confirm('入力内容が破棄されます。よろしいですか？')) {
      router.push('/admin/products'); // 管理者向け商品一覧ページへ遷移
    }
  };

  return (
    <form
      onSubmit={onSubmit} encType="multipart/form-data" 
      className="w-full space-y-6 p-8 bg-white shadow-lg rounded-xl"
    >
      <label className={labelStyle} htmlFor="name">
        商品名<span className={badgeStyle}>必須</span>
      </label>
      <input type="text" id="name" name="name" required
        defaultValue={initialValues.name} className={inputStyle}
      />

      <label className={labelStyle} htmlFor="imageFile">
        商品画像<span className={badgeStyle}>必須</span>
      </label>
      <div className="flex flex-col gap-6 mt-2">
        <input type="file" id="imageFile" name="imageFile"
          required={!initialValues.image_url}
          accept="image/*" // 画像ファイルのみ許可
          onChange={handleImageChange}
          className="text-gray-600 file:bg-gray-50 file:border file:border-gray-300 file:px-4 file:py-2 file:rounded-sm file:cursor-pointer"
        />
        {previewUrl && (
          <div className="px-8" id="imagePreview">
            <img src={previewUrl} alt="preview" className="object-cover rounded-md shadow-md" />
          </div>
        )}
      </div>

      <label className={labelStyle} htmlFor="description">
        説明
      </label>
      <textarea id="description" name="description" rows={5}
        defaultValue={initialValues.description ?? ''} className={inputStyle}
      />

      <label className={labelStyle} htmlFor="price">
        価格(税込)<span className={badgeStyle}>必須</span>
      </label>
      <input type="number" id="price" name="price" required
        min="0" step="1" // 0以上の整数のみ許可
        defaultValue={initialValues.price} className={inputStyle}
      />

      <label className={labelStyle} htmlFor="stock">
        在庫数<span className={badgeStyle}>必須</span>
      </label>
      <input type="number" id="stock" name="stock" required
        min="0" step="1" // 0以上の整数のみ許可
        defaultValue={initialValues.stock} className={inputStyle}
      />

      <label className={labelStyle} htmlFor="isFeatured" >
        <input
          type="checkbox" id="isFeatured" name="isFeatured"
          defaultChecked={initialValues.is_featured} className="mr-2"
        />
        注目商品として表示する
      </label>

      <div className="flex justify-end space-x-4 mt-6">
        <button type="button" onClick={handleCancel}
          className="w-1/2 bg-gray-200 hover:bg-gray-300 text-gray-700 py-2 rounded-sm"
        >
          キャンセル
        </button>
        <button type="submit"
          className="w-1/2 bg-indigo-500 hover:bg-indigo-600 text-white font-semibold py-2 rounded-sm"
        >
          {submitLabel}
        </button>
      </div>
    </form>
  );
}