import React from 'react';
import Link from 'next/link';
import { notFound } from 'next/navigation';
import { type ReviewsResponse } from '@/types/review';
import Pagination from '@/components/Pagination';

// レビュー一覧ページ
export default async function ReviewsPage({
  params,
  searchParams,
}: {
  params: Promise<{ id: string }>
  searchParams?: Promise<{ [key: string]: string | string[] | undefined }>
}) {
  const resolvedParams = await params; // 非同期でparamを取得
  const productId = resolvedParams.id;

  const sp = await searchParams;
  const page = Number(sp?.page ?? '1');
  const perPage = 10; // 表示件数は固定

  // クエリパラメータを1つの文字列にまとめる
  const query = new URLSearchParams({
    page: String(page),
    perPage: String(perPage)
  });

  // レビューAPIからデータを取得
  const res = await fetch(`${process.env.API_BASE_URL}/api/products/${productId}/reviews?${query.toString()}`, {
    cache: 'no-store'
  });

  if (!res.ok) {
    notFound();
  }

  const reviewsPageData: ReviewsResponse = await res.json();
  if (!Array.isArray(reviewsPageData.reviews)) {
    notFound();
  }

  return (
    <main className="p-8">
      <div className="my-4">
        <Link href={`/products/${productId}`} className="text-indigo-600 hover:underline">
          ← 商品詳細ページに戻る
        </Link>
      </div>
      <h1 className="text-center mb-6">レビュー一覧</h1>

      <section className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
        {reviewsPageData.reviews.map((r) => (
          <div key={r.id} className="p-4 border border-gray-200 rounded-md shadow-sm">
            <div className="flex items-center text-yellow-500 text-lg mb-2">
              {Array.from({ length: 5 }, (_, i) => (
                <span key={i}>{i < r.score ? '★' : '☆'}</span>
              ))}
            </div>
            <p className="text-gray-800 mb-2">{r.content}</p>
            <p className="text-xs text-gray-500">
              {r.user_name} さん {new Date(r.created_at).toLocaleDateString()}
            </p>
          </div>
        ))}
      </section>

      <section className="mb-8">
        {reviewsPageData.pagination.totalPages > 0 && (
          <Pagination
            currentPage={reviewsPageData.pagination.currentPage}
            totalPages={reviewsPageData.pagination.totalPages}
          />
        )}
      </section>
    </main>
  );
}