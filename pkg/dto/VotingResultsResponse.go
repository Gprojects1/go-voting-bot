package dto

type VotingResultsResponse struct {
	Question   string   `json:"question"`
	Options    []string `json:"options"`
	Results    []Result `json:"results"`
	TotalVotes int      `json:"total_votes"`
}

type Result struct {
	Option     string  `json:"option"`
	VoteCount  int     `json:"vote_count"`
	Percentage float64 `json:"percentage"`
}
