import { api, Profile } from '@sdk/client/usr'

// Current user's profile, shared by the settings page and the header avatar so
// both stay in sync after an avatar upload or profile edit.
export const useUserStore = defineStore('user', () => {
  const profile = ref<Profile>(inst(Profile))
  const loading = ref(false)

  async function $load(force = false) {
    if (!force && profile.value.id) return
    loading.value = true
    try {
      Object.assign(profile.value, await api.profile.find())
    } catch {
      // 401 is handled by the SDK; any other failure keeps the previous value
      // so the page still renders.
    } finally {
      loading.value = false
    }
  }

  function $reset() {
    profile.value = inst(Profile)
  }

  return { profile, loading, $load, $reset }
})
