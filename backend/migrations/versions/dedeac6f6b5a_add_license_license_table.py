"""add license_license table

Revision ID: dedeac6f6b5a
Revises: 821951c30f57
Create Date: 2026-09-14 00:17:26.096809

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = 'dedeac6f6b5a'
down_revision: Union[str, Sequence[str], None] = '821951c30f57'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None

def upgrade() -> None:
    op.create_table(
        "license_license",
        sa.Column("id", sa.Integer(), autoincrement=True, nullable=False),
        sa.Column("license_key", sa.String(length=255), nullable=False),
        sa.Column("license_type", sa.String(length=50), nullable=False),
        sa.Column("max_users", sa.Integer(), nullable=True),
        sa.Column("authorized_modules", sa.JSON(), nullable=True),
        sa.Column("machine_code", sa.String(length=255), nullable=True),
        sa.Column("issued_at", sa.DateTime(), nullable=False),
        sa.Column("expires_at", sa.DateTime(), nullable=False),
        sa.Column(
            "is_active",
            sa.SmallInteger(),
            nullable=False,
            server_default="1",
        ),
        sa.Column("activated_at", sa.DateTime(), nullable=True),
        sa.Column(
            "created_at",
            sa.DateTime(),
            nullable=False,
            server_default=sa.func.now(),
        ),
        sa.Column(
            "updated_at",
            sa.DateTime(),
            nullable=False,
            server_default=sa.func.now(),
        ),
        sa.PrimaryKeyConstraint("id"),
    )
    op.create_index(
        "ix_license_license_license_key",
        "license_license",
        ["license_key"],
        unique=True,
    )
    op.create_index(
        "ix_license_license_expires_at",
        "license_license",
        ["expires_at"],
    )
    # v1.2（R-3）：部分唯一索引，保证同一时刻最多一条 is_active=1 记录
    op.create_index(
        "uq_license_license_active",
        "license_license",
        ["is_active"],
        unique=True,
        sqlite_where=sa.text("is_active = 1"),
        postgresql_where=sa.text("is_active = 1"),
    )
def downgrade() -> None:
    op.drop_index(
        "uq_license_license_active", table_name="license_license"
    )
    op.drop_index(
        "ix_license_license_expires_at", table_name="license_license"
    )
    op.drop_index(
        "ix_license_license_license_key", table_name="license_license"
    )
    op.drop_table("license_license")
