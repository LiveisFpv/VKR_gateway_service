package handlers

import (
	"VKR_gateway_service/internal/app"
	"VKR_gateway_service/internal/transport/http/presenters"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SearchPapers(ctx *gin.Context, a *app.App) {
	ctx.JSON(http.StatusNotImplemented, presenters.Error(fmt.Errorf("search not implemented")))
}

func PaperAdd(ctx *gin.Context, a *app.App) {
	ctx.JSON(http.StatusNotImplemented, presenters.Error(fmt.Errorf("addpaper not implemented")))
}
