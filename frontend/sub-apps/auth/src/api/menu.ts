import request from './index';

/** 菜单节点（与后端 models.Menu 对齐） */
export interface Menu {
  id: number;
  parentId: number;
  menuName: string;
  name: string;
  microApp: string;
  orderNum: number;
  path: string;
  component: string;
  query: string;
  /** 0=是外链，1=否 */
  isFrame: number;
  /** 0=缓存，1=不缓存 */
  isCache: number;
  menuType: 'M' | 'C' | 'F';
  /** 0=显示，1=隐藏 */
  visible: number;
  /** 0=正常，1=停用 */
  status: number;
  perms: string;
  icon: string;
  moduleId: string;
  remark: string;
  createBy?: string;
  createTime?: string;
  updateBy?: string;
  updateTime?: string;
  children?: Menu[];
}

/** 表单提交体（去掉 id、时间戳、children） */
export type MenuForm = Omit<
  Menu,
  'id' | 'createBy' | 'createTime' | 'updateBy' | 'updateTime' | 'children'
>;

export interface MenuQuery {
  menuName?: string;
  menuType?: string;
  status?: number;
  moduleId?: string;
  microApp?: string;
}

export function getMenuTree() {
  return request.get('/menus/tree');
}

export function listMenus(params: MenuQuery) {
  return request.get('/menus', { params });
}

export function getMenu(id: number) {
  return request.get(`/menus/detail/${id}`);
}

export function createMenu(data: MenuForm) {
  return request.post('/menus', data);
}

export function updateMenu(id: number, data: MenuForm) {
  return request.put(`/menus/detail/${id}`, data);
}

export function deleteMenu(id: number) {
  return request.delete(`/menus/detail/${id}`);
}

/** 微应用选项（与后端 ListKnownMicroApps 响应对齐） */
export interface MicroAppOption {
  id: string;
  name: string;
  /** builtin / module / menu_history */
  source: 'builtin' | 'module' | 'menu_history';
}
/** 拉取已知微应用列表（内置 + 已注册模块 + 历史菜单） */
export function listKnownMicroApps() {
  return request.get('/menus/micro-apps');
}
