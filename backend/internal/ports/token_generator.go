package ports

type TokenGenerator interface {
	NewRawToken() (string, error)
	HashToken(raw string) string
	NewID() string
}
