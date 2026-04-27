package ports

type PasswordHasher interface {
	Hash(plaintext string) (string, error)
	Verify(hash, plaintext string) (bool, error)
}
