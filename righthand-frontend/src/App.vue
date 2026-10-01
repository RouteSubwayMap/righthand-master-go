<template>
  <router-view />
</template>

<script setup>
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: 'Helvetica Neue', Helvetica, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', Arial, sans-serif;
}

#app {
  width: 100%;
  height: 100vh;
}

/* 统一表单元素样式 */
.el-form-item__content .el-select,
.el-form-item__content .el-input,
.el-form-item__content .el-date-picker,
.el-form-item__content .el-cascader {
  width: 100%;
}

/* 行内表单的下拉框默认宽度 */
.el-form--inline .el-select {
  width: 180px;
}

/* 日期选择器在行内表单中的宽度 */
.el-form--inline .el-date-picker {
  width: 180px;
}

/* 输入框在行内表单中的宽度 */
.el-form--inline .el-input__wrapper {
  min-width: 180px;
}
</style>
