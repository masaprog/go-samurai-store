'use client'; // クライアント（ブラウザ）側で動作

import UserForm from '@/components/UserForm';
import { useState } from 'react';
import { useRouter } from 'next/navigation';

interface Props {
  initialValues: { name: string; email: string };
}

// 会員編集フォーム
export default function UserEditForm({ initialValues }: Props) {
  const router = useRouter();
  const [errorMessage, setErrorMessage] = useState(''); // エラーメッセージ

  // フォーム送信時のイベントハンドラ
  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault(); // デフォルトの送信動作をキャンセル
    setErrorMessage(''); // 送信前にエラーをクリア

    const formData = new FormData(e.currentTarget);
    const name = formData.get('name') as string;
    const email = formData.get('email') as string;

    // 入力チェック
    if (!name?.trim() || !email?.trim()) {
      setErrorMessage('すべての項目を入力してください。');
      return;
    }

    try { // 会員編集APIにPUTリクエストを送信
      const res = await fetch('/api/users', {
        method: 'PUT',
        body: JSON.stringify({ name, email }),
        headers: { 'Content-Type': 'application/json' }
      });

      if (res.ok) { // 更新成功時はマイページへ
        router.push('/account?edited=1');
      } else { // 更新失敗時はエラー情報を表示
        const data = await res.json();
        setErrorMessage(data.error || '更新に失敗しました。');
      }
    } catch {
      setErrorMessage('通信エラーが発生しました。');
    }
  };

  return (
    <>
      {errorMessage && (
        <p className="text-red-600 text-center mt-8">{errorMessage}</p>
      )}
      <UserForm
        onSubmit={handleSubmit}
        initialValues={initialValues}
        submitLabel="更新"
      />
    </>
  );
}