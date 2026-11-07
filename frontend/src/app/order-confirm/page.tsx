'use client'; // クライアント（ブラウザ）側で動作

import { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useCart, CartItem } from '@/hooks/useCart';
import CartItemCard from '@/components/CartItemCard';

// 配送費用
const SHIPPING_COST = 500; // 500円固定とする

// 注文確認ページ
export default function OrderConfirmPage() {
  const router = useRouter();
  const { cartItems, totalPrice } = useCart(); // カート情報と制御関数を取得
  const finalPrice = totalPrice + SHIPPING_COST; // 送料込みの最終的な合計金額

  const [address, setAddress] = useState(''); // お届け先住所
  const [isAgreed, setIsAgreed] = useState(false); // 同意のチェック有無
  const [errorMessage, setErrorMessage] = useState(''); // エラーメッセージ

  // 決済ボタン押下時のイベントハンドラ
  const handleConfirmPayment = async () => {
    if (!address.trim()) {
      setErrorMessage('お届け先住所を入力してください。');
      return;
    }
    if (!isAgreed) {
      setErrorMessage('利用規約とプライバシーポリシーに同意してください。');
      return;
    }

    // Stripe Checkoutセッションを作成
    const checkoutRes = await fetch('/api/orders/checkout', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ items: cartItems, address }),
    });
    if (!checkoutRes.ok) {
      setErrorMessage('決済ページの生成に失敗しました。');
      return;
    }
    const checkoutData = await checkoutRes.json();
    if (checkoutData.url) {
      router.push(checkoutData.url); // 決済ページへ遷移
    }
  };

  // キャンセル時のイベントハンドラ
  const handleCancel = () => {
    // 確認画面を表示
    if (confirm('入力内容が破棄されます。よろしいですか？')) {
      router.push('/cart'); // カートページへ遷移
    }
  };

  return (
    <main className="container mx-auto px-4 py-8">
      <h1 className="text-center mb-6">ご注文内容の確認</h1>
      {errorMessage && (
        <p className="text-red-600 text-center mb-4">{errorMessage}</p>
      )}

      {cartItems.length === 0 ? (
        <div className="text-center py-16">
          <p className="text-gray-600">カートに商品がありません。</p>
          <Link href="/products" className="text-indigo-600 hover:underline">← 商品一覧に戻る</Link>
        </div>
      ) : (
        <>
          <div className="flex flex-col space-y-6">
            {cartItems.map((item: CartItem) => (
              <CartItemCard key={item.id} item={item} />
            ))}
          </div>

          <div className="mt-8 border-t border-gray-300 pt-6">
            <label className="block font-bold mb-1" htmlFor="address">
              お届け先住所<span className="ml-2 px-2 py-0.5 bg-red-500 text-white text-xs font-semibold rounded-md">必須</span>
            </label>
            <textarea
              id="address" value={address} placeholder="お届け先住所を入力してください" required
              className="w-full border border-gray-300 px-3 py-2 rounded-sm focus:ring-2 focus:ring-indigo-500"
              onChange={(e) => setAddress(e.target.value)}
            />
          </div>

          <div className="mt-6">
            <div className="flex justify-between items-center pb-2 font-semibold">
              <span>商品合計：</span><span>¥{totalPrice.toLocaleString()}</span>
            </div>
            <div className="flex justify-between items-center py-4 border-b border-gray-300 font-semibold">
              <span>送料：</span><span>¥{SHIPPING_COST.toLocaleString()}</span>
            </div>
            <div className="flex justify-between items-center pt-4 text-green-600 text-2xl font-bold">
              <span>支払い合計：</span><span>¥{finalPrice.toLocaleString()}</span>
            </div>

            <p className="text-gray-500 text-sm mt-2 text-right">
              ※価格はすべて税込表記です。また、お支払いはStripeによるクレジットカード決済となります。
            </p>
          </div>

          <div className="mt-8">
            <p className="text-sm">
              ご注文の確定前に、必ず
              <Link href="/terms" target="_blank" rel="noopener noreferrer" className="text-indigo-600 font-bold underline">
                利用規約
              </Link>
              および
              <Link href="/privacy" target="_blank" rel="noopener noreferrer" className="text-indigo-600 font-bold underline">
                プライバシーポリシー
              </Link>
              をお読みください。<br/>同意のチェックをもって、上記規約に同意したものとみなされます。
            </p>
            <div className="mt-6 flex items-center">
              <input
                type="checkbox" id="agreement-checkbox" checked={isAgreed}
                onChange={(e) => setIsAgreed(e.target.checked)}
                className="h-4 w-4 text-indigo-600 border-gray-300 rounded focus:ring-indigo-500"
              />
              <label htmlFor="agreement-checkbox" className="ml-2 text-base font-semibold text-gray-800 leading-snug">
                利用規約およびプライバシーポリシーに同意する
              </label>
            </div>
          </div>

          <div className="mt-8 flex justify-between items-center">
            <button
              onClick={handleConfirmPayment}
              disabled={!isAgreed || !address.trim()}
              className={`py-2 px-6 rounded-sm text-white ${
                isAgreed && address.trim()
                  ? 'bg-indigo-500 hover:bg-indigo-600'
                  : 'bg-gray-400 cursor-not-allowed'
              }`}
            >
              決済する
            </button>
            <button
              type="button" onClick={handleCancel}
              className="bg-gray-200 hover:bg-gray-300 text-gray-700 py-2 px-6 rounded-sm"
            >
              キャンセル
            </button>
          </div>
        </>
      )}
    </main>
  );
}