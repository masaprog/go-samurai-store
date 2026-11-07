'use client'; // クライアント（ブラウザ）側で動作

// 並べ替えコンポーネントに渡すデータ（props）の型定義
interface Props {
  sort: string; // 並べ替え条件
  perPage: number; // 1ページあたりの表示件数
  keyword: string; // 検索キーワード
}

// 並べ替えコンポーネント
export default function Sort({ sort, perPage, keyword }: Props) {
  return (
    <form action="/products" method="GET" className="flex flex-col md:flex-row gap-4">
      <input type="hidden" name="page" value="1" />
      <input type="hidden" name="perPage" value={perPage} />
      <input type="hidden" name="keyword" value={keyword} />
      <select
        name="sort"
        value={sort}
        className="border border-gray-300 rounded px-4 py-2 w-full md:w-48"
        onChange={(e) => e.currentTarget.form?.submit()}
      >
        <option value="new">新着順</option>
        <option value="priceAsc">価格が安い順</option>
      </select>
    </form>
  );
}