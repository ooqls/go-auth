package tokenv1

type TokenData struct {
	Token      string
	ValidHours int
}

func NewTokenData(token string, validMinutes int) *TokenData {
	return &TokenData{
		Token:      token,
		ValidHours: validMinutes,
	}
}
