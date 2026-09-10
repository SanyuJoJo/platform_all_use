"""
Pytest 全局夹具。
- session 级夹具确保认证种子数据就绪；
- 提供独立测试用户（禁用用户、改密用户）的创建与清理；
- 所有夹具均在数据库层面操作，不依赖其他测试的执行顺序。
"""
import pytest
from sqlalchemy import delete, select
from src.core.database import AsyncSessionLocal
from src.core.security import hash_password
from src.modules.auth.models import Role, User, UserRole
from src.modules.auth.service import ensure_auth_seed_data
@pytest.fixture(scope="session", autouse=True)
async def init_seed_data():
    """确保认证种子数据存在（幂等）。"""
    async with AsyncSessionLocal() as session:
        await ensure_auth_seed_data(session)
    yield
@pytest.fixture
async def disabled_user():
    """
    创建一个独立禁用用户 `disabled_user`，测试结束后删除。
    避免直接修改 guest 状态而污染其他测试。
    """
    username = "disabled_user"
    async with AsyncSessionLocal() as session:
        existing = await session.scalar(select(User).where(User.username == username))
        if existing:
            await session.execute(
                delete(UserRole).where(UserRole.user_id == existing.id)
            )
            await session.delete(existing)
            await session.commit()
        user = User(
            username=username,
            password_hash=hash_password("123456"),
            nickname="禁用用户",
            email="disabled@example.com",
            status=0,
        )
        session.add(user)
        await session.commit()
        await session.refresh(user)
        user_id = user.id
    yield {"id": user_id, "username": username, "password": "123456"}
    async with AsyncSessionLocal() as session:
        await session.execute(delete(UserRole).where(UserRole.user_id == user_id))
        user = await session.get(User, user_id)
        if user:
            await session.delete(user)
            await session.commit()
@pytest.fixture
async def password_test_user():
    """
    创建独立用户 `pwd_test_user` 用于改密测试，测试结束后删除。
    """
    username = "pwd_test_user"
    async with AsyncSessionLocal() as session:
        existing = await session.scalar(select(User).where(User.username == username))
        if existing:
            await session.execute(
                delete(UserRole).where(UserRole.user_id == existing.id)
            )
            await session.delete(existing)
            await session.commit()
        user = User(
            username=username,
            password_hash=hash_password("123456"),
            nickname="改密测试用户",
            email="pwd_test@example.com",
            status=1,
        )
        session.add(user)
        await session.commit()
        await session.refresh(user)
        user_id = user.id
        guest_role = await session.scalar(select(Role).where(Role.code == "guest"))
        if guest_role:
            session.add(UserRole(user_id=user_id, role_id=guest_role.id))
            await session.commit()
    yield {"id": user_id, "username": username, "password": "123456"}
    async with AsyncSessionLocal() as session:
        await session.execute(delete(UserRole).where(UserRole.user_id == user_id))
        user = await session.get(User, user_id)
        if user:
            await session.delete(user)
            await session.commit()
