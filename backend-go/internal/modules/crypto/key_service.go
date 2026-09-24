package crypto

import (
	"gorm.io/gorm"

	"backend-go/internal/models"
)

type KeyService struct {
	db *gorm.DB
}

func NewKeyService(db *gorm.DB) *KeyService { return &KeyService{db: db} }

func (s *KeyService) List(page, pageSize int, algorithm, state string) (map[string]interface{}, error) {
	q := s.db.Model(&models.KeyMeta{})
	if algorithm != "" {
		q = q.Where("algorithm = ?", algorithm)
	}
	if state != "" {
		q = q.Where("state = ?", state)
	}
	var keys []models.KeyMeta
	return paginateQuery(q, &keys, page, pageSize)
}
