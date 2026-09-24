package crypto

import (
	"gorm.io/gorm"

	"backend-go/internal/models"
)

type CRLService struct {
	db *gorm.DB
}

func NewCRLService(db *gorm.DB) *CRLService { return &CRLService{db: db} }

func (s *CRLService) List(page, pageSize int, caID string) (map[string]interface{}, error) {
	q := s.db.Model(&models.CRL{})
	if caID != "" {
		q = q.Where("ca_id = ?", caID)
	}
	var crls []models.CRL
	return paginateQuery(q, &crls, page, pageSize)
}
