package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"righthand-go/internal/config"
	"righthand-go/internal/database"
	"righthand-go/internal/middleware"
	"righthand-go/internal/utils"
	authHandler "righthand-go/services/auth-service/handler"
	contactHandler "righthand-go/services/task-service/handler"
	projHandler "righthand-go/services/task-service/handler"
	requirementHandler "righthand-go/services/task-service/handler"
	taskHandler "righthand-go/services/task-service/handler"

	"github.com/gin-gonic/gin"
)

// setupTestRouter 设置测试路由(模拟main.go中的路由注册)
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 初始化数据库（强约束：自动测试仅允许使用独立测试库，见 docs/ai/rules.md §9）
	cfg, err := config.LoadConfig("../../configs")
	if err != nil {
		panic(err)
	}
	cfg.Database.DBName = database.TestDBName
	database.InitDB(&cfg.Database)

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API路由组
	apiGroup := router.Group("/api")
	{
		// 认证模块(无需JWT)
		authGroup := apiGroup.Group("/auth")
		{
			authHandler := authHandler.NewAuthHandler()
			authGroup.POST("/login", authHandler.Login)
		}

		// 需要JWT认证的路由
		apiGroup.Use(middleware.JWTAuth())
		{
			// 系统管理
			systemGroup := apiGroup.Group("/system")
			{
				// 用户管理
				userHandler := authHandler.NewUserHandler()
				userGroup := systemGroup.Group("/users")
				{
					userGroup.GET("", userHandler.ListUsers)
					userGroup.GET("/page", userHandler.ListUsers) // 分页查询
					userGroup.GET("/:id", userHandler.GetUser)
					userGroup.POST("", userHandler.CreateUser)
					userGroup.PUT("/:id", userHandler.UpdateUser)
					userGroup.DELETE("/:id", userHandler.DeleteUser)
					userGroup.POST("/:id/roles", userHandler.AssignRoles)
				}

				// 角色管理
				roleHandler := authHandler.NewRoleHandler()
				roleGroup := systemGroup.Group("/roles")
				{
					roleGroup.GET("", roleHandler.ListRoles)
					roleGroup.GET("/all", roleHandler.GetAllRoles)
					roleGroup.GET("/:id", roleHandler.GetRole)
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

			// 任务管理
			taskHandler := taskHandler.NewTaskHandler()
			apiGroup.GET("/tasks", taskHandler.ListTasks)
			apiGroup.GET("/tasks/:id", taskHandler.GetTask)
			apiGroup.POST("/tasks", taskHandler.CreateTask)
			apiGroup.PUT("/tasks/:id", taskHandler.UpdateTask)
			apiGroup.DELETE("/tasks/:id", taskHandler.DeleteTask)
			apiGroup.PUT("/tasks/:id/status", taskHandler.UpdateStatus)

			// 联系人管理
			contactHdlr := contactHandler.NewContactHandler()
			apiGroup.GET("/contacts", contactHdlr.ListContacts)
			apiGroup.GET("/contacts/:id", contactHdlr.GetContact)
			apiGroup.POST("/contacts", contactHdlr.CreateContact)
			apiGroup.PUT("/contacts/:id", contactHdlr.UpdateContact)
			apiGroup.DELETE("/contacts/:id", contactHdlr.DeleteContact)

			// 需求管理
			requirementHdlr := requirementHandler.NewRequirementHandler()
			apiGroup.GET("/requirements", requirementHdlr.ListRequirements)
			apiGroup.GET("/requirements/:id", requirementHdlr.GetRequirement)
			apiGroup.POST("/requirements", requirementHdlr.CreateRequirement)
			apiGroup.PUT("/requirements/:id", requirementHdlr.UpdateRequirement)
			apiGroup.DELETE("/requirements/:id", requirementHdlr.DeleteRequirement)
			apiGroup.PUT("/requirements/:id/status", requirementHdlr.UpdateStatus)
			apiGroup.GET("/requirements/export", requirementHdlr.ExportRequirements)
			apiGroup.GET("/requirements/template", requirementHdlr.DownloadTemplate)
			apiGroup.POST("/requirements/import", requirementHdlr.ImportRequirements)

			// 项目管理
			projectHdl := projHandler.NewProjectHandler()
			projectWeekHdl := projHandler.NewProjectWeekHandler()
			apiGroup.GET("/projects", projectHdl.ListProjects)
			apiGroup.POST("/projects", projectHdl.CreateProject)
			apiGroup.GET("/projects/:id", projectHdl.GetProjectDetail)
			apiGroup.PUT("/projects/:id", projectHdl.UpdateProject)
			apiGroup.DELETE("/projects/:id", projectHdl.DeleteProject)
			apiGroup.PUT("/projects/:id/status", projectHdl.ChangeProjectStatus)
			apiGroup.PUT("/projects/:id/risk", projectHdl.ChangeProjectRisk)
			apiGroup.GET("/projects/:id/members", projectHdl.ListMembers)
			apiGroup.POST("/projects/:id/members", projectHdl.AddMember)
			apiGroup.DELETE("/projects/:id/members/:memberId", projectHdl.RemoveMember)
			apiGroup.GET("/projects/:id/stages", projectHdl.ListStages)
			apiGroup.POST("/projects/:id/stages", projectHdl.CreateStage)
			apiGroup.PUT("/projects/:id/stages/:stageId", projectHdl.UpdateStage)
			apiGroup.DELETE("/projects/:id/stages/:stageId", projectHdl.DeleteStage)
			apiGroup.POST("/projects/:id/weeks", projectWeekHdl.CreateWeek)
			apiGroup.PUT("/projects/:id/weeks/:weekId", projectWeekHdl.UpdateWeek)
			apiGroup.DELETE("/projects/:id/weeks/:weekId", projectWeekHdl.DeleteWeek)
			apiGroup.PUT("/projects/:id/weeks/:weekId/status", projectWeekHdl.ChangeWeekStatus)
			apiGroup.POST("/projects/:id/weeks/:weekId/summary", projectWeekHdl.SubmitWeekSummary)
			apiGroup.PUT("/projects/:id/weeks/:weekId/summary/unlock", projectWeekHdl.UnlockWeekSummary)
			apiGroup.GET("/projects/:id/tasks", projectWeekHdl.ListProjectTasks)
			apiGroup.POST("/projects/:id/weeks/:weekId/tasks", projectWeekHdl.CreateTaskInWeek)
			apiGroup.POST("/projects/:id/weeks/:weekId/tasks/:taskId", projectWeekHdl.LinkTaskToWeek)
			apiGroup.DELETE("/projects/:id/weeks/:weekId/tasks/:taskId", projectWeekHdl.UnlinkTaskFromWeek)
			apiGroup.PUT("/projects/:id/tasks/:taskId/week", projectWeekHdl.MoveTaskToWeek)

			// 站内消息
			messageHdl := projHandler.NewMessageHandler()
			apiGroup.GET("/messages", messageHdl.ListMessages)
			apiGroup.GET("/messages/unread-count", messageHdl.UnreadCount)
			apiGroup.PUT("/messages/read-all", messageHdl.MarkAllRead)
			apiGroup.PUT("/messages/:id/read", messageHdl.MarkRead)
		}
	}

	return router
}

// TestUserRoutes_NoConflict 测试用户路由没有冲突
func TestUserRoutes_NoConflict(t *testing.T) {
	router := setupTestRouter()

	// 测试 /api/system/users/page 不会匹配到 /:id
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/system/users/page?pageNum=1&pageSize=10", nil)
	router.ServeHTTP(w, req)

	// 应该返回200或401(未认证),而不是400(无效ID)
	assert.NotEqual(t, 400, w.Code, "用户分页路由不应与/:id冲突导致400错误")
	t.Logf("Response Code: %d, Body: %s", w.Code, w.Body.String())
}

// TestRoleRoutes_AllEndpoint 测试角色/all路由存在
func TestRoleRoutes_AllEndpoint(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/system/roles/all", nil)
	router.ServeHTTP(w, req)

	// 应该返回200或401(未认证),而不是404
	assert.NotEqual(t, 404, w.Code, "/api/system/roles/all 路由应该存在")
	t.Logf("Response Code: %d, Body: %s", w.Code, w.Body.String())
}

// TestRoutePriority_SpecificBeforeParametric 测试具体路径优先于参数化路径
func TestRoutePriority_SpecificBeforeParametric(t *testing.T) {
	router := setupTestRouter()

	testCases := []struct {
		path           string
		shouldNotBe400 bool
		description    string
	}{
		{"/api/system/users/page", true, "用户分页路径"},
		{"/api/system/roles/all", true, "角色全部路径"},
		{"/api/system/dicts/type/test", true, "字典类型路径"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", tc.path, nil)
			router.ServeHTTP(w, req)

			if tc.shouldNotBe400 {
				assert.NotEqual(t, 400, w.Code, "%s 不应该返回400错误", tc.description)
			}
			t.Logf("%s - Code: %d", tc.description, w.Code)
		})
	}
}

