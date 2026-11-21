/// <reference types="vite/client" />

// Русская локаль ECharts лежит в пакете без типов.
declare module 'echarts/i18n/langRU-obj.js' {
  const locale: Parameters<typeof import('echarts/core').registerLocale>[1]
  export default locale
}
