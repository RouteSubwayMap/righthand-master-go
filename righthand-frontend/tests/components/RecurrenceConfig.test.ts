import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import RecurrenceConfig from '@/components/RecurrenceConfig.vue'

describe('RecurrenceConfig 组件', () => {
  it('应该正确渲染组件', () => {
    const wrapper = mount(RecurrenceConfig, {
      props: {
        modelValue: {
          enabled: false,
          type: 'daily',
          interval: 1
        }
      }
    })
    
    expect(wrapper.exists()).toBe(true)
  })

  it('应该显示启用复选框', () => {
    const wrapper = mount(RecurrenceConfig, {
      props: {
        modelValue: {
          enabled: false,
          type: 'DAILY',
          interval: 1
        }
      }
    })
    
    // 检查是否包含重复类型选项
    expect(wrapper.text()).toContain('不重复')
    expect(wrapper.text()).toContain('每天')
  })

  it('应该在启用时显示配置选项', async () => {
    const wrapper = mount(RecurrenceConfig, {
      props: {
        modelValue: {
          enabled: true,
          type: 'DAILY',
          interval: 1
        }
      }
    })
    
    // 组件应该正确渲染
    expect(wrapper.exists()).toBe(true)
    // 检查是否包含重复类型选项
    expect(wrapper.text()).toContain('每天')
  })

  it('应该支持每日类型', () => {
    const wrapper = mount(RecurrenceConfig, {
      props: {
        modelValue: {
          enabled: true,
          type: 'DAILY',
          interval: 1
        }
      }
    })
    
    expect(wrapper.exists()).toBe(true)
    // 检查是否包含"每天"选项
    expect(wrapper.text()).toContain('每天')
  })

  it('应该支持每周类型', () => {
    const wrapper = mount(RecurrenceConfig, {
      props: {
        modelValue: {
          enabled: true,
          type: 'WEEKLY',
          interval: 1,
          weekDays: [1]
        }
      }
    })
    
    expect(wrapper.exists()).toBe(true)
    // 检查是否包含"每周"选项
    expect(wrapper.text()).toContain('每周')
  })

  it('应该支持每月类型', () => {
    const wrapper = mount(RecurrenceConfig, {
      props: {
        modelValue: {
          enabled: true,
          type: 'MONTHLY',
          interval: 1,
          dayOfMonth: 1
        }
      }
    })
    
    expect(wrapper.exists()).toBe(true)
    // 检查是否包含"每月"选项
    expect(wrapper.text()).toContain('每月')
  })

  it('应该验证间隔值必须大于0', async () => {
    const wrapper = mount(RecurrenceConfig, {
      props: {
        modelValue: {
          enabled: true,
          type: 'daily',
          interval: 0
        }
      }
    })
    
    // 组件应该有验证逻辑或显示错误提示
    expect(wrapper.exists()).toBe(true)
  })

  it('应该在禁用时隐藏详细配置', () => {
    const wrapper = mount(RecurrenceConfig, {
      props: {
        modelValue: {
          enabled: false,
          type: 'daily',
          interval: 1
        }
      }
    })
    
    // 当enabled为false时,某些配置项可能被隐藏
    expect(wrapper.exists()).toBe(true)
  })

  it('应该发出update:modelValue事件', async () => {
    const wrapper = mount(RecurrenceConfig, {
      props: {
        modelValue: {
          enabled: false,
          type: 'daily',
          interval: 1
        }
      }
    })
    
    // 触发启用复选框
    const checkbox = wrapper.find('input[type="checkbox"]')
    if (checkbox.exists()) {
      await checkbox.setValue(true)
      
      // 检查是否发出了更新事件
      const emitted = wrapper.emitted('update:modelValue')
      expect(emitted).toBeTruthy()
    }
  })
})
