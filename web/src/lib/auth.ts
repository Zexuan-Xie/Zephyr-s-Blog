const tokenKey = 'xlab-blog-token';

export function getToken(): string | null {
  return localStorage.getItem(tokenKey);
}

export function setToken(token: string): void {
  localStorage.setItem(tokenKey, token);
}

export function clearToken(): void {
  localStorage.removeItem(tokenKey);
}

export function sanitizeReturnTo(candidate: string | null, defaultPath = '/recent'): string {
  if (!candidate?.startsWith('/') || candidate.startsWith('//')) {
    return defaultPath;
  }

  const path = candidate.split(/[?#]/, 1)[0].replace(/\/+$/, '') || '/';
  if (path === '/login' || path === '/register') {
    return defaultPath;
  }

  return candidate;
}

export function getReturnTo(defaultPath = '/recent'): string {
  const params = new URLSearchParams(window.location.search);
  return sanitizeReturnTo(params.get('return_to'), defaultPath);
}
