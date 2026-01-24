package handlers

import (
	pb "VKR_gateway_service/gen/go"
	"VKR_gateway_service/internal/app"
	"VKR_gateway_service/internal/transport/http/presenters"
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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

// CreateChat
// @Summary Create chat
// @Description Create a new chat for the user
// @Tags chat
// @Accept json
// @Produce json
// @Param data body presenters.CreateChatRequest true "Chat data"
// @Success 200 {object} presenters.ChatResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 401 {object} presenters.ErrorResponse
// @Failure 403 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /chats [post]
func CreateChat(ctx *gin.Context, a *app.App) {
	var in presenters.CreateChatRequest
	if err := ctx.ShouldBindJSON(&in); err != nil {
		ctx.JSON(http.StatusBadRequest, presenters.Error(err))
		return
	}
	userID, statusCode, err := resolveUserID(ctx, in.UserId)
	if err != nil {
		ctx.JSON(statusCode, presenters.Error(err))
		return
	}

	req := &pb.Chat{
		UserId: userID,
		Title:  in.Title,
	}
	rctx, cancel := requestContext(ctx, a)
	defer cancel()
	resp, err := a.AI.CreateNewChat(rctx, req)
	if err != nil {
		if a.Logger != nil {
			a.Logger.WithError(err).WithField("user_id", userID).Error("AI CreateNewChat RPC failed")
		}
		if s, ok := status.FromError(err); ok {
			ctx.JSON(mapGRPCToHTTP(s.Code()), presenters.Error(fmt.Errorf(s.Message())))
			return
		}
		ctx.JSON(http.StatusBadGateway, presenters.Error(err))
		return
	}
	chat := resp.GetChat()
	if chat == nil {
		ctx.JSON(http.StatusBadGateway, presenters.Error(fmt.Errorf("empty chat response")))
		return
	}
	ctx.JSON(http.StatusOK, mapChat(chat))
}

// GetUserChats
// @Summary Get user chats
// @Description Get all chats for a user
// @Tags chat
// @Accept json
// @Produce json
// @Param user_id query int false "User ID"
// @Success 200 {object} presenters.ChatsResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 401 {object} presenters.ErrorResponse
// @Failure 403 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /chats [get]
func GetUserChats(ctx *gin.Context, a *app.App) {
	userID, err := parseOptionalQueryInt64(ctx, "user_id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, presenters.Error(err))
		return
	}
	userID, statusCode, err := resolveUserID(ctx, userID)
	if err != nil {
		ctx.JSON(statusCode, presenters.Error(err))
		return
	}

	req := &pb.UserChatsReq{UserId: userID}
	rctx, cancel := requestContext(ctx, a)
	defer cancel()
	resp, err := a.AI.GetUserChats(rctx, req)
	if err != nil {
		if a.Logger != nil {
			a.Logger.WithError(err).WithField("user_id", userID).Error("AI GetUserChats RPC failed")
		}
		if s, ok := status.FromError(err); ok {
			ctx.JSON(mapGRPCToHTTP(s.Code()), presenters.Error(fmt.Errorf(s.Message())))
			return
		}
		ctx.JSON(http.StatusBadGateway, presenters.Error(err))
		return
	}

	out := presenters.ChatsResponse{Chats: make([]presenters.ChatResponse, 0, len(resp.GetChats()))}
	for _, chat := range resp.GetChats() {
		out.Chats = append(out.Chats, mapChat(chat))
	}
	ctx.JSON(http.StatusOK, out)
}

// GetChatHistory
// @Summary Get chat history
// @Description Get chat history by chat ID
// @Tags chat
// @Accept json
// @Produce json
// @Param chat_id path int true "Chat ID"
// @Param user_id query int false "User ID"
// @Success 200 {object} presenters.ChatHistoryResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 401 {object} presenters.ErrorResponse
// @Failure 403 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /chats/{chat_id}/history [get]
func GetChatHistory(ctx *gin.Context, a *app.App) {
	chatID, err := parsePathInt64(ctx, "chat_id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, presenters.Error(err))
		return
	}
	userID, err := parseOptionalQueryInt64(ctx, "user_id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, presenters.Error(err))
		return
	}
	userID, statusCode, err := resolveUserID(ctx, userID)
	if err != nil {
		ctx.JSON(statusCode, presenters.Error(err))
		return
	}
	if !authorizeChatAccess(ctx, a, userID, chatID) {
		return
	}

	req := &pb.HistoryReq{ChatId: chatID}
	rctx, cancel := requestContext(ctx, a)
	defer cancel()
	resp, err := a.AI.GetChatHistory(rctx, req)
	if err != nil {
		if a.Logger != nil {
			a.Logger.WithError(err).WithField("chat_id", chatID).Error("AI GetChatHistory RPC failed")
		}
		if s, ok := status.FromError(err); ok {
			ctx.JSON(mapGRPCToHTTP(s.Code()), presenters.Error(fmt.Errorf(s.Message())))
			return
		}
		ctx.JSON(http.StatusBadGateway, presenters.Error(err))
		return
	}

	out := presenters.ChatHistoryResponse{ChatMessages: make([]presenters.ChatHistoryMessage, 0, len(resp.GetChatMessages()))}
	for _, msg := range resp.GetChatMessages() {
		out.ChatMessages = append(out.ChatMessages, presenters.ChatHistoryMessage{
			SearchQuery: msg.GetSearchQuery(),
			CreatedAt:   msg.GetCreatedAt(),
			Papers:      mapPapers(msg.GetPapers().GetPapers()),
		})
	}
	ctx.JSON(http.StatusOK, out)
}

// CreateChatHistory
// @Summary Add chat history entry
// @Description Create a new chat history entry by chat ID and search text
// @Tags chat
// @Accept json
// @Produce json
// @Param chat_id path int true "Chat ID"
// @Param user_id query int false "User ID"
// @Param data body presenters.ChatHistoryCreateRequest true "Search query"
// @Success 200 {object} presenters.SearchPaperResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 401 {object} presenters.ErrorResponse
// @Failure 403 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /chats/{chat_id}/history [post]
func CreateChatHistory(ctx *gin.Context, a *app.App) {
	chatID, err := parsePathInt64(ctx, "chat_id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, presenters.Error(err))
		return
	}
	userID, err := parseOptionalQueryInt64(ctx, "user_id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, presenters.Error(err))
		return
	}
	userID, statusCode, err := resolveUserID(ctx, userID)
	if err != nil {
		ctx.JSON(statusCode, presenters.Error(err))
		return
	}
	if !authorizeChatAccess(ctx, a, userID, chatID) {
		return
	}
	var in presenters.ChatHistoryCreateRequest
	if err := ctx.ShouldBindJSON(&in); err != nil {
		ctx.JSON(http.StatusBadRequest, presenters.Error(err))
		return
	}

	req := &pb.SearchRequest{
		InputData: in.Text,
		ChatId:    chatID,
	}
	rctx, cancel := requestContext(ctx, a)
	defer cancel()
	resp, err := a.AI.SearchPaper(rctx, req)
	if err != nil {
		if a.Logger != nil {
			a.Logger.WithError(err).WithFields(map[string]interface{}{
				"chat_id": chatID,
				"user_id": userID,
			}).Error("AI SearchPaper RPC failed")
		}
		if s, ok := status.FromError(err); ok {
			ctx.JSON(mapGRPCToHTTP(s.Code()), presenters.Error(fmt.Errorf(s.Message())))
			return
		}
		ctx.JSON(http.StatusBadGateway, presenters.Error(err))
		return
	}

	out := presenters.SearchPaperResponse{Papers: mapPapers(resp.GetPapers())}
	ctx.JSON(http.StatusOK, out)
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

func requestContext(ctx *gin.Context, a *app.App) (context.Context, context.CancelFunc) {
	rctx := ctx.Request.Context()
	if a != nil && a.Config != nil && a.Config.GRPCTimeout > 0 {
		return context.WithTimeout(rctx, a.Config.GRPCTimeout)
	}
	return rctx, func() {}
}

func parsePathInt64(ctx *gin.Context, name string) (int64, error) {
	raw := ctx.Param(name)
	if raw == "" {
		return 0, fmt.Errorf("%s path param is required", name)
	}
	return parsePositiveInt64(raw, name)
}

func parseOptionalQueryInt64(ctx *gin.Context, name string) (int64, error) {
	raw := ctx.Query(name)
	if raw == "" {
		return 0, nil
	}
	return parsePositiveInt64(raw, name)
}

func parsePositiveInt64(raw, field string) (int64, error) {
	val, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || val <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", field)
	}
	return val, nil
}

func resolveUserID(ctx *gin.Context, userID int64) (int64, int, error) {
	authID, ok := authUserID(ctx)
	if userID > 0 {
		if ok && authID != userID {
			return 0, http.StatusForbidden, fmt.Errorf("user_id does not match token")
		}
		return userID, 0, nil
	}
	if ok {
		return authID, 0, nil
	}
	return 0, http.StatusUnauthorized, fmt.Errorf("user_id is required")
}

func authUserID(ctx *gin.Context) (int64, bool) {
	val, ok := ctx.Get("user_id")
	if !ok {
		return 0, false
	}
	switch v := val.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case float64:
		return int64(v), true
	case string:
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, false
		}
		return id, true
	default:
		return 0, false
	}
}

