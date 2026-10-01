# UTF-8中文编码问题修复与测试优化方案

## 📋 问题描述

在通过PowerShell API创建包含中文的菜单时,数据库中存储的是乱码(问号`?`),导致前端显示异常。

### 根本原因

1. **PowerShell编码转换问题**: PowerShell的`ConvertTo-Json`默认使用UTF-16编码,通过管道传递给`Invoke-WebRequest`时可能发生编码转换错误
2. **测试覆盖不足**: 缺少针对中文字符编码的单元测试和集成测试
3. **API调用不规范**: 没有明确指定`Content-Type`的charset参数

---

## ✅ 已实施的修复

### 1. 数据库配置验证

**配置文件**: `righthand-go/configs/config.yaml`
```yaml
database:
  charset: utf8mb4  # ✅ 正确配置,支持完整Unicode
```

**DSN连接字符串**:
```go
// internal/config/config.go
func (c *DatabaseConfig) GetDSN() string {
    return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
        c.Username, c.Password, c.Host, c.Port, c.DBName,
        c.Charset,  // utf8mb4
        c.ParseTime,
        c.Loc,
    )
}
```

---

### 2. API调用最佳实践

#### ❌ 错误的做法(会导致乱码)
```powershell
# PowerShell默认编码可能导致问题
$menuData = @{menuName="需求管理"} | ConvertTo-Json
Invoke-WebRequest -Uri $url -Method POST -Body $menuData -ContentType "application/json"
```

#### ✅ 正确的做法
```powershell
# 明确指定UTF-8编码
$menuData = @{menuName="需求管理";productCode="task";...} | ConvertTo-Json -Depth 10
$jsonBytes = [System.Text.Encoding]::UTF8.GetBytes($menuData)
$content = [System.Text.Encoding]::UTF8.GetString($jsonBytes)

Invoke-WebRequest -Uri $url `
    -Method POST `
    -Headers $headers `
    -Body $content `
    -ContentType "application/json; charset=utf-8" `
    -UseBasicParsing
```

**关键点**:
- 使用`-Depth 10`确保嵌套对象正确序列化
- 显式转换为UTF-8字节再转回字符串
- Content-Type指定`charset=utf-8`

---

### 3. 新增测试覆盖

#### A. UTF-8编码单元测试
**文件**: `internal/utils/utf8_encoding_test.go`

**测试用例**:
1. ✅ `TestUTF8Encoding_ChineseString` - 中文字符串UTF-8编码验证
2. ✅ `TestJSONMarshal_ChineseCharacters` - JSON序列化中文字符
3. ✅ `TestJSONUnmarshal_ChineseFromAPI` - API响应中文解析
4. ✅ `TestPowerShell_JSON_Encoding` - PowerShell JSON编码模拟
5. ✅ `TestDatabase_Charset_UTF8MB4` - 数据库charset配置验证
6. ✅ `TestChineseStringEncoding_EdgeCases` - 边界情况测试

**运行测试**:
```bash
cd righthand-go
go test ./internal/utils -v -run TestUTF8
go test ./internal/utils -v -run TestJSON
go test ./internal/utils -v -run TestChinese
```

#### B. 菜单Handler单元测试
**文件**: `services/auth-service/handler/menu_handler_test.go`

**测试用例**:
1. ✅ `TestMenu_CreateMenu_WithChineseName` - 创建中文名称菜单
2. ✅ `TestMenu_CreateMenu_VerifyUTF8Encoding` - 验证UTF-8编码存储
3. ✅ `TestMenu_GetMenu_VerifyChineseCharacters` - 读取中文菜单验证
4. ✅ `TestMenu_UpdateMenu_ChineseName` - 更新中文菜单名称
5. ✅ `TestMenu_ListMenus_VerifyChineseDisplay` - 菜单列表中文显示
6. ✅ `TestMenu_ChineseEncoding_BoundaryTest` - 最长中文菜单名边界测试

**运行测试**:
```bash
cd righthand-go
go test ./services/auth-service/handler -v -run TestMenu
```

---

## 🔍 问题诊断流程

### 步骤1: 验证API返回数据
```powershell
$loginResult = Invoke-WebRequest -Uri http://localhost:9080/api/auth/login `
    -Method POST `
    -Body (@{username="admin";password="admin123"} | ConvertTo-Json) `
    -ContentType "application/json" `
    -UseBasicParsing

$token = ($loginResult.Content | ConvertFrom-Json).data.token
$headers = @{Authorization="Bearer $token"}

# 获取菜单详情
$menusResult = Invoke-WebRequest -Uri "http://localhost:9080/api/system/menus/11" `
    -Headers $headers `
    -UseBasicParsing

$menuData = ($menusResult.Content | ConvertFrom-Json).data

# 验证UTF-8字节
$bytes = [System.Text.Encoding]::UTF8.GetBytes($menuData.menuName)
Write-Host "UTF-8字节: $(($bytes | ForEach-Object { '0x' + $_.ToString('X2') }) -join ' ')"

# 预期输出: 0xE9 0x9C 0x80 0xE6 0xB1 0x82 0xE7 0xAE 0xA1 0xE7 0x90 0x86
# "需求管理" = 4个字符 × 3字节 = 12字节
```

