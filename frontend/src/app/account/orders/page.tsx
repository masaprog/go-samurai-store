'use client'; // クライアント（ブラウザ）側で動作

import Link from 'next/link';
import { useEffect, useState } from 'react';

// 注文ごとの商品データの型定義
interface OrderItem {
  productName: string;
  quantity: number;
  unitPrice: number;
}

// 最終的なレスポンス（商品明細を配列化した注文データ）の型定義
interface OrderData {
  id: number;
  totalPrice: number;
  status: '未処理' | '処理中' | '出荷済み' | '完了' | 'キャンセル' | '返金済み';
  paymentStatus: '未決済' | '決済処理中' | '決済成功' | '決済失敗' | '返金処理中' | '返金済み';
  createdAt: string;
  items: OrderItem[]; // 商品明細の配列
}

// 注文履歴一覧ページ
export default function OrdersPage() {
  const [orders, setOrders] = useState<OrderData[]>([]);
  const [errorMessage, setErrorMessage] = useState('');
  const [loading, setLoading] = useState(true);

  // 初期表示時に注文データを取得
  useEffect(() => {
    const getOrders = async () => {
      try {
        const res = await fetch('/api/orders');
        const data = await res.json();
        if (!res.ok) {
          setErrorMessage(data.error || '注文履歴の取得に失敗しました。');
          setLoading(false); // 読み込み終了
          return;
        }
        setOrders(data.orders); // 注文履歴データを更新
      } catch (err) {
        console.error(err);
        setErrorMessage('通信エラーが発生しました。');
      } finally {
        setLoading(false); // 読み込み終了
      }
    };

    // 注文履歴データを取得
    getOrders();
  }, []);

  if (loading) return <div className="text-center py-12 text-gray-600 text-lg">注文履歴データを読み込み中です...</div>;
  if (errorMessage) return <p className="text-center py-12 text-red-600">{errorMessage}</p>;
  if (orders.length === 0) return <p className="text-center py-12 text-gray-500">注文履歴はありません。</p>;

  // ステータスに応じて表示スタイルを決定
  const getStatusStyle = (status: OrderData['status'] | OrderData['paymentStatus']) => {
    switch (status) {
      case '未処理':
      case '処理中':
      case '未決済':
      case '決済処理中':
      case '返金処理中': // 進行中または要確認の状態
        return 'text-yellow-500';
      case '出荷済み':
      case '完了':
      case '決済成功': // ポジティブな完了状態
        return 'text-green-500';
      case 'キャンセル':
      case '決済失敗':
      case '返金済み': // ネガティブな最終状態
        return 'text-red-500';
      default:
        return 'text-gray-500'; // デフォルトの色
    }
  };

  // テーブルの共通スタイル
  const tableStyle = 'px-3 py-2 border-b';

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="my-4">
        <Link href="/account" className="text-indigo-600 hover:underline">
          ← マイページに戻る
        </Link>
      </div>
      <h1 className="text-center mb-8">注文履歴一覧</h1>
      <div className="space-y-6">
        {orders.map((order) => (
          <div key={order.id} className="border rounded-lg shadow-sm p-4">
            <div className="flex justify-between mb-4">
              <div>
                <p className="text-lg font-semibold">注文ID：{order.id}</p>
                <p>注文日：{order.createdAt ? new Date(order.createdAt).toLocaleDateString() : '-'}</p>
              </div>
              <div className="text-right font-semibold">
                <p className="text-blue-600 text-xl">合計金額：¥{order.totalPrice.toLocaleString()}（送料込み）</p>
                <p className={getStatusStyle(order.status)}>注文状況：{order.status}</p>
                <p className={getStatusStyle(order.paymentStatus)}>決済状況：{order.paymentStatus}</p>
              </div>
            </div>

            <table className="w-full text-left border-t border-gray-200 shadow-lg rounded-lg overflow-hidden">
              <thead>
                <tr className="bg-gray-100 text-gray-700">
                  <th className={tableStyle}>商品名</th>
                  <th className={tableStyle}>数量</th>
                  <th className={tableStyle}>単価（税込）</th>
                </tr>
              </thead>
              <tbody>
                {order.items.map((item, index) => (
                  <tr key={index} className="hover:bg-gray-50">
                    <td className={tableStyle}>{item.productName}</td>
                    <td className={tableStyle}>{item.quantity}</td>
                    <td className={tableStyle}>¥{item.unitPrice.toLocaleString()}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ))}
      </div>
    </div>
  );
}