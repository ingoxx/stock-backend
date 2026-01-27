package domain

type VerifyData struct {
}

type VerifyRepository interface {
	GetAuthData(vd string) error
}
