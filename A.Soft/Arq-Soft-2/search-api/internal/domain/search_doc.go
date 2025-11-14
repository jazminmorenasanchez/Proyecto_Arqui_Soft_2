package domain

type SearchDoc struct {
	ID         string   `json:"id"` // = session_id
	ActivityID string   `json:"activity_id"`
	SessionID  string   `json:"session_id"`
	Name       string   `json:"name"`
	Sport      string   `json:"sport"`
	Site       string   `json:"site"`
	Instructor string   `json:"instructor"`
	StartAt    string   `json:"start_dt"` // ISO8601
	EndAt      string   `json:"end_dt"`
	Difficulty int      `json:"difficulty"`
	Price      float64  `json:"price"`
	Tags       []string `json:"tags"`
	UpdatedAt  string   `json:"updated_dt"`
}

type Result struct {
	Total int         `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
	Docs  []SearchDoc `json:"docs"`
}
