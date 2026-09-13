"""add audit_log_operation table

Revision ID: 821951c30f57
Revises: 278f016f7011
Create Date: 2026-09-13 23:28:31.055429

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = '821951c30f57'
down_revision: Union[str, Sequence[str], None] = '278f016f7011'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None

def upgrade() -> None:
    op.create_table(
        "audit_log_operation",
        sa.Column("id", sa.Integer(), autoincrement=True, nullable=False),
        sa.Column("user_id", sa.Integer(), nullable=True),
        sa.Column("username", sa.String(length=50), nullable=True),
        sa.Column("module_id", sa.String(length=50), nullable=False),
        sa.Column("action", sa.String(length=50), nullable=False),
        sa.Column("resource", sa.String(length=50), nullable=True),
        sa.Column("resource_id", sa.String(length=50), nullable=True),
        sa.Column("detail", sa.Text(), nullable=True),
        sa.Column("ip", sa.String(length=45), nullable=True),
        sa.Column("user_agent", sa.String(length=255), nullable=True),
        sa.Column(
            "status",
            sa.String(length=20),
            nullable=False,
            server_default="success",
        ),
        sa.Column("error_code", sa.Integer(), nullable=True),
        sa.Column("request_id", sa.String(length=36), nullable=True),
        sa.Column(
            "created_at",
            sa.DateTime(),
            nullable=False,
            server_default=sa.func.now(),
        ),
        sa.ForeignKeyConstraint(
            ["user_id"],
            ["auth_user.id"],
            ondelete="SET NULL",
        ),
        sa.PrimaryKeyConstraint("id"),
    )
    op.create_index(
        "ix_audit_log_operation_user_id",
        "audit_log_operation",
        ["user_id"],
    )
    op.create_index(
        "ix_audit_log_operation_module_id",
        "audit_log_operation",
        ["module_id"],
    )
    op.create_index(
        "ix_audit_log_operation_action",
        "audit_log_operation",
        ["action"],
    )
    op.create_index(
        "ix_audit_log_operation_created_at",
        "audit_log_operation",
        ["created_at"],
    )
def downgrade() -> None:
    op.drop_index(
        "ix_audit_log_operation_created_at",
        table_name="audit_log_operation",
    )
    op.drop_index(
        "ix_audit_log_operation_action",
        table_name="audit_log_operation",
    )
    op.drop_index(
        "ix_audit_log_operation_module_id",
        table_name="audit_log_operation",
    )
    op.drop_index(
        "ix_audit_log_operation_user_id",
        table_name="audit_log_operation",
    )
    op.drop_table("audit_log_operation")
