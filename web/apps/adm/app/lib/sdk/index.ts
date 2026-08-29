import Apis, { Api } from './gen'
import type { Method } from 'alova'

function ok(_response: Response, method: Method) {
  const t = useNuxtApp().$i18n.t
  if (method.type !== 'GET') useSnackBar().success(t('success'))
}

function unauthorized(_response: Response, _method: Method) {
  useAuth().$signOut()
}

function fallback(response: Response, _method: Method, v: any) {
  const t = useNuxtApp().$i18n.t
  if (v?.code) {
    useSnackBar().warn(t(v.code))
  } else {
    useSnackBar().warn(v?.message)
  }
  throw response.statusText
}

Api.NewEvent(200, ok)
Api.NewEvent(401, unauthorized)
Api.NewEvent(-1, fallback)

export default Apis
export { Apis }
export { Api }
export * from './gen/models'
export * from './etc'
