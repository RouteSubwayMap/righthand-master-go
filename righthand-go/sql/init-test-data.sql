-- ============================================
-- RightHand 系统 - 测试数据初始化脚本
-- 说明：此文件包含示例测试数据，用于开发和测试环境
-- 注意：生产环境请勿执行此脚本
-- ============================================

USE righthand;

-- ============================================
-- 1. 需求管理测试数据
-- ============================================

INSERT INTO requirement (
    title, bank_name, requirement_no, priority, risk_level, requirement_type, 
    status, schedule_date, main_system, developer, involved_systems,
    follower_id, follower_name, submitter_id, submitter_name, submit_time,
    completion_date, estimated_workload, group_type, description, remark, user_id
) VALUES
('个人贷款流程优化', '工商银行', 'REQ-2024-001', 'HIGH', 'MEDIUM', 'INTERNAL', 
 'DEVELOPING', '2024-03-15', '信贷核心系统', '张三', '信贷系统,风控系统',
 1, '管理员', 1, '管理员', NOW(), NULL, 10.50, 'LOAN_MIDDLE', 
 '优化个人贷款申请流程，减少用户等待时间，提升用户体验。包括前端界面优化、后端接口性能提升、数据库查询优化等。', '高优先级需求', 1),

('风险控制模型升级', '建设银行', 'REQ-2024-002', 'HIGH', 'HIGH', 'BANK', 
 'PENDING_SCHEDULING', '2024-04-01', '风控引擎', '李四', '风控系统,决策引擎',
 1, '管理员', 1, '管理员', NOW(), NULL, 20.00, 'RISK', 
 '升级风险控制模型，提高准确率和响应速度。引入机器学习算法，实现实时风险评估。', '重要需求，需要与业务部门密切配合', 1),

('移动端UI优化', '农业银行', 'REQ-2024-003', 'LOW', 'LOW', 'INTERNAL', 
 'COMPLETED', '2024-02-20', '移动APP', '王五', '前端系统',
 1, '管理员', 1, '管理员', '2024-01-15 10:00:00', '2024-02-20 18:00:00', 5.00, 'LOAN_PRE', 
 '优化移动端界面用户体验，包括首页布局调整、交互流程优化、动画效果增强等。', '已完成并通过验收', 1),

('企业网银批量转账功能', '中国银行', 'REQ-2024-004', 'MEDIUM', 'MEDIUM', 'BANK', 
 'TESTING', '2024-03-30', '企业网银系统', '赵六', '企业网银,支付系统',
 1, '管理员', 1, '管理员', '2024-02-01 09:00:00', NULL, 15.00, 'LOAN_MIDDLE', 
 '为企业客户提供批量转账功能，支持Excel导入、模板下载、转账结果导出等功能。', '客户需求强烈', 1),

('智能客服系统集成', '交通银行', 'REQ-2024-005', 'MEDIUM', 'LOW', 'INTERNAL', 
 'PENDING_CLARIFY', NULL, '客服系统', '孙七', '客服系统,AI平台',
 1, '管理员', 1, '管理员', '2024-03-01 14:00:00', NULL, 25.00, 'RISK', 
 '集成智能客服系统，实现7x24小时自动应答，降低人工客服压力。需要对接NLP平台和知识库系统。', '待与AI团队确认技术方案', 1);

-- ============================================
-- 2. 统一事项测试数据（含周期性任务）
-- ============================================

-- 普通任务
INSERT INTO task_item (user_id, title, description, item_type, category, priority, tags, start_time, deadline, status_code, progress) VALUES
(1, '完成项目需求文档', '编写并评审项目需求文档，确保需求清晰完整', 'TASK', '工作', 'HIGH', '文档,需求', '2026-05-20 09:00:00', '2026-05-25 18:00:00', 'DOING', 60),
(1, '修复登录bug', '修复用户登录时偶尔出现的验证码失效问题', 'TASK', '工作', 'HIGH', 'bug,紧急', '2026-05-21 10:00:00', '2026-05-22 12:00:00', 'TODO', 0),
(1, '代码重构优化', '对用户模块代码进行重构，提高代码质量和可维护性', 'TASK', '工作', 'MEDIUM', '重构,优化', '2026-05-22 09:00:00', '2026-05-30 18:00:00', 'TODO', 0),
(1, '学习Vue3新特性', '学习Vue3 Composition API和响应式原理', 'TASK', '学习', 'LOW', '学习,前端', NULL, '2026-06-15 23:59:59', 'DONE', 100);

-- 周期性任务示例
INSERT INTO task_item (user_id, title, description, item_type, category, priority, recurrence_type, recurrence_rule, recurrence_end_type, recurrence_count, start_time, deadline, status_code) VALUES
(1, '团队周会', '参加团队周会，讨论项目进度和问题', 'TASK', '工作', 'HIGH', 
 'WEEKLY', '{"weekDays": [1]}', 'COUNT', 52, 
 '2026-05-20 10:00:00', '2026-05-20 11:00:00', 'TODO'),

