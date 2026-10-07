package usecase

type CommentCreatedEvent struct {
	CommentID int64 `json:"comment_id"`
	MovieID   int64 `json:"movie_id"`
	UserID    int64 `json:"user_id"`
	CreatedAt int64 `json:"created_at"` // Unix timestamp
}
