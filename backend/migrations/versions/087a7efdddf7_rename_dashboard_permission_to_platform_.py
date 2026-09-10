"""rename dashboard permission to platform format

Revision ID: 087a7efdddf7
Revises: 041c6783a402
Create Date: 2026-09-10 21:48:15.634745

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = '087a7efdddf7'
down_revision: Union[str, Sequence[str], None] = '041c6783a402'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None

def upgrade() -> None:
    """dashboard:view → platform:dashboard:view"""
    op.execute(
        sa.text(
            "UPDATE auth_permission "
            "SET code = 'platform:dashboard:view', "
            "    module_id = 'platform', "
            "    resource = 'dashboard', "
            "    action = 'view' "
            "WHERE code = 'dashboard:view'"
        )
    )


def downgrade() -> None:
    """platform:dashboard:view → dashboard:view"""
    op.execute(
        sa.text(
            "UPDATE auth_permission "
            "SET code = 'dashboard:view', "
            "    module_id = 'platform', "
            "    resource = 'dashboard', "
            "    action = 'view' "
            "WHERE code = 'platform:dashboard:view'"
        )
    )
