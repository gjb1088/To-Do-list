package sessions

import "net/http"

type Session struct {
	Values map[interface{}]interface{}
}

type CookieStore struct{}

func NewCookieStore(keyPairs ...[]byte) *CookieStore {
	return &CookieStore{}
}

func (c *CookieStore) Get(r *http.Request, name string) (*Session, error) {
	return &Session{Values: make(map[interface{}]interface{})}, nil
}

func (s *Session) Save(r *http.Request, w http.ResponseWriter) error {
	return nil
}
