import { FormEvent, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { getReturnTo, setToken } from '../lib/auth';

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
        setError(mode === 'login' ? 'Invalid email or password.' : 'Unable to register this account.');
        return;
      }

      const json = (await response.json()) as { token?: string };
      if (!json.token) {
        setError('Authentication succeeded without a usable session. Please try again.');
        return;
      }

      setToken(json.token);
      onAuthenticated?.();
      navigate(getReturnTo('/recent'), { replace: true });
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
