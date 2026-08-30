import request from './index';
export async function getModules() {
  const res = await request.get('/modules?status=active');
  const items = res.data?.items || res.data;
  return Array.isArray(items) ? items : [];
}
