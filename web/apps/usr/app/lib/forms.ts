import { ResetPasswordIn } from '@sdk/client/usr'

// `confirm` is a client-only form field: the API only takes old/new, so the
// form model extends the generated request model instead of the SDK mirroring a
// field the backend does not have.
export type ResetPasswordForm = ResetPasswordIn & { confirm: string }

export const ResetPasswordForm: ResetPasswordForm = { ...ResetPasswordIn, confirm: '' }
