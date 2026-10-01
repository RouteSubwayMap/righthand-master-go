package main

import (
	"fmt"
	"log"
	"righthand-go/internal/config"
	"righthand-go/internal/database"
	"righthand-go/internal/logger"
	"righthand-go/internal/middleware"
	"righthand-go/internal/scheduler"
	authHandler "righthand-go/services/auth-service/handler"
	taskHandler "righthand-go/services/task-service/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志（控制台 + 按日期切割文件；失败仅告警并降级为控制台输出）
	if err := logger.Init(&cfg.Log); err != nil {
		log.Printf("警告: 日志初始化失败，降级为控制台输出: %v", err)
	}
	defer logger.Close()

	// 初始化数据库
	if err := database.InitDB(&cfg.Database); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 初始化Casbin
	if err := middleware.InitCasbin(database.GetDB()); err != nil {
		log.Printf("警告: Casbin初始化失败: %v", err)
		log.Println("说明: 已知问题，当前不影响功能——CasbinMiddleware 尚未挂载到路由，权限控制由 JWT 认证 + 业务层校验 + 前端权限码承担")
		log.Println("      常见原因: 部署目录 configs/ 缺少 model.conf（发布包未收集该文件）或程序工作目录不正确；正式启用 Casbin 前需补齐该文件并校准策略")
	} else {
		log.Println("Casbin权限控制初始化成功")
	}

	// 启动周期性任务调度器
	recurrenceScheduler := scheduler.NewRecurrenceScheduler(database.GetDB())
	recurrenceScheduler.Start()
	defer recurrenceScheduler.Stop()

	// 启动项目自动化调度器（M3：每日/每周五 01:00）
	projectScheduler := scheduler.NewProjectScheduler(database.GetDB())
	projectScheduler.Start()
	defer projectScheduler.Stop()

	// 创建Gin引擎（中间件链顺序固定：CORS → Recovery → Logger）
	r := gin.New()
	r.Use(middleware.CORS())
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())

	// 手动触发定时任务的接口（用于测试）
	r.GET("/api/scheduler/trigger", func(c *gin.Context) {
		recurrenceScheduler.ManualTrigger()
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "周期性任务已手动触发，请查看服务器日志",
		})
	})

	// 手动触发项目自动化任务（用于测试/联调）
	r.GET("/api/scheduler/project/trigger", func(c *gin.Context) {
		projectScheduler.ManualTrigger()
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "项目自动化任务已手动触发，请查看服务器日志",
		})
	})

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "righthand-server",
		})
	})

	// ==================== 认证模块路由 ====================
	authGroup := r.Group("/api/auth")
	{
		handler := authHandler.NewAuthHandler()

		// 公开路由
		authGroup.POST("/login", handler.Login)

		// 需要认证的路由
		protected := authGroup.Use(middleware.JWTAuth())
		{
			protected.GET("/info", handler.GetInfo)
			protected.GET("/menus", handler.GetMenus)
			protected.POST("/logout", handler.Logout)
		}
	}

	// ==================== 系统管理路由 ====================
	systemGroup := r.Group("/api/system")
	systemGroup.Use(middleware.JWTAuth())
	{
		// 用户管理
		userHandler := authHandler.NewUserHandler()
		userGroup := systemGroup.Group("/users")
		{
			userGroup.GET("", userHandler.ListUsers)
			userGroup.GET("/page", userHandler.ListUsers) // 分页查询（与GET ""相同）
			userGroup.GET("/:id", userHandler.GetUser)
			userGroup.GET("/:id/roles", userHandler.GetUserRoles) // 获取用户角色列表
			userGroup.POST("", userHandler.CreateUser)
			userGroup.PUT("/:id", userHandler.UpdateUser)
			userGroup.PUT("/:id/reset-password", userHandler.ResetPassword) // 重置密码
			userGroup.POST("/change-password", userHandler.ChangePassword)  // 修改密码
			userGroup.DELETE("/:id", userHandler.DeleteUser)
			userGroup.POST("/:id/roles", userHandler.AssignRoles)
		}

		// 角色管理
		roleHandler := authHandler.NewRoleHandler()
		roleGroup := systemGroup.Group("/roles")
		{
			roleGroup.GET("", roleHandler.ListRoles)
			roleGroup.GET("/all", roleHandler.GetAllRoles) // 获取所有角色（不分页）
			roleGroup.GET("/:id", roleHandler.GetRole)
			roleGroup.GET("/:id/menus", roleHandler.GetRoleMenus) // 获取角色菜单列表
			roleGroup.POST("", roleHandler.CreateRole)
			roleGroup.PUT("/:id", roleHandler.UpdateRole)
			roleGroup.DELETE("/:id", roleHandler.DeleteRole)
			roleGroup.POST("/:id/menus", roleHandler.AssignMenus)
		}

		// 菜单管理
		menuHandler := authHandler.NewMenuHandler()
		menuGroup := systemGroup.Group("/menus")
		{
			menuGroup.GET("", menuHandler.GetMenus)
			menuGroup.GET("/tree", menuHandler.GetMenuTree) // 获取菜单树形结构
			menuGroup.GET("/:id", menuHandler.GetMenu)
			menuGroup.POST("", menuHandler.CreateMenu)
			menuGroup.PUT("/:id", menuHandler.UpdateMenu)
			menuGroup.DELETE("/:id", menuHandler.DeleteMenu)
		}

		// 字典管理
		dictHandler := authHandler.NewDictHandler()
		dictGroup := systemGroup.Group("/dicts")
		{
			dictGroup.GET("", dictHandler.ListDicts)
			dictGroup.GET("/type/:type", dictHandler.GetDictsByType)
			dictGroup.GET("/:id", dictHandler.GetDict)
			dictGroup.POST("", dictHandler.CreateDict)
			dictGroup.PUT("/:id", dictHandler.UpdateDict)
			dictGroup.DELETE("/:id", dictHandler.DeleteDict)
		}
	}

	// ==================== 统一事项模块路由 ====================
	tasksGroup := r.Group("/api/tasks")
	tasksGroup.Use(middleware.JWTAuth())
	{
		// 事项管理
		taskHdl := taskHandler.NewTaskHandler()
		{
			tasksGroup.GET("", taskHdl.ListTasks)
			tasksGroup.GET("/:id", taskHdl.GetTask)
			tasksGroup.POST("", taskHdl.CreateTask)
			tasksGroup.PUT("/:id", taskHdl.UpdateTask)
			tasksGroup.DELETE("/:id", taskHdl.DeleteTask)
			tasksGroup.PUT("/:id/status", taskHdl.UpdateStatus)
			tasksGroup.PUT("/:id/transfer", taskHdl.TransferTask)
		}

		// 联系人管理
		contactHdl := taskHandler.NewContactHandler()
		contactsGroup := r.Group("/api/contacts")
		contactsGroup.Use(middleware.JWTAuth())
		{
			contactsGroup.GET("", contactHdl.ListContacts)
			contactsGroup.GET("/:id", contactHdl.GetContact)
			contactsGroup.POST("", contactHdl.CreateContact)
			contactsGroup.PUT("/:id", contactHdl.UpdateContact)
			contactsGroup.DELETE("/:id", contactHdl.DeleteContact)
			contactsGroup.GET("/groups", contactHdl.GetContactGroups)
		}

		// 批量任务清单
		batchTaskHdl := taskHandler.NewBatchTaskHandler()
		batchTaskGroup := r.Group("/api/batch-tasks")
		batchTaskGroup.Use(middleware.JWTAuth())
		{
			batchTaskGroup.GET("", batchTaskHdl.ListBatchTasks)
			batchTaskGroup.GET("/:id", batchTaskHdl.GetBatchTask)
			batchTaskGroup.POST("", batchTaskHdl.CreateBatchTask)
			batchTaskGroup.POST("/import", batchTaskHdl.BatchImportTasks)
			batchTaskGroup.PUT("/:id", batchTaskHdl.UpdateBatchTask)
			batchTaskGroup.DELETE("/:id", batchTaskHdl.DeleteBatchTask)
		}

		// 生产环境节点清单
		prodNodeHdl := taskHandler.NewProdNodeHandler()
		prodNodeGroup := r.Group("/api/prod-nodes")
		prodNodeGroup.Use(middleware.JWTAuth())
		{
			prodNodeGroup.GET("", prodNodeHdl.ListProdNodes)
			prodNodeGroup.GET("/:id", prodNodeHdl.GetProdNode)
			prodNodeGroup.POST("", prodNodeHdl.CreateProdNode)
			prodNodeGroup.POST("/import", prodNodeHdl.BatchImportNodes)
			prodNodeGroup.PUT("/:id", prodNodeHdl.UpdateProdNode)
			prodNodeGroup.DELETE("/:id", prodNodeHdl.DeleteProdNode)
		}

		// 银行机构参数管理
		bankParamHdl := taskHandler.NewBankParamHandler()
		bankParamGroup := r.Group("/api/bank-params")
		bankParamGroup.Use(middleware.JWTAuth())
		{
			bankParamGroup.GET("", bankParamHdl.ListBankParams)
			bankParamGroup.GET("/:id", bankParamHdl.GetBankParam)
			bankParamGroup.POST("", bankParamHdl.CreateBankParam)
			bankParamGroup.POST("/import", bankParamHdl.BatchImportParams)
			bankParamGroup.PUT("/:id", bankParamHdl.UpdateBankParam)
			bankParamGroup.DELETE("/:id", bankParamHdl.DeleteBankParam)
		}

		// 仪表盘统计
		dashboardGroup := r.Group("/api/dashboard")
		dashboardGroup.Use(middleware.JWTAuth())
		{
			dashboardGroup.GET("/stats", taskHdl.GetDashboardStats)
		}

		// 需求管理
		requirementHdl := taskHandler.NewRequirementHandler()
		requirementsGroup := r.Group("/api/requirements")
		requirementsGroup.Use(middleware.JWTAuth())
		{
			requirementsGroup.GET("", requirementHdl.ListRequirements)
			requirementsGroup.GET("/:id", requirementHdl.GetRequirement)
			requirementsGroup.GET("/:id/tasks", requirementHdl.GetRequirementTasks)
			requirementsGroup.POST("", requirementHdl.CreateRequirement)
			requirementsGroup.PUT("/:id", requirementHdl.UpdateRequirement)
			requirementsGroup.DELETE("/:id", requirementHdl.DeleteRequirement)
			requirementsGroup.PUT("/:id/status", requirementHdl.UpdateStatus)
			requirementsGroup.GET("/export", requirementHdl.ExportRequirements)
			requirementsGroup.GET("/template", requirementHdl.DownloadTemplate)
			requirementsGroup.POST("/import", requirementHdl.ImportRequirements)
		}
	}

	// ==================== 项目管理模块路由 ====================
	projectGroup := r.Group("/api/projects")
	projectGroup.Use(middleware.JWTAuth())
	{
		projectHdl := taskHandler.NewProjectHandler()
		projectWeekHdl := taskHandler.NewProjectWeekHandler()

		// 项目
		projectGroup.GET("", projectHdl.ListProjects)
		projectGroup.POST("", projectHdl.CreateProject)
		projectGroup.GET("/:id", projectHdl.GetProjectDetail)
		projectGroup.PUT("/:id", projectHdl.UpdateProject)
		projectGroup.DELETE("/:id", projectHdl.DeleteProject)
		projectGroup.PUT("/:id/status", projectHdl.ChangeProjectStatus)
		projectGroup.PUT("/:id/risk", projectHdl.ChangeProjectRisk)

		// 成员
		projectGroup.GET("/:id/members", projectHdl.ListMembers)
		projectGroup.POST("/:id/members", projectHdl.AddMember)
		projectGroup.DELETE("/:id/members/:memberId", projectHdl.RemoveMember)

		// 项目阶段
		projectGroup.GET("/:id/stages", projectHdl.ListStages)
		projectGroup.POST("/:id/stages", projectHdl.CreateStage)
		projectGroup.PUT("/:id/stages/:stageId", projectHdl.UpdateStage)
		projectGroup.DELETE("/:id/stages/:stageId", projectHdl.DeleteStage)

		// 周计划
		projectGroup.POST("/:id/weeks", projectWeekHdl.CreateWeek)
		projectGroup.PUT("/:id/weeks/:weekId", projectWeekHdl.UpdateWeek)
		projectGroup.DELETE("/:id/weeks/:weekId", projectWeekHdl.DeleteWeek)
		projectGroup.PUT("/:id/weeks/:weekId/status", projectWeekHdl.ChangeWeekStatus)
		projectGroup.POST("/:id/weeks/:weekId/summary", projectWeekHdl.SubmitWeekSummary)
		projectGroup.PUT("/:id/weeks/:weekId/summary/unlock", projectWeekHdl.UnlockWeekSummary)

		// 任务-周关联
		projectGroup.GET("/:id/tasks", projectWeekHdl.ListProjectTasks)
		projectGroup.POST("/:id/weeks/:weekId/tasks", projectWeekHdl.CreateTaskInWeek)
		projectGroup.POST("/:id/weeks/:weekId/tasks/:taskId", projectWeekHdl.LinkTaskToWeek)
		projectGroup.DELETE("/:id/weeks/:weekId/tasks/:taskId", projectWeekHdl.UnlinkTaskFromWeek)
		projectGroup.PUT("/:id/tasks/:taskId/week", projectWeekHdl.MoveTaskToWeek)
	}

	// ==================== 站内消息路由 ====================
	messagesGroup := r.Group("/api/messages")
	messagesGroup.Use(middleware.JWTAuth())
	{
		messageHdl := taskHandler.NewMessageHandler()
		messagesGroup.GET("", messageHdl.ListMessages)
		messagesGroup.GET("/unread-count", messageHdl.UnreadCount)
		messagesGroup.PUT("/read-all", messageHdl.MarkAllRead)
		messagesGroup.PUT("/:id/read", messageHdl.MarkRead)
	}

	// 启动服务器
	port := cfg.Server.Port
	addr := fmt.Sprintf(":%d", port)

	log.Printf("RightHand服务启动在端口 %d", port)
	log.Printf("访问地址: http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("启动服务失败: %v", err)
	}
}