(1, '月度工作总结', '编写月度工作总结报告，汇报工作成果和下月计划', 'TASK', '工作', 'MEDIUM', 
 'MONTHLY', '{"dayOfMonth": 1}', 'NEVER', 10, 
 '2026-06-01 09:00:00', '2026-06-01 18:00:00', 'TODO'),

(1, '年度体检', '参加公司组织的年度健康体检', 'MEMO', '生活', 'HIGH', 
 'YEARLY', '{"month": 6, "dayOfMonth": 15}', 'DATE', 1, 
 '2026-06-15 08:00:00', '2026-06-15 12:00:00', 'TODO');

-- 备忘录示例
INSERT INTO task_item (user_id, title, description, item_type, category, priority, tags, deadline, status_code) VALUES
(1, '购买办公用品', '采购打印机墨盒、A4纸等办公用品', 'MEMO', '工作', 'MEDIUM', '采购,办公', '2026-05-23 17:00:00', 'TODO'),
(1, '周末爬山活动', '组织团队周末爬山活动，增进团队凝聚力', 'MEMO', '生活', 'LOW', '团建,运动', '2026-05-25 08:00:00', 'TODO');

-- ============================================
-- 3. 联系人测试数据
-- ============================================

INSERT INTO contact (user_id, name, gender, phone, email, contact_group, company, position, remark) VALUES
(1, '张三', 1, '13800138001', 'zhangsan@example.com', 'R&D', '某某科技公司', '高级开发工程师', '负责后端开发'),
(1, '李四', 0, '13800138002', 'lisi@example.com', 'PRODUCT', '某某科技公司', '产品经理', '产品需求对接人'),
(1, '王五', 1, '13800138003', 'wangwu@example.com', 'ACCOUNT_MANAGER', '某某银行', '客户经理', '银行业务对接人'),
(1, '赵六', 1, '13800138004', 'zhaoliu@example.com', 'R&D', '某某云服务公司', '架构师', '技术咨询顾问');

-- ============================================
-- 4. 站内消息测试数据
-- ============================================

INSERT INTO sys_message (user_id, title, content, type, is_read, business_id, create_time) VALUES
(1, '任务分配通知', '您有一个新的任务：完成项目需求文档，请及时处理', 'TASK', 0, 1, NOW()),
(1, '系统维护通知', '系统将于今晚22:00-23:00进行维护，期间可能无法访问', 'SYSTEM', 1, NULL, DATE_SUB(NOW(), INTERVAL 1 DAY)),
(1, '任务截止提醒', '您的任务"修复登录bug"将在明天截止，请尽快完成', 'REMINDER', 0, 2, NOW());

-- ============================================
-- 5. 项目管理演示数据（日期基于当前周动态计算，保证与运行日相关）
-- ============================================

-- 本周一（与前端 Mock 种子口径一致）
SET @monday = DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY);

-- 5.1 演示项目：1 个进行中（含已延期周）+ 1 个启动
INSERT INTO project (project_name, project_key, description, plan_desc, manager_id, start_date, end_date, status, risk_level, delivery_status, risk_note, progress_pct, week_count, done_week_count, create_time, create_user) VALUES
('新一代信贷核心系统对接', 'NCBC', '与新一代信贷核心系统完成放款、还款、额度三大模块的接口对接与联调。', '第1-2周需求调研与接口设计；第3-4周接口开发与联调；第5周验收测试。', 1,
 DATE_SUB(@monday, INTERVAL 21 DAY), DATE_ADD(@monday, INTERVAL 11 DAY), 'IN_PROGRESS', 'MEDIUM', 'NORMAL', '接口方联调环境不稳定，存在延期风险', 40, 5, 2, NOW(), 1),
('数据中台二期建设', 'DATA-MID-2', '数据中台二期：指标平台、数据质量监控、数据资产目录。', '待推进后逐周排期。', 2,
 DATE_ADD(@monday, INTERVAL 7 DAY), DATE_ADD(@monday, INTERVAL 42 DAY), 'STARTING', 'LOW', 'NORMAL', '', 0, 0, 0, NOW(), 2);

SET @p1 = (SELECT id FROM project WHERE project_key = 'NCBC');
SET @p2 = (SELECT id FROM project WHERE project_key = 'DATA-MID-2');

-- 5.2 项目成员（PM 自动加入）
INSERT INTO project_member (project_id, user_id, member_role, create_time, create_user) VALUES
(@p1, 1, 'PM', NOW(), 1),
(@p1, 2, 'MEMBER', NOW(), 1),
(@p2, 2, 'PM', NOW(), 2);