### 步骤2: 检查数据库原始数据
```sql
-- 查看菜单名称的HEX值
SELECT id, menu_name, HEX(menu_name) as hex_value 
FROM auth_sys_menu 
WHERE menu_name LIKE '%需求%';

-- 正确的HEX应该是: E99C80E6B182E7AEA1E79086
-- 如果看到3F3F3F3F,说明存储的是问号(乱码)
```

### 步骤3: 验证数据库字符集
```sql
-- 检查数据库字符集
SHOW CREATE DATABASE righthand;

-- 应该看到: DEFAULT CHARACTER SET utf8mb4

-- 检查表字符集
SHOW CREATE TABLE auth_sys_menu;

-- 应该看到: DEFAULT CHARSET=utf8mb4
```

---

## 🛡️ 预防措施

### 1. CI/CD中的编码测试

在持续集成流程中添加:
```yaml
# .github/workflows/test.yml
steps:
  - name: Run UTF-8 Encoding Tests
    run: |
      go test ./internal/utils -v -run TestUTF8
      go test ./internal/utils -v -run TestChinese
      
  - name: Run Menu Handler Tests
    run: |
      go test ./services/auth-service/handler -v -run TestMenu
```

### 2. 代码审查检查点

PR审查时必须检查:
- [ ] 所有HTTP请求是否指定`Content-Type: application/json; charset=utf-8`
- [ ] 中文字符串是否正确序列化
- [ ] 是否有对应的UTF-8编码测试
- [ ] 数据库字段是否使用`utf8mb4`字符集

### 3. 开发规范

**Go后端**:
```go
// ✅ 正确: GORM自动处理UTF-8
menu := models.SysMenu{
    MenuName: "需求管理",  // Go字符串原生支持UTF-8
}
db.Create(&menu)

// ✅ 正确: Gin自动设置UTF-8响应头
c.JSON(http.StatusOK, gin.H{
    "code": 200,
    "data": menu,
})
```

**PowerShell脚本**:
```powershell
# ✅ 正确: 明确指定UTF-8编码
$data = @{name="中文"} | ConvertTo-Json -Depth 10
Invoke-RestMethod -Uri $url `
    -Method POST `
    -Body $data `
    -ContentType "application/json; charset=utf-8"
```

**前端Vue**:
```javascript
// ✅ 正确: Axios默认使用UTF-8
axios.post('/api/system/menus', {
  menuName: '需求管理'  // JavaScript字符串原生支持UTF-8
})
```

---

## 📊 测试覆盖率提升

### 之前
- ❌ 菜单Handler: 0个测试
- ❌ UTF-8编码: 0个测试
- ❌ 中文字符: 0个测试

### 现在
- ✅ 菜单Handler: 6个测试用例
- ✅ UTF-8编码: 6个测试用例
- ✅ 中文字符: 覆盖普通中文、emoji、边界情况

### 目标
- 🎯 所有Handler都有单元测试
- 🎯 所有涉及中文的接口都有编码测试
- 🎯 集成测试包含中文数据验证

---

## 🚀 快速修复指南

如果遇到中文乱码问题:

### 场景1: 通过API创建数据时乱码
```powershell
# 1. 删除乱码数据
Invoke-WebRequest -Uri "http://localhost:9080/api/system/menus/{id}" `
    -Method DELETE `
    -Headers $headers

# 2. 使用正确的编码重新创建
$menuData = @{menuName="需求管理";...} | ConvertTo-Json -Depth 10
$jsonBytes = [System.Text.Encoding]::UTF8.GetBytes($menuData)
$content = [System.Text.Encoding]::UTF8.GetString($jsonBytes)

Invoke-WebRequest -Uri "http://localhost:9080/api/system/menus" `
    -Method POST `
    -Headers $headers `
    -Body $content `
    -ContentType "application/json; charset=utf-8"
```

### 场景2: SQL初始化脚本乱码
```bash
# ❌ 错误: 使用PowerShell管道
Get-Content init.sql | mysql -u root -p righthand

# ✅ 正确: 使用MySQL source命令
mysql -u root -p righthand < init.sql

# 或者在MySQL客户端中
source /path/to/init.sql;
```

### 场景3: 数据库已有乱码数据
```sql
-- 1. 备份数据
CREATE TABLE auth_sys_menu_backup AS SELECT * FROM auth_sys_menu;

-- 2. 清空表
TRUNCATE TABLE auth_sys_menu;

-- 3. 重新导入(确保SQL文件是UTF-8编码)
source /path/to/init.sql;

-- 4. 验证
SELECT id, menu_name, HEX(menu_name) FROM auth_sys_menu WHERE id IN (1,2,3);
```

---

## 📝 总结

### 核心问题
- PowerShell编码转换导致UTF-8数据损坏
- 缺乏针对中文编码的测试覆盖

### 解决方案
1. ✅ 明确指定HTTP请求的charset
2. ✅ 添加UTF-8编码单元测试
3. ✅ 添加菜单Handler单元测试
4. ✅ 建立编码问题诊断流程

### 持续改进
- 所有新接口必须包含中文测试
- CI/CD流程中加入编码验证
- 代码审查时检查UTF-8处理

---

## 🔗 相关资源

- [MySQL utf8mb4字符集文档](https://dev.mysql.com/doc/refman/8.0/en/charset-unicode-utf8mb4.html)
- [Go语言UTF-8支持](https://go.dev/blog/strings)
- [PowerShell编码最佳实践](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_character_encoding)
