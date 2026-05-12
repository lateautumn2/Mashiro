import { clearAuthAndRedirect } from './auth';

const DEFAULT_API_BASE_URL = '/api';

function normalizeBaseUrl(url: string) {
  return url.replace(/\/$/, '');
}

export class ApiError extends Error {
  status?: number;

  constructor(message: string, status?: number) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

export function getApiBaseUrl() {
  const envUrl = import.meta.env.VITE_API_BASE_URL;
  if (envUrl) {
    return normalizeBaseUrl(envUrl);
  }

  return DEFAULT_API_BASE_URL;
}

function buildHeaders(headers: HeadersInit | undefined, requireAuth: boolean) {
  const nextHeaders = new Headers(headers);

  if (requireAuth) {
    const token = localStorage.getItem('token');
    if (!token) {
      throw new Error('未找到登录令牌，请先登录后再访问管理接口。');
    }

    nextHeaders.set('Authorization', `Bearer ${token}`);
  }

  return nextHeaders;
}

export async function apiRequest<T>(
  path: string,
  init: RequestInit = {},
  options: { auth?: boolean } = {},
) {
  const headers = buildHeaders(init.headers, options.auth ?? false);
  if (init.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  const response = await fetch(`${getApiBaseUrl()}${path}`, {
    ...init,
    headers,
  }).catch((error: unknown) => {
    throw new ApiError(
      error instanceof Error
        ? `无法连接后端服务，请确认后端已启动并监听 ${getApiBaseUrl()}`
        : '无法连接后端服务',
    );
  });

  const contentType = response.headers.get('content-type') ?? '';
  const payload = contentType.includes('application/json')
    ? await response.json()
    : await response.text();

  if (!response.ok) {
    if (response.status === 401 && options.auth) {
      clearAuthAndRedirect();
      throw new ApiError('登录已过期，请重新登录', 401);
    }

    const message =
      typeof payload === 'object' && payload && 'error' in payload
        ? String(payload.error)
        : `请求失败：${response.status}`;
    throw new ApiError(message, response.status);
  }

  return payload as T;
}
