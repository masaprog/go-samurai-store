'use client'; // クライアント（ブラウザ）側で動作

import Link from 'next/link';
import Image from 'next/image';
import { useState, useEffect, useRef } from 'react';
import { useSearchParams } from 'next/navigation';
import { useCart } from '@/hooks/useCart';
import { type AuthUser } from '@/lib/auth';

// ヘッダーコンポーネントに渡すデータ（props）の型定義
export interface HeaderProps {
  user: AuthUser | null;
};

// 共通ヘッダー
export default function Header({ user }: HeaderProps) {
  const [isMenuOpen, setIsMenuOpen] = useState(false); // メニュー開閉状態
  const menuRef = useRef<HTMLDivElement>(null); // メニューのDOM参照用
  // メニューを閉じる関数
  const closeMenu = () => setIsMenuOpen(false);
  // メニューの開閉状態を反転させる関数
  const toggleMenu = () => setIsMenuOpen(prev => !prev);

  // カート内商品の総数を取得
  const { totalQuantity } = useCart();
  // 表示用のカート数量（初期値は0）
  const [displayQuantity, setDisplayQuantity] = useState(0);

  const searchParams = useSearchParams();
  const perPage = searchParams.get('perPage') || '16';
  const sort = searchParams.get('sort') || 'new';
  const keyword = searchParams.get('keyword') || '';

  // メニュー外のクリックを検知し、メニューを閉じる
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      // メニュー外がクリックされたら非表示
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        closeMenu();
      }
    };
    // メニュー外のクリックを検知するイベントリスナー
    document.addEventListener('click', handleClickOutside);

    // クリーンアップ処理（イベントリスナーを削除）
    return () => document.removeEventListener('click', handleClickOutside);
  }, []);

  // カート内商品の数量変更時に表示を更新
  useEffect(() => {
    setDisplayQuantity(totalQuantity);
  }, [totalQuantity]);

  // メニュー項目の共通スタイル
  const menuItemStyle = 'block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100';

  return (
    <header className="bg-white shadow-md">
      <div className="container mx-auto px-4 py-3 flex items-center justify-between">
        <div className="flex-shrink-0">
          <Link href="/">
            <Image src="/images/samurai-store-logo.png" alt="SAMURAI Store" width={910} height={200} className="w-[250px] h-[55px] object-contain"/>
          </Link>
        </div>

        <nav className="flex-grow text-center mt-8">
          <ul className="inline-flex divide-x divide-gray-300 list-none">
            <li className="border-r border-gray-300">
              <Link href="/" className="block w-[120px] py-3 hover:bg-gray-200 rounded-sm">ホーム</Link>
            </li>
            <li className="border-r border-gray-300">
              <Link href="/products" className="block w-[120px] py-3 hover:bg-gray-200 rounded-sm">商品一覧</Link>
            </li>
            <li>
              <Link href="/contact" className="block w-[120px] py-3 hover:bg-gray-200 rounded-sm">お問い合わせ</Link>
            </li>
          </ul>
        </nav>

        <div className="flex items-center space-x-6 flex-shrink-0">
          <form action="/products" method="GET" className="hidden sm:block">
            <input type="hidden" name="page" value="1" />
            <input type="hidden" name="perPage" value={perPage} />
            <input type="hidden" name="sort" value={sort} />
            <input
              type="text"
              name="keyword"
              placeholder="検索..."
              defaultValue={keyword}
              className="border border-gray-300 rounded-md py-1 px-3 text-sm focus:ring-2 focus:ring-indigo-500"
            />
          </form>
          <Link href="/account/favorites">
            <Image src="/icons/heart-icon.svg" alt="Favorites" width={24} height={24} className="w-6 h-6" />
          </Link>
          <Link href="/cart" className="relative">
            <Image src="/icons/cart-icon.svg" alt="Cart" width={24} height={24} className="w-6 h-6" />
            {displayQuantity > 0 && (
              <span className="absolute -top-2 -right-2 w-[20px] h-[20px] bg-yellow-500 text-black flex items-center justify-center rounded-full ring-2 ring-white text-xs font-semibold">
                {displayQuantity > 9 ? '9+' : displayQuantity}
              </span>
            )}
          </Link>
          <div className="relative" ref={menuRef}>
            <button onClick={toggleMenu}  className="cursor-pointer" >
              <Image src="/icons/account-icon.svg" alt="Account" width={24} height={24} className="w-6 h-6" />
              {user && (
                <span className="absolute -top-0.5 -right-0.5 w-2.5 h-2.5 bg-green-500 rounded-full ring-2 ring-white" />
              )}
            </button>
            {isMenuOpen && (
              <div className="absolute right-0 w-48 bg-white rounded-md shadow-lg py-1 z-10 border border-gray-300">
                {user ? (
                  <>
                    <Link href="/account" onClick={closeMenu} className={menuItemStyle}>
                      マイページ
                    </Link>
                    <form method="POST" action="/api/auth/logout">
                      <button type="submit" className={`${menuItemStyle} w-full text-left`}>
                        ログアウト
                      </button>
                    </form>
                  </>
                ) : (
                  <>
                    <Link href="/login" onClick={closeMenu} className={menuItemStyle}>
                      ログイン
                    </Link>
                    <Link href="/register" onClick={closeMenu} className={menuItemStyle}>
                      会員登録
                    </Link>
                  </>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </header>
  );
}