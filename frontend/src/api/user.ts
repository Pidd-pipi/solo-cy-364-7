import request from '@/utils/request'
import type { PageResult, User } from '@/types'

export interface LoginResult {
  token: string
  user: User
}

export function register(data: { username: string; password: string; name: string; role: string; store_id?: number }) {
  return request.post('/auth/register', data)
}

export function login(data: { username: string; password: string }): Promise<LoginResult> {
  return request.post('/auth/login', data)
}

export function getMe(): Promise<User> {
  return request.get('/users/me')
}

export function updateProfile(data: { name?: string; store_id?: number }): Promise<User> {
  return request.put('/users/me', data)
}

export function listUsers(params: { page?: number; page_size?: number } = {}): Promise<PageResult<User>> {
  return request.get('/users', { params })
}
