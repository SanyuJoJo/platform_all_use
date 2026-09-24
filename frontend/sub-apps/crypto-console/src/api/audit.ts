import request from './index';
export const auditApi = {
  listAudits(params?: Record<string, unknown>) {
    return request.get('/audits', { params });
  },
  getAudit(auditId: string) {
    return request.get(`/audits/${auditId}`);
  },
};
