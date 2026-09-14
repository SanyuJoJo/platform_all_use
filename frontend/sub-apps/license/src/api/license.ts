import request from './index';
import type { LicenseStatus, ModuleAuthStatus } from '@/types';

export const licenseApi = {
  // 泛型只到业务数据结构，不再嵌套 ApiResponse
  getStatus() {
    return request.get<LicenseStatus>('/license/status');
  },
  importLicenseFile(file: File) {
    const formData = new FormData();
    formData.append('license_file', file);
    return request.post<unknown>('/license/import', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
  },
  importLicenseCode(activationCode: string) {
    return request.post<unknown>('/license/import', {
      activation_code: activationCode,
    });
  },
  activate(activationCode: string, machineCode: string) {
    return request.post<unknown>('/license/activate', {
      activation_code: activationCode,
      machine_code: machineCode,
    });
  },
  getModulesAuth() {
    return request.get<ModuleAuthStatus[]>('/license/modules');
  },
};
