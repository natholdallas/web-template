import { Admin, api } from '@sdk/client/adm'

// Current admin's profile, shared by the settings page, the app bar and the
// sidebar so every avatar surface stays in sync without a reload.
export const useProfileStore = defineStore('profile', () => {
  const profile = ref<Admin>(inst(Admin))
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
    profile.value = inst(Admin)
  }

  return { profile, loading, $load, $reset }
})
