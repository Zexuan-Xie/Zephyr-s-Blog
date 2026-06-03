import { FormEvent, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { getReturnTo, setToken } from '../lib/auth';

interface AuthPageProps {
  mode: 'login' | 'register';
}

export function AuthPage({ mode }: AuthPageProps) {
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    const payload = Object.fromEntries(data.entries());
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
    if (json.token) {
      setToken(json.token);
    }
    navigate(getReturnTo('/recent'), { replace: true });
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
