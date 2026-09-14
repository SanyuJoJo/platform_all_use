#!/usr/bin/env node
/**
 * 前端一键生产构建脚本
 *
 * 功能：
 * 1. 清理 deploy 目录
 * 2. 构建 main-app（走 package.json 的 build，含 vue-tsc）
 * 3. 构建 sub-apps 下所有非 _template 子应用（走 package.json 的 build，含 vue-tsc）
 * 4. 复制 dist 到 deploy
 * 5. 生成 deploy-manifest.json
 * 6. 生成 backend-module-entry.patch.json
 *
 * 约定：
 * - 子应用目录名（sub-apps 下的文件夹）与后端模块 ID 可能不同。
 *   例如目录 `module-manager` 对应后端模块 ID `module_manager`。
 *   此处维护映射表，新增子应用时需在此登记。
 */
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { execSync } from 'node:child_process';
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const ROOT = path.resolve(__dirname, '..');
const DEPLOY_DIR = path.join(ROOT, 'deploy');
const SUB_APPS_DIR = path.join(ROOT, 'sub-apps');
/**
 * 子应用目录名 → 后端模块 ID 映射
 * 新增子应用时需同步登记，否则默认将目录名中的 `-` 替换为 `_`
 */
const SUB_APP_MODULE_ID_MAP = {
  auth: 'auth',
  'module-manager': 'module_manager',
  'audit-log': 'audit_log',
  license: 'license',
};
function log(msg) {
  console.log(`\n[build-all] ${msg}`);
}
function run(cmd, options = {}) {
  log(`执行：${cmd}`);
  execSync(cmd, {
    cwd: ROOT,
    stdio: 'inherit',
    env: {
      ...process.env,
      ...options.env,
    },
  });
}
function rmrf(target) {
  fs.rmSync(target, { recursive: true, force: true });
}
function ensureDir(dir) {
  fs.mkdirSync(dir, { recursive: true });
}
function copyDir(src, dest) {
  if (!fs.existsSync(src)) {
    throw new Error(`源目录不存在：${src}`);
  }
  rmrf(dest);
  fs.cpSync(src, dest, { recursive: true });
}
function readJson(file) {
  return JSON.parse(fs.readFileSync(file, 'utf-8'));
}
function getSubApps() {
  if (!fs.existsSync(SUB_APPS_DIR)) return [];
  return fs
    .readdirSync(SUB_APPS_DIR)
    .filter((name) => {
      if (name.startsWith('_')) return false;
      const full = path.join(SUB_APPS_DIR, name);
      return (
        fs.statSync(full).isDirectory() &&
        fs.existsSync(path.join(full, 'package.json'))
      );
    })
    .map((dir) => {
      const pkgPath = path.join(SUB_APPS_DIR, dir, 'package.json');
      const pkg = readJson(pkgPath);
      const moduleId = SUB_APP_MODULE_ID_MAP[dir] || dir.replace(/-/g, '_');
      return {
        dir,
        pkgName: pkg.name,
        moduleId,
        entry: `/sub-apps/${dir}/`,
        activeRule: `/${moduleId}`,
        staticBase: `/sub-apps/${dir}/`,
      };
    });
}
function main() {
  log('开始前端生产构建');
  rmrf(DEPLOY_DIR);
  ensureDir(DEPLOY_DIR);
  // ============================================================
  // 1. 构建主应用（走 package.json 的 build，含 vue-tsc 类型检查）
  // ============================================================
  run('pnpm --filter main-app build', {
    env: {
      VITE_BASE_PATH: '/',
      VITE_API_BASE_URL: '',
      VITE_USE_MOCK: 'false',
    },
  });
  copyDir(
    path.join(ROOT, 'main-app', 'dist'),
    path.join(DEPLOY_DIR, 'main-app')
  );
  // ============================================================
  // 2. 构建子应用（走各子应用 package.json 的 build，含 vue-tsc 类型检查）
  // ============================================================
  const subApps = getSubApps();
  if (subApps.length === 0) {
    log('未发现子应用，跳过子应用构建');
  }
  const deploySubApps = [];
  for (const app of subApps) {
    const publicPath = `/sub-apps/${app.dir}/`;
    log(`构建子应用：${app.pkgName}（模块ID=${app.moduleId}），base=${publicPath}`);
    run(`pnpm --filter ${app.pkgName} build`, {
      env: {
        VITE_PUBLIC_PATH: publicPath,
        VITE_MODULE_ID: app.moduleId,
        VITE_SUBAPP_DIR: app.dir,
        VITE_QIANKUN_USE_DEV_MODE: 'false',
        VITE_API_BASE_URL: '',
        VITE_USE_MOCK: 'false',
      },
    });
    const distDir = path.join(SUB_APPS_DIR, app.dir, 'dist');
    const targetDir = path.join(DEPLOY_DIR, 'sub-apps', app.dir);
    copyDir(distDir, targetDir);
    deploySubApps.push({
      dir: app.dir,
      pkgName: app.pkgName,
      moduleId: app.moduleId,
      entry: app.entry,
      activeRule: app.activeRule,
      staticBase: app.staticBase,
    });
  }
  // ============================================================
  // 3. 生成 deploy-manifest.json
  // ============================================================
  const manifest = {
    generatedAt: new Date().toISOString(),
    mainApp: {
      dir: 'main-app',
      base: '/',
      index: '/index.html',
    },
    subApps: deploySubApps,
    fastapi: {
      mainAppMount: '/',
      subAppsMount: '/sub-apps',
      apiPrefix: '/api/v1',
      spaFallback: true,
    },
  };
  fs.writeFileSync(
    path.join(DEPLOY_DIR, 'deploy-manifest.json'),
    JSON.stringify(manifest, null, 2),
    'utf-8'
  );
  // ============================================================
  // 4. 生成后端模块 entry_frontend 更新参考
  // ============================================================
  const patch = deploySubApps.map((app) => ({
    module_id: app.moduleId,
    entry_frontend: app.entry,
    active_rule: app.activeRule,
    static_base: app.staticBase,
  }));
  fs.writeFileSync(
    path.join(DEPLOY_DIR, 'backend-module-entry.patch.json'),
    JSON.stringify(patch, null, 2),
    'utf-8'
  );
  log('前端生产构建完成');
  console.log(`\n产物目录：${DEPLOY_DIR}`);
  console.log(`主应用：/`);
  deploySubApps.forEach((app) => {
    console.log(`子应用 ${app.moduleId}：${app.entry}`);
  });
}
main();
