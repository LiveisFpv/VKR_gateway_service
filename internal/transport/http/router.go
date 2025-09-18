package http

import (
	"VKR_gateway_service/internal/app"
	"VKR_gateway_service/internal/transport/http/handlers"

	"github.com/gin-gonic/gin"
)

func AIRouter(r *gin.RouterGroup, a *app.App) {
	r.POST("/paper/add", func(ctx *gin.Context) { handlers.PaperAdd(ctx, a) })
	r.GET("/search/papers", func(ctx *gin.Context) { handlers.SearchPapers(ctx, a) })
}

func SSORouter(r *gin.RouterGroup, a *app.App) {

}
