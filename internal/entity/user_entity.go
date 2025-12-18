package entity

// User defines the data model for a system user, containing authentication details and encrypted secrets.
type User struct {
	ID              string
	Login           string
	PasswordHash    string
	EncryptedSecret string
}
