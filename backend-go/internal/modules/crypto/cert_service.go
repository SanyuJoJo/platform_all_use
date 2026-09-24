package crypto

import (
	"errors"

	"gorm.io/gorm"

	"backend-go/internal/exception"
	"backend-go/internal/models"
)

// CertService 证书元数据服务。
type CertService struct {
	db     *gorm.DB
	files  *FileStore
	parser *CertParser
}

func NewCertService(db *gorm.DB, caller *CoreCaller, files *FileStore, parser *CertParser) *CertService {
	return &CertService{db: db, files: files, parser: parser}
}

func (s *CertService) List(page, pageSize int, certType, caID string) (map[string]interface{}, error) {
	q := s.db.Model(&models.Certificate{})
	if certType != "" {
		q = q.Where("cert_type = ?", certType)
	}
	if caID != "" {
		q = q.Where("ca_id = ?", caID)
	}
	var certs []models.Certificate
	return paginateQuery(q, &certs, page, pageSize)
}

func (s *CertService) Get(certID string) (*models.Certificate, error) {
	var c models.Certificate
	if err := s.db.Where("cert_id = ?", certID).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.New(exception.CodeNotFound, "证书不存在", 404, nil)
		}
		return nil, exception.New(exception.CodeInternalError, "查询证书失败", 500, nil)
	}
	return &c, nil
}
