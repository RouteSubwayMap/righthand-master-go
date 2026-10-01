-- ============================================
-- RightHand 系统 - 数据库完整初始化脚本
-- 数据库：righthand
-- 包含：所有表结构、基本数据、权限策略、菜单配置
-- 版本：Go后端统一版（表名简化版）
-- ============================================

SET FOREIGN_KEY_CHECKS = 0;

-- ============================================
-- 1. 创建数据库
-- ============================================
CREATE DATABASE IF NOT EXISTS righthand
DEFAULT CHARACTER SET utf8mb4
DEFAULT COLLATE utf8mb4_unicode_ci;

USE righthand;

-- ============================================
-- 2. 认证授权模块表结构（移除 auth_ 前缀）
-- ============================================

-- 2.1 系统用户表
DROP TABLE IF EXISTS sys_user;
CREATE TABLE sys_user (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键 ID',
    username VARCHAR(50) UNIQUE NOT NULL COMMENT '用户名',
    password VARCHAR(100) NOT NULL COMMENT '密码（加密）',
    real_name VARCHAR(50) COMMENT '真实姓名',
    email VARCHAR(100) COMMENT '邮箱',
    phone VARCHAR(20) COMMENT '手机号',
    avatar VARCHAR(255) COMMENT '头像 URL',
    status TINYINT DEFAULT 1 COMMENT '状态（0 禁用 1 启用）',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    create_user BIGINT DEFAULT NULL COMMENT '创建用户ID',
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    update_user BIGINT DEFAULT NULL COMMENT '更新用户ID',
    deleted TINYINT DEFAULT 0 COMMENT '逻辑删除标志'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统用户表';

-- 2.2 系统角色表
DROP TABLE IF EXISTS sys_role;
CREATE TABLE sys_role (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键 ID',
    role_name VARCHAR(50) NOT NULL COMMENT '角色名',
    role_code VARCHAR(50) NOT NULL COMMENT '角色编码',
    description VARCHAR(255) COMMENT '描述',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    create_user BIGINT DEFAULT NULL COMMENT '创建用户ID',
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    update_user BIGINT DEFAULT NULL COMMENT '更新用户ID'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统角色表';

-- 2.3 用户角色关联表
DROP TABLE IF EXISTS sys_user_role;
CREATE TABLE sys_user_role (
    user_id BIGINT NOT NULL COMMENT '用户 ID',
    role_id BIGINT NOT NULL COMMENT '角色 ID',
    PRIMARY KEY (user_id, role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户角色关联表';

-- 2.4 系统菜单表
DROP TABLE IF EXISTS sys_menu;
CREATE TABLE sys_menu (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键 ID',
    menu_name VARCHAR(50) NOT NULL COMMENT '菜单名',
    product_code VARCHAR(50) DEFAULT 'auth' COMMENT '产品编码（auth/memo/task 等）',
    menu_type TINYINT NOT NULL COMMENT '菜单类型（0 目录 1 菜单 2 按钮）',
    parent_id BIGINT DEFAULT 0 COMMENT '父菜单 ID',
    path VARCHAR(100) COMMENT '路由路径',
    component VARCHAR(100) COMMENT '组件路径',
    permission VARCHAR(50) COMMENT '权限标识',
    icon VARCHAR(50) COMMENT '图标',
    sort_order INT DEFAULT 0 COMMENT '排序',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    create_user BIGINT DEFAULT NULL COMMENT '创建用户ID',
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    update_user BIGINT DEFAULT NULL COMMENT '更新用户ID',
    INDEX idx_product_code (product_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统菜单表';

-- 2.5 角色菜单关联表
DROP TABLE IF EXISTS sys_role_menu;
CREATE TABLE sys_role_menu (
    role_id BIGINT NOT NULL COMMENT '角色 ID',
    menu_id BIGINT NOT NULL COMMENT '菜单 ID',
    PRIMARY KEY (role_id, menu_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色菜单关联表';

-- 2.6 数据字典表
DROP TABLE IF EXISTS sys_dict;
CREATE TABLE sys_dict (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键 ID',
    dict_code VARCHAR(50) NOT NULL COMMENT '字典编码',
    dict_name VARCHAR(100) NOT NULL COMMENT '字典名称',
    dict_type VARCHAR(50) NOT NULL COMMENT '字典类型（memo_type/priority 等）',
    dict_value VARCHAR(50) NOT NULL COMMENT '字典值',
    dict_label VARCHAR(100) NOT NULL COMMENT '字典标签',
    sort_order INT DEFAULT 0 COMMENT '排序',
    status TINYINT DEFAULT 1 COMMENT '状态（0 禁用 1 启用）',
    remark VARCHAR(255) COMMENT '备注',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    create_user BIGINT DEFAULT NULL COMMENT '创建用户ID',
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    update_user BIGINT DEFAULT NULL COMMENT '更新用户ID',
    deleted TINYINT DEFAULT 0 COMMENT '逻辑删除标志',
    UNIQUE KEY uk_dict_type_value (dict_type, dict_value)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据字典表';

-- ============================================
-- 3. 备忘录模块表结构（移除 memo_ 前缀）
-- ============================================

-- 3.1 联系人表
DROP TABLE IF EXISTS contact;
CREATE TABLE contact (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键 ID',
    user_id BIGINT NOT NULL COMMENT '所属用户 ID',
    name VARCHAR(50) NOT NULL COMMENT '姓名',
    gender TINYINT COMMENT '性别（0 女 1 男）',
    phone VARCHAR(20) COMMENT '手机号',
    email VARCHAR(100) COMMENT '邮箱',
    contact_group VARCHAR(50) COMMENT '分组（字典值）',
    company VARCHAR(100) COMMENT '公司',
    position VARCHAR(100) COMMENT '职务',
    remark VARCHAR(255) COMMENT '备注',
    avatar VARCHAR(255) COMMENT '头像',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    create_user BIGINT DEFAULT NULL COMMENT '创建用户ID',
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    update_user BIGINT DEFAULT NULL COMMENT '更新用户ID',
    deleted TINYINT DEFAULT 0 COMMENT '逻辑删除标志'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='联系人表';

-- ============================================
-- 4. 任务管理模块表结构（移除 task_ 前缀）
-- ============================================

-- 4.1 统一事项表（任务和备忘录合并）
DROP TABLE IF EXISTS task_item;
CREATE TABLE task_item (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键 ID',
    user_id BIGINT NOT NULL COMMENT '用户 ID',
    
    -- 基本信息
    title VARCHAR(200) NOT NULL COMMENT '标题',
    description TEXT COMMENT '内容/描述',
    item_type VARCHAR(20) DEFAULT 'TASK' COMMENT '类型（TASK/MEMO/REQUIREMENT）',
    
    -- 分类和标签
    category VARCHAR(50) COMMENT '分类（工作/生活/学习等）',
    priority VARCHAR(20) DEFAULT 'MEDIUM' COMMENT '优先级（HIGH/MEDIUM/LOW）',
    tags VARCHAR(500) COMMENT '标签（逗号分隔）',
    
    -- 时间信息
    start_time DATETIME COMMENT '开始时间',
    deadline DATETIME COMMENT '截止时间',
    completed_time DATETIME COMMENT '完成时间',
    
    -- 简化状态（固定流程：TODO -> DOING -> DONE/CANCELLED）
    status_code VARCHAR(20) DEFAULT 'TODO' COMMENT '状态（TODO/DOING/DONE/CANCELLED）',
    progress INT DEFAULT 0 COMMENT '进度（0-100）',
    
    -- 周期性任务字段
    recurrence_type VARCHAR(20) DEFAULT 'NONE' COMMENT '重复类型（NONE/DAILY/WEEKLY/MONTHLY/YEARLY）',
    recurrence_rule JSON COMMENT '重复规则（JSON格式）',
    recurrence_end_type VARCHAR(20) DEFAULT 'NEVER' COMMENT '结束类型（NEVER/DATE/COUNT）',
    recurrence_end_date DATETIME COMMENT '结束日期',
    recurrence_count INT DEFAULT 10 COMMENT '重复次数',
    parent_recurrence_id BIGINT COMMENT '父任务ID',
    
    -- 审计字段
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    create_user BIGINT COMMENT '创建人',
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    update_user BIGINT COMMENT '更新人',
    deleted TINYINT DEFAULT 0 COMMENT '逻辑删除标志',
    
    INDEX idx_user_id (user_id),
    INDEX idx_status_code (status_code),
    INDEX idx_item_type (item_type),
    INDEX idx_deadline (deadline),
    INDEX idx_parent_recurrence (parent_recurrence_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='统一事项表';

-- 4.2 需求表
DROP TABLE IF EXISTS requirement;
CREATE TABLE requirement (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键 ID',
    title VARCHAR(200) NOT NULL COMMENT '需求名称',
    bank_name VARCHAR(100) COMMENT '所属银行',
    requirement_no VARCHAR(50) COMMENT '需求编号',
    priority VARCHAR(20) DEFAULT 'MEDIUM' COMMENT '优先级（HIGH/MEDIUM/LOW）',
    risk_level VARCHAR(20) DEFAULT 'MEDIUM' COMMENT '风险等级（HIGH/MEDIUM/LOW）',
    requirement_type VARCHAR(50) DEFAULT 'INTERNAL' COMMENT '需求类型（INTERNAL:内部优化/BANK:行内需求）',
    status VARCHAR(50) DEFAULT 'PENDING_CLARIFY' COMMENT '状态（PENDING_CLARIFY:待澄清/PENDING_SCHEDULING:待排期/DEVELOPING:开发中/TESTING:测试中/PENDING_RELEASE:待投产/COMPLETED:已完成）',
    schedule_date VARCHAR(50) COMMENT '排期',
    main_system VARCHAR(100) COMMENT '主 R 系统',
    developer VARCHAR(100) COMMENT '开发主 R|参与者',
    involved_systems VARCHAR(500) COMMENT '涉及系统',
    follower_id BIGINT COMMENT '需求跟进人 ID',
    follower_name VARCHAR(100) COMMENT '需求跟进人姓名',
    submitter_id BIGINT COMMENT '提交人 ID',
    submitter_name VARCHAR(100) COMMENT '提交人姓名',
    submit_time DATETIME COMMENT '需求提出时间',
    completion_date DATETIME COMMENT '需求完成时间',
    estimated_workload DECIMAL(10,2) COMMENT '预计工作量（人天）',
    group_type VARCHAR(50) COMMENT '所属小组（LOAN_MIDDLE:贷中/LOAN_PRE:贷前/RISK:风险）',
    last_modifier_id BIGINT COMMENT '最后修改人 ID',
    last_modifier_name VARCHAR(100) COMMENT '最后修改人姓名',
    description TEXT COMMENT '需求描述',
    remark VARCHAR(1000) COMMENT '备注',
    user_id BIGINT NOT NULL COMMENT '所属用户 ID（数据隔离）',
    deleted TINYINT DEFAULT 0 COMMENT '逻辑删除标记',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    create_user BIGINT DEFAULT NULL COMMENT '创建用户ID',
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    update_user BIGINT DEFAULT NULL COMMENT '更新用户ID',
    INDEX idx_user_id (user_id),
    INDEX idx_bank_name (bank_name),
    INDEX idx_requirement_no (requirement_no),
    INDEX idx_status (status),
    INDEX idx_priority (priority),
    INDEX idx_risk_level (risk_level),
    INDEX idx_group_type (group_type),
    INDEX idx_deleted (deleted)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='需求表';

-- 4.3 任务 - 需求关联表
DROP TABLE IF EXISTS task_requirement_rel;
CREATE TABLE task_requirement_rel (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键 ID',
    task_id BIGINT NOT NULL COMMENT '任务 ID',
    requirement_id BIGINT NOT NULL COMMENT '需求 ID',
    requirement_no VARCHAR(50) COMMENT '需求编号（冗余字段，便于查询）',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    create_user BIGINT DEFAULT NULL COMMENT '创建用户ID',
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    update_user BIGINT DEFAULT NULL COMMENT '更新用户ID',
    UNIQUE KEY uk_task_requirement (task_id, requirement_id),
    INDEX idx_requirement_id (requirement_id),
    INDEX idx_requirement_no (requirement_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务 - 需求关联表';

-- 4.4 站内消息表（保留，有用）
DROP TABLE IF EXISTS sys_message;
CREATE TABLE sys_message (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '消息 ID',
    user_id BIGINT NOT NULL COMMENT '接收用户 ID',
    title VARCHAR(200) NOT NULL COMMENT '消息标题',
    content TEXT COMMENT '消息内容',
    type VARCHAR(50) NOT NULL DEFAULT 'SYSTEM' COMMENT '消息类型：SYSTEM-系统消息，TASK-任务消息，REMINDER-提醒消息，PROJECT-项目消息',
    is_read TINYINT DEFAULT 0 COMMENT '是否已读：0-未读，1-已读',
    business_id BIGINT COMMENT '关联业务 ID（如任务 ID）',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    create_user BIGINT DEFAULT NULL COMMENT '创建用户ID',
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    update_user BIGINT DEFAULT NULL COMMENT '更新用户ID',
    read_time DATETIME COMMENT '阅读时间',
    deleted TINYINT DEFAULT 0 COMMENT '逻辑删除：0-未删除，1-已删除',
    INDEX idx_user_id (user_id),
    INDEX idx_is_read (is_read),
    INDEX idx_create_time (create_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='站内消息表';

-- ============================================
-- 4.5 业务扩展模块表
-- ============================================

-- 4.5.1 银行机构参数表
DROP TABLE IF EXISTS bank_param;
CREATE TABLE `bank_param` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `bank_type` VARCHAR(50) NOT NULL COMMENT '银行类型（联名卡/标卡行/分支行）',
  `bank_name` VARCHAR(100) NOT NULL COMMENT '银行名称',
  `bank_code` VARCHAR(50) NOT NULL COMMENT '银行机构号',
  `reserve1` VARCHAR(200) DEFAULT '' COMMENT '备用字段1',
  `reserve2` VARCHAR(200) DEFAULT '' COMMENT '备用字段2',
  `reserve3` VARCHAR(200) DEFAULT '' COMMENT '备用字段3',
  `reserve4` VARCHAR(200) DEFAULT '' COMMENT '备用字段4',
  `reserve5` VARCHAR(200) DEFAULT '' COMMENT '备用字段5',
  `remark` VARCHAR(500) DEFAULT '' COMMENT '备注',
  `create_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `create_user` BIGINT DEFAULT NULL COMMENT '创建人',
  `update_time` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `update_user` BIGINT DEFAULT NULL COMMENT '更新人',
  `deleted` TINYINT DEFAULT 0 COMMENT '逻辑删除（0-未删除 1-已删除）',
  PRIMARY KEY (`id`),
  INDEX `idx_bank_type` (`bank_type`),
  INDEX `idx_bank_name` (`bank_name`),
  INDEX `idx_bank_code` (`bank_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='银行机构参数';

-- 4.5.2 批量任务清单表
DROP TABLE IF EXISTS batch_task;
CREATE TABLE `batch_task` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `bank_type` VARCHAR(50) NOT NULL COMMENT '银行类型（联名卡/标卡行/分支行）',
  `bank_name` VARCHAR(100) NOT NULL COMMENT '银行名称',
  `system_name` VARCHAR(50) NOT NULL COMMENT '批量所属系统',
  `task_type` VARCHAR(50) NOT NULL COMMENT '批量任务类型（定时任务/批量Job）',
  `task_name` VARCHAR(200) NOT NULL COMMENT '批量任务名（Job标识）',
  `task_desc` VARCHAR(500) DEFAULT '' COMMENT '批量任务描述',
  `exec_time` VARCHAR(100) DEFAULT '' COMMENT '任务执行时间（cron表达式或时间点）',
  `upstream_system` VARCHAR(200) DEFAULT '' COMMENT '上游依赖系统',
  `upstream_info` VARCHAR(500) DEFAULT '' COMMENT '上游依赖信息',
  `downstream_system` VARCHAR(200) DEFAULT '' COMMENT '下游依赖系统',
  `downstream_info` VARCHAR(500) DEFAULT '' COMMENT '下游依赖信息',
  `remark` VARCHAR(500) DEFAULT '' COMMENT '备注',
  `create_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `create_user` BIGINT DEFAULT NULL COMMENT '创建人',
  `update_time` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `update_user` BIGINT DEFAULT NULL COMMENT '更新人',
  `deleted` TINYINT DEFAULT 0 COMMENT '逻辑删除（0-未删除 1-已删除）',
  PRIMARY KEY (`id`),
  INDEX `idx_bank_name` (`bank_name`),
  INDEX `idx_system_name` (`system_name`),
  INDEX `idx_task_type` (`task_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='批量任务清单';

-- 4.5.3 生产环境节点清单表
DROP TABLE IF EXISTS prod_node;
CREATE TABLE `prod_node` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `bank_type` varchar(50) DEFAULT '' COMMENT '银行类型',
  `bank_name` VARCHAR(100) NOT NULL COMMENT '银行名称',
  `env_type` VARCHAR(50) NOT NULL COMMENT '环境类型（如二代生产）',
  `node_name` VARCHAR(100) NOT NULL COMMENT '节点简称',
  `node_ip` VARCHAR(50) NOT NULL COMMENT '节点IP',
  `username` VARCHAR(50) DEFAULT '' COMMENT '用户名',
  `is_mutual_trust` VARCHAR(20) DEFAULT '' COMMENT '是否互信节点',
  `create_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `create_user` BIGINT DEFAULT NULL COMMENT '创建人',
  `update_time` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `update_user` BIGINT DEFAULT NULL COMMENT '更新人',
  `deleted` TINYINT DEFAULT 0 COMMENT '逻辑删除（0-未删除 1-已删除）',
  PRIMARY KEY (`id`),
  INDEX `idx_bank_name` (`bank_name`),
  INDEX `idx_env_type` (`env_type`),
  INDEX `idx_node_ip` (`node_ip`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='生产环境节点清单';

-- ============================================
-- 4.6 项目管理模块表结构
-- ============================================

-- 4.6.1 项目表
DROP TABLE IF EXISTS project;
CREATE TABLE `project` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `project_name` VARCHAR(200) NOT NULL COMMENT '项目名称',
  `project_key` VARCHAR(50) NOT NULL COMMENT '项目Key（服务层在未删除范围内校验唯一）',
  `description` TEXT COMMENT '项目描述',
  `plan_desc` TEXT COMMENT '整体计划说明',
  `manager_id` BIGINT NOT NULL COMMENT '项目经理用户ID',
  `start_date` DATE NOT NULL COMMENT '开始日期',
  `end_date` DATE NOT NULL COMMENT '结束日期',
  `status` VARCHAR(20) NOT NULL DEFAULT 'STARTING' COMMENT '状态：STARTING/IN_PROGRESS/PAUSED/CLOSED/TERMINATED',
  `risk_level` VARCHAR(20) NOT NULL DEFAULT 'MEDIUM' COMMENT '风险等级：HIGH/MEDIUM/LOW',
  `delivery_status` VARCHAR(20) NOT NULL DEFAULT 'NORMAL' COMMENT '交付状态：NORMAL/AT_RISK/CANNOT_FINISH',
  `risk_note` VARCHAR(1000) DEFAULT '' COMMENT '风险说明',
  `progress_pct` INT NOT NULL DEFAULT 0 COMMENT '进度百分比（仅服务端 RecalcProjectProgress 计算）',
  `week_count` INT NOT NULL DEFAULT 0 COMMENT '非跳过周计划数（进度分母）',
  `done_week_count` INT NOT NULL DEFAULT 0 COMMENT '已完成周数（进度分子）',
  `paused_at` DATETIME DEFAULT NULL COMMENT '最近暂停时间',
  `closed_at` DATETIME DEFAULT NULL COMMENT '结项/终止时间',
  `create_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `create_user` BIGINT DEFAULT NULL COMMENT '创建人',
  `update_time` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `update_user` BIGINT DEFAULT NULL COMMENT '更新人',
  `deleted` TINYINT DEFAULT 0 COMMENT '逻辑删除（0-未删除 1-已删除）',
  PRIMARY KEY (`id`),
  INDEX `idx_manager` (`manager_id`),
  INDEX `idx_status_risk` (`status`, `risk_level`),
  INDEX `idx_date_range` (`start_date`, `end_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='项目';

-- 4.6.2 项目周计划表
DROP TABLE IF EXISTS project_weekly_plan;
CREATE TABLE `project_weekly_plan` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `project_id` BIGINT NOT NULL COMMENT '项目ID',
  `week_no` INT NOT NULL COMMENT '项目内周序号（服务端 max+1 生成）',
  `title` VARCHAR(200) NOT NULL COMMENT '周主题',
  `plan_desc` TEXT COMMENT '本周计划内容',
  `start_date` DATE NOT NULL COMMENT '周开始日期',
  `end_date` DATE NOT NULL COMMENT '周结束日期',
  `owner_id` BIGINT NOT NULL COMMENT '周负责人用户ID',
  `status` VARCHAR(20) NOT NULL DEFAULT 'NOT_STARTED' COMMENT '状态：NOT_STARTED/IN_PROGRESS/DONE/SKIPPED',
  `remark` VARCHAR(1000) DEFAULT '' COMMENT '备注',
  `summary` TEXT COMMENT '周总结JSON：finished/nextPlan/risks/support',
  `summary_by` BIGINT DEFAULT NULL COMMENT '总结提交人',
  `summary_at` DATETIME DEFAULT NULL COMMENT '总结提交时间（非空即锁定）',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序（默认同 week_no）',
  `create_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `create_user` BIGINT DEFAULT NULL COMMENT '创建人',
  `update_time` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `update_user` BIGINT DEFAULT NULL COMMENT '更新人',
  `deleted` TINYINT DEFAULT 0 COMMENT '逻辑删除（0-未删除 1-已删除）',
  PRIMARY KEY (`id`),
  INDEX `idx_project_sort` (`project_id`, `sort_order`),
  INDEX `idx_owner_status` (`owner_id`, `status`),
  INDEX `idx_date` (`end_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='项目周计划';

-- 4.6.3 项目成员表
DROP TABLE IF EXISTS project_member;
CREATE TABLE `project_member` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `project_id` BIGINT NOT NULL COMMENT '项目ID',
  `user_id` BIGINT NOT NULL COMMENT '用户ID',
  `member_role` VARCHAR(20) NOT NULL DEFAULT 'MEMBER' COMMENT '成员角色：PM/MEMBER',
  `create_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `create_user` BIGINT DEFAULT NULL COMMENT '创建人',
  `update_time` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `update_user` BIGINT DEFAULT NULL COMMENT '更新人',
  `deleted` TINYINT DEFAULT 0 COMMENT '逻辑删除（0-未删除 1-已删除）',
  PRIMARY KEY (`id`),
  INDEX `idx_project_user` (`project_id`, `user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='项目成员';

-- 4.6.4 项目变更日志表（仅记项目级关键变更，无软删除字段）
DROP TABLE IF EXISTS project_change_log;
CREATE TABLE `project_change_log` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `project_id` BIGINT NOT NULL COMMENT '项目ID',
  `action` VARCHAR(50) NOT NULL COMMENT '动作：CREATE/UPDATE/STATUS/RISK_CHANGE/DELETE/MEMBER_ADD/MEMBER_REMOVE/WEEK_ADD/WEEK_EDIT/WEEK_DELETE/WEEK_STATUS/SUMMARY_SUBMIT/SUMMARY_UNLOCK/STAGE_ADD/STAGE_EDIT/STAGE_DELETE',
  `detail` VARCHAR(1000) DEFAULT '' COMMENT '变更详情',
  `operator_id` BIGINT NOT NULL COMMENT '操作人用户ID',
  `create_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  INDEX `idx_project_time` (`project_id`, `create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='项目变更日志';

-- 4.6.5 任务-周计划关联表（多周关联：一个事项可同时出现在多个周清单；物理删除，对齐 task_requirement_rel 先例）
DROP TABLE IF EXISTS task_week_plan_rel;
CREATE TABLE `task_week_plan_rel` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `task_id` BIGINT NOT NULL COMMENT '事项ID（task_item.id）',
  `week_plan_id` BIGINT NOT NULL COMMENT '周计划ID（project_weekly_plan.id）',
  `create_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `create_user` BIGINT DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_task_week` (`task_id`, `week_plan_id`),
  INDEX `idx_week_plan` (`week_plan_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务-周计划关联表';

-- 4.6.6 项目阶段表（同一项目未删除范围内阶段名称唯一，服务层校验）
DROP TABLE IF EXISTS project_stage;
CREATE TABLE `project_stage` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `project_id` BIGINT NOT NULL COMMENT '项目ID（project.id）',
  `stage_name` VARCHAR(100) NOT NULL COMMENT '阶段名称（同一项目未删除范围内唯一，服务层校验）',
  `description` VARCHAR(500) DEFAULT '' COMMENT '阶段说明',
  `start_date` DATE NOT NULL COMMENT '阶段开始日期',
  `end_date` DATE NOT NULL COMMENT '阶段结束日期',
  `create_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `create_user` BIGINT DEFAULT NULL COMMENT '创建人',
  `update_time` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `update_user` BIGINT DEFAULT NULL COMMENT '更新人',
  `deleted` TINYINT DEFAULT 0 COMMENT '逻辑删除（0-未删除 1-已删除）',
  PRIMARY KEY (`id`),
  INDEX `idx_project_start` (`project_id`, `start_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='项目阶段';

-- ============================================
-- 5. Casbin 权限策略表（保留原表名）
-- ============================================
DROP TABLE IF EXISTS casbin_rule;
CREATE TABLE casbin_rule (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    ptype VARCHAR(100) NOT NULL COMMENT '策略类型',
    v0 VARCHAR(100) DEFAULT NULL COMMENT '主体',
    v1 VARCHAR(100) DEFAULT NULL COMMENT '资源',
    v2 VARCHAR(100) DEFAULT NULL COMMENT '操作',
    v3 VARCHAR(100) DEFAULT NULL,
    v4 VARCHAR(100) DEFAULT NULL,
    v5 VARCHAR(100) DEFAULT NULL,
    UNIQUE KEY unique_index (ptype, v0, v1, v2, v3, v4, v5)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Casbin权限策略表';

-- ============================================
-- 6. 插入基本数据
-- ============================================

-- 6.1 默认管理员用户 (密码 admin123)
INSERT INTO sys_user (username, password, real_name, email, phone, status)
VALUES ('admin', '$2a$10$cj0u1E/WQNOZc2WARM4fVezAiYNr1tnAwIS.O1Lo.HZ23nmLsU7W.', '管理员', 'admin@righthand.com', '13800138000', 1);

-- 6.2 默认普通用户 test（密码 test123）
INSERT INTO sys_user (username, password, real_name, email, phone, status)
VALUES ('test', '$2a$10$cj0u1E/WQNOZc2WARM4fVezAiYNr1tnAwIS.O1Lo.HZ23nmLsU7W.', '测试用户', 'test@righthand.com', '13800138001', 1);

-- 6.3 默认角色
INSERT INTO sys_role (role_name, role_code, description) VALUES
('超级管理员', 'ADMIN', '拥有所有权限'),
('普通用户', 'USER', '普通用户权限');

-- 6.4 用户角色关联
INSERT INTO sys_user_role (user_id, role_id) VALUES 
(1, 1),  -- admin -> ADMIN
(2, 2);  -- test -> USER

-- 6.5 菜单数据（包含需求管理）

-- 一级菜单
INSERT INTO sys_menu (menu_name, parent_id, sort_order, path, component, menu_type, icon, product_code) VALUES
('仪表盘', 0, 1, '/dashboard', 'Dashboard', 1, 'Odometer', 'auth'),
('事项管理', 0, 2, '/tasks', 'tasks/Tasks', 1, 'List', 'auth'),
('需求管理', 0, 4, '/requirements', 'requirements/Requirements', 1, 'ClipboardList', 'auth'),
('联系人', 0, 5, '/contacts', 'Contacts', 1, 'User', 'auth');

-- 系统管理作为目录单独插入
INSERT INTO sys_menu (menu_name, parent_id, sort_order, path, component, menu_type, icon, product_code) VALUES
('系统管理', 0, 5, '/auth', NULL, 0, 'Setting', 'auth');

-- 获取系统管理的ID
SET @system_menu_id = LAST_INSERT_ID();

-- 二级菜单（系统管理下）
INSERT INTO sys_menu (menu_name, parent_id, sort_order, path, component, menu_type, icon, product_code) VALUES
('用户管理', @system_menu_id, 1, 'users', 'system/UserManagement', 1, 'User', 'auth'),
('角色管理', @system_menu_id, 2, 'roles', 'system/RoleManagement', 1, 'UserFilled', 'auth'),
('菜单管理', @system_menu_id, 3, 'menus', 'system/MenuManagement', 1, 'Menu', 'auth'),
('字典管理', @system_menu_id, 4, 'dicts', 'system/DictManagement', 1, 'Document', 'auth');

-- 6.6 为admin角色分配所有菜单权限
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT 1, id FROM sys_menu;

-- 6.7 为普通用户角色分配除系统管理外的所有菜单权限
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT 2, id FROM sys_menu 
WHERE id NOT IN (
    SELECT id FROM sys_menu WHERE parent_id = @system_menu_id  -- 系统管理子菜单
    UNION
    SELECT @system_menu_id  -- 系统管理目录本身
);

-- 6.8 扩展模块菜单与权限配置

-- 6.8.1 银行机构参数管理菜单与按钮权限
INSERT INTO sys_menu (menu_name, parent_id, sort_order, path, component, menu_type, icon, product_code) VALUES
('银行信息维护', 0, 8, '/bank-params', 'BankParams', 1, 'OfficeBuilding', 'auth');

SET @bank_param_menu_id = (SELECT id FROM sys_menu WHERE menu_name = '银行信息维护' AND parent_id = 0);

INSERT INTO sys_menu (menu_name, parent_id, sort_order, path, component, menu_type, icon, product_code, permission) VALUES
('查看参数', @bank_param_menu_id, 0, NULL, NULL, 2, NULL, 'auth', 'bankParam:view'),
('新增参数', @bank_param_menu_id, 1, NULL, NULL, 2, NULL, 'auth', 'bankParam:add'),
('编辑参数', @bank_param_menu_id, 2, NULL, NULL, 2, NULL, 'auth', 'bankParam:edit'),
('删除参数', @bank_param_menu_id, 3, NULL, NULL, 2, NULL, 'auth', 'bankParam:delete');

-- 为管理员角色分配菜单和按钮权限
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT 1, id FROM sys_menu WHERE id = @bank_param_menu_id OR parent_id = @bank_param_menu_id;

-- 为普通用户角色分配菜单和按钮权限
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT 2, id FROM sys_menu WHERE id = @bank_param_menu_id OR parent_id = @bank_param_menu_id;

-- 6.8.2 批量任务清单菜单与按钮权限
INSERT INTO sys_menu (menu_name, parent_id, sort_order, path, component, menu_type, icon, product_code) VALUES
('批量任务管理', 0, 6, '/batch-tasks', 'BatchTasks', 1, 'Tickets', 'auth');

SET @batch_task_menu_id = (SELECT id FROM sys_menu WHERE menu_name = '批量任务管理' AND parent_id = 0);

INSERT INTO sys_menu (menu_name, parent_id, sort_order, path, component, menu_type, icon, product_code, permission) VALUES
('查看批量任务', @batch_task_menu_id, 0, NULL, NULL, 2, NULL, 'auth', 'batchTask:view'),
('新增批量任务', @batch_task_menu_id, 1, NULL, NULL, 2, NULL, 'auth', 'batchTask:add'),
('编辑批量任务', @batch_task_menu_id, 2, NULL, NULL, 2, NULL, 'auth', 'batchTask:edit'),
('删除批量任务', @batch_task_menu_id, 3, NULL, NULL, 2, NULL, 'auth', 'batchTask:delete');

-- 为管理员角色分配菜单和按钮权限
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT 1, id FROM sys_menu WHERE id = @batch_task_menu_id OR parent_id = @batch_task_menu_id;

-- 为普通用户角色分配菜单和按钮权限
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT 2, id FROM sys_menu WHERE id = @batch_task_menu_id OR parent_id = @batch_task_menu_id;

-- 6.8.3 生产环境节点清单菜单与按钮权限
INSERT INTO sys_menu (menu_name, parent_id, sort_order, path, component, menu_type, icon, product_code) VALUES
('环境清单', 0, 7, '/prod-nodes', 'ProdNodes', 1, 'Monitor', 'auth');

SET @prod_node_menu_id = (SELECT id FROM sys_menu WHERE menu_name = '环境清单' AND parent_id = 0);

INSERT INTO sys_menu (menu_name, parent_id, sort_order, path, component, menu_type, icon, product_code, permission) VALUES
('查看节点', @prod_node_menu_id, 0, NULL, NULL, 2, NULL, 'auth', 'prodNode:view'),
('新增节点', @prod_node_menu_id, 1, NULL, NULL, 2, NULL, 'auth', 'prodNode:add'),
('编辑节点', @prod_node_menu_id, 2, NULL, NULL, 2, NULL, 'auth', 'prodNode:edit'),
('删除节点', @prod_node_menu_id, 3, NULL, NULL, 2, NULL, 'auth', 'prodNode:delete');

-- 为管理员角色分配菜单和按钮权限
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT 1, id FROM sys_menu WHERE id = @prod_node_menu_id OR parent_id = @prod_node_menu_id;

-- 为普通用户角色分配菜单和按钮权限
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT 2, id FROM sys_menu WHERE id = @prod_node_menu_id OR parent_id = @prod_node_menu_id;

-- 6.8.4 项目管理菜单与按钮权限（一级菜单，排列在「事项管理」之后）
INSERT INTO sys_menu (menu_name, parent_id, sort_order, path, component, menu_type, icon, product_code) VALUES
('项目管理', 0, 3, '/projects', 'project/ProjectList', 1, 'Management', 'auth');

SET @project_menu_id = LAST_INSERT_ID();

INSERT INTO sys_menu (menu_name, parent_id, sort_order, path, component, menu_type, icon, product_code, permission) VALUES
('新建项目', @project_menu_id, 1, NULL, NULL, 2, NULL, 'auth', 'project:add'),
('编辑项目', @project_menu_id, 2, NULL, NULL, 2, NULL, 'auth', 'project:edit'),
('删除项目', @project_menu_id, 3, NULL, NULL, 2, NULL, 'auth', 'project:delete'),
('项目状态流转', @project_menu_id, 4, NULL, NULL, 2, NULL, 'auth', 'project:status'),
('项目成员管理', @project_menu_id, 5, NULL, NULL, 2, NULL, 'auth', 'project:member'),
('周计划管理', @project_menu_id, 6, NULL, NULL, 2, NULL, 'auth', 'project:week'),
('周总结解锁', @project_menu_id, 7, NULL, NULL, 2, NULL, 'auth', 'project:summary'),
('阶段管理', @project_menu_id, 8, NULL, NULL, 2, NULL, 'auth', 'project:stage');

-- 为管理员角色分配菜单和按钮权限
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT 1, id FROM sys_menu
WHERE id = @project_menu_id OR parent_id = @project_menu_id;

-- 为普通用户角色分配菜单和按钮权限
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT 2, id FROM sys_menu
WHERE id = @project_menu_id OR parent_id = @project_menu_id;

-- 6.9 默认字典数据
INSERT INTO sys_dict (dict_code, dict_name, dict_type, dict_value, dict_label, sort_order, status, remark) VALUES
-- 备忘类型
('MEMO_TYPE_WORK', '备忘类型 - 工作', 'memo_type', 'WORK', '工作', 1, 1, '工作类型'),
('MEMO_TYPE_LIFE', '备忘类型 - 生活', 'memo_type', 'LIFE', '生活', 2, 1, '生活类型'),
('MEMO_TYPE_STUDY', '备忘类型 - 学习', 'memo_type', 'STUDY', '学习', 3, 1, '学习类型'),

-- 优先级
('PRIORITY_HIGH', '优先级 - 高', 'priority', 'HIGH', '高', 1, 1, '高优先级'),
('PRIORITY_MEDIUM', '优先级 - 中', 'priority', 'MEDIUM', '中', 2, 1, '中优先级'),
('PRIORITY_LOW', '优先级 - 低', 'priority', 'LOW', '低', 3, 1, '低优先级'),

-- 任务类型
('TASK_TYPE_TASK', '任务类型 - 普通任务', 'task_type', 'TASK', '普通任务', 1, 1, '普通任务'),
('TASK_TYPE_REQUIREMENT', '任务类型 - 需求', 'task_type', 'REQUIREMENT', '需求', 2, 1, '需求'),
('TASK_TYPE_BUG', '任务类型 - 缺陷', 'task_type', 'BUG', 'Bug', 3, 1, '缺陷'),
('TASK_TYPE_INCIDENT', '任务类型 - 生产事件', 'task_type', 'INCIDENT', '生产事件', 4, 1, '生产事件'),

-- 状态
('STATUS_UNFINISHED', '状态 - 未完成', 'status', '0', '未完成', 1, 1, '未完成状态'),
('STATUS_FINISHED', '状态 - 已完成', 'status', '1', '已完成', 2, 1, '已完成状态'),

-- 性别
('GENDER_FEMALE', '性别 - 女', 'gender', '0', '女', 1, 1, '女性'),
('GENDER_MALE', '性别 - 男', 'gender', '1', '男', 2, 1, '男性'),

-- 需求类型
('REQ_TYPE_INTERNAL', '需求类型 - 内部优化', 'requirement_type', 'INTERNAL', '内部优化', 1, 1, '内部优化需求'),
('REQ_TYPE_BANK', '需求类型 - 行内需求', 'requirement_type', 'BANK', '行内需求', 2, 1, '行内业务需求'),

-- 需求状态
('REQ_STATUS_PENDING', '需求状态 - 待评审', 'requirement_status', 'PENDING', '待评审', 1, 1, '等待评审'),
('REQ_STATUS_APPROVED', '需求状态 - 已批准', 'requirement_status', 'APPROVED', '已批准', 2, 1, '评审通过'),
('REQ_STATUS_REJECTED', '需求状态 - 已拒绝', 'requirement_status', 'REJECTED', '已拒绝', 3, 1, '评审不通过'),
('REQ_STATUS_IN_PROGRESS', '需求状态 - 进行中', 'requirement_status', 'IN_PROGRESS', '进行中', 4, 1, '正在实施'),
('REQ_STATUS_COMPLETED', '需求状态 - 已完成', 'requirement_status', 'COMPLETED', '已完成', 5, 1, '已完成'),

-- 联系人群组
('CONTACT_GROUP_CUSTOMER_MANAGER', '联系人群组 - 客户经理', 'contact_group', 'ACCOUNT_MANAGER', '客户经理', 1, 1, '客户经理'),
('CONTACT_GROUP_R_AND_D', '联系人群组 - 研发', 'contact_group', 'R&D', '研发', 2, 1, '研发人员'),
('CONTACT_GROUP_PRODUCT_MANAGER', '联系人群组 - 产品经理', 'contact_group', 'PRODUCT', '产品经理', 3, 1, '产品经理');

-- 6.10 扩展模块字典数据

-- 6.10.1 风险等级和需求所属小组字典
INSERT INTO sys_dict (dict_code, dict_name, dict_type, dict_value, dict_label, sort_order, status, remark, create_time, create_user, update_time, update_user, deleted) VALUES
('RISK_LEVEL_HIGH', '风险等级 - 高', 'risk_level', 'HIGH', '高', 1, 1, '风险等级-高', NOW(), 1, NOW(), 1, 0),
('RISK_LEVEL_MEDIUM', '风险等级 - 中', 'risk_level', 'MEDIUM', '中', 2, 1, '风险等级-中', NOW(), 1, NOW(), 1, 0),
('RISK_LEVEL_LOW', '风险等级 - 低', 'risk_level', 'LOW', '低', 3, 1, '风险等级-低', NOW(), 1, NOW(), 1, 0),
('REQ_GROUP_LOAN_MIDDLE', '所属小组 - 贷中', 'requirement_group', 'LOAN_MIDDLE', '贷中', 1, 1, '所属小组-贷中', NOW(), 1, NOW(), 1, 0),
('REQ_GROUP_LOAN_PRE', '所属小组 - 贷前', 'requirement_group', 'LOAN_PRE', '贷前', 2, 1, '所属小组-贷前', NOW(), 1, NOW(), 1, 0),
('REQ_GROUP_RISK', '所属小组 - 风险', 'requirement_group', 'RISK', '风险', 3, 1, '所属小组-风险', NOW(), 1, NOW(), 1, 0);

-- 6.10.2 批量任务/生产环境节点相关字典
INSERT INTO sys_dict (dict_code, dict_name, dict_type, dict_value, dict_label, sort_order, status, remark) VALUES
-- 所属系统（system_name）
('SYSTEM_NAME_POSEIDON', '所属系统 - poseidon', 'system_name', 'poseidon', 'poseidon', 1, 1, NULL),
('SYSTEM_NAME_PMS', '所属系统 - pms', 'system_name', 'pms', 'pms', 2, 1, NULL),
('SYSTEM_NAME_CTS', '所属系统 - cts', 'system_name', 'cts', 'cts', 3, 1, NULL),
('SYSTEM_NAME_WWDAS', '所属系统 - wwdas', 'system_name', 'wwdas', 'wwdas', 4, 1, NULL),
('SYSTEM_NAME_RAS', '所属系统 - ras', 'system_name', 'ras', 'ras', 5, 1, NULL),
('SYSTEM_NAME_RDS', '所属系统 - rds', 'system_name', 'rds', 'rds', 6, 1, NULL),
('SYSTEM_NAME_CDS', '所属系统 - cds', 'system_name', 'cds', 'cds', 7, 1, NULL),
('SYSTEM_NAME_PIC', '所属系统 - pic', 'system_name', 'pic', 'pic', 8, 1, NULL),

-- 批量任务类型（batch_task_type）
('BATCH_TASK_TYPE_SCHEDULED', '批量任务类型 - 定时任务', 'batch_task_type', '定时任务', '定时任务', 1, 1, '定时调度任务'),
('BATCH_TASK_TYPE_BATCH_JOB', '批量任务类型 - 批量Job', 'batch_task_type', '批量Job', '批量Job', 2, 1, '批量批处理任务'),

-- 环境类型（env_type）
('ENV_TYPE_GEN2_PROD', '环境类型 - 二代生产', 'env_type', '二代生产', '二代生产', 1, 1, '二代生产环境'),
('ENV_TYPE_GEN1_PROD', '环境类型 - 一代生产', 'env_type', '一代生产', '一代生产', 2, 1, '一代生产环境'),
('ENV_TYPE_TEST', '环境类型 - 测试环境', 'env_type', '测试环境', '测试环境', 3, 1, '测试环境'),
('ENV_TYPE_DEV', '环境类型 - 开发环境', 'env_type', '开发环境', '开发环境', 4, 1, '开发环境');

-- ============================================
-- 7. Casbin 权限策略初始化
-- ============================================

-- 7.1 系统管理模块权限（仅管理员）
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
-- 用户管理
('p', 'role:admin', '/api/system/users', 'GET'),
('p', 'role:admin', '/api/system/users', 'POST'),
('p', 'role:admin', '/api/system/users', 'PUT'),
('p', 'role:admin', '/api/system/users', 'DELETE'),
('p', 'role:admin', '/api/system/users/*/roles', 'POST'),
('p', 'role:user', '/api/system/users', 'GET'),

-- 角色管理
('p', 'role:admin', '/api/system/roles', 'GET'),
('p', 'role:admin', '/api/system/roles', 'POST'),
('p', 'role:admin', '/api/system/roles', 'PUT'),
('p', 'role:admin', '/api/system/roles', 'DELETE'),
('p', 'role:admin', '/api/system/roles/*/menus', 'POST'),

-- 菜单管理
('p', 'role:admin', '/api/system/menus', 'GET'),
('p', 'role:admin', '/api/system/menus', 'POST'),
('p', 'role:admin', '/api/system/menus', 'PUT'),
('p', 'role:admin', '/api/system/menus', 'DELETE'),

-- 字典管理
('p', 'role:admin', '/api/system/dicts', 'GET'),
('p', 'role:admin', '/api/system/dicts', 'POST'),
('p', 'role:admin', '/api/system/dicts', 'PUT'),
('p', 'role:admin', '/api/system/dicts', 'DELETE'),
('p', 'role:admin', '/api/system/dicts/type/*', 'GET');

-- 7.2 备忘录模块权限
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
-- 备忘录管理
('p', 'role:admin', '/api/memo/memos', 'GET'),
('p', 'role:admin', '/api/memo/memos', 'POST'),
('p', 'role:admin', '/api/memo/memos', 'PUT'),
('p', 'role:admin', '/api/memo/memos', 'DELETE'),
('p', 'role:admin', '/api/memo/memos/*/complete', 'PUT'),
('p', 'role:user', '/api/memo/memos', 'GET'),
('p', 'role:user', '/api/memo/memos', 'POST'),
('p', 'role:user', '/api/memo/memos', 'PUT'),
('p', 'role:user', '/api/memo/memos', 'DELETE'),
('p', 'role:user', '/api/memo/memos/*/complete', 'PUT'),

-- 联系人管理
('p', 'role:admin', '/api/contacts', 'GET'),
('p', 'role:admin', '/api/contacts', 'POST'),
('p', 'role:admin', '/api/contacts', 'PUT'),
('p', 'role:admin', '/api/contacts', 'DELETE'),
('p', 'role:user', '/api/contacts', 'GET'),
('p', 'role:user', '/api/contacts', 'POST'),
('p', 'role:user', '/api/contacts', 'PUT'),
('p', 'role:user', '/api/contacts', 'DELETE'),

-- 仪表盘统计
('p', 'role:admin', '/api/memo/dashboard/stats', 'GET'),
('p', 'role:user', '/api/memo/dashboard/stats', 'GET');

-- 7.3 任务模块权限
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
-- 任务管理
('p', 'role:admin', '/api/task/tasks', 'GET'),
('p', 'role:admin', '/api/task/tasks', 'POST'),
('p', 'role:admin', '/api/task/tasks', 'PUT'),
('p', 'role:admin', '/api/task/tasks', 'DELETE'),
('p', 'role:admin', '/api/task/tasks/*/transition', 'PUT'),
('p', 'role:user', '/api/task/tasks', 'GET'),
('p', 'role:user', '/api/task/tasks', 'POST'),
('p', 'role:user', '/api/task/tasks', 'PUT'),
('p', 'role:user', '/api/task/tasks', 'DELETE'),
('p', 'role:user', '/api/task/tasks/*/transition', 'PUT'),

-- 需求管理
('p', 'role:admin', '/api/requirements', 'GET'),
('p', 'role:admin', '/api/requirements', 'POST'),
('p', 'role:admin', '/api/requirements', 'PUT'),
('p', 'role:admin', '/api/requirements', 'DELETE'),
('p', 'role:admin', '/api/requirements/export', 'GET'),
('p', 'role:admin', '/api/requirements/import', 'POST'),
('p', 'role:user', '/api/requirements', 'GET'),
('p', 'role:user', '/api/requirements', 'POST'),
('p', 'role:user', '/api/requirements', 'PUT'),
('p', 'role:user', '/api/requirements', 'DELETE'),
('p', 'role:user', '/api/requirements/export', 'GET'),
('p', 'role:user', '/api/requirements/import', 'POST'),

-- 工作流管理（仅管理员）
('p', 'role:admin', '/api/task/workflows', 'GET'),
('p', 'role:admin', '/api/task/workflows/*/statuses', 'GET'),
('p', 'role:admin', '/api/task/workflows/*/transitions', 'PUT'),

-- 消息中心
('p', 'role:admin', '/api/task/messages', 'GET'),
('p', 'role:admin', '/api/task/messages/*/read', 'PUT'),
('p', 'role:user', '/api/task/messages', 'GET'),
('p', 'role:user', '/api/task/messages/*/read', 'PUT');

-- 7.4 事项管理权限（统一API）
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
('p', 'role:admin', '/api/tasks', 'GET'),
('p', 'role:admin', '/api/tasks', 'POST'),
('p', 'role:admin', '/api/tasks', 'PUT'),
('p', 'role:admin', '/api/tasks', 'DELETE'),
('p', 'role:user', '/api/tasks', 'GET'),
('p', 'role:user', '/api/tasks', 'POST'),
('p', 'role:user', '/api/tasks', 'PUT'),
('p', 'role:user', '/api/tasks', 'DELETE');

-- 7.5 仪表盘权限
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
('p', 'role:admin', '/api/dashboard/stats', 'GET'),
('p', 'role:user', '/api/dashboard/stats', 'GET');

-- 7.6 扩展模块权限策略

-- 银行机构参数
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
('p', 'role:admin', '/api/bank-params', 'GET'),
('p', 'role:admin', '/api/bank-params', 'POST'),
('p', 'role:admin', '/api/bank-params', 'PUT'),
('p', 'role:admin', '/api/bank-params', 'DELETE'),
('p', 'role:user', '/api/bank-params', 'GET'),
('p', 'role:user', '/api/bank-params', 'POST'),
('p', 'role:user', '/api/bank-params', 'PUT'),
('p', 'role:user', '/api/bank-params', 'DELETE');

-- 批量任务清单
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
('p', 'role:admin', '/api/batch-tasks', 'GET'),
('p', 'role:admin', '/api/batch-tasks', 'POST'),
('p', 'role:admin', '/api/batch-tasks', 'PUT'),
('p', 'role:admin', '/api/batch-tasks', 'DELETE'),
('p', 'role:user', '/api/batch-tasks', 'GET'),
('p', 'role:user', '/api/batch-tasks', 'POST'),
('p', 'role:user', '/api/batch-tasks', 'PUT'),
('p', 'role:user', '/api/batch-tasks', 'DELETE');

-- 生产环境节点清单
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
('p', 'role:admin', '/api/prod-nodes', 'GET'),
('p', 'role:admin', '/api/prod-nodes', 'POST'),
('p', 'role:admin', '/api/prod-nodes', 'PUT'),
('p', 'role:admin', '/api/prod-nodes', 'DELETE'),
('p', 'role:admin', '/api/prod-nodes/import', 'POST'),
('p', 'role:user', '/api/prod-nodes', 'GET'),
('p', 'role:user', '/api/prod-nodes', 'POST'),
('p', 'role:user', '/api/prod-nodes', 'PUT'),
('p', 'role:user', '/api/prod-nodes', 'DELETE'),
('p', 'role:user', '/api/prod-nodes/import', 'POST');

-- 7.7 项目管理与消息中心权限策略
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
-- 项目管理（具体权限由服务层 CanAct 按角色计算，此处仅放行登录用户访问）
('p', 'role:admin', '/api/projects', 'GET'),
('p', 'role:admin', '/api/projects', 'POST'),
('p', 'role:admin', '/api/projects/*', 'GET'),
('p', 'role:admin', '/api/projects/*', 'POST'),
('p', 'role:admin', '/api/projects/*', 'PUT'),
('p', 'role:admin', '/api/projects/*', 'DELETE'),
('p', 'role:user', '/api/projects', 'GET'),
('p', 'role:user', '/api/projects', 'POST'),
('p', 'role:user', '/api/projects/*', 'GET'),
('p', 'role:user', '/api/projects/*', 'POST'),
('p', 'role:user', '/api/projects/*', 'PUT'),
('p', 'role:user', '/api/projects/*', 'DELETE'),

-- 站内消息
('p', 'role:admin', '/api/messages', 'GET'),
('p', 'role:admin', '/api/messages/*', 'GET'),
('p', 'role:admin', '/api/messages/*', 'PUT'),
('p', 'role:user', '/api/messages', 'GET'),
('p', 'role:user', '/api/messages/*', 'GET'),
('p', 'role:user', '/api/messages/*', 'PUT');

SET FOREIGN_KEY_CHECKS = 1;

-- ============================================
-- 脚本执行完成
-- ============================================
SELECT '数据库初始化完成' AS result;