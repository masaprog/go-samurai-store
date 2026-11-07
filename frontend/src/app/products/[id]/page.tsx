import React from 'react';
import { notFound } from 'next/navigation'; // 404ページ表示用
import { cookies } from "next/headers";
import Link from 'next/link';
import Image from 'next/image';
import { type ProductData } from '@/types/product';
import { type ReviewsResponse } from '@/types/review';
import { isLoggedIn, AUTH_TOKEN } from '@/lib/auth';
import CartControls from '@/app/products/[id]/CartControls';
import ReviewControls from '@/app/products/[id]/ReviewControls';
import FavoriteControls from '@/app/products/[id]/FavoriteControls';

// 商品データの型定義
type Product = ProductData; // 基本型から変更なし

// 商品詳細ページに必要なデータ群
interface ProductDetailPageProps {
  params: Promise<{
    id: string; // URLから取得する商品ID
  }>
}

// 商品データを取得
async function getProduct(id: string): Promise<Product | null> {
  try {
    // 商品APIから商品データを取得
    const res = await fetch(`${process.env.API_BASE_URL}/api/products/${id}`, {
      cache: 'no-store',
    });

    // 取得失敗時
    if (!res.ok) return null;

    // APIから返されたデータをJavaScriptの配列に変換
    const product = await res.json();
    return product;
  } catch (err) {
    console.error('商品取得エラー：', err);
    return null;
  }
}

// 商品IDに紐づくお気に入り状態を取得
async function getFavoriteStatus(id: string): Promise<boolean> {
  try {
    // クッキーを取得
    const cookieStore = await cookies();
    const token = cookieStore.get(AUTH_TOKEN)?.value;
    const headers: HeadersInit = token ? { 'Cookie': `${AUTH_TOKEN}=${token}` } : {};

    // お気に入り取得APIから商品IDの登録有無を取得
    const res = await fetch(`${process.env.API_BASE_URL}/api/favorites/${id}`, {
      cache: 'no-store',
      headers: headers,
    });

    if (!res.ok) return false;

    // お気に入り登録有無を返す
    const { isFavorite } = await res.json();
    return isFavorite;
  } catch (err) {
    console.error('お気に入り状態取得エラー：', err);
    return false;
  }
}

// 商品IDに紐づくレビュー一覧を取得
async function getReviews(id: string): Promise<ReviewsResponse | []> {
  // レビューAPIからレビュー一覧を取得
  const res = await fetch(`${process.env.API_BASE_URL}/api/products/${id}/reviews`, {
    cache: 'no-store',
  });
  if (!res.ok) return []; // 失敗時は空配列
  return await res.json();
}

// レビューの星表示を決定
function displayStars(avgRating: number) {
  const rating = Math.round(avgRating); // 四捨五入
  const filledStars = '★'.repeat(rating); // 評価分塗りつぶす
  const emptyStars = '☆'.repeat(5 - rating); // 残りは空の星
  return `${filledStars}${emptyStars}`;
}

