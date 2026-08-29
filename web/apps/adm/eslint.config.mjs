// @ts-check
import withNuxt from './.nuxt/eslint.config.mjs'
import vuetify from 'eslint-config-vuetify'

export default withNuxt(
  vuetify({
    ts: true,
  }),
)
// Your custom configs here
