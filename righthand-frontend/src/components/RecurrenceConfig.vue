<template>
  <div class="recurrence-config">
    <el-form-item label="重复类型">
      <el-select v-model="modelValue.recurrenceType" placeholder="不重复" style="width: 100%" @change="handleTypeChange">
        <el-option label="不重复" value="NONE" />
        <el-option label="每天" value="DAILY" />
        <el-option label="每周" value="WEEKLY" />
        <el-option label="每月" value="MONTHLY" />
      </el-select>
    </el-form-item>

    <!-- 每周重复：选择星期几 -->
    <el-form-item v-if="modelValue.recurrenceType === 'WEEKLY'" label="重复星期">
      <el-checkbox-group v-model="modelValue.weekDays">
        <el-checkbox :label="1">周一</el-checkbox>
        <el-checkbox :label="2">周二</el-checkbox>
        <el-checkbox :label="3">周三</el-checkbox>
        <el-checkbox :label="4">周四</el-checkbox>
        <el-checkbox :label="5">周五</el-checkbox>
        <el-checkbox :label="6">周六</el-checkbox>
        <el-checkbox :label="0">周日</el-checkbox>
      </el-checkbox-group>
    </el-form-item>

    <!-- 每月重复：选择月末或指定几号 -->
    <el-form-item v-if="modelValue.recurrenceType === 'MONTHLY'">
      <el-radio-group v-model="monthDayType">
        <el-radio label="END">月末</el-radio>
        <el-radio label="SPECIFIC">指定几号</el-radio>
      </el-radio-group>
      <el-input-number
        v-if="monthDayType === 'SPECIFIC'"
        v-model="modelValue.monthDay"
        :min="1"
        :max="31"
        placeholder="几号"
        style="width: 100%; margin-top: 10px"
      />
    </el-form-item>

    <!-- 结束条件 -->
    <el-form-item v-if="modelValue.recurrenceType !== 'NONE'" label="结束条件">
      <el-radio-group v-model="modelValue.endType">
        <el-radio label="NEVER">永不结束</el-radio>
        <el-radio label="DATE">到指定日期</el-radio>
      </el-radio-group>
    </el-form-item>

    <!-- 结束日期 -->
    <el-form-item v-if="modelValue.endType === 'DATE'" label="结束日期">
      <el-date-picker
        v-model="modelValue.endDate"
        type="date"
        placeholder="选择结束日期"
        format="YYYY-MM-DD"
        value-format="YYYY-MM-DD"
        style="width: 100%"
      />
    </el-form-item>

    <!-- 提示信息 -->
    <el-alert
      v-if="modelValue.recurrenceType !== 'NONE'"
      :title="recurrenceSummary"
      type="info"
      :closable="false"
      show-icon
    />
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({
      recurrenceType: 'NONE',
      weekDays: [],
      monthDay: 1,
      endType: 'NEVER',
      endDate: null
    })
  }
})

const emit = defineEmits(['update:modelValue'])

// 每月重复日期类型：月末(0) 或 指定几号(>0)
const monthDayType = computed({
  get() {
    return props.modelValue.monthDay === 0 ? 'END' : 'SPECIFIC'
  },
  set(val) {
    if (val === 'END') {
      props.modelValue.monthDay = 0
    } else if (props.modelValue.monthDay === 0) {
      props.modelValue.monthDay = 1
    }
  }
})

// 类型改变时重置相关字段
const handleTypeChange = (type) => {
  if (type === 'NONE') {
    props.modelValue.endType = 'NEVER'
    props.modelValue.weekDays = []
    props.modelValue.monthDay = 1
  } else if (type === 'WEEKLY' && props.modelValue.weekDays.length === 0) {
    // 默认选中今天
    const today = new Date().getDay()
    props.modelValue.weekDays = [today === 0 ? 0 : today]
  }
}

// 计算重复描述
const recurrenceSummary = computed(() => {
  const { recurrenceType, weekDays, monthDay, endType, endDate } = props.modelValue
  
  if (recurrenceType === 'NONE') return ''
  
  let desc = ''
  switch (recurrenceType) {
    case 'DAILY':
      desc = '每天重复'
      break
    case 'WEEKLY':
      const dayNames = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
      const selectedDays = weekDays.map(d => dayNames[d]).join('、')
      desc = `每周${selectedDays}重复`
      break
    case 'MONTHLY':
      desc = monthDay === 0 ? '每月月末重复' : `每月${monthDay}号重复`
      break
  }
  
  if (endType === 'DATE' && endDate) {
    desc += `，到 ${endDate} 结束`
  } else {
    desc += '，永不结束'
  }
  
  return desc
})
</script>

<style scoped lang="scss">
.recurrence-config {
  :deep(.el-alert) {
    margin-top: 10px;
  }
}
</style>
