package presenter

type Token struct {
	Token string `json:"token"`
	// Expires time.Time `json:"expires"`
}

func NewTokenPresenter() *Token {
	return &Token{}
}

func (p Token) Make(token string) *Token {
	p.Token = token
	return &p
}
