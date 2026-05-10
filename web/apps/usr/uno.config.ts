import shadcn from '@natholdallas/shadcn/config'
import {
  defineConfig,
  mergeConfigs,
  presetIcons,
  transformerCompileClass,
  transformerDirectives,
  transformerVariantGroup,
} from 'unocss'

export default defineConfig(
  mergeConfigs([
    shadcn(),
    {
      presets: [presetIcons()],
      transformers: [transformerVariantGroup(), transformerDirectives(), transformerCompileClass()],
      theme: {
        font: {
          mono: '"Maple Mono", ui-monospace, monospace',
        },
      },
      shortcuts: {
        'px-container': 'px-6',
        'py-container': 'py-12',
      },
    },
  ]),
)
