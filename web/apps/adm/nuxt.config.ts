const seoEnabled = process.env.ENABLE_SEO === 'true' ? true : false
const pwaEnabled = process.env.ENABLE_PWA === 'true' ? true : false
const sitemapRoutes = seoEnabled ? ['/sitemap.xml'] : []

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  ssr: false,
  compatibilityDate: '2026-05-27',

  // dev
  devtools: {
    enabled: false,
  },
  devServer: {
    port: 3001,
  },

  css: [
    // css
    '~/assets/styles/css/tailwind.css',
    '~/assets/styles/css/main.css',
    '~/assets/styles/css/fonts.css',
    '~/assets/styles/css/transition.css',
    '~/assets/styles/css/utilities.css',
  ],
  modules: [
    // modules
    '@natholdallas/alova',
    '@natholdallas/i18n',
    '@natholdallas/infra',
    '@natholdallas/vuetify',
    '@natholdallas/tauri',
  ],
  imports: {
    presets: [
      {
        from: 'lodash',
        imports: ['isArray', 'isEmpty', 'cloneDeep', 'toNumber'],
      },
    ],
  },

  app: {
    rootAttrs: { id: 'root' },
    head: {
      link: [
        // link
        { rel: 'stylesheet', href: '/layers.css' },
      ],
      meta: [
        // meta
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
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
    // enabled: seoEnabled,
    // url: 'https://example.com',
    indexable: seoEnabled,
  },

  seo: {
    enabled: seoEnabled,
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
})
