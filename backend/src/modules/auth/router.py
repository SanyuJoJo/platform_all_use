"""
认证模块 - 路由定义
"""
from typing import Optional, Tuple
from fastapi import APIRouter, Body, Depends, Header, Request
from sqlalchemy.ext.asyncio import AsyncSession
from src.core.database import get_db
from src.core.exceptions import PlatformException
from src.core.response import success_response
from src.modules.auth.dependencies import CurrentUser, get_current_user
from src.modules.auth.schemas import ChangePasswordReq, LoginReq, RefreshReq
from src.modules.auth.service import (
    authenticate_user,
    change_password as change_password_service,
    create_tokens_for_user,
    log_auth_event,
    refresh_access_token,
    revoke_refresh_token,
)
router = APIRouter(prefix="/api/v1/auth", tags=["Auth"])
def _extract_client_info(request: Request) -> Tuple[Optional[str], Optional[str]]:
    """从请求中提取客户端 IP 与 User-Agent。"""
    ip = request.client.host if request.client else None
    user_agent = request.headers.get("user-agent")
    return ip, user_agent
@router.post("/login")
async def login(
    req: LoginReq,
    request: Request,
    db: AsyncSession = Depends(get_db),
):
    """用户登录。"""
    ip, user_agent = _extract_client_info(request)
    user_info = await authenticate_user(
        db, req.username, req.password, ip, user_agent
    )
    tokens = await create_tokens_for_user(db, user_info, ip, user_agent)
    return success_response(
        data={**tokens, "user": user_info},
        message="登录成功",
    )
@router.post("/refresh")
async def refresh(
    req: Optional[RefreshReq] = Body(None),
    authorization: Optional[str] = Header(None),
    db: AsyncSession = Depends(get_db),
):
    """
    刷新 Access Token（v1.2 起同时轮换 Refresh Token）。
    优先从 Authorization: Bearer <refresh_token> 读取，
    也兼容请求体 {"refresh_token": "..."}。
    响应 data 包含新的 access_token 与 refresh_token；
    旧 refresh_token 立即失效。
    v1.3（N-1）：移除未使用的 `request` 参数。
    """
    token = None
    if authorization and authorization.lower().startswith("bearer "):
        token = authorization.split(" ", 1)[1]
    elif req and req.refresh_token:
        token = req.refresh_token
    if not token:
        raise PlatformException(
            code=10001, message="缺少refresh token", status_code=401
        )
    data = await refresh_access_token(db, token)
    return success_response(data=data, message="刷新成功")
@router.post("/logout")
async def logout(
    req: Optional[RefreshReq] = Body(None),
    current_user: CurrentUser = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    """
    用户登出。
    - 若请求体提供 refresh_token，则将其标记为已撤销，后续无法用于刷新；
    - Access Token 为无状态 JWT，服务端无法立即失效，前端应主动清除本地存储。
    v1.3（N-1）：移除未使用的 `request` 参数。
    """
    if req and req.refresh_token:
        await revoke_refresh_token(db, req.refresh_token)
    log_auth_event(
        "logout",
        user_id=current_user["id"],
        username=current_user["username"],
        status="success",
    )
    return success_response(message="登出成功", data=None)
@router.get("/me")
async def me(current_user: CurrentUser = Depends(get_current_user)):
    """获取当前用户信息。"""
    return success_response(data=current_user)
@router.put("/me/password")
async def change_password(
    request: Request,
    req: ChangePasswordReq,
    current_user: CurrentUser = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    """修改当前用户密码。修改成功后，该用户全部 Refresh Token 将被撤销。"""
    ip, _ = _extract_client_info(request)
    await change_password_service(
        db=db,
        user_id=current_user["id"],
        old_password=req.old_password,
        new_password=req.new_password,
        confirm_password=req.confirm_password,
        ip=ip,
    )
    return success_response(message="密码修改成功", data=None)
