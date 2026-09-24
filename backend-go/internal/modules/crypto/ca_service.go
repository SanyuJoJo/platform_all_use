package crypto
import (
	"errors"

	"gorm.io/gorm"

	"backend-go/internal/exception"
	"backend-go/internal/models"
)
// CAService 只负责 CA 相关业务。
type CAService struct {
	db     *gorm.DB
	caller *CoreCaller
	files  *FileStore
	parser *CertParser
	keys   *KeyCrypto
}
func NewCAService(
	db *gorm.DB, caller *CoreCaller, files *FileStore,
	parser *CertParser, keys *KeyCrypto,
) *CAService {
	return &CAService{db: db, caller: caller, files: files, parser: parser, keys: keys}
}
func (s *CAService) List(page, pageSize int) (map[string]interface{}, error) {
	var cas []models.CA
	return paginateQuery(
		s.db.Model(&models.CA{}).Where("status <> ?", "DELETED"),
		&cas, page, pageSize,
	)
}
func (s *CAService) Get(caID string) (*models.CA, error) {
	var ca models.CA
	if err := s.db.Where("ca_id = ?", caID).First(&ca).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.New(exception.CodeNotFound, "CA 不存在", 404, nil)
		}
		return nil, exception.New(exception.CodeInternalError, "查询 CA 失败", 500, nil)
	}
	return &ca, nil
}
func (s *CAService) GetDetail(caID string) (*CADetail, error) {
	ca, err := s.Get(caID)
	if err != nil {
		return nil, err
	}
	abs, err := s.files.guard.Resolve(ca.CertPath, "cert_path")
	if err != nil {
		return nil, err
	}
	return s.parser.Parse(abs)
}
func (s *CAService) Delete(caID string) error {
	res := s.db.Model(&models.CA{}).
		Where("ca_id = ? AND status <> ?", caID, "DELETED").
		Updates(map[string]interface{}{"status": "DELETED", "updated_at": timeNowUTC()})
	if res.Error != nil {
		return exception.New(exception.CodeInternalError, "删除 CA 失败", 500, nil)
	}
	if res.RowsAffected == 0 {
		return exception.New(exception.CodeNotFound, "CA 不存在或已删除", 404, nil)
	}
	return nil
}
