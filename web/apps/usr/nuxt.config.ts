// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  future: { compatibilityVersion: 4 },
  compatibilityDate: 'latest',
  devtools: { enabled: false },
  experimental: {
    serverAppConfig: false,
    typedPages: true,
  },

  css: [
    '~/assets/main.css', // css
    '~/assets/fonts.css',
    '~/assets/transition.css',
    '~/assets/utilities.css',
  ],
  modules: [
    '@natholdallas/alova',
    '@natholdallas/i18n',
    '@natholdallas/infra',
    '@natholdallas/shadcn',
    '@natholdallas/pinia',
    '@natholdallas/unocss',
    '@natholdallas/tauri',
  ],

  ssr: false,
  build: { analyze: true },
  devServer: { port: 3000 },
  app: {
    buildAssetsDir: 'static',
    rootAttrs: { id: 'root' },
    head: {
      meta: [
        {
          name: 'viewport',
          content: 'width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no',
        },
      ],
    },
  },
  runtimeConfig: {
    public: {
      apiBase: '',
    },
  },

  imports: {
    presets: [{ from: 'lodash', imports: ['isArray', 'isEmpty', 'cloneDeep'] }],
  },

  nitro: {
    compressPublicAssets: true,
    output: {
      dir: 'dist',
      serverDir: 'dist/server',
      publicDir: 'dist/public',
    },
  },
  vite: {
    build: {
      terserOptions: {
        compress: {
          drop_console: true,
          drop_debugger: true,
        },
      },
    },
    esbuild: {
      drop: ['debugger', 'console'],
    },
    optimizeDeps: {
      include: [
        'alova/client',
        'lodash', // CJS
        'copy-to-clipboard',
        'alova',
        'alova/fetch',
        'alova/vue',
        '@vueuse/core',
        'lucide-vue-next',
        'vue-sonner',
        'class-variance-authority',
        'vee-validate',
        'reka-ui',
        'lodash/cloneDeep', // CJS
        'clsx',
        'tailwind-merge',
        'lodash/isArray', // CJS
        '@vee-validate/zod',
        'zod',
        'dayjs', // CJS
        'dayjs/plugin/updateLocale', // CJS
        'dayjs/locale/en', // CJS
        'dayjs/locale/zh-cn', // CJS
        'dayjs/plugin/relativeTime', // CJS
        'dayjs/plugin/utc', // CJS
        'dayjs/plugin/timezone', // CJS
        'dayjs/plugin/quarterOfYear', // CJS
      ],
    },
  },

  i18n: {
    strategy: 'no_prefix',
    locales: [
      {
        code: 'en-US',
        language: 'en-US',
        name: 'English',
        file: { path: 'en_us.ts', cache: true },
      },
      {
        code: 'zh-CN',
        language: 'zh-CN',
        name: 'Chinese',
        file: { path: 'zh_cn.ts', cache: true },
      },
    ],
    defaultLocale: 'en-US',
    skipSettingLocaleOnNavigate: false,
    autoDeclare: true,
    langDir: 'locale',
    restructureDir: 'app',
    experimental: {
      typedOptionsAndMessages: 'default',
    },
    detectBrowserLanguage: {
      useCookie: true,
      fallbackLocale: 'en-US',
    },
  },

  shadcn: {
    prefix: 'Ui',
    componentDir: './app/components/ui',
  },
})
