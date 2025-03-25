package dto

type VotingRequest struct {
	Text string `form:"text" binding:"required"`
}
