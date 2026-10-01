import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { defineComponent } from 'vue'

// 创建一个简单的测试组件
const TestComponent = defineComponent({
  template: '<div><h1>{{ title }}</h1><p v-if="show">{{ message }}</p></div>',
  props: ['title', 'message', 'show']
})

describe('Vue组件测试示例', () => {
  it('应该正确渲染标题', () => {
    const wrapper = mount(TestComponent, {
      props: {
        title: '测试标题'
      }
    })
    
    expect(wrapper.find('h1').text()).toBe('测试标题')
  })

  it('应该根据条件显示消息', () => {
    const wrapper = mount(TestComponent, {
      props: {
        title: '测试',
        message: '这是一条消息',
        show: true
      }
    })
    
    expect(wrapper.find('p').exists()).toBe(true)
    expect(wrapper.find('p').text()).toBe('这是一条消息')
  })

  it('应该在show为false时隐藏消息', () => {
    const wrapper = mount(TestComponent, {
      props: {
        title: '测试',
        message: '这是一条消息',
        show: false
      }
    })
    
    expect(wrapper.find('p').exists()).toBe(false)
  })
})
