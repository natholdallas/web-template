import type { Method } from 'alova'
import { toast } from 'vue-sonner'
import { Model } from './etc'

export const api = sneakyfetch(useRuntimeConfig().public.apiBase + '/user/api/v1')
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
  if (method.type !== 'GET') toast.success(t('success'))
}

function unauthorized(_response: Response, _method: Method) {
  useAuth().$signOut()
}

function fallback(response: Response, _method: Method, v: any) {
  const t = useNuxtApp().$i18n.t
  if (v.code) {
    toast.warning(t(v.code))
  } else {
    toast.error(v.message)
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

export type ResetPasswordIn = typeof ResetPasswordIn
export const ResetPasswordIn = {
  old: '',
  new: '',
  confirm: '',
}

const users = api.Group('/user', authorization)
const userInc = ['username']
export const FindUser = () => users.Get<User>()
export const UpdateUser = (data: User) => users.Put({ params: [data.id], data, includes: userInc })
export const RstPwd = (data: ResetPasswordIn) => users.Post({ sub: '/reset/password', data })