// 商品詳細ページ
export default async function ProductDetailPage(props: ProductDetailPageProps) {
  const resolvedParams = await props.params; // 非同期で取得されるためawaitが必要
  const productId = resolvedParams.id; // URLパラメータから商品IDを取得

  // 商品データ、レビューデータ、お気に入り状態を並行して取得
  const [product, reviewsResponse, isFavorite] = await Promise.all([
    getProduct(productId),
    getReviews(productId),
    getFavoriteStatus(productId)
  ]);

  // 商品が見つからない場合は404ページを表示
  if (!product) {
    notFound(); // returnは不要
  }

  // レビュー表示に必要な情報を取得
  const reviews = Array.isArray(reviewsResponse) ? [] : reviewsResponse.reviews;
  const rating = Array.isArray(reviewsResponse) ? 0 : reviewsResponse.review_avg;
  const reviewCount = Array.isArray(reviewsResponse) ? 0 : reviewsResponse.pagination.totalItems;

  // 在庫数が存在しない場合の対策
  const stock = product.stock ?? 0;

  // 在庫状況に応じて表示テキストとスタイルを切り替え
  let stockText = '売り切れ';
  let stockStyle = 'text-red-600';
  if (stock > 10) {
    stockText = '在庫あり';
    stockStyle = 'text-green-600';
  } else if (stock > 0) {
    stockText = '在庫わずか';
    stockStyle = 'text-orange-500';
  }

  // 画像の指定がなければダミー画像を表示
  const finalImageUrl = product.image_url
    ? `/uploads/${product.image_url}`
    : '/images/no-image.jpg';

  // ログイン状態を取得
  const loggedIn = await isLoggedIn();

  return (
    <main className="container mx-auto px-4 py-8">
      <div className="flex flex-col md:flex-row gap-8">
        <Image
          src={finalImageUrl}
          alt={product.name || '商品画像'}
          width={800}
          height={800}
          className="w-full object-contain md:w-1/2 max-h-[600px]"
        />
        <div className="w-full md:w-1/2 space-y-6 pt-4">
          <h1>{product.name}</h1>
          <p className="text-gray-700 whitespace-pre-line">{product.description}</p>
          <p className="text-3xl font-bold text-indigo-600">¥{product.price.toLocaleString()}<span className="text-base font-normal text-gray-500">（税込）</span></p>
          {reviewCount > 0 ? (
            <div className="flex items-center mb-4">
              <span className="text-yellow-500 text-xl mr-2">{displayStars(rating)}</span>
              <span className="text-gray-700 text-base">{rating.toFixed(1)}</span>
              <span className="text-gray-500 text-sm ml-2">（レビュー{reviewCount}件）</span>
            </div>
          ) : (
            <p className="text-gray-500 text-sm mb-4">まだレビューがありません</p>
          )}
          <p className={`text-base font-medium ${stockStyle}`}>在庫状況：{stockText}</p>

          <div className="space-y-6 mt-8">
            {stock > 0 && (
              <CartControls
                cartItem={{
                  id: product.id.toString(),
                  title: product.name,
                  price: product.price,
                  imageUrl: product.image_url ?? '',
                }}
                stock={stock}
                loggedIn={loggedIn}
              />
            )}
            {loggedIn && (
              <FavoriteControls
                productId={product.id}
                initialFavorite={isFavorite}
              />
            )}
          </div>
        </div>
      </div>

      <div className="flex flex-col md:flex-row gap-8 mt-6 p-4 border border-gray-300 rounded-md shadow-sm">
        <div className="w-full md:w-1/2">
          <h2 className="mt-2">レビュー一覧</h2>
          {reviewCount > 0 ? (
            <ul className="space-y-4 list-none">
              {reviews.slice(0, 3).map(r => (
                <li key={r.id} className="border-b border-gray-300 pb-2">
                  <div className="flex items-center text-sm text-yellow-500 mb-1">
                    {displayStars(r.score)}
                  </div>
                  <p className="text-gray-800">{r.content}</p>
                  <p className="text-xs text-gray-500 mt-1">
                    {r.user_name} さん {new Date(r.created_at).toLocaleDateString()}
                  </p>
                </li>
              ))}
              {reviewCount > 3 && (
                <div className="text-center mt-4">
                  <Link href={`/products/${productId}/reviews`} className="text-indigo-600 hover:underline">
                    すべてのレビューを見る（{reviewCount}件）
                  </Link>
                </div>
              )}
            </ul>
          ) : (
            <p className="text-gray-500">まだレビューがありません。</p>
          )}
        </div>
        <div className="w-full md:w-1/2 border-l border-gray-200 pl-6">
          <ReviewControls productId={product.id} loggedIn={loggedIn} />
        </div>
      </div>
      <div className="mt-8 pt-4 border-t border-gray-200">
        <Link href="/products" className="text-indigo-600 hover:underline">
          ← 商品一覧に戻る
        </Link>
      </div>
    </main>
  );
}