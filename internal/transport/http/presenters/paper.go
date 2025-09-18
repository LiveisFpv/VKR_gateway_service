package presenters

type AddPaperRequest struct {
	Id               string            `json:"id"`
	Title            string            `json:"title"`
	Abstract         string            `json:"abstract"`
	Year             int               `json:"year"`
	Best_oa_location string            `json:"best_oa_location"`
	ReferencedPapers []ReferencedPaper `json:"referenced_paper"`
	RelatedPaper     []RelatedPaper    `json:"related_paper"`
}

type ReferencedPaper struct {
	Id string `json:"id"`
}

type RelatedPaper struct {
	Id string `json:"id"`
}
