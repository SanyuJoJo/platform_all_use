import request from './index';
import type { LicenseStatus, ModuleAuthStatus, ApiResponse } from '@/types';
export const licenseApi = {
  getStatus() {
    return request.get<ApiResponse<LicenseStatus>>('/license/status');
  },
  importLicenseFile(file: File) {
    const formData = new FormData();
    formData.append('license_file', file);
    return request.post<ApiResponse<any>>('/license/import', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
  },
  importLicenseCode(activationCode: string) {
    return request.post<ApiResponse<any>>('/license/import', { activation_code: activationCode });
  },
  activate(activationCode: string, machineCode: string) {
    return request.post<ApiResponse<any>>('/license/activate', {
      activation_code: activationCode,
      machine_code: machineCode,
    });
  },
  getModulesAuth() {
    return request.get<ApiResponse<ModuleAuthStatus[]>>('/license/modules');
  },
};
