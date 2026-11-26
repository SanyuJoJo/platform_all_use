# Qiankun → TDesign 单应用迁移约定

本文件记录 `frontend-qiankun/`（naive-ui + qiankun 微前端）迁移到 `frontend/`
（Vue3 + TDesign 单体应用）的约定。迁移业务模块时请严格遵守。

## 1. 目标结构

所有子应用页面合并进 `frontend/src/views/`，路径必须与后端菜单的
`component` 字段一一对应（后端 `GET /api/v1/menus/routers` 下发）：

| 后端 component | 新文件 |
| --- | --- |
| `dashboard/index` | `src/views/dashboard/index.vue` |
| `system/user/index` | `src/views/system/user/index.vue` |
| `system/role/index` | `src/views/system/role/index.vue` |
| `system/permission/index` | `src/views/system/permission/index.vue` |
| `system/menu/index` | `src/views/system/menu/index.vue` |
| `modules/index` | `src/views/modules/index.vue` |
| `logs/index` | `src/views/logs/index.vue` |
| `license/index` | `src/views/license/index.vue` |
| `ca/index` | `src/views/ca/index.vue` |
| `certs/index` | `src/views/certs/index.vue` |
| `csr/index` | `src/views/csr/index.vue` |
| `crl/index` | `src/views/crl/index.vue` |
| `keys/index` | `src/views/keys/index.vue` |
| `tasks/index` | `src/views/tasks/index.vue` |
| `audits/index` | `src/views/audits/index.vue` |
| `list/index` | `src/views/list/index.vue` |
| `tcps/list/index` | `src/views/tcps/list/index.vue` |
| `media/index` | `src/views/media/index.vue` |
| `live/index` | `src/views/live/index.vue` |

路由由 `src/router/dynamic.ts` 根据 `component` 自动解析，**不要**再为每个模块
创建 router 文件、qiankun 生命周期或 `createWebHistory`。

## 2. 共享基础设施（禁止修改）

- `src/utils/request.ts`：全局 axios 实例，已处理 token、401、`code !== 0`。
  返回 `ApiResponse<T>`（`res.data` 为业务数据）；失败 reject
  `{ code, message, data }`，用 `error.message` 提示。
- `src/store/user.ts`：`token / user / permissions / hasPermission / login / logout`。
- `src/store/permission.ts` + `src/store/menu.ts`：动态路由与侧边栏菜单。
- `src/router/index.ts`、`src/router/dynamic.ts`、`src/layouts/enterprise/*`、
  `src/components/AppMenu.vue`、`src/utils/menu.ts`、`src/utils/icon-map.ts`、
  `src/main.ts`、`src/directives/permission.ts`、`src/composables/usePermission.ts`。
- `v-permission="'auth:user:create'"` 已全局注册，可直接使用。

需要新增共享能力时，在最终汇报里说明，不要自行修改上述文件。

## 3. naive-ui → TDesign 对照

| naive-ui | TDesign |
| --- | --- |
| `n-space` | `t-space`（`direction="vertical"`，`size`） |
| `n-button type="primary"` | `t-button theme="primary"` |
| `n-button text` | `t-button variant="text"` |
| `n-input v-model:value` | `t-input v-model` |
| `n-input type="textarea"` | `t-textarea v-model` |
| `n-input-number v-model:value` | `t-input-number v-model` |
| `n-select v-model:value :options` | `t-select v-model :options` |
| `n-radio-group` / `n-radio` | `t-radio-group` / `t-radio`（`value`） |
| `n-checkbox` | `t-checkbox` |
| `n-switch v-model:value` | `t-switch v-model` |
| `n-data-table :columns :data` | `t-table row-key="id" :columns :data` |
| `n-modal preset="dialog"` | `t-dialog v-model:visible :header` |
| `n-modal preset="card"` | `t-dialog v-model:visible :header :footer="false"` 或默认 footer |
| `n-drawer` | `t-drawer v-model:visible` |
| `n-form :model` / `n-form-item path` | `t-form :data` / `t-form-item name` |
| `n-tag type` | `t-tag theme`（`default/success/warning/danger/primary`） |
| `n-popconfirm @positive-click` | `t-popconfirm content="..." @confirm` |
| `n-alert type` | `t-alert theme`（`info/success/warning/error`） |
| `n-spin :show` | `t-loading :loading`（或用 `v-loading`） |
| `n-tabs` / `n-tab-pane` | `t-tabs v-model` / `t-tab-panel value label` |
| `n-descriptions` / `n-descriptions-item` | `t-descriptions` / `t-descriptions-item` |
| `n-tree` | `t-tree`（`checkStrictly`、`keys`、`#label`） |
| `n-tree-select` | `t-tree-select` |
| `n-upload` | `t-upload`（自定义分片上传时可用原生 `input type=file`） |
| `n-empty` | `t-empty` |
| `n-pagination` | 表格内置 `:pagination` |
| `n-tooltip` | `t-tooltip` |
| `n-divider` | `t-divider` |
| `n-icon` + `@vicons` | `t-icon name="..."` |
| `useMessage()` | `MessagePlugin` from `tdesign-vue-next` |
| `useDialog()` | `DialogPlugin` / `Dialog` 组件 |
| `n-dropdown` | `t-dropdown :options` |

