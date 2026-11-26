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
    port: 3001,
  },

  css: [
    // css
    '~/assets/styles/css/vuetify.css',
    '~/assets/styles/css/tailwind.css',
    '~/assets/styles/css/main.css',
    '~/assets/styles/css/fonts.css',
    '~/assets/styles/css/transition.css',
    '~/assets/styles/css/utilities.css',
    '~/assets/styles/css/admin.css',
  ],
  modules: [
    // modules
    '@natholdallas/alova',
    '@natholdallas/i18n',
    '@natholdallas/infra',
    '@natholdallas/vuetify',
    '@natholdallas/pinia',
    '@natholdallas/tauri',
  ],

  // Deep-merged over the shared module's `moduleDependencies` defaults.
  vuetify: {
    moduleOptions: {
      // `vuetify-nuxt-module` injects this order into <head> before any app
      // stylesheet, so `assets/styles/css/vuetify.css` cannot establish it:
      // a later `@layer` statement can only append. Keeping Tailwind's layers
      // here is what makes preflight (tailwind-reset) lose to Vuetify's
      // component rules (padding, heights) and Tailwind utilities win after.
      cascadeLayers: [
        'tailwind-theme',
        'tailwind-reset',
        'vuetify-core',
        'vuetify-components',
        'vuetify-overrides',
        'vuetify-utilities',
        'tailwind-utilities',
        'vuetify-final',
      ],
    },
    vuetifyOptions: {
      theme: {
        defaultTheme: 'system',
        utilities: true,
        themes: {
          light: {
            dark: false,
            colors: {
              background: '#F5F6FA',
              surface: '#FFFFFF',
              'surface-bright': '#FFFFFF',
              'surface-light': '#F8FAFC',
              'surface-variant': '#EEF1F7',
              'on-surface': '#0F172A',
              'on-surface-variant': '#475569',
              'on-background': '#0F172A',
              primary: '#4F46E5',
              'primary-darken-1': '#4338CA',
              'on-primary': '#FFFFFF',
              secondary: '#0EA5E9',
              'secondary-darken-1': '#0284C7',
              'on-secondary': '#FFFFFF',
              accent: '#7C3AED',
              error: '#E11D48',
              'on-error': '#FFFFFF',
              info: '#0284C7',
              'on-info': '#FFFFFF',
              success: '#059669',
              'on-success': '#FFFFFF',
              warning: '#D97706',
              'on-warning': '#FFFFFF',
            },
          },
          dark: {
            dark: true,
            colors: {
              background: '#080B14',
              surface: '#0F1526',
              'surface-bright': '#16203A',
              'surface-light': '#0C1220',
              'surface-variant': '#1B2338',
              'on-surface': '#E2E8F0',
              'on-surface-variant': '#94A3B8',
              'on-background': '#E2E8F0',
              primary: '#818CF8',
              'primary-darken-1': '#6366F1',
              'on-primary': '#0B1020',
              secondary: '#38BDF8',
              'secondary-darken-1': '#0EA5E9',
              'on-secondary': '#0B1020',
              accent: '#A78BFA',
              error: '#FB7185',
              'on-error': '#0B1020',
              info: '#38BDF8',
              'on-info': '#0B1020',
              success: '#34D399',
              'on-success': '#0B1020',
              warning: '#FBBF24',
              'on-warning': '#0B1020',
            },
          },
        },
      },
    },
  },

  app: {
    rootAttrs: { id: 'root' },
    head: {
      meta: [
        {
          name: 'viewport',
          content: 'width=device-width, initial-scale=1',
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


  // Vuetify emits a single entry chunk measuring 557 kB, past Vite's default
  // 500 kB warning threshold. For a statically hosted SPA admin shell that
  // limit only produces noise, so raise it rather than chase 57 kB with a
  // vendor-splitting config.
  vite: {
    build: {
      chunkSizeWarningLimit: 800,
      // Rolldown's bundlerTimings check (Vite 8 default) prints a PLUGIN_TIMINGS
      // block on every build: Vuetify's styles plugin resolves all ~62 CSS
      // modules and the reported per-callback times overlap, so the numbers are
      // not actionable. Keep the build, drop the log.
      rolldownOptions: {
        checks: { bundlerTimings: false },
      },
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

  postcss: {
    plugins: {
      // cssnano's postcss-calc cannot parse the relative-colour / calc syntax
      // Tailwind emits and logs `Lexical error` warnings. Keep every other
      // cssnano transform; skip only the calc rewrite.
      cssnano: { preset: ['default', { calc: false }] },
    },
  },
})
