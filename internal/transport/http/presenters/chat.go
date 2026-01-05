package presenters

type CreateChatRequest struct {
	UserId int64  `json:"user_id"`
	Title  string `json:"title"`
}

type Chat struct {
	ChatId    int64  `json:"chat_id"`
	UserId    int64  `json:"user_id"`
	UpdatedAt string `json:"updated_at"`
	Title     string `json:"title"`
}

type ChatResponse struct {
	Chat Chat `json:"chat"`
}

type ChatsResponse struct {
	Chats []Chat `json:"chats"`
}

type ChatHistoryCreateRequest struct {
	Text string `json:"text" binding:"required"`
}

type ChatHistoryMessage struct {
	SearchQuery string  `json:"search_query"`
	CreatedAt   string  `json:"created_at"`
	Papers      []Paper `json:"papers"`
}

type ChatHistoryResponse struct {
	ChatMessages []ChatHistoryMessage `json:"chat_messages"`
}
