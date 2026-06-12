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

const applicationOrigin = 'http://xlab.local';
const asciiControlOrBackslash = /[\u0000-\u001f\u007f\\]/;

function decodePercentEscapes(value: string): string {
  return value.replace(/%([0-9a-f]{2})/gi, (_, hex: string) =>
    String.fromCharCode(Number.parseInt(hex, 16)));
}

export function sanitizeReturnTo(candidate: string | null, defaultPath = '/recent'): string {
  if (!candidate?.startsWith('/') || candidate.startsWith('//')) {
    return defaultPath;
  }

  let inspected = candidate;
  while (true) {
    if (asciiControlOrBackslash.test(inspected) || inspected.startsWith('//')) {
      return defaultPath;
    }

    const decoded = decodePercentEscapes(inspected);
    if (decoded === inspected) {
      break;
    }
    inspected = decoded;
  }

  let target: URL;
  try {
    target = new URL(inspected, applicationOrigin);
  } catch {
    return defaultPath;
  }
  if (target.origin !== applicationOrigin) {
    return defaultPath;
  }

  const path = target.pathname.replace(/\/+$/, '') || '/';
  if (path === '/login' || path === '/register') {
    return defaultPath;
  }

  return candidate;
}

export function getReturnTo(defaultPath = '/recent'): string {
  const params = new URLSearchParams(window.location.search);
  return sanitizeReturnTo(params.get('return_to'), defaultPath);
}
