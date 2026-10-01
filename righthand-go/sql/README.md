# 数据库初始化脚本

## 📁 文件说明

本项目采用**双文件策略**管理数据库初始化：

1. **init-database.sql** - 数据库完整初始化（表结构 + 基本数据）
2. **init-test-data.sql** - 测试数据初始化（仅用于开发/测试环境）

---

## 🚀 快速开始

### 生产环境部署

```bash
# 仅执行初始化脚本
mysql -u root -p < sql/init-database.sql
```

### 开发/测试环境部署

```bash
# 先执行初始化，再加载测试数据
mysql -u root -p < sql/init-database.sql
mysql -u root -p < sql/init-test-data.sql
```

---

## 📄 init-database.sql

**数据库完整初始化脚本**

- **数据库名称**: `righthand`
- **字符集**: `utf8mb4`
- **表总数**: 11 个
- **包含内容**: 所有表结构、默认用户、角色、菜单、字典、Casbin权限策略

#### 模块划分

| 模块 | 说明 |
|------|------|
| SYS | 系统模块（用户、角色、菜单、字典、消息） |
| MEMO | 备忘录业务模块 |
| TASK | 任务管理模块 |
| CASBIN | Casbin权限策略表 |

#### 主要功能

✅ **系统模块**
- 用户管理（sys_user）
- 角色管理（sys_role）
- 用户角色关联（sys_user_role）
- 菜单管理（sys_menu）
- 角色菜单关联（sys_role_menu）
- 数据字典（sys_dict）
- 站内消息（sys_message）

✅ **备忘录模块**
- 联系人管理（contact）

✅ **任务管理模块**
- 统一事项表（task_item）- **任务和备忘录合并**
- 需求管理（requirement）
- 任务需求关联（task_requirement_rel）

✅ **权限控制**
- Casbin RBAC权限策略（casbin_rule）
- 预配置 admin/user 角色权限

#### 默认账号

- **管理员**: admin / admin123
- **测试用户**: test / test123

#### 默认菜单

- 仪表盘 (/dashboard)
- 事项管理 (/tasks) - **支持周期性任务设置**
- 需求管理 (/requirements)
- 联系人 (/contacts)
- 系统管理 (/auth)
  - 用户管理
  - 角色管理
  - 菜单管理
  - 字典管理

---

## 🧪 init-test-data.sql

**测试数据初始化脚本**

⚠️ **注意**: 此文件仅用于开发和测试环境，**生产环境请勿执行**！

#### 包含的测试数据

✅ **需求管理** (5条)
- 个人贷款流程优化
- 风险控制模型升级
- 移动端UI优化
- 企业网银批量转账功能
- 智能客服系统集成

✅ **统一事项** (9条)
- 普通任务 (4条)
- 周期性任务示例 (3条):
  - 团队周会（每周重复）
  - 月度工作总结（每月重复）
  - 年度体检（每年重复）
- 备忘录 (2条)

✅ **联系人** (4条)
- 研发人员、产品经理、客户经理等

✅ **站内消息** (3条)
- 任务分配通知
- 系统维护通知
- 任务截止提醒

---

## 🔧 使用方法

### 方式一：命令行执行

```bash
# Windows PowerShell
Get-Content sql\init-database.sql | mysql -u root -p

# Linux/Mac
mysql -u root -p < sql/init-database.sql
```

### 方式二：MySQL客户端

```sql
-- 连接MySQL后执行
SOURCE D:/git_local/righthand/righthand-go/sql/init-database.sql;
SOURCE D:/git_local/righthand/righthand-go/sql/init-test-data.sql;
```

### 方式三：图形化工具

使用 MySQL Workbench、Navicat、DataGrip 等工具直接执行 SQL 文件。

---

## ⚙️ 特性说明

### 周期性任务支持

系统支持以下重复类型：
- **NONE**: 不重复
- **DAILY**: 每天重复
- **WEEKLY**: 每周重复（可指定星期几）
- **MONTHLY**: 每月重复（可指定几号）
- **YEARLY**: 每年重复（可指定月份和日期）

结束条件：
- **NEVER**: 永不结束
- **DATE**: 到指定日期结束
- **COUNT**: 重复指定次数后结束

### 统一事项模型

系统将任务和备忘录合并为统一的 `task_item` 表，通过 `item_type` 字段区分：
- **TASK**: 任务
- **MEMO**: 备忘录
- **REQUIREMENT**: 需求

简化状态流转：
```
TODO → DOING → DONE/CANCELLED
```

### Casbin RBAC权限

基于角色的访问控制：
- **role:admin**: 拥有所有权限
- **role:user**: 基础业务权限（排除系统管理）

权限粒度：API路径 + HTTP方法

---

## ⚠️ 注意事项

1. ✅ 执行前请备份现有数据
2. ✅ 确保 MySQL 服务已启动（建议版本 8.0+）
3. ✅ 需要有创建数据库的权限
4. ✅ 使用 UTF8MB4 字符集以支持中文和emoji
5. ✅ 生产环境不要执行 init-test-data.sql
6. ✅ 建议在事务中执行，失败时回滚

---

## 🔗 相关配置

数据库连接配置请参考: `../configs/config.yaml`

典型配置：
```yaml
database:
  host: localhost
  port: 3306
  username: root
  password: your_password
  dbname: righthand
  charset: utf8mb4
```

---

## 📝 更新日志

### v2.1 (当前版本)
- ✨ 表名简化：移除 auth_、memo_、task_ 前缀
- ✨ 清理无用表：公告管理、教育经历、工作经历、项目经历、生活事件
- ✨ 新增需求管理菜单和权限
- ✨ 优化普通用户权限配置（排除系统管理）

### v2.0
- ✨ 合并所有SQL脚本为两个文件
- ✨ 添加周期性任务支持
- ✨ 任务和备忘录合并为统一事项表
- ✨ 集成Casbin RBAC权限控制
- ✨ 添加完整的测试数据

### v1.0 (历史版本)
- 多文件分散管理
- 任务和备忘录独立表
- Java版遗留表清理
