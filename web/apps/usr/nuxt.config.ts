const seoEnabled = process.env.ENABLE_SEO === 'true' ? true : false
const sitemapRoutes = seoEnabled ? ['/sitemap.xml'] : []

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  ssr: false,
  compatibilityDate: '2026-05-27',

  // Declare the runtime config shape so `useRuntimeConfig().public.apiBase` is
  // typed as a string. The value itself comes from NUXT_PUBLIC_API_BASE at build
  // time (main.sh derives it from domains[]/app_env); an empty default means
  // "same origin", which is what a reverse-proxied deployment wants.
  runtimeConfig: {
    public: {
      apiBase: '',
    },
  },

  // dev
  devtools: {
    enabled: false,
  },
  devServer: {
    port: 3000,
  },

  css: [
    // css
    '~/assets/main.css',
    '~/assets/tailwind.css',
    '~/assets/transition.css',
    '~/assets/utilities.css',
  ],
  modules: [
    // modules
    '@natholdallas/alova',
    '@natholdallas/i18n',
    '@natholdallas/infra',
    '@natholdallas/pinia',
    '@natholdallas/shadcn',
    '@natholdallas/tauri',
  ],
  app: {
    rootAttrs: { id: 'root' },
    head: {
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
    // Absolute origin for canonical URLs, sitemap entries and OG images. main.sh
    // derives NUXT_PUBLIC_SITE_URL from domain_<app> at build time; without it
    // the SEO output has no absolute URLs.
    url: process.env.NUXT_PUBLIC_SITE_URL ?? '',
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

  postcss: {
    plugins: {
      // cssnano's postcss-calc cannot parse the relative-colour / calc syntax
      // Tailwind emits and logs `Lexical error` warnings. Keep every other
      // cssnano transform; skip only the calc rewrite.
      cssnano: { preset: ['default', { calc: false }] },
    },
  },
})
