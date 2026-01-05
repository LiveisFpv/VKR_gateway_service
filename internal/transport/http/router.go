package http

import (
	"VKR_gateway_service/internal/app"
	"VKR_gateway_service/internal/transport/http/handlers"

	"github.com/gin-gonic/gin"
)

func AIRouter(r *gin.RouterGroup, a *app.App) {
	r.POST("/paper/add", func(ctx *gin.Context) { handlers.PaperAdd(ctx, a) })
	// r.GET("/search/papers", func(ctx *gin.Context) { handlers.SearchPapers(ctx, a) })
}

func ChatRouter(r *gin.RouterGroup, a *app.App) {
	r.POST("", func(ctx *gin.Context) { handlers.CreateChat(ctx, a) })
	r.GET("", func(ctx *gin.Context) { handlers.GetUserChats(ctx, a) })
	r.GET("/:chat_id/history", func(ctx *gin.Context) { handlers.GetChatHistory(ctx, a) })
	r.POST("/:chat_id/history", func(ctx *gin.Context) { handlers.CreateChatHistory(ctx, a) })
}

func SSORouter(r *gin.RouterGroup, a *app.App) {

}
