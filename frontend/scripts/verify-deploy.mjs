#!/usr/bin/env node
/**
 * 校验前端 deploy 产物是否完整
 *
 * 检查项：
 * 1. deploy 目录存在
 * 2. deploy-manifest.json 存在且合法
 * 3. backend-module-entry.patch.json 存在
 * 4. 主应用 index.html 存在
 * 5. 主应用 assets 目录存在
 * 6. 每个子应用 index.html 存在
 * 7. 每个子应用 assets 目录存在
 * 8. 每个子应用 index.html 中资源路径包含 /sub-apps/{dir}/assets/
 * 9. manifest 中每个子应用的 entry / staticBase 与实际目录匹配
 * 10. manifest 中 moduleId 与 backend-module-entry.patch.json 一一对应
 */
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const ROOT = path.resolve(__dirname, '..');
const DEPLOY_DIR = path.join(ROOT, 'deploy');
let hasError = false;
function fail(msg) {
  hasError = true;
  console.error(`❌ ${msg}`);
}
function ok(msg) {
  console.log(`✅ ${msg}`);
}
function warn(msg) {
  console.warn(`⚠️  ${msg}`);
}
function checkFile(file, label) {
  if (!fs.existsSync(file)) {
    fail(`${label}不存在：${file}`);
    return false;
  }
  return true;
}
function main() {
  console.log('[verify-deploy] 开始校验\n');
  // 1. deploy 目录
  if (!fs.existsSync(DEPLOY_DIR)) {
    fail(`deploy 目录不存在：${DEPLOY_DIR}`);
    process.exit(1);
  }
  ok('deploy 目录存在');
  // 2. deploy-manifest.json
  const manifestPath = path.join(DEPLOY_DIR, 'deploy-manifest.json');
  if (!checkFile(manifestPath, 'deploy-manifest.json')) {
    process.exit(1);
  }
  let manifest;
  try {
    manifest = JSON.parse(fs.readFileSync(manifestPath, 'utf-8'));
    ok('deploy-manifest.json 解析成功');
  } catch (e) {
    fail(`deploy-manifest.json 解析失败：${e.message}`);
    process.exit(1);
  }
  // 3. backend-module-entry.patch.json
  const patchPath = path.join(DEPLOY_DIR, 'backend-module-entry.patch.json');
  if (!checkFile(patchPath, 'backend-module-entry.patch.json')) {
    // 不直接退出，继续其他检查
  } else {
    ok('backend-module-entry.patch.json 存在');
    try {
      JSON.parse(fs.readFileSync(patchPath, 'utf-8'));
      ok('backend-module-entry.patch.json 解析成功');
    } catch (e) {
      fail(`backend-module-entry.patch.json 解析失败：${e.message}`);
    }
  }
  // 4. 主应用 index.html
  const mainIndex = path.join(DEPLOY_DIR, 'main-app', 'index.html');
  if (checkFile(mainIndex, '主应用 index.html')) {
    ok('主应用 index.html 存在');
  }
  // 5. 主应用 assets
  const mainAssets = path.join(DEPLOY_DIR, 'main-app', 'assets');
  if (!fs.existsSync(mainAssets)) {
    fail('主应用 main-app/assets 目录不存在');
  } else {
    ok('主应用 assets 目录存在');
  }
  // 6-8. 子应用逐个检查
  const subApps = manifest.subApps || [];
  if (subApps.length === 0) {
    warn('manifest 中未声明子应用，跳过子应用检查');
  }
  const manifestModuleIds = new Set();
  for (const app of subApps) {
    manifestModuleIds.add(app.moduleId);
    const appDir = path.join(DEPLOY_DIR, 'sub-apps', app.dir);
    const index = path.join(appDir, 'index.html');
    const assetsDir = path.join(appDir, 'assets');
    // index.html
    if (checkFile(index, `子应用 ${app.moduleId} index.html`)) {
      ok(`子应用 ${app.moduleId} index.html 存在`);
      // 8. 检查资源路径
      const html = fs.readFileSync(index, 'utf-8');
      const expectedPrefix = `/sub-apps/${app.dir}/assets/`;
      // 允许绝对或相对路径两种写法，但至少应包含 /sub-apps/{dir}/
      const looseCheck = new RegExp(
        `/sub-apps/${app.dir.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}/`
      );
      if (!html.includes(expectedPrefix) && !looseCheck.test(html)) {
        fail(
          `子应用 ${app.moduleId} index.html 资源路径可能不正确，应包含 ${expectedPrefix}`
        );
      } else {
        ok(`子应用 ${app.moduleId} 资源路径正确`);
      }
    }
    // assets
    if (!fs.existsSync(assetsDir)) {
      fail(`子应用 ${app.moduleId} assets 目录不存在：${assetsDir}`);
    } else {
      ok(`子应用 ${app.moduleId} assets 目录存在`);
    }
    // 9. entry / staticBase 与实际目录一致
    const expectedEntry = `/sub-apps/${app.dir}/`;
    if (app.entry !== expectedEntry) {
      fail(
        `子应用 ${app.moduleId} entry=${app.entry}，与目录 ${app.dir} 不一致（期望 ${expectedEntry}）`
      );
    } else {
      ok(`子应用 ${app.moduleId} entry 与目录一致`);
    }
    if (app.staticBase && app.staticBase !== expectedEntry) {
      fail(
        `子应用 ${app.moduleId} staticBase=${app.staticBase}，与目录 ${app.dir} 不一致`
      );
    }
    // activeRule 合法性
    if (!app.activeRule || !app.activeRule.startsWith('/')) {
      fail(`子应用 ${app.moduleId} activeRule 不合法：${app.activeRule}`);
    }
  }
  // 10. patch.json 与 manifest 一一对应
  if (fs.existsSync(patchPath)) {
    try {
      const patch = JSON.parse(fs.readFileSync(patchPath, 'utf-8'));
      const patchIds = new Set((patch || []).map((p) => p.module_id));
      for (const id of manifestModuleIds) {
        if (!patchIds.has(id)) {
          fail(`patch.json 缺少模块 ${id}`);
        }
      }
      for (const id of patchIds) {
        if (!manifestModuleIds.has(id)) {
          fail(`patch.json 存在 manifest 未声明的模块 ${id}`);
        }
      }
      if (patchIds.size === manifestModuleIds.size) {
        ok('manifest 与 patch.json 模块 ID 一一对应');
      }
    } catch (e) {
      fail(`patch.json 校验失败：${e.message}`);
    }
  }
  // 汇总
  console.log('\n[verify-deploy] 校验完成');
  if (hasError) {
    console.error('\n❌ 校验失败，请检查以上错误');
    process.exit(1);
  } else {
    console.log('\n✅ 全部检查通过');
  }
}
main();
