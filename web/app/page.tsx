'use client';

import { FormEvent, useEffect, useMemo, useState } from 'react';

import {
  AuthResponse,
  Post,
  User,
  createPost,
  listPosts,
  login,
  signup
} from '@/lib/api';

type AuthMode = 'login' | 'signup';

const tokenStorageKey = 'c2dp-token';
const userStorageKey = 'c2dp-user';

export default function Home() {
  const [authMode, setAuthMode] = useState<AuthMode>('signup');
  const [email, setEmail] = useState('');
  const [identifier, setIdentifier] = useState('');
  const [password, setPassword] = useState('');
  const [username, setUsername] = useState('');
  const [caption, setCaption] = useState('');
  const [image, setImage] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [user, setUser] = useState<User | null>(null);
  const [posts, setPosts] = useState<Post[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);
  const [isAuthLoading, setIsAuthLoading] = useState(false);
  const [isPostLoading, setIsPostLoading] = useState(false);
  const [isFeedLoading, setIsFeedLoading] = useState(true);

  const isLoggedIn = Boolean(token && user);

  useEffect(() => {
    const savedToken = localStorage.getItem(tokenStorageKey);
    const savedUser = localStorage.getItem(userStorageKey);
    if (savedToken && savedUser) {
      setToken(savedToken);
      setUser(JSON.parse(savedUser) as User);
    }
  }, []);

  useEffect(() => {
    void refreshPosts();
  }, []);

  useEffect(() => {
    if (!image) {
      setImagePreview(null);
      return;
    }

    const previewURL = URL.createObjectURL(image);
    setImagePreview(previewURL);
    return () => URL.revokeObjectURL(previewURL);
  }, [image]);

  const headline = useMemo(() => {
    if (isLoggedIn) return `Welcome back, ${user?.username}`;
    return authMode === 'signup' ? 'Create your account' : 'Log in to post';
  }, [authMode, isLoggedIn, user?.username]);

  async function refreshPosts() {
    setIsFeedLoading(true);
    setError(null);
    try {
      setPosts(await listPosts());
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not load posts');
    } finally {
      setIsFeedLoading(false);
    }
  }

  function persistSession(auth: AuthResponse) {
    localStorage.setItem(tokenStorageKey, auth.token);
    localStorage.setItem(userStorageKey, JSON.stringify(auth.user));
    setToken(auth.token);
    setUser(auth.user);
  }

  async function handleAuth(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setNotice(null);
    setIsAuthLoading(true);

    try {
      const auth =
        authMode === 'signup'
          ? await signup({ email, password, username })
          : await login({ identifier, password });
      persistSession(auth);
      setPassword('');
      setNotice(authMode === 'signup' ? 'Account created.' : 'Logged in.');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Authentication failed');
    } finally {
      setIsAuthLoading(false);
    }
  }

  async function handleCreatePost(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!token) return;
    if (!image) {
      setError('Choose an image before posting.');
      return;
    }

    setError(null);
    setNotice(null);
    setIsPostLoading(true);
    const form = event.currentTarget;

    try {
      const post = await createPost({ caption, image, token });
      setPosts((currentPosts) => [post, ...currentPosts]);
      setCaption('');
      setImage(null);
      setNotice('Post uploaded.');
      form.reset();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not create post');
    } finally {
      setIsPostLoading(false);
    }
  }

  function logout() {
    localStorage.removeItem(tokenStorageKey);
    localStorage.removeItem(userStorageKey);
    setToken(null);
    setUser(null);
    setNotice('Logged out.');
  }

  return (
    <main className="min-h-screen bg-[radial-gradient(circle_at_top_left,#e7f0ff,transparent_34rem),linear-gradient(180deg,#ffffff_0%,#f4f7fb_38%)]">
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-8 px-4 py-6 sm:px-6 lg:px-8">
        <header className="flex flex-col gap-4 rounded-3xl border border-white/70 bg-white/85 p-5 shadow-sm backdrop-blur md:flex-row md:items-center md:justify-between">
          <div>
            <p className="text-sm font-semibold uppercase tracking-[0.22em] text-chelsea">
              C2DP
            </p>
            <h1 className="mt-2 text-3xl font-bold tracking-tight text-slate-950 sm:text-4xl">
              Image posts, bare bones.
            </h1>
          </div>
          {isLoggedIn ? (
            <button
              className="rounded-full border border-slate-200 px-4 py-2 text-sm font-semibold text-slate-700 transition hover:border-chelsea hover:text-chelsea"
              onClick={logout}
              type="button">
              Log out
            </button>
          ) : null}
        </header>

        <section className="grid gap-6 lg:grid-cols-[380px_1fr]">
          <aside className="h-fit rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
            <h2 className="text-xl font-bold text-slate-950">{headline}</h2>
            <p className="mt-2 text-sm leading-6 text-slate-500">
              Sign up, log in, and upload a required image with every post.
            </p>

            {error ? (
              <div className="mt-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm font-medium text-red-700">
                {error}
              </div>
            ) : null}
            {notice ? (
              <div className="mt-4 rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm font-medium text-emerald-700">
                {notice}
              </div>
            ) : null}

            {isLoggedIn ? (
              <form className="mt-6 space-y-4" onSubmit={handleCreatePost}>
                <label className="block">
                  <span className="text-sm font-semibold text-slate-700">
                    Image
                  </span>
                  <input
                    accept="image/*"
                    className="mt-2 block w-full rounded-2xl border border-slate-200 bg-white px-3 py-2 text-sm file:mr-3 file:rounded-full file:border-0 file:bg-chelsea file:px-3 file:py-2 file:text-sm file:font-semibold file:text-white"
                    onChange={(event) =>
                      setImage(event.target.files?.[0] ?? null)
                    }
                    required
                    type="file"
                  />
                </label>

                {imagePreview ? (
                  <img
                    alt="Selected upload preview"
                    className="aspect-square w-full rounded-3xl object-cover"
                    src={imagePreview}
                  />
                ) : (
                  <div className="flex aspect-square w-full items-center justify-center rounded-3xl border border-dashed border-slate-300 bg-slate-50 text-sm font-medium text-slate-400">
                    Image preview
                  </div>
                )}

                <label className="block">
                  <span className="text-sm font-semibold text-slate-700">
                    Caption
                  </span>
                  <textarea
                    className="mt-2 min-h-28 w-full resize-none rounded-2xl border border-slate-200 px-4 py-3 text-sm outline-none ring-chelsea/20 transition placeholder:text-slate-400 focus:border-chelsea focus:ring-4"
                    maxLength={2200}
                    onChange={(event) => setCaption(event.target.value)}
                    placeholder="Say something about the image"
                    value={caption}
                  />
                </label>

                <button
                  className="w-full rounded-2xl bg-chelsea px-5 py-3 text-sm font-bold text-white shadow-lg shadow-chelsea/20 transition hover:bg-chelsea-dark disabled:cursor-not-allowed disabled:opacity-60"
                  disabled={isPostLoading}
                  type="submit">
                  {isPostLoading ? 'Uploading...' : 'Post image'}
                </button>
              </form>
            ) : (
              <>
                <div className="mt-6 grid grid-cols-2 rounded-2xl bg-slate-100 p-1">
                  {(['signup', 'login'] as const).map((mode) => (
                    <button
                      className={`rounded-xl px-3 py-2 text-sm font-bold transition ${
                        authMode === mode
                          ? 'bg-white text-chelsea shadow-sm'
                          : 'text-slate-500'
                      }`}
                      key={mode}
                      onClick={() => setAuthMode(mode)}
                      type="button">
                      {mode === 'signup' ? 'Sign up' : 'Log in'}
                    </button>
                  ))}
                </div>

                <form className="mt-5 space-y-4" onSubmit={handleAuth}>
                  {authMode === 'signup' ? (
                    <label className="block">
                      <span className="text-sm font-semibold text-slate-700">
                        Username
                      </span>
                      <input
                        className="mt-2 w-full rounded-2xl border border-slate-200 px-4 py-3 text-sm outline-none ring-chelsea/20 transition focus:border-chelsea focus:ring-4"
                        onChange={(event) => setUsername(event.target.value)}
                        required
                        value={username}
                      />
                    </label>
                  ) : null}

                  <label className="block">
                    <span className="text-sm font-semibold text-slate-700">
                      {authMode === 'signup' ? 'Email' : 'Email or username'}
                    </span>
                    <input
                      className="mt-2 w-full rounded-2xl border border-slate-200 px-4 py-3 text-sm outline-none ring-chelsea/20 transition focus:border-chelsea focus:ring-4"
                      onChange={(event) => authMode === 'signup' ? setEmail(event.target.value) : setIdentifier(event.target.value)}
                      required
                      type={authMode === 'signup' ? 'email' : 'text'}
                      autoComplete={authMode === 'signup' ? 'email' : 'username'}
                      autoCapitalize="none"
                      spellCheck={false}
                      value={authMode === 'signup' ? email : identifier}
                    />
                  </label>

                  <label className="block">
                    <span className="text-sm font-semibold text-slate-700">
                      Password
                    </span>
                    <input
                      className="mt-2 w-full rounded-2xl border border-slate-200 px-4 py-3 text-sm outline-none ring-chelsea/20 transition focus:border-chelsea focus:ring-4"
                      minLength={8}
                      onChange={(event) => setPassword(event.target.value)}
                      required
                      type="password"
                      value={password}
                    />
                  </label>

                  <button
                    className="w-full rounded-2xl bg-chelsea px-5 py-3 text-sm font-bold text-white shadow-lg shadow-chelsea/20 transition hover:bg-chelsea-dark disabled:cursor-not-allowed disabled:opacity-60"
                    disabled={isAuthLoading}
                    type="submit">
                    {isAuthLoading
                      ? 'Please wait...'
                      : authMode === 'signup'
                        ? 'Create account'
                        : 'Log in'}
                  </button>
                </form>
              </>
            )}
          </aside>

          <section className="min-w-0">
            <div className="mb-4 flex items-center justify-between">
              <div>
                <h2 className="text-2xl font-bold text-slate-950">Feed</h2>
                <p className="text-sm text-slate-500">
                  Every post has an image.
                </p>
              </div>
              <button
                className="rounded-full border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700 transition hover:border-chelsea hover:text-chelsea"
                onClick={() => void refreshPosts()}
                type="button">
                Refresh
              </button>
            </div>

            {isFeedLoading ? (
              <div className="rounded-3xl border border-slate-200 bg-white p-8 text-center text-sm font-medium text-slate-500">
                Loading posts...
              </div>
            ) : posts.length ? (
              <div className="grid gap-5 sm:grid-cols-2">
                {posts.map((post) => (
                  <article
                    className="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm"
                    key={post.id}>
                    <img
                      alt={post.caption || `Post by ${post.user.username}`}
                      className="aspect-square w-full bg-slate-100 object-cover"
                      src={post.image_url}
                    />
                    <div className="space-y-3 p-4">
                      <div className="flex items-center justify-between gap-3">
                        <p className="font-bold text-slate-950">
                          @{post.user.username}
                        </p>
                        <time className="text-xs font-medium text-slate-400">
                          {new Date(post.created_at).toLocaleDateString()}
                        </time>
                      </div>
                      {post.caption ? (
                        <p className="text-sm leading-6 text-slate-600">
                          {post.caption}
                        </p>
                      ) : (
                        <p className="text-sm italic text-slate-400">
                          No caption.
                        </p>
                      )}
                    </div>
                  </article>
                ))}
              </div>
            ) : (
              <div className="rounded-3xl border border-dashed border-slate-300 bg-white p-10 text-center">
                <p className="text-lg font-bold text-slate-900">No posts yet</p>
                <p className="mt-2 text-sm text-slate-500">
                  Sign up and upload the first image.
                </p>
              </div>
            )}
          </section>
        </section>
      </div>
    </main>
  );
}