### 表单校验

```ts
const formRef = ref<FormInstanceFunctions>()
const result = await formRef.value?.validate()
if (result !== true) return
```

规则写在 `t-form :rules`，字段用 `t-form-item name="field"`；
自定义校验返回 `true` 或 `{ result: false, message: '...' }`。

### 表格渲染函数

TDesign Vue Next 表格列没有通用 cell 插槽，沿用 `columns` 的 `render`：

```ts
import { h } from 'vue'
import { Tag, Button, Popconfirm, Space } from 'tdesign-vue-next'
import type { PrimaryTableCol } from 'tdesign-vue-next'

const columns: PrimaryTableCol<User>[] = [
  { colKey: 'id', title: 'ID', width: 80 },
  {
    colKey: 'status',
    title: '状态',
    cell: (_h, { row }) =>
      h(Tag, { theme: row.status === 1 ? 'success' : 'danger', variant: 'light' },
        () => (row.status === 1 ? '启用' : '禁用')),
  },
  {
    colKey: 'actions',
    title: '操作',
    cell: (_h, { row }) =>
      h(Space, { size: 4 }, () => [
        h(Button, { size: 'small', variant: 'text', theme: 'primary',
          onClick: () => handleEdit(row) }, () => '编辑'),
        h(Popconfirm, { content: '确认删除？', onConfirm: () => handleDelete(row) },
          () => h(Button, { size: 'small', variant: 'text', theme: 'danger' }, () => '删除')),
      ]),
  },
]
```

### 分页

```vue
<t-table
  row-key="id"
  :data="rows"
  :columns="columns"
  :loading="loading"
  :pagination="{ current: page, pageSize, total, showJumper: true }"
  @page-change="onPageChange"
/>
```

```ts
function onPageChange(info: PageInfo) {
  page.value = info.current
  pageSize.value = info.pageSize
  fetchList()
}
```

## 4. API / 类型 / 状态

- 所有 API 模块统一 `import request from '@/utils/request'`。
- 类型放 `src/types/<module>.ts`，并从 `src/types/index.ts` 重新导出（如需要）。
- 跨页面状态放 `src/store/<name>.ts`（Pinia setup store）。
- 不要保留任何 `qiankun`、`vite-plugin-qiankun`、`__POWERED_BY_QIANKUN__`、
  `props.routers`、`onGlobalStateChange` 相关代码。
- 子应用自己的 `403/404` 页面不再需要（主应用统一处理）。

## 5. 迁移原则

1. **行为不变**：接口地址、请求参数、业务分支、权限码、错误码处理全部保留。
2. **只换 UI**：naive-ui 组件替换为 TDesign，布局/交互可小幅优化，但不要删功能。
3. **中文文案保持**，风格与现有页面一致（`t-card` + 工具栏 + `t-table`）。
4. 页面标题统一使用 `t-card` 的 `title`，不再手写 `<h1>`。
5. 提交前自检：`grep -r "naive-ui" src/` 必须无结果。
