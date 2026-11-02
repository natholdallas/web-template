import en from '@natholdallas/i18n/locale/en_us'
import { en as $vuetify } from 'vuetify/locale'

export default {
  $vuetify,
  ...en,

  'app.name': 'webtplmst',

  'admin': 'Admin',
  'admin.username': 'Username',
  'admin.password': 'Password',
  'admin.desc': 'Manage administrator accounts',

  'user': 'User',
  'user.username': 'Username',
  'user.password': 'Password',
  'user.desc': 'Manage user accounts',

  'reset.password': 'Reset Password',
  'reset.password.desc': 'The new password is generated randomly, please copy and keep it safe:',

  'locale.en.us': 'English',
  'locale.zh.cn': 'Simplified Chinese',

  'urls.index': 'Home',
  'urls.admin': 'Admin',
  'urls.user': 'Users',
  'urls.media': 'Media',
  'urls.settings': 'Settings',
  'urls.entrance': 'Entrance',

  'settings.profile': 'Profile',
  'settings.profile.desc': 'Update your account information',
  'settings.username': 'Username',
  'settings.no.username': 'No username',
  'settings.password': 'Reset Password',
  'settings.password.old': 'Current Password',
  'settings.password.new': 'New Password',
  'settings.password.desc': 'Change your login password',
  'settings.theme': 'Theme',
  'settings.theme.desc': 'Choose how the app should look',
  'settings.theme.system': 'System',
  'settings.theme.dark': 'Dark',
  'settings.theme.light': 'Light',
  'settings.locale': 'Language',
  'settings.locale.desc': 'Choose your preferred language',
  'settings.desc': 'Manage your account, security and preferences',

  // media
  'media.desc': 'Manage static assets uploaded to the platform',
  'media.preview': 'Preview',
  'media.name': 'Filename',
  'media.kind': 'Kind',
  'media.size': 'Size',
  'media.kind.image': 'Image',
  'media.kind.video': 'Video',
  'media.kind.file': 'File',
  'upload': 'Upload',
  'open': 'Open',
  'err.invalid.media': 'Invalid media file',
  'err.upload.failed': 'Upload failed',

  'dashboard.title': 'Dashboard',
  'dashboard.total.admins': 'Total Admins',
  'dashboard.total.users': 'Total Users',
}