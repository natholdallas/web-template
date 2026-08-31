import zh from '@natholdallas/i18n/locale/zh_cn'
import { zhHans as $vuetify } from 'vuetify/locale'

export default {
  $vuetify,
  ...zh,

  'app.name': 'webtplmst',

  'admin': '管理员',
  'admin.username': '用户名',
  'admin.password': '密码',

  'user': '用户',
  'user.username': '用户名',
  'user.password': '密码',

  'reset.password': '重置密码',
  'reset.password.desc': '新密码为随机生成，请复制并妥善保存：',

  'locale.en.us': '英语',
  'locale.zh.cn': '简体中文',

  'urls.index': '首页',
  'urls.admin': '管理员账号管理',
  'urls.user': '用户管理',
  'urls.settings': '设置',
  'urls.entrance': '登录',

  'settings.profile': '个人信息',
  'settings.profile.desc': '修改您的账号信息',
  'settings.username': '用户名',
  'settings.password': '重置密码',
  'settings.password.old': '当前密码',
  'settings.password.new': '新密码',
  'settings.password.desc': '修改您的登录密码',
  'settings.theme': '主题',
  'settings.theme.dark': '深色',
  'settings.theme.light': '浅色',
  'settings.locale': '语言',

  'dashboard.title': '仪表盘',
  'dashboard.total.admins': '管理员总数',
  'dashboard.total.users': '用户总数',
}
