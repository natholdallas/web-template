import { defineConfig, presetIcons, presetWind4, transformerDirectives } from 'unocss'

export default defineConfig({
  presets: [presetWind4(), presetIcons()],
  transformers: [transformerDirectives()],
  theme: {
    breakpoints: {
      sm: '600px',
      md: '960px',
      lg: '1280px',
    },
  },
})
