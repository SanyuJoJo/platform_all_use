-- +goose Up
-- 扩展 platform_certificate：新增完整 DN、摘要、算法、key_ref 字段。
ALTER TABLE platform_certificate ADD COLUMN subject VARCHAR(512);
ALTER TABLE platform_certificate ADD COLUMN issuer VARCHAR(512);
ALTER TABLE platform_certificate ADD COLUMN fingerprint VARCHAR(128);
ALTER TABLE platform_certificate ADD COLUMN public_key_algorithm VARCHAR(64);
ALTER TABLE platform_certificate ADD COLUMN signature_algorithm VARCHAR(64);
ALTER TABLE platform_certificate ADD COLUMN key_ref VARCHAR(64);

CREATE INDEX IF NOT EXISTS ix_platform_certificate_cert_type
    ON platform_certificate (cert_type);
CREATE INDEX IF NOT EXISTS ix_platform_certificate_key_ref
    ON platform_certificate (key_ref);

-- +goose Down
DROP INDEX IF EXISTS ix_platform_certificate_key_ref;
DROP INDEX IF EXISTS ix_platform_certificate_cert_type;
-- SQLite 不支持 DROP COLUMN，如需回滚请重建表。
