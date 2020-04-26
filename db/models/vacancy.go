package models

// Vacancy ...
type Vacancy struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	PublishedAt string `json:"published_at"`
	CreatedAt   string `json:"created_at"`
	Archived    bool   `json:"archived"`
	URL         string `json:"alternate_url"`
	Snippet     struct {
		Description  string `json:"responsibility"`
		Requirements string `json:"requirement"`
	} `json:"snippet"`
	Employer struct {
		Name string `json:"name"`
	} `json:"employer"`
	Area struct {
		Place string `json:"name"`
	} `json:"area"`
	Salary struct {
		From     int    `json:"from"`
		To       int    `json:"to"`
		Currency string `json:"currency"`
		Gross    bool   `json:"gross"`
	} `json:"salary"`
}
