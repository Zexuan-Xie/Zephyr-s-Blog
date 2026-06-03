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

export function getReturnTo(defaultPath = '/recent'): string {
  const params = new URLSearchParams(window.location.search);
  return params.get('return_to') || defaultPath;
}
