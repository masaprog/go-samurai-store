import Link from 'next/link';
import { cookies } from 'next/headers';
import { AUTH_TOKEN } from '@/lib/auth';

// inquiriesテーブルのデータ型定義
type Inquiry = {
  id: number;
  name: string;
  email: string;
  message: string;
  created_at: string;
};

// お問い合わせ一覧ページ
export default async function InquiriesPage() {
  // クッキーを取得
  const cookieStore = await cookies();
  const token = cookieStore.get(AUTH_TOKEN)?.value;

  // HTTPリクエストヘッダーを設定
  const headers: HeadersInit = {};
  if (token) {
    headers['Cookie'] = `${AUTH_TOKEN}=${token}`;
  }

  // お問い合わせAPIからデータを取得
  const res = await fetch(`${process.env.API_BASE_URL}/api/inquiries`, {
    cache: 'no-store',
    headers: headers
  });

  // APIから返されたデータを取得
  const inquiries: Inquiry[] = await res.json()
  if (!Array.isArray(inquiries)) {
    console.error('お問い合わせデータの取得に失敗しました。');
    return <p className="text-center text-gray-500 text-lg py-10">お問い合わせデータの取得に失敗しました。</p>;
  }

  // テーブルの共通スタイル
  const tableStyle = 'px-5 py-3 border-b border-gray-300';

  return (
    <main className="container mx-auto px-4 py-8">
      <div className="my-4">
        <Link href="/admin/products" className="text-indigo-600 hover:underline">
          ← 商品一覧ページに戻る
        </Link>
      </div>
      <h1 className="text-center">お問い合わせ一覧</h1>

      <div className="shadow-lg rounded-lg overflow-hidden">
        <table className="min-w-full leading-normal">
          <thead>
            <tr className="bg-gray-200 text-gray-700 text-left">
              <th className={tableStyle}>ID</th>
              <th className={tableStyle}>氏名</th>
              <th className={tableStyle}>メールアドレス</th>
              <th className={tableStyle}>お問い合わせ内容</th>
              <th className={tableStyle}>送信日時</th>
            </tr>
          </thead>
          <tbody>
            {inquiries.length === 0 ? (
              <tr>
                <td colSpan={5} className={`${tableStyle} text-center text-gray-500`}>
                  お問い合わせは見つかりませんでした。
                </td>
              </tr>
            ) : (
              inquiries.map((inquiry) => (
                <tr key={inquiry.id} className="hover:bg-gray-100">
                  <td className={tableStyle}>{inquiry.id}</td>
                  <td className={tableStyle}>{inquiry.name}</td>
                  <td className={tableStyle}>{inquiry.email}</td>
                  <td className={tableStyle}>{inquiry.message}</td>
                  <td className={tableStyle}>
                    {new Date(inquiry.created_at).toLocaleDateString()}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </main>
  );
}