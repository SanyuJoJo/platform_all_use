package crypto

import (
	"gorm.io/gorm/clause"
)

// clauseOnConflict 构造 GORM 的 UPSERT 子句。
//
// 用法：
//   s.db.Clauses(clauseOnConflict("ca_id", []string{
//       "subject_cn", "algorithm", "cert_path",
//   })).Create(&ca)
//
// SQLite 3.24+ / PostgreSQL 9.5+ / MySQL 8.0+ 均支持。
func clauseOnConflict(column string, updateColumns []string) clause.OnConflict {
	return clause.OnConflict{
		Columns:   []clause.Column{{Name: column}},
		DoUpdates: clause.AssignmentColumns(updateColumns),
	}
}
