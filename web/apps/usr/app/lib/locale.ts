export enum Locale {
  EN_US = 'en-US', // 英语
  ZH_CN = 'zh-CN', // 简中
}

export const locales = {
  [Locale.EN_US]: { k: 'locale.en.us', v: Locale.EN_US },
  [Locale.ZH_CN]: { k: 'locale.zh.cn', v: Locale.ZH_CN },
}
