/* ============================================================================
 * 模块管理类型
 * ==========================================================================*/

/** 模块对象（列表接口返回） */
export interface Module {
  id: string;
  name: string;
  version: string;
  description: string;
  author: string;
  status: 'active' | 'inactive';
  dependencies: string[];
  installed_at: string;
  updated_at: string;
}

/** 模块配置（键值对） */
export interface Config {
  [key: string]: any;
}

export interface InstallPayload {
  install_type: 'zip' | 'path';
  file_path?: string;
  source_path?: string;
}