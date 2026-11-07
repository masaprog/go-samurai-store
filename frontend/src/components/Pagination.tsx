'use client'; // クライアント（ブラウザ）側で動作

import React from 'react';
import { useRouter, useSearchParams } from 'next/navigation';

// ページネーション共通コンポーネントに渡すデータ（props）の型定義
interface PaginationProps {
  currentPage: number;
  totalPages: number;
}

// ページネーション共通コンポーネント
export default function Pagination({ currentPage, totalPages }: PaginationProps) {
  const router = useRouter();
  const searchParams = useSearchParams();

  // ページ切り替え用のイベントハンドラ
  const handlePageChange = (newPage: number) => {
    // 現在のクエリパラメータを取得
    const params = new URLSearchParams(searchParams.toString());
    // クエリパラメータのページ番号（page）を更新
    params.set('page', String(newPage));
    // URLを更新し、リロードなしで遷移
    router.push(`?${params.toString()}`);
  };

  // ページネーション系ボタンの共通スタイルを定義
  const baseClasses = 'min-w-9 h-9 rounded border border-gray-300 mx-1 cursor-pointer';
  const hover = 'hover:bg-gray-100 hover:text-gray-800';
  const active = 'bg-indigo-500 text-white border-indigo-500';
  const disabled = 'opacity-50';

  return (
    <nav className="flex justify-center items-center mt-8" aria-label="ページネーション">
      {/* 前へ（<） */}
      <button
        onClick={() => handlePageChange(currentPage - 1)}
        disabled={currentPage === 1}
        className={`${baseClasses} text-gray-700 ${hover} ${currentPage === 1 ? disabled : ''}`}
      >
        &lt;
      </button>

      {/* 各ページ番号 */}
      {Array.from({ length: totalPages }, (_, i) => {
        // ページ番号は1から始まるため、配列のインデックス（0始まり）に1加算
        const page = i + 1;
        // 現在のページ番号に一致するページだけをアクティブに
        const isActive = (page === currentPage);
        return (
          <button key={page} onClick={() => handlePageChange(page)} disabled={isActive}
            className={`${baseClasses} ${isActive ? active : 'text-gray-700 ' + hover}`}
          >
            {page}
          </button>
        );
      })}

      {/* 次へ（>） */}
      <button
        onClick={() => handlePageChange(currentPage + 1)}
        disabled={currentPage === totalPages}
        className={`${baseClasses} text-gray-700 ${hover} ${currentPage === totalPages ? disabled : ''}`}
      >
        &gt;
      </button>
    </nav>
  );
}