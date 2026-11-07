'use client'; // クライアント（ブラウザ）側で動作

import { useRouter } from 'next/navigation';

// 削除リンクコンポーネント
export default function DeleteLink({ id, name }: { id: number; name: string }) {
  const router = useRouter();

  // 削除リンク押下時のイベントハンドラ
  const handleDelete = async () => {
    if (!confirm(`「${name}」を削除すると元に戻せません。\n削除してもよろしいですか？`)) {
      return; // 削除をキャンセル
    }

    try { // 商品削除APIにDELETEリクエストを送信
      const res = await fetch(`/api/products/${id}`, { method: 'DELETE' });

      if (res.ok) { // 削除成功時は管理者向け商品一覧ページを再表示
        router.push(`/admin/products?deleted=1`); // 削除成功をクエリパラメータで通知
      } else { // 削除失敗時はエラー情報を表示
        const data = await res.json();
        alert(data.error || '削除に失敗しました。');
      }
    } catch {
      alert('通信エラーが発生しました。');
    }
  }

  return (
    <button
      onClick={handleDelete}
      className="text-red-600 hover:text-red-700 cursor-pointer"
    >
      削除
    </button>
  )
}