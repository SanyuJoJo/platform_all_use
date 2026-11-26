import request from '@/utils/request';

export const cryptoAuditApi = {
  listAudits(params?: Record<string, unknown>) {
    return request.get('/audits', { params });
  },
  getAudit(auditId: string) {
    return request.get(`/audits/${auditId}`);
  },
};