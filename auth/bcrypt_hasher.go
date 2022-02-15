package auth

import "golang.org/x/crypto/bcrypt"

type BcryptHasher struct{}

func (b *BcryptHasher) HashAndSalt(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func (b *BcryptHasher) VerifyPassword(passwordHash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
}
