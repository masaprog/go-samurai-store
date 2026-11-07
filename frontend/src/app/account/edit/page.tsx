import Link from 'next/link';
import { getAuthUser } from '@/lib/auth';
import UserEditForm from '@/app/account/edit/UserEditForm';

// 会員編集ページ
export default async function UserEditPage() {
  // 会員情報を取得
  const user = await getAuthUser();
  if (!user) {
    return <p className="text-center mt-10">ログインしてください。</p>;
  }

  return (
    <main className="max-w-md mx-auto py-10">
      <div className="my-4">
        <Link href="/account" className="text-indigo-600 hover:underline">
          ← マイページに戻る
        </Link>
      </div>
      <h1 className="text-center mb-4">会員情報の編集</h1>
      <UserEditForm initialValues={{ name: user.name, email: user.email }} />
    </main>
  );
}