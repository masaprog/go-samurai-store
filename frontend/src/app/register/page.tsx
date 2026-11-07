'use client'; // クライアント（ブラウザ）側で動作

import UserForm from '@/components/UserForm';
import { useState } from 'react';
import { useRouter } from 'next/navigation';

// 会員編集ページ
export default function UserRegisterPage() {
  const router = useRouter();
  const [errorMessage, setErrorMessage] = useState(''); // エラーメッセージ

  // フォーム送信時のイベントハンドラ
  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault(); // デフォルトの送信動作をキャンセル
    setErrorMessage(''); // 送信前にエラーをクリア

    const formData = new FormData(e.currentTarget);
    const name = formData.get('name') as string;
    const email = formData.get('email') as string;
    const password = formData.get('password') as string;
    const confirmPassword = formData.get('confirmPassword') as string;

    // 入力データのバリデーション
    if (!name?.trim() || !email?.trim() || !password?.trim() || !confirmPassword?.trim()) {
      setErrorMessage('すべての項目を入力してください。');
      return;
    }
    if (password !== confirmPassword) {
      setErrorMessage('パスワードが一致しません。');
      return;
    }

    try { // 会員登録APIにPOSTリクエストを送信
      const res = await fetch('/api/users', {
        method: 'POST',
        body: JSON.stringify({ name, email, password }),
        headers: { 'Content-Type': 'application/json' }
      });

      if (res.ok) { // 登録成功時はトップページへ遷移
        router.push('/?registered=1'); // 登録成功をクエリパラメータで通知
      } else { // 登録失敗時はエラー情報を表示
        const data = await res.json();
        setErrorMessage(data.error || '登録に失敗しました。');
      }
    } catch {
      setErrorMessage('通信エラーが発生しました。');
    }
  };

  return (
    <main className="max-w-md mx-auto py-10">
      <h1 className="text-center mb-4">会員登録</h1>
      {errorMessage && (
        <p className="text-red-600 text-center mt-8">{errorMessage}</p>
      )}
      <UserForm
        onSubmit={handleSubmit}
        withPassword={true}
        submitLabel="登録"
      />
    </main>
  );
}