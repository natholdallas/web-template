// @ts-check
import withNuxt from './.nuxt/eslint.config.mjs'

export default withNuxt(
  {
    ignores: ['app/lib/sdk/**'],
  },
  {
    files: ['**/*.vue'],
    rules: {
      // Vuetify uses dotted slot names (`#item.data-table-expand`) which
      // eslint-plugin-vue misreads as modifiers.
      'vue/valid-v-slot': 'off',
    },
  },
)
// Your custom configs here
