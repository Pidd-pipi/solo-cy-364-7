// UserRole 用户角色枚举（与 backend/internal/constants/user.go 对应）
export const UserRole = {
  HQ: 'hq',
  STORE_MANAGER: 'store_manager',
  ADMIN: 'admin'
} as const

export type UserRoleValue = (typeof UserRole)[keyof typeof UserRole]

export const USER_ROLE_TEXT: Record<UserRoleValue, string> = {
  [UserRole.HQ]: '总部',
  [UserRole.STORE_MANAGER]: '店长',
  [UserRole.ADMIN]: '管理员'
}

export const USER_ROLE_OPTIONS: { label: string; value: UserRoleValue }[] = [
  { label: '管理员', value: UserRole.ADMIN },
  { label: '总部', value: UserRole.HQ },
  { label: '店长', value: UserRole.STORE_MANAGER }
]

export function hasRole(userRole: string | undefined, roles: UserRoleValue[]): boolean {
  return !!userRole && (roles as string[]).includes(userRole)
}
