'use client'; // クライアント（ブラウザ）側で動作

import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useState } from 'react';

// パスワード変更ページ
export default function PasswordChangePage() {
  const router = useRouter();
  const [errorMessage, setErrorMessage] = useState('');

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault(); // デフォルトの送信動作をキャンセル
    setErrorMessage(''); // 送信前にエラーをクリア

    const formData = new FormData(e.currentTarget);
    const oldPassword = formData.get('oldPassword') as string;
    const newPassword = formData.get('newPassword') as string;
    const confirmPassword = formData.get('confirmPassword') as string;

    // 入力データのバリデーション
    if (!oldPassword?.trim() || !newPassword?.trim() || !confirmPassword?.trim()) {
      setErrorMessage('すべての項目を入力してください。');
      return;
    }
    if (newPassword !== confirmPassword) {
      setErrorMessage('新しいパスワードが一致しません。');
      return;
    }

    try { // パスワード変更APIにPUTリクエストを送信
      const res = await fetch('/api/users/password', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ oldPassword, newPassword }),
      });

      if (res.ok) { // 変更成功時はマイページへ
        router.push('/account?password-changed=1');
      } else { // 変更失敗時はエラー情報を表示
        const data = await res.json();
        setErrorMessage(data.error || 'パスワード変更に失敗しました。');
      }
    } catch {
      setErrorMessage('通信エラーが発生しました。');
    }
  };

  // 入力欄の共通スタイル
  const inputStyle = 'w-full border border-gray-300 px-3 py-2 rounded-sm focus:ring-2 focus:ring-indigo-500';
  // ラベルの共通スタイル
  const labelStyle = "block font-bold mb-1";
  // バッジの共通スタイル
  const badgeStyle = "ml-2 px-2 py-0.5 bg-red-500 text-white text-xs font-semibold rounded-md";

  return (
    <main className="max-w-md mx-auto py-10">
      <div className="my-4">
        <Link href="/account" className="text-indigo-600 hover:underline">
          ← マイページに戻る
        </Link>
      </div>
      <h1 className="text-center mb-6">パスワード変更</h1>
      {errorMessage && (
        <p className="text-red-600 text-center mt-8">{errorMessage}</p>
      )}

      <form onSubmit={handleSubmit} className="w-full space-y-6 p-8 bg-white shadow-lg rounded-xl">
        <label className={labelStyle} htmlFor="oldPassword">
          現在のパスワード<span className={badgeStyle}>必須</span>
        </label>
        <input type="password" id="oldPassword" name="oldPassword" required className={inputStyle} />

        <label className={labelStyle} htmlFor="newPassword">
          新しいパスワード<span className={badgeStyle}>必須</span>
        </label>
        <input type="password" id="newPassword" name="newPassword" required className={inputStyle} />

        <label className={labelStyle} htmlFor="confirmPassword">
          新しいパスワード（確認用）<span className={badgeStyle}>必須</span>
        </label>
        <input type="password" id="confirmPassword" name="confirmPassword" required className={inputStyle} />

        <button type="submit" className="w-full mt-2 bg-indigo-500 hover:bg-indigo-600 text-white py-2 rounded-sm">
          更新
        </button>
      </form>
    </main>
  );
}