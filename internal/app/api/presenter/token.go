package presenter

// Token ...
type Token struct {
	Token string `json:"token"`
	// Expires time.Time `json:"expires"`
}

// NewTokenPresenter create new token presenter
func NewTokenPresenter() *Token {
	return &Token{}
}

// Make make token presenter
func (p Token) Make(token string) *Token {
	p.Token = token
	return &p
}
