// Vitest 测试配置文件
import { config } from '@vue/test-utils'
import ElementPlus from 'element-plus'
import { createPinia } from 'pinia'

// 全局注册 Element Plus（挂载真实组件的测试需要，否则 el-* 组件无法解析）
// 同时提供独立 Pinia 实例（组件内 useStore 需要 active pinia）
config.global.plugins = [ElementPlus, createPinia()]

// happy-dom 未实现 window.confirm/alert，补默认实现以支持 vi.spyOn 与删除确认流转
if (typeof window.confirm !== 'function') {
  window.confirm = () => true
}
if (typeof window.alert !== 'function') {
  window.alert = () => undefined
}

// v-permission 指令全局桩：测试环境不拦截元素渲染（权限显隐由组件逻辑断言）
config.global.directives = {
  permission: {}
}

// 全局配置
config.global.mocks = {
  $t: (key: string) => key,
  $te: (key: string) => false
}
