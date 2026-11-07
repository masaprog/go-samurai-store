import { cookies } from 'next/headers';

// 認証済みユーザーの型定義
export type AuthUser = {
  userId: number;
  name: string;
  email: string;
  isAdmin: boolean;
};

// 認証に用いるトークン名
export const AUTH_TOKEN = 'authToken';

// 認証済みユーザーの情報を取得
export async function getAuthUser(): Promise<AuthUser | null> {
  const cookieStore = await cookies(); // クッキーを非同期で取得
  const token = cookieStore.get(AUTH_TOKEN)?.value;
  if (!token) { // トークンが存在しない
    return null;
  }

  // トークン検証結果を取得
  try {
    const apiUrl = process.env.API_BASE_URL || 'http://backend:8080';
    const res = await fetch(`${apiUrl}/api/users/me`, {
      headers: {
        'Cookie': `${AUTH_TOKEN}=${token}`,
      },
      // 常に最新の情報を取得するためキャッシュを無効化
      cache: 'no-store',
    });

    if (!res.ok) {
      return null;
    }

    const user: AuthUser = await res.json();
    return user;

  } catch (err) {
    console.error('認証情報の取得に失敗しました:', err);
    return null;
  }
}

// ユーザーがログイン済みかどうかをチェック
export async function isLoggedIn(): Promise<boolean> {
  const user = await getAuthUser();
  return user !== null;
}

// ユーザーが管理者かどうかをチェック
export async function isAdmin(): Promise<boolean> {
  const user = await getAuthUser();
  return user?.isAdmin ?? false;
}