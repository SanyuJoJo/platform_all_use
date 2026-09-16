package health
import (
	"github.com/gin-gonic/gin"
	"backend-go/internal/database"
	"backend-go/internal/response"
)
// Handler 返回 /health 与 Python 版一致的统一响应。
func Handler(c *gin.Context) {
	dbStatus := "unavailable"
	if database.HealthCheck() {
		dbStatus = "connected"
	}
	response.Success(c, response.HealthData{
		Status:   "ok",
		Database: dbStatus,
	}, "success")
}
