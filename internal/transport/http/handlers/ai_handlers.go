package handlers

import (
	pb "VKR_gateway_service/gen/go"
	"VKR_gateway_service/internal/app"
	"VKR_gateway_service/internal/transport/http/presenters"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SearchPapers
// @Summary Search papers
// @Description Semantic search for papers
// @Tags ai
// @Accept json
// @Produce json
// @Param text query string true "Search text"
// @Success 200 {object} presenters.SearchPaperResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 401 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /ai/search/papers [get]
func SearchPapers(ctx *gin.Context, a *app.App) {
	text := ctx.Query("text")
	if text == "" {
		ctx.JSON(http.StatusBadRequest, presenters.Error(fmt.Errorf("text query param is required")))
		return
	}

	req := &pb.SearchRequest{InputData: text}
	rctx := ctx.Request.Context()
	if a.Config.GRPCTimeout > 0 {
		var cancel context.CancelFunc
		rctx, cancel = context.WithTimeout(rctx, a.Config.GRPCTimeout)
		defer cancel()
	}
	resp, err := a.AI.SearchPaper(rctx, req)
	if err != nil {
		if a.Logger != nil {
			a.Logger.WithError(err).WithField("text", text).Error("AI SearchPaper RPC failed")
		}
		if s, ok := status.FromError(err); ok {
			ctx.JSON(mapGRPCToHTTP(s.Code()), presenters.Error(fmt.Errorf(s.Message())))
			return
		}
		ctx.JSON(http.StatusBadGateway, presenters.Error(err))
		return
	}

	out := presenters.SearchPaperResponse{Papers: make([]presenters.Paper, 0, len(resp.GetPapers()))}
	for _, p := range resp.GetPapers() {
		out.Papers = append(out.Papers, presenters.Paper{
			Id:               p.GetID(),
			Title:            p.GetTitle(),
			Abstract:         p.GetAbstract(),
			Year:             int(p.GetYear()),
			Best_oa_location: p.GetBestOaLocation(),
		})
	}
	ctx.JSON(http.StatusOK, out)
}

// PaperAdd
// @Summary Add paper
// @Description Add a paper to the index
// @Tags ai
// @Accept json
// @Produce json
// @Param data body presenters.AddPaperRequest true "Paper data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 401 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /ai/paper/add [post]
func PaperAdd(ctx *gin.Context, a *app.App) {
	var in presenters.AddPaperRequest
	if err := ctx.ShouldBindJSON(&in); err != nil {
		ctx.JSON(http.StatusBadRequest, presenters.Error(err))
		return
	}

	req := &pb.AddRequest{
		ID:             in.Id,
		Title:          in.Title,
		Abstract:       in.Abstract,
		Year:           int64(in.Year),
		BestOaLocation: in.Best_oa_location,
	}
	if len(in.ReferencedPapers) > 0 {
		req.ReferencedWorks = make([]*pb.ReferencedWorks, 0, len(in.ReferencedPapers))
		for _, r := range in.ReferencedPapers {
			req.ReferencedWorks = append(req.ReferencedWorks, &pb.ReferencedWorks{ID: r.Id})
		}
	}
	if len(in.RelatedPaper) > 0 {
		req.RelatedWorks = make([]*pb.RelatedWorks, 0, len(in.RelatedPaper))
		for _, r := range in.RelatedPaper {
			req.RelatedWorks = append(req.RelatedWorks, &pb.RelatedWorks{ID: r.Id})
		}
	}

	rctx := ctx.Request.Context()
	if a.Config.GRPCTimeout > 0 {
		var cancel context.CancelFunc
		rctx, cancel = context.WithTimeout(rctx, a.Config.GRPCTimeout)
		defer cancel()
	}
	resp, err := a.AI.AddPaper(rctx, req)
	if err != nil {
		if a.Logger != nil {
			a.Logger.WithError(err).WithField("id", in.Id).Error("AI AddPaper RPC failed")
		}
		if s, ok := status.FromError(err); ok {
			ctx.JSON(mapGRPCToHTTP(s.Code()), presenters.Error(fmt.Errorf(s.Message())))
			return
		}
		ctx.JSON(http.StatusBadGateway, presenters.Error(err))
		return
	}
	if msg := resp.GetError(); msg != "" {
		ctx.JSON(http.StatusBadRequest, &presenters.ErrorResponse{Error: msg})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func mapGRPCToHTTP(c codes.Code) int {
	switch c {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.NotFound:
		return http.StatusNotFound
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.Unavailable:
		return http.StatusBadGateway
	case codes.PermissionDenied, codes.Unauthenticated:
		return http.StatusUnauthorized
	default:
		return http.StatusBadGateway
	}
}
