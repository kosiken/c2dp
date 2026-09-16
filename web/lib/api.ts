export type User = {
  created_at: string;
  email: string;
  id: number;
  updated_at: string;
  username: string;
};

export type Post = {
  caption: string;
  created_at: string;
  id: number;
  image_url: string;
  updated_at: string;
  user: User;
  user_id: number;
};

export type AuthResponse = {
  token: string;
  user: User;
};

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080';

type RequestOptions = {
  body?: BodyInit;
  headers?: HeadersInit;
  method?: string;
  token?: string;
};

export async function apiRequest<T>(
  path: string,
  { body, headers, method = 'GET', token }: RequestOptions = {}
): Promise<T> {
  const response = await fetch(`${API_URL}${path}`, {
    body,
    headers: {
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...headers
    },
    method
  });

  if (!response.ok) {
    const error = await response.json().catch(() => null);
    throw new Error(error?.message ?? `Request failed with ${response.status}`);
  }

  return response.json() as Promise<T>;
}

export function signup(input: {
  email: string;
  password: string;
  username: string;
}) {
  return apiRequest<AuthResponse>('/api/v1/auth/signup', {
    body: JSON.stringify(input),
    headers: { 'Content-Type': 'application/json' },
    method: 'POST'
  });
}

export function login(input: { identifier: string; password: string }) {
  return apiRequest<AuthResponse>('/api/v1/auth/login', {
    body: JSON.stringify(input),
    headers: { 'Content-Type': 'application/json' },
    method: 'POST'
  });
}

export function listPosts() {
  return apiRequest<Post[]>('/api/v1/posts');
}

export function createPost(input: {
  caption: string;
  image: File;
  token: string;
}) {
  const formData = new FormData();
  formData.append('caption', input.caption);
  formData.append('image', input.image);

  return apiRequest<Post>('/api/v1/posts', {
    body: formData,
    method: 'POST',
    token: input.token
  });
}
