package service

import "github.com/ingoxx/stock-backend/internal/domain"

type VerifyService struct {
	repo domain.VerifyRepository
}

func NewVerifyService(repo domain.VerifyRepository) *VerifyService {
	return &VerifyService{repo: repo}
}

func (vs *VerifyService) GetAuthData(vd string) error {
	return vs.repo.GetAuthData(vd)
}

func (vs *VerifyService) GetJwtToken(user, jt string) error {
	return vs.repo.GetJwtToken(user, jt)
}

func (vs *VerifyService) DelJwtToken(user string) error {
	return vs.repo.DelJwtToken(user)
}
