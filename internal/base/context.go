package base

import "context"

type contextKey string

const (
	// KeyUserID represents the current logged-in UserID
	KeyUserID contextKey = "UserID"
)

// CurrentUser gets current user id from the context
func CurrentUser(ctx context.Context) *int {
	currentUser := ctx.Value(KeyUserID)
	if currentUser != nil {
		v := currentUser.(int)
		return &v
	}
	return nil
}
