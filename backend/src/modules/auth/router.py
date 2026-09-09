from fastapi import APIRouter, Depends, HTTPException, status
from fastapi.security import OAuth2PasswordBearer, OAuth2PasswordRequestForm
from src.core.dependencies import get_current_user
from src.core.response import success_response
from src.core.exceptions import PlatformException
from .schemas import LoginReq, TokenResp, UserInfo
from .service import authenticate_user, create_tokens_for_user, refresh_access_token, get_user_info
router = APIRouter(prefix="/api/v1/auth", tags=["Auth"])
@router.post("/login")
async def login(req: LoginReq):
    user = authenticate_user(req.username, req.password)
    access_token, refresh_token = create_tokens_for_user(user)
    user_info = UserInfo(
        id=user["id"],
        username=user["username"],
        nickname=user["nickname"],
        email=user.get("email"),
        avatar=user.get("avatar"),
        status=user["status"],
        roles=user["roles"],
        permissions=user["permissions"]
    )
    return success_response(
        data={
            "access_token": access_token,
            "refresh_token": refresh_token,
            "token_type": "bearer",
            "expires_in": 1440,
            "user": user_info.dict()
        },
        message="登录成功"
    )
@router.post("/refresh")
async def refresh(refresh_token: str = Depends(OAuth2PasswordBearer(tokenUrl="/api/v1/auth/login", auto_error=True))):
    # 注意：这里直接从请求头获取 refresh_token，实际应用可能需要放在请求体
    # 简单起见，我们用依赖注入获取，但需注意 OAuth2PasswordBearer 期望的是 access_token
    # 我们改为从请求体或查询参数获取，或使用 Header
    # 但为了快速演示，我们从请求头 Authorization 中读取 refresh_token（但规范要求用 refresh_token 换 access_token）
    # 此处我们接收一个请求体：{"refresh_token": "xxx"} 更合适，但为了简化，我们使用查询参数
    pass
# 更规范的做法：用 POST /refresh 接收 json body
@router.post("/refresh")
async def refresh_token(req: dict):  # 简单接收 {"refresh_token": "..."}
    token = req.get("refresh_token")
    if not token:
        raise PlatformException(code=10001, message="缺少refresh_token", status_code=400)
    new_token = refresh_access_token(token)
    return success_response(data={"access_token": new_token, "expires_in": 1440})
@router.get("/me")
async def me(current_user: dict = Depends(get_current_user)):
    # get_current_user 已从 token 解析出用户信息（模拟）
    # 但我们需要完整用户信息，可调用 service 获取
    user = get_user_info(current_user["id"])
    if not user:
        raise PlatformException(code=10002, message="用户不存在", status_code=404)
    user_info = UserInfo(
        id=user["id"],
        username=user["username"],
        nickname=user["nickname"],
        email=user.get("email"),
        avatar=user.get("avatar"),
        status=user["status"],
        roles=user["roles"],
        permissions=user["permissions"]
    )
    return success_response(data=user_info.dict())
# 注意：为方便测试，也支持 /logout（可返回成功）
@router.post("/logout")
async def logout():
    return success_response(message="登出成功", data=None)

# 在现有 router 后追加以下代码
@router.put("/me/password")
async def change_password(
    req: dict,  # 实际应使用 Pydantic schema，为简洁直接使用 dict
    current_user: dict = Depends(get_current_user),
):
    old = req.get("old_password")
    new = req.get("new_password")
    confirm = req.get("confirm_password")
    if not old or not new or not confirm:
        raise PlatformException(code=90001, message="参数不完整", status_code=400)
    if new != confirm:
        raise PlatformException(code=10004, message="两次密码不一致", status_code=400)
    # 模拟验证旧密码（假设所有用户旧密码均为 "123456"）
    if old != "123456":
        raise PlatformException(code=10003, message="原密码错误", status_code=400)
    # 模拟更新（实际需更新数据库）
    return success_response(message="密码修改成功", data=None)
