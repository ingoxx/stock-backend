package domain

type VerifyData struct {
}

type VerifyRepository interface {
	GetAuthData(vd string) error
	GetJwtToken(user, jt string) error
	DelJwtToken(user string) error
}