-- 5.3 周计划（项目1：5 周，前 2 周已完成并锁定总结；第 3 周已延期）
INSERT INTO project_weekly_plan (project_id, week_no, title, plan_desc, start_date, end_date, owner_id, status, summary, summary_by, summary_at, sort_order, create_time, create_user) VALUES
(@p1, 1, '需求调研与接口设计', '梳理三大模块接口清单，输出接口设计文档。', DATE_SUB(@monday, INTERVAL 21 DAY), DATE_SUB(@monday, INTERVAL 17 DAY), 1, 'DONE', '{"finished":"完成放款、还款模块接口清单梳理，输出设计文档 v1.0。","nextPlan":"启动额度模块接口设计。","risks":"接口方需求变更频繁，已锁定基线。","support":"需要产品经理确认额度规则细节。"}', 1, DATE_SUB(@monday, INTERVAL 17 DAY), 1, NOW(), 1),
(@p1, 2, '额度模块设计与评审', '完成额度模块接口设计与内部评审。', DATE_SUB(@monday, INTERVAL 14 DAY), DATE_SUB(@monday, INTERVAL 10 DAY), 1, 'DONE', '{"finished":"额度模块接口设计通过评审，进入开发。","nextPlan":"开始放款模块接口开发。","risks":"暂无。","support":"无。"}', 1, DATE_SUB(@monday, INTERVAL 10 DAY), 2, NOW(), 1),
(@p1, 3, '放款模块开发联调', '放款模块接口开发完成并联调通过。', DATE_SUB(@monday, INTERVAL 7 DAY), DATE_SUB(@monday, INTERVAL 3 DAY), 2, 'IN_PROGRESS', NULL, NULL, NULL, 3, NOW(), 1),
(@p1, 4, '还款模块开发联调', '还款模块接口开发与联调。', @monday, DATE_ADD(@monday, INTERVAL 4 DAY), 1, 'IN_PROGRESS', NULL, NULL, NULL, 4, NOW(), 1),
(@p1, 5, '整体验收与上线', '全链路验收测试，灰度上线。', DATE_ADD(@monday, INTERVAL 7 DAY), DATE_ADD(@monday, INTERVAL 11 DAY), 1, 'NOT_STARTED', NULL, NULL, NULL, 5, NOW(), 1),
(@p2, 1, '需求梳理与指标定义', '梳理二期范围，定义核心指标口径。', DATE_ADD(@monday, INTERVAL 7 DAY), DATE_ADD(@monday, INTERVAL 11 DAY), 2, 'NOT_STARTED', NULL, NULL, NULL, 1, NOW(), 2);

-- 5.4 任务-周关联（事项 2 跨第1、2周演示多周标签；事项 1/3 分别挂第4、5周）
INSERT INTO task_week_plan_rel (task_id, week_plan_id, create_time, create_user)
SELECT 2, id, NOW(), 1 FROM project_weekly_plan WHERE project_id = @p1 AND week_no = 1
UNION ALL
SELECT 2, id, NOW(), 1 FROM project_weekly_plan WHERE project_id = @p1 AND week_no = 2
UNION ALL
SELECT 1, id, NOW(), 1 FROM project_weekly_plan WHERE project_id = @p1 AND week_no = 4
UNION ALL
SELECT 3, id, NOW(), 1 FROM project_weekly_plan WHERE project_id = @p1 AND week_no = 5;

-- 5.5 项目变更日志
INSERT INTO project_change_log (project_id, action, detail, operator_id, create_time) VALUES
(@p1, 'CREATE', '创建项目「新一代信贷核心系统对接」', 1, DATE_SUB(@monday, INTERVAL 21 DAY)),
(@p1, 'STATUS', '项目状态：启动 → 进行中（开始）', 1, DATE_SUB(@monday, INTERVAL 21 DAY)),
(@p2, 'CREATE', '创建项目「数据中台二期建设」', 2, DATE_SUB(@monday, INTERVAL 2 DAY));

-- 5.6 项目站内消息（周总结提醒 / 周计划延期提醒）
INSERT INTO sys_message (user_id, title, content, type, is_read, business_id, create_time) VALUES
(1, '每周总结提醒', CONCAT('「新一代信贷核心系统对接」第4周（', DATE_FORMAT(@monday, '%Y-%m-%d'), ' ~ ', DATE_FORMAT(DATE_ADD(@monday, INTERVAL 4 DAY), '%Y-%m-%d'), '）尚未提交周总结，请及时填写。'), 'PROJECT', 0, @p1, NOW()),
(2, '周计划延期提醒', '「新一代信贷核心系统对接」第3周「放款模块开发联调」已延期，请尽快跟进。', 'PROJECT', 0, @p1, DATE_SUB(NOW(), INTERVAL 1 DAY));

-- ============================================
-- 验证测试数据
-- ============================================

SELECT '✅ 测试数据初始化完成！' AS message;
SELECT COUNT(*) AS requirement_count FROM requirement;
SELECT COUNT(*) AS task_item_count FROM task_item;
SELECT COUNT(*) AS contact_count FROM contact;
SELECT COUNT(*) AS message_count FROM sys_message;
SELECT COUNT(*) AS project_count FROM project;
SELECT COUNT(*) AS project_weekly_plan_count FROM project_weekly_plan;
SELECT COUNT(*) AS task_week_plan_rel_count FROM task_week_plan_rel;