func authorizeChatAccess(ctx *gin.Context, a *app.App, userID, chatID int64) bool {
	req := &pb.UserChatsReq{UserId: userID}
	rctx, cancel := requestContext(ctx, a)
	defer cancel()
	resp, err := a.AI.GetUserChats(rctx, req)
	if err != nil {
		if a.Logger != nil {
			a.Logger.WithError(err).WithField("user_id", userID).Error("AI GetUserChats RPC failed")
		}
		if s, ok := status.FromError(err); ok {
			ctx.JSON(mapGRPCToHTTP(s.Code()), presenters.Error(fmt.Errorf(s.Message())))
			return false
		}
		ctx.JSON(http.StatusBadGateway, presenters.Error(err))
		return false
	}
	for _, chat := range resp.GetChats() {
		if chat.GetChatId() == chatID {
			return true
		}
	}
	ctx.JSON(http.StatusForbidden, presenters.Error(fmt.Errorf("chat access denied")))
	return false
}

func mapChat(chat *pb.Chat) presenters.ChatResponse {
	if chat == nil {
		return presenters.ChatResponse{}
	}
	return presenters.ChatResponse{
		ChatId:    chat.GetChatId(),
		UserId:    chat.GetUserId(),
		UpdatedAt: chat.GetUpdatedAt(),
		Title:     chat.GetTitle(),
	}
}

func mapPapers(papers []*pb.PaperResponse) []presenters.Paper {
	out := make([]presenters.Paper, 0, len(papers))
	for _, p := range papers {
		out = append(out, presenters.Paper{
			Id:               p.GetID(),
			Title:            p.GetTitle(),
			Abstract:         p.GetAbstract(),
			Year:             int(p.GetYear()),
			Best_oa_location: p.GetBestOaLocation(),
		})
	}
	return out
}
