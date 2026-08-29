import Apis, { Api } from './gen'
import type { Method } from 'alova'
import { toast } from 'vue-sonner'
import { ResetPasswordIn as GenResetPasswordIn } from './gen/models'
import type { ResetPasswordIn as GenResetPasswordInT } from './gen/models'

function ok(_response: Response, method: Method) {
  const t = useNuxtApp().$i18n.t
  if (method.type !== 'GET') toast.success(t('success'))
}

function unauthorized(_response: Response, _method: Method) {
  useAuth().$signOut()
}

function fallback(response: Response, _method: Method, v: any) {
  const t = useNuxtApp().$i18n.t
  if (v?.code) {
    toast.warning(t(v.code))
  } else {
    toast.error(v?.message)
  }
  throw response.statusText
}

Api.NewEvent(200, ok)
Api.NewEvent(401, unauthorized)
Api.NewEvent(-1, fallback)

// ResetPasswordIn adds a client-only confirm field for the form.
export type ResetPasswordIn = GenResetPasswordInT & { confirm: string }
export const ResetPasswordIn: ResetPasswordIn = { ...GenResetPasswordIn, confirm: '' }

export default Apis
export { Apis }
export { Api }
export { Auth, AuthIn, User, UserIn } from './gen/models'
export * from './etc'
