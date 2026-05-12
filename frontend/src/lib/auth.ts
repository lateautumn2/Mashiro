const TOKEN_KEY = 'token';
const USERNAME_KEY = 'username';

export function getToken() {
  return localStorage.getItem(TOKEN_KEY);
}

export function getUsername() {
  return localStorage.getItem(USERNAME_KEY) ?? 'admin';
}

export function isAuthenticated() {
  return Boolean(getToken());
}

export function isTokenExpired(): boolean {
  const token = getToken();
  if (!token) return true;

  try {
    const payloadBase64 = token.split('.')[1];
    if (!payloadBase64) return true;

    const payloadJson = atob(payloadBase64);
    const payload = JSON.parse(payloadJson);

    if (payload.exp && typeof payload.exp === 'number') {
      return Date.now() >= payload.exp * 1000;
    }

    return false;
  } catch {
    return true;
  }
}

export function saveAuth(token: string, username: string) {
  localStorage.setItem(TOKEN_KEY, token);
  localStorage.setItem(USERNAME_KEY, username);
}

export function clearAuth() {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(USERNAME_KEY);
}

export function clearAuthAndRedirect() {
  clearAuth();
  window.location.href = '/admin/login';
}