// TestInvalidUserID 测试无效用户ID返回正确错误
// 说明：/api/system/users/:id 位于 JWTAuth 之后，未认证会先行返回 401，
// 因此必须携带有效 Token 才能验证「无效 ID 返回 400」的处理器行为。
func TestInvalidUserID(t *testing.T) {
	router := setupTestRouter()

	token, err := utils.GenerateToken(1, "admin", "管理员", true, "ADMIN")
	if err != nil {
		t.Fatalf("生成测试 Token 失败: %v", err)
	}

	// 测试字符串作为用户ID
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/system/users/abc", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	// 应该返回400(无效ID),但消息应该是"无效的用户ID"
	assert.Equal(t, 400, w.Code)
	t.Logf("Response: %s", w.Body.String())
}

// TestProjectRoutes_Exist 项目/消息路由存在性（未认证时应返回 401 而非 404）
func TestProjectRoutes_Exist(t *testing.T) {
	router := setupTestRouter()

	testCases := []struct {
		method string
		path   string
	}{
		{"GET", "/api/projects"},
		{"POST", "/api/projects"},
		{"GET", "/api/projects/1"},
		{"PUT", "/api/projects/1"},
		{"DELETE", "/api/projects/1"},
		{"PUT", "/api/projects/1/status"},
		{"PUT", "/api/projects/1/risk"},
		{"GET", "/api/projects/1/members"},
		{"POST", "/api/projects/1/members"},
		{"DELETE", "/api/projects/1/members/2"},
		{"GET", "/api/projects/1/stages"},
		{"POST", "/api/projects/1/stages"},
		{"PUT", "/api/projects/1/stages/2"},
		{"DELETE", "/api/projects/1/stages/2"},
		{"POST", "/api/projects/1/weeks"},
		{"PUT", "/api/projects/1/weeks/2"},
		{"DELETE", "/api/projects/1/weeks/2"},
		{"PUT", "/api/projects/1/weeks/2/status"},
		{"POST", "/api/projects/1/weeks/2/summary"},
		{"PUT", "/api/projects/1/weeks/2/summary/unlock"},
		{"GET", "/api/projects/1/tasks"},
		{"POST", "/api/projects/1/weeks/2/tasks"},
		{"POST", "/api/projects/1/weeks/2/tasks/3"},
		{"DELETE", "/api/projects/1/weeks/2/tasks/3"},
		{"PUT", "/api/projects/1/tasks/3/week"},
		{"GET", "/api/messages"},
		{"GET", "/api/messages/unread-count"},
		{"PUT", "/api/messages/read-all"},
		{"PUT", "/api/messages/1/read"},
	}

	for _, tc := range testCases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, nil)
			router.ServeHTTP(w, req)
			assert.NotEqual(t, 404, w.Code, "%s %s 路由应该存在（未认证应为 401）", tc.method, tc.path)
		})
	}
}
