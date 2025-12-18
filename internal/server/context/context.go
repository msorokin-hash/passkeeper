package context

import "context"

type contextKey string

// UserContext represents the authenticated user's data stored inside
// a request context. It contains the user's ID, login, and decrypted secret.
// This structure is used by server handlers to access user-specific metadata.
type UserContext struct {
	ID     string
	Login  string
	Secret string
}

const userContextKey contextKey = "user"

// WrapContextWithUser returns a new context containing the provided UserContext.
// It is typically used by authentication middleware to attach user information
// to the request context so that downstream handlers can access it.
func WrapContextWithUser(ctx context.Context, user *UserContext) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// GetUserFromContext retrieves the UserContext stored within the given context.
// If no user information is present, or the stored value has an unexpected type,
// the function returns nil.
func GetUserFromContext(ctx context.Context) *UserContext {
	val := ctx.Value(userContextKey)
	user, ok := val.(*UserContext)
	if !ok {
		return nil
	}

	return user
}
