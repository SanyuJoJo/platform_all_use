import { MockMethod } from 'vite-plugin-mock';
const adminUser = {
  id: 1,
  username: 'admin',
  nickname: '管理员',
  email: 'admin@example.com',
  avatar: '',
  status: 1,
  roles: ['admin'],
  permissions: ['auth:user:view', 'auth:user:create', 'auth:role:view','dashboard:view'],
};
export default [
  {
    url: '/api/v1/auth/login',
    method: 'post',
    response: ({ body }) => {
      const { username, password } = body;
      if (username === 'admin' && password === '123456') {
        return {
          code: 0,
          message: '登录成功',
          data: {
            access_token: 'mock-jwt-token',
            refresh_token: 'mock-refresh-token',
            token_type: 'bearer',
            expires_in: 1440,
            user: adminUser,
          },
          timestamp: new Date().toISOString(),
          requestId: 'mock-request-id',
        };
      } else {
        return {
          code: 10001,
          message: '用户名或密码错误',
          data: null,
          timestamp: new Date().toISOString(),
          requestId: 'mock-request-id',
        };
      }
    },
  },
  {
    url: '/api/v1/auth/me',
    method: 'get',
    response: () => ({
      code: 0,
      message: 'success',
      data: adminUser,
      timestamp: new Date().toISOString(),
      requestId: 'mock-request-id',
    }),
  },
  {
    url: '/api/v1/modules',
    method: 'get',
    response: () => ({
      code: 0,
      message: 'success',
      data: {
        items: [
          {
            id: 'auth',
            name: '认证授权',
            version: '1.0.0',
            status: 'active',
            manifest: {
              menus: [
                {
                  id: 'auth:dashboard',
                  parent_id: null,
                  title: '认证授权',
                  icon: 'Lock',
                  path: '/dashboard',
                  permission: 'auth:user:view',
                  order: 10,
                },
                {
                  id: 'auth:users',
                  parent_id: 'auth:dashboard',
                  title: '用户管理',
                  icon: 'Person',
                  path: '/users',
                  permission: 'auth:user:view',
                  order: 10,
                },
                {
                  id: 'auth:roles',
                  parent_id: 'auth:dashboard',
                  title: '角色管理',
                  icon: 'Shield',
                  path: '/roles',
                  permission: 'auth:role:view',
                  order: 20,
                },
              ],
            },
          },
        ],
        total: 1,
        page: 1,
        page_size: 20,
        pages: 1,
      },
      timestamp: new Date().toISOString(),
      requestId: 'mock-request-id',
    }),
  },
] as MockMethod[];
