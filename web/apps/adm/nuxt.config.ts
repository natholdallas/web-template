const seoEnabled = process.env.ENABLE_SEO === 'true' ? true : false
const pwaEnabled = process.env.ENABLE_PWA === 'true' ? true : false
const sitemapRoutes = seoEnabled ? ['/sitemap.xml'] : []

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  ssr: false,
  devtools: { enabled: false },
  devServer: { port: 3001 },
  compatibilityDate: '2026-05-27',

  css: ['~/assets/main.css', '~/assets/fonts.css', '~/assets/transition.css', '~/assets/utilities.css'],
  modules: [
    '@natholdallas/alova',
    '@natholdallas/i18n',
    '@natholdallas/infra',
    '@natholdallas/vuetify',
    '@natholdallas/pinia',
    '@natholdallas/unocss',
    '@natholdallas/pwa',
  ],
  imports: {
    presets: [{ from: 'lodash', imports: ['isArray', 'isEmpty', 'cloneDeep', 'toNumber'] }],
  },

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

  site: {
    // url: 'https://example.com',
    indexable: seoEnabled,
  },

  seo: {
    meta: {
      description: '',
    },
  },

  sitemap: {
    enabled: seoEnabled,
  },

  schemaOrg: {
    enabled: seoEnabled,
  },

  pwa: {
    disable: !pwaEnabled,
    manifest: {
      name: 'App',
      short_name: 'App',
      theme_color: '#0a0a0a',
      description: 'App',
    },
  },

  nitro: {
    compressPublicAssets: true,
    output: {
      dir: 'dist',
      publicDir: 'dist/public',
      serverDir: 'dist/server',
    },
    prerender: {
      routes: [...sitemapRoutes],
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
      drop: ['debugger'],
    },
    optimizeDeps: {
      include: [
        'alova/client',
        'lodash',
        'copy-to-clipboard',
        'alova',
        'alova/fetch',
        'alova/vue',
        'lodash/isArray', // CJS
        'lodash/cloneDeep', // CJS
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

  typescript: {
    tsConfig: {
      compilerOptions: {
        // experimentalDecorators: true,
        // emitDecoratorMetadata: true
      },
    },
  },
})
