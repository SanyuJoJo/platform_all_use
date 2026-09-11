"""add module_manager tables

Revision ID: 278f016f7011
Revises: 087a7efdddf7
Create Date: 2026-09-11 16:14:57.990492

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = '278f016f7011'
down_revision: Union[str, Sequence[str], None] = '087a7efdddf7'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None

def upgrade() -> None:
    op.create_table(
        "module_manager_module",
        sa.Column("id", sa.String(length=50), nullable=False, comment="模块 ID"),
        sa.Column("name", sa.String(length=50), nullable=False, comment="模块名称"),
        sa.Column("version", sa.String(length=20), nullable=False, comment="语义化版本"),
        sa.Column("description", sa.Text(), nullable=True, comment="模块描述"),
        sa.Column("author", sa.String(length=100), nullable=True, comment="作者"),
        sa.Column("homepage", sa.String(length=255), nullable=True, comment="主页"),
        sa.Column(
            "status",
            sa.String(length=20),
            nullable=False,
            server_default="inactive",
            comment="active/inactive",
        ),
        sa.Column(
            "entry_backend",
            sa.String(length=100),
            nullable=False,
            comment="后端入口，如 router:router",
        ),
        sa.Column(
            "entry_frontend",
            sa.String(length=255),
            nullable=True,
            comment="前端入口 URL",
        ),
        sa.Column("config", sa.JSON(), nullable=True, comment="模块配置"),
        sa.Column(
            "manifest",
            sa.JSON(),
            nullable=True,
            comment="完整模块清单（含 menus/permissions/config_schema）",
        ),
        sa.Column(
            "installed_at",
            sa.DateTime(),
            nullable=False,
            server_default=sa.func.now(),
            comment="安装时间",
        ),
        sa.Column(
            "updated_at",
            sa.DateTime(),
            nullable=False,
            server_default=sa.func.now(),
            comment="更新时间",
        ),
        sa.PrimaryKeyConstraint("id"),
    )

    op.create_table(
        "module_manager_dependency",
        sa.Column("id", sa.Integer(), autoincrement=True, nullable=False),
        sa.Column("module_id", sa.String(length=50), nullable=False),
        sa.Column("dependency_id", sa.String(length=50), nullable=False),
        sa.Column(
            "created_at",
            sa.DateTime(),
            nullable=False,
            server_default=sa.func.now(),
        ),
        sa.ForeignKeyConstraint(
            ["module_id"],
            ["module_manager_module.id"],
            ondelete="CASCADE",
        ),
        sa.PrimaryKeyConstraint("id"),
        sa.UniqueConstraint(
            "module_id",
            "dependency_id",
            name="uq_module_manager_dependency_module_dep",
        ),
    )
    op.create_index(
        "ix_module_manager_dependency_module_id",
        "module_manager_dependency",
        ["module_id"],
    )
    op.create_index(
        "ix_module_manager_dependency_dependency_id",
        "module_manager_dependency",
        ["dependency_id"],
    )


def downgrade() -> None:
    op.drop_index(
        "ix_module_manager_dependency_dependency_id",
        table_name="module_manager_dependency",
    )
    op.drop_index(
        "ix_module_manager_dependency_module_id",
        table_name="module_manager_dependency",
    )
    op.drop_table("module_manager_dependency")
    op.drop_table("module_manager_module")
