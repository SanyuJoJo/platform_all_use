import request from '@/utils/request'

export interface UpgradeStatus {
  current_version: string
  install_root: string
  pending?: UpgradeManifest[]
  last_upgrade_at?: string
  last_upgrade_result?: string
  update_available: boolean
}

export interface UpgradeManifest {
  name: string
  from_version: string
  to_version: string
  min_platform_version: string
  channel: string
  requires: string
  created_at: string
  backend_artifact: string
  frontend_artifact: string
  backend_sha256: string
  frontend_sha256: string
}

export interface UpgradePackage {
  id: string
  from_version: string
  to_version: string
  channel: string
  requires: string
  created_at: string
  backend_artifact: string
  backend_sha256: string
  frontend_sha256: string
  verified: boolean
}

export const upgradeApi = {
  getStatus() {
    return request.get<UpgradeStatus>('/upgrade/status')
  },
  listPackages() {
    return request.get<{ items: UpgradePackage[] }>('/upgrade/packages')
  },
  run(packageId: string) {
    return request.post<UpgradeStatus>('/upgrade/run', { package_id: packageId })
  },
  rollback(id: string) {
    return request.post<UpgradeStatus>('/upgrade/rollback', { id })
  },
}