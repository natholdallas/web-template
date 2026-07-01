import type { Method } from 'alova'
import { BaseQueries, Model, Page } from './etc'

export const api = sneakyfetch(useRuntimeConfig().public.apiBase + '/adm/api/v1')
api.NewEvent(200, ok)
api.NewEvent(401, unauthorized)
api.NewEvent(-1, fallback)

export function authorization(config: RequestConfig) {
  if (!config.headers) config.headers = {}
  const { token } = useAuth()
  if (token) {
    config.headers['Authorization'] = `Bearer ${token}`
  }
}

function ok(_response: Response, method: Method) {
  const t = useNuxtApp().$i18n.t
  if (method.type !== 'GET') useSnackBar().success(t('success'))
}

function unauthorized(_response: Response, _method: Method) {
  useAuth().$signOut()
}

function fallback(response: Response, _method: Method, v: any) {
  const t = useNuxtApp().$i18n.t
  if (v.code) {
    useSnackBar().warn(t(v.code))
  } else {
    useSnackBar().warn(v.message)
  }
  throw response.statusText
}

export type AuthIn = typeof AuthIn
export const AuthIn = {
  username: '',
  password: '',
}

export type Auth = typeof Auth
export const Auth = {
  id: 0,
  token: '',
}

const auths = api.Group('/auth')
export const SignIn = (data: AuthIn) => auths.Post<Auth>({ sub: '/in', data })

export type User = typeof User
export const User = {
  ...Model,
  username: '',
  password: '',
}

const users = api.Group('/user', authorization)
const userInc = ['username', 'password']
export const ListUser = (queries: BaseQueries) => users.Get<Page<User>>({ queries })
export const FindUser = (id: number) => users.Get<User>({ params: [id] })
export const CreateUser = (data: User) => users.Post({ data, includes: userInc })
export const UpdateUser = (data: User) => users.Put({ params: [data.id], data, includes: userInc })
export const RemoveUser = (id: number) => users.Delete({ params: [id] })

export type Admin = typeof Admin
export const Admin = {
  ...Model,
  username: '',
  password: '',
}

const admins = api.Group('/admin', authorization)
const adminInc = ['username', 'password']
export const ListAdmin = (queries: BaseQueries) => admins.Get<Page<Admin>>({ queries })
export const FindAdmin = (id: number) => admins.Get<Admin>({ params: [id] })
export const CreateAdmin = (data: Admin) => admins.Post({ data, includes: adminInc })
export const UpdateAdmin = (data: Admin) => admins.Put({ params: [data.id], data, includes: adminInc })
export const RemoveAdmin = (id: number) => admins.Delete({ params: [id] })
