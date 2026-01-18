import { api } from '@/lib/api';

// =============================================================================
// Types
// =============================================================================

export type UserRole = 'operador' | 'gestor' | 'admin';

export interface User {
  id: string;
  email: string;
  nome: string;
  role: UserRole;
  ativo: boolean;
  mobile_phone?: string;
  email_notifications: boolean;
  hospitals?: { id: string; nome: string }[];
  tenant_id?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateUserInput {
  email: string;
  password: string;
  nome: string;
  role: UserRole;
  hospital_ids?: string[];
  mobile_phone?: string;
  email_notifications?: boolean;
}

export interface UpdateUserInput {
  password?: string;
  nome?: string;
  role?: UserRole;
  hospital_ids?: string[];
  mobile_phone?: string;
  email_notifications?: boolean;
  ativo?: boolean;
}

export interface PaginatedUsersResponse {
  data: User[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

// =============================================================================
// API Functions
// =============================================================================

/**
 * List all users for the current tenant
 */
export async function listUsers(params?: {
  page?: number;
  per_page?: number;
  search?: string;
  role?: UserRole;
}): Promise<PaginatedUsersResponse> {
  const searchParams = new URLSearchParams();
  if (params?.page) searchParams.set('page', params.page.toString());
  if (params?.per_page) searchParams.set('per_page', params.per_page.toString());
  if (params?.search) searchParams.set('search', params.search);
  if (params?.role) searchParams.set('role', params.role);

  const queryString = searchParams.toString();
  const url = `/users${queryString ? `?${queryString}` : ''}`;

  const { data } = await api.get<User[] | PaginatedUsersResponse>(url);

  // Handle both array response and paginated response
  if (Array.isArray(data)) {
    return {
      data,
      total: data.length,
      page: 1,
      per_page: data.length,
      total_pages: 1,
    };
  }

  return data;
}

/**
 * Get a single user by ID
 */
export async function getUser(id: string): Promise<User> {
  const { data } = await api.get<User>(`/users/${id}`);
  return data;
}

/**
 * Create a new user
 */
export async function createUser(input: CreateUserInput): Promise<User> {
  const { data } = await api.post<User>('/users', input);
  return data;
}

/**
 * Update a user
 */
export async function updateUser(id: string, input: UpdateUserInput): Promise<User> {
  const { data } = await api.patch<User>(`/users/${id}`, input);
  return data;
}

/**
 * Delete a user
 */
export async function deleteUser(id: string): Promise<void> {
  await api.delete(`/users/${id}`);
}
