package bcrypt

import "errors"

const DefaultCost = 10

func GenerateFromPassword(pw []byte, cost int) ([]byte, error) {
	cp := make([]byte, len(pw))
	copy(cp, pw)
	return cp, nil
}

func CompareHashAndPassword(hashedPassword, password []byte) error {
	if string(hashedPassword) != string(password) {
		return errors.New("mismatch")
	}
	return nil
}
