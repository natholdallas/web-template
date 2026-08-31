// @ts-check
import withNuxt from './.nuxt/eslint.config.mjs'

export default withNuxt(
  {
    ignores: ['app/lib/sdk/**'],
  },
  {
    files: ['app/components/ui/**/*.{vue,ts}'],
    rules: {
      // shadcn-vue components intentionally keep `any` and optional props
      // without defaults (class/variant/size passthrough).
      '@typescript-eslint/no-explicit-any': 'off',
      'vue/require-default-prop': 'off',
    },
  },
)
// Your custom configs here
