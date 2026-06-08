import request from './request'

export interface LoginResult {
  token: string
  admin: { id: number; username: string; nickname: string; role: string }
}

export function adminLogin(username: string, password: string): Promise<LoginResult> {
  return request.post('/api/v1/admin/auth/login', { username, password }) as Promise<LoginResult>
}
