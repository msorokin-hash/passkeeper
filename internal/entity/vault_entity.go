package entity

import "time"

type DataType string

// DataType constants define the types of items stored in the vault.
const (
	Password DataType = "PASSWORD"
	Text     DataType = "TEXT"
	BankCard DataType = "BANK_CARD"
	File     DataType = "FILE"
)

// VaultItem represents a single encrypted entry in the user's vault.
type VaultItem struct {
	ID            string
	UserID        string
	EncryptedData []byte
	Meta          string
	Type          DataType
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// PasswordMeta contains metadata for password items.
type PasswordMeta struct {
	Resource string `json:"resource"`
	Login    string `json:"login"`
	Comment  string `json:"comment"`
}

// TextMeta contains metadata for text/note items.
type TextMeta struct {
	Name    string `json:"name"`
	Comment string `json:"comment"`
}

// BankCardMeta contains metadata for bank card items.
type BankCardMeta struct {
	Bank    string `json:"bank"`
	Comment string `json:"comment"`
}

// FileMeta contains metadata for file items.
type FileMeta struct {
	Name      string `json:"name"`
	Extension string `json:"extension"`
	Comment   string `json:"comment"`
}

// BankCardData contains sensitive bank card information.
type BankCardData struct {
	Number     string `json:"number"`
	ValidMonth int    `json:"validMonth"`
	ValidYear  int    `json:"validYear"`
	Holder     string `json:"holder"`
	CSV        string `json:"csv"`
}

// PasswordItem is the decrypted representation of a password item returned by the service.
type PasswordItem struct {
	Type DataType
	Meta PasswordMeta
	Data string
}

// TextItem is the decrypted representation of a text item returned by the service.
type TextItem struct {
	Type DataType
	Meta TextMeta
	Data string
}

// BankCardItem is the decrypted representation of a bank card item returned by the service.
type BankCardItem struct {
	Type DataType
	Meta BankCardMeta
	Data BankCardData
}

// FileItem is the representation of a file item returned by the service.
type FileItem struct {
	Type DataType
	Meta FileMeta
}

// ItemInfo provides summary information for displaying vault item lists.
type ItemInfo struct {
	ID   string
	Meta any
}
