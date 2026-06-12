import { FormEvent, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import type { CurrentUser } from '../lib/types';
import { sanitizeReturnTo, setToken } from '../lib/auth';

interface AuthPageProps {
  mode: 'login' | 'register';
  onAuthenticated?: () => void;
}

export function AuthPage({ mode, onAuthenticated }: AuthPageProps) {
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    const data = new FormData(event.currentTarget);
    const payload = Object.fromEntries(data.entries());

    try {
      const response = await fetch(`/api/auth/${mode}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        setError(await formatAuthError(response, mode));
        return;
      }

      const json = (await response.json()) as { token?: string; user?: CurrentUser };
      if (!json.token || !json.user) {
        setError('Authentication succeeded without a usable session. Please try again.');
        return;
      }

      setToken(json.token);
      onAuthenticated?.();
      const requestedReturn = new URLSearchParams(window.location.search).get('return_to');
      const defaultPath = mode === 'login' && json.user.role === 'admin' ? '/admin' : '/recent';
      navigate(sanitizeReturnTo(requestedReturn, defaultPath), { replace: true });
    } catch {
      setError('Unable to reach the server. Check your connection and try again.');
    }
  }

  return (
    <section className="glass auth-panel">
      <p className="eyebrow">{mode === 'login' ? 'LOGIN' : 'REGISTER'}</p>
      <h1>{mode === 'login' ? 'Welcome back' : 'Create reader account'}</h1>
      <form className="auth-form" onSubmit={submit}>
        {mode === 'register' ? <input name="display_name" placeholder="Display name" required /> : null}
        <input name="email" type="email" placeholder="Email" required />
        <input name="password" type="password" placeholder="Password" required minLength={8} />
        {error ? <p className="form-error">{error}</p> : null}
        <button className="primary-button" type="submit">{mode === 'login' ? 'Log in' : 'Register'}</button>
      </form>
      <p className="muted">
        {mode === 'login' ? 'Need an account? ' : 'Already registered? '}
        <Link to={mode === 'login' ? '/register' : '/login'}>{mode === 'login' ? 'Register' : 'Log in'}</Link>
      </p>
    </section>
  );
}


async function formatAuthError(response: Response, mode: 'login' | 'register'): Promise<string> {
  const serverMessage = await readServerError(response);

  if (response.status === 401) {
    return 'Invalid email or password.';
  }
  if (response.status === 409) {
    return 'This email is already registered. Log in instead.';
  }
  if (response.status === 400) {
    if (/password/i.test(serverMessage)) {
      return 'Password must be at least 8 characters.';
    }
    if (/email/i.test(serverMessage)) {
      return 'Enter a valid email address.';
    }
    return mode === 'login' ? 'Check your email and password.' : 'Check the registration details and try again.';
  }
  if (response.status >= 500) {
    return 'The server could not complete authentication. Try again later.';
  }

  return mode === 'login' ? 'Unable to log in. Try again.' : 'Unable to register this account.';
}

async function readServerError(response: Response): Promise<string> {
  try {
    const payload: unknown = await response.json();
    if (typeof payload === 'object' && payload !== null && 'error' in payload && typeof payload.error === 'string') {
      return payload.error;
    }
  } catch {
    // Keep status-based guidance when the error body is empty or malformed.
  }
  return '';
}
