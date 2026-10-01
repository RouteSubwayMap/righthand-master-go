import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getAllDicts } from '@/api/dict'

export const useDictStore = defineStore('dict', () => {
  const dictData = ref({})

  function setDict(type, data) {
    dictData.value[type] = data
  }

  function getDictByType(type) {
    return dictData.value[type] || []
  }

  function getDictLabel(type, value) {
    const list = dictData.value[type] || []
    const item = list.find(d => d.dictValue === String(value))
    return item ? item.dictLabel : value
  }

  // 加载所有字典数据
  async function loadAllDicts() {
    try {
      const res = await getAllDicts()
      // 后端返回的是分页格式: { total, pageNum, pageSize, list }
      // 需要将 list 数组按 dict_type 分组
      const list = res.data?.list || res.data || []
      
      // 按 dict_type 分组
      const grouped = {}
      list.forEach(item => {
        const type = item.dictType
        if (!grouped[type]) {
          grouped[type] = []
        }
        grouped[type].push(item)
      })
      
      // 将分组后的数据存储
      Object.keys(grouped).forEach(type => {
        dictData.value[type] = grouped[type]
      })
      
      console.log('字典数据已加载:', dictData.value)
    } catch (error) {
      console.error('加载字典数据失败:', error)
    }
  }

  return {
    dictData,
    setDict,
    getDictByType,
    getDictLabel,
    loadAllDicts
  }
})
