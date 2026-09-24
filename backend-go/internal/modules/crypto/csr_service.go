package crypto

import (
	"gorm.io/gorm"

	"backend-go/internal/models"
)

type CSRService struct {
	db *gorm.DB
}

func NewCSRService(db *gorm.DB) *CSRService { return &CSRService{db: db} }

func (s *CSRService) List(page, pageSize int) (map[string]interface{}, error) {
	var csrs []models.CSR
	return paginateQuery(s.db.Model(&models.CSR{}), &csrs, page, pageSize)
}
