/**
 * media 模块 API 统一出口。
 *
 * 复用全局请求实例（@/utils/request），保持与其它模块一致的
 * token 注入、401 跳转、业务错误处理行为。
 */
import request from '@/utils/request';

export default request;

export * from './media';