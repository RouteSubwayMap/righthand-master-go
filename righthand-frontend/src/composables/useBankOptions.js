import { ref, computed } from 'vue'
import { getBankParamList } from '@/api/bankParam'

// 模块级缓存：银行机构参数全量数据，避免每个页面重复请求
const bankParams = ref([])
let loaded = false
let inflight = null

/**
 * 银行类型 / 银行名称 下拉选项
 * 数据来源：银行机构参数管理（bank_param 表），不再使用数据字典关联
 */
export function useBankOptions() {
  // 加载银行机构参数（全量），带模块级缓存
  async function loadBankParams(force = false) {
    if (loaded && !force) return
    if (inflight) return inflight
    inflight = (async () => {
      try {
        const res = await getBankParamList({ pageNum: 1, pageSize: 100000 })
        bankParams.value = res.data?.records || []
        loaded = true
      } catch (e) {
        console.error('加载银行机构参数失败', e)
        bankParams.value = []
      } finally {
        inflight = null
      }
    })()
    return inflight
  }

  // 银行类型选项（去重）
  const bankTypeOptions = computed(() => {
    const set = new Set()
    bankParams.value.forEach(p => {
      if (p.bankType) set.add(p.bankType)
    })
    return Array.from(set).map(v => ({ label: v, value: v }))
  })

  // 银行名称选项（按银行类型联动，未选类型则返回全部，去重）
  function getBankNameOptions(bankType) {
    const list = bankType
      ? bankParams.value.filter(p => p.bankType === bankType)
      : bankParams.value
    const set = new Set()
    list.forEach(p => {
      if (p.bankName) set.add(p.bankName)
    })
    return Array.from(set).map(v => ({ label: v, value: v }))
  }

  // 进入页面即自动加载一次（HMR 重新执行 setup 时也会触发，受 inflight/loaded 保护不会重复请求）
  loadBankParams()

  return {
    bankParams,
    bankTypeOptions,
    getBankNameOptions,
    loadBankParams
  }
}
