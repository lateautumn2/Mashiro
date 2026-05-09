package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// CORS
	r.Use(cors.Default())

	api := r.Group("/api")
	{
		// Auth
		api.POST("/login", controllers.Login)

		// Agent Report (Uses custom header token)
		agent := api.Group("/agent")
		{
			agent.GET("/install.sh", controllers.AgentInstallShell)
			agent.GET("/install.ps1", controllers.AgentInstallPowerShell)
			agent.GET("/package.zip", controllers.AgentPackageArchive)
		}
		agent.Use(middleware.AgentAuthRequired())
		{
			agent.POST("/report", controllers.ReportAgentData)
			agent.GET("/latency/tasks", controllers.AgentGetLatencyTasks)
			agent.POST("/latency/results", controllers.AgentReportLatencyResults)
		}

		// Frontend Dashboard (Public or read-only)
		dashboard := api.Group("/dashboard")
		{
			dashboard.GET("/stats", controllers.GetDashboardStats)
			dashboard.GET("/servers", controllers.GetServersList)
		}

		// Admin API (Requires JWT)
		admin := api.Group("/admin")
		admin.Use(middleware.AuthRequired())
		{
			// Server Management
			admin.GET("/servers", controllers.AdminGetServers)
			admin.POST("/servers", controllers.AdminAddServer)
			admin.PUT("/servers/:id", controllers.AdminUpdateServer)
			admin.DELETE("/servers/:id", controllers.AdminDeleteServer)
			admin.GET("/servers/:id/command", controllers.AdminGetDeployCommand)
			admin.GET("/latency-tasks", controllers.AdminListLatencyTasks)
			admin.POST("/latency-tasks", controllers.AdminCreateLatencyTask)
			admin.PUT("/latency-tasks/:id", controllers.AdminUpdateLatencyTask)
			admin.DELETE("/latency-tasks/:id", controllers.AdminDeleteLatencyTask)

			// Config (TG Bot, Latency)
			admin.GET("/config", controllers.GetConfig)
			admin.POST("/config", controllers.UpdateConfig)

			// Auth
			admin.POST("/password", controllers.ChangePassword)
		}
	}

	return r
}
