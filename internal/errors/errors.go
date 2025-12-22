package errors

import "errors"

var (
	// ErrLoginTaken is returned when trying to register with an existing login or email.
	ErrLoginTaken = errors.New("login or email is already taken")

	// ErrNoUser is returned when a requested user is not found in the database.
	ErrNoUser = errors.New("user not found in db")

	// ErrNoData is returned when requested data is not found in the database.
	ErrNoData = errors.New("data not found in db")

	// ErrInternalDatabase is returned for unexpected database errors
	// (connection issues, timeouts, deadlocks, etc.).
	ErrInternalDatabase = errors.New("internal database error")

	// ErrInvalidUserData is returned when user input fails validation.
	ErrInvalidUserData = errors.New("not valid user data")

	// ErrInvalidJwtToken is returned for JWT token issues
	// (expired, malformed, or invalid signature).
	ErrInvalidJwtToken = errors.New("invalid jwt token")
)
