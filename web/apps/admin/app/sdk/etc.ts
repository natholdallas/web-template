export enum Language {
  EN_US = 'en-US', // 英语
  ZH_CN = 'zh-CN', // 简中
}

export const languages = {
  [Language.EN_US]: { k: 'locale.en.us', v: Language.EN_US },
  [Language.ZH_CN]: { k: 'locale.zh.cn', v: Language.ZH_CN },
}
export const languageItems = Object.entries(languages).map(([k, v]) => ({ k, v }))

export type IDModel = typeof IDModel
export const IDModel = {
  id: 0,
}

export type Model = typeof Model
export const Model = {
  id: 0,
  createdAt: '',
  updatedAt: '',
}

export type PageQueries = Partial<typeof PageQueries>
export const PageQueries = {
  page: 1,
  size: 20,
}

export type SortQueries = Partial<typeof SortQueries>
export const SortQueries = {
  column: <string | undefined>undefined,
  desc: <boolean | undefined>undefined,
}

export type BaseQueries = Partial<typeof BaseQueries>
export const BaseQueries = {
  ...PageQueries,
  ...SortQueries,
}

export type Page<T> = { content: T[] } & Omit<typeof Page, 'content'>
export const Page = {
  total: 0,
  page: 0,
  content: [],
}
