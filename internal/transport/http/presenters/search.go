package presenters

type SearchPaperRequest struct {
	Text string `json:"text" binding:"required"`
}

type Paper struct {
	Id               string `json:"id"`
	Title            string `json:"title"`
	Abstract         string `json:"abstract"`
	Year             int    `json:"year"`
	Best_oa_location string `json:"best_oa_location"`
}

type SearchPaperResponse struct {
	Papers []Paper `json:"papers"`
}
