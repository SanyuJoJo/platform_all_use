<template>
  <t-dialog
    v-model:visible="showRef"
    :header="editing ? '编辑菜单' : '新增菜单'"
    :confirm-btn="null"
    :cancel-btn="null"
    :width="760"
    :mask-closable="false"
  >
    <t-form
      ref="formRef"
      :data="form"
      :rules="rules"
      label-align="left"
      :label-width="100"
    >
      <t-row :gutter="16">
        <t-col :span="12">
          <t-form-item label="上级菜单" name="parentId">
            <t-tree-select
              v-model="form.parentId"
              :data="parentOptions"
              :keys="treeKeys"
              placeholder="选择上级菜单"
              clearable
              @change="handleParentChange"
            />
          </t-form-item>
        </t-col>
        <t-col :span="12">
          <t-form-item label="菜单类型" name="menuType">
            <t-select
              v-model="form.menuType"
              :options="menuTypeOptions"
              :disabled="lockMenuType"
            />
          </t-form-item>
        </t-col>
      </t-row>

      <t-alert
        v-if="parentIsMenu && form.menuType !== 'F'"
        theme="warning"
        message="上级是「菜单」类型，只能新增「按钮」，已将类型自动切换为按钮。"
        style="margin-bottom: 12px"
      />

      <t-alert
        v-if="form.menuType === 'F'"
        theme="info"
        message="按钮不生成路由、不出现在侧边栏。它只承载权限标识，用于前端 v-permission 控制按钮显隐，以及后端接口权限校验。"
        style="margin-bottom: 12px"
      />

      <t-row :gutter="16">
        <t-col :span="12">
          <t-form-item label="菜单名称" name="menuName">
            <t-input
              v-model="form.menuName"
              :placeholder="form.menuType === 'F' ? '如：新增用户' : '如：用户管理'"
            />
          </t-form-item>
        </t-col>
        <t-col :span="12">
          <t-form-item label="路由名称" name="name">
            <t-input
              v-model="form.name"
              :disabled="form.menuType === 'F'"
              placeholder="英文 PascalCase（留空则按路由地址派生）"
            />
          </t-form-item>
        </t-col>
      </t-row>

      <t-row v-if="form.menuType !== 'F'" :gutter="16">
        <t-col :span="12">
          <t-form-item label="路由地址" name="path">
            <t-input v-model="form.path" placeholder="如：/system/users" />
          </t-form-item>
        </t-col>
        <t-col :span="12">
          <t-form-item label="组件路径" name="component">
            <t-input v-model="form.component" placeholder="如：system/user/index" />
          </t-form-item>
        </t-col>
      </t-row>

      <t-row v-if="form.menuType !== 'F'" :gutter="16">
        <t-col :span="12">
          <t-form-item label="微应用" name="microApp">
            <t-select
              v-model="form.microApp"
              :options="microAppOptions"
              :loading="microAppLoading"
              filterable
              clearable
              placeholder="选择或输入 microApp id"
              @change="handleMicroAppChange"
            />
          </t-form-item>
        </t-col>
        <t-col :span="12">
          <t-form-item label="模块 ID" name="moduleId">
            <t-select
              v-model="form.moduleId"
              :options="moduleIdOptions"
              :loading="microAppLoading"
              filterable
              clearable
              placeholder="选择或输入模块 ID"
            />
          </t-form-item>
        </t-col>
      </t-row>

      <t-alert
        v-if="microAppTip && form.menuType !== 'F'"
        theme="info"
        :message="microAppTip"
        style="margin-bottom: 12px"
      />

      <t-row :gutter="16">
        <t-col :span="12">
          <t-form-item label="权限标识" name="perms">
            <t-input
              v-model="form.perms"
              :placeholder="form.menuType === 'F' ? '如：auth:user:create' : '如：auth:user:list（进入页面所需权限码）'"
            />
          </t-form-item>
        </t-col>
        <t-col v-if="form.menuType !== 'F'" :span="12">
          <t-form-item label="图标" name="icon">
            <t-select
              v-model="form.icon"
              :options="iconOptions"
              filterable
              clearable
              placeholder="选择或输入图标名"
            />
          </t-form-item>
        </t-col>
      </t-row>

      <t-row :gutter="16">
        <t-col :span="8">
          <t-form-item label="显示排序" name="orderNum">
            <t-input-number v-model="form.orderNum" :min="0" style="width: 100%" />
          </t-form-item>
        </t-col>
        <t-col v-if="form.menuType !== 'F'" :span="8">
          <t-form-item label="显示状态" name="visible">
            <t-select v-model="form.visible" :options="visibleOptions" />
          </t-form-item>
        </t-col>
        <t-col :span="8">
          <t-form-item label="菜单状态" name="status">
            <t-select v-model="form.status" :options="statusOptions" />
          </t-form-item>
        </t-col>
      </t-row>

      <t-row v-if="form.menuType !== 'F'" :gutter="16">
        <t-col :span="12">
          <t-form-item label="是否外链" name="isFrame">
            <t-select v-model="form.isFrame" :options="frameOptions" />
          </t-form-item>
        </t-col>
        <t-col :span="12">
          <t-form-item label="是否缓存" name="isCache">
            <t-select v-model="form.isCache" :options="cacheOptions" />
          </t-form-item>
        </t-col>
      </t-row>

      <t-form-item label="备注" name="remark">
        <t-textarea v-model="form.remark" :autosize="{ minRows: 2 }" placeholder="备注" />
      </t-form-item>
    </t-form>
    <template #footer>
      <div style="text-align: right">
        <t-space>
          <t-button @click="close">取消</t-button>
          <t-button theme="primary" :loading="submitting" @click="submit">保存</t-button>
        </t-space>
      </div>
    </template>
  </t-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, type PropType } from 'vue';
import { MessagePlugin, type FormInstanceFunctions } from 'tdesign-vue-next';
import {
  createMenu,
  updateMenu,
  listKnownMicroApps,
  type Menu,
  type MenuForm,
  type MicroAppOption,
} from '@/api/system-menu';

const props = defineProps({
  show: { type: Boolean, default: false },
  editing: { type: Object as PropType<Menu | null>, default: null },
  allMenus: { type: Array as PropType<Menu[]>, default: () => [] },
  defaultMenuType: {
    type: String as PropType<'M' | 'C' | 'F' | undefined>,
    default: undefined,
  },
  lockMenuType: { type: Boolean, default: false },
});

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void;
  (e: 'success'): void;
}>();

const formRef = ref<FormInstanceFunctions>();
const submitting = ref(false);

const treeKeys = { value: 'key', label: 'label', children: 'children' };

const knownMicroApps = ref<MicroAppOption[]>([]);
const microAppLoading = ref(false);
const MICRO_APP_RE = /^[a-z][a-z0-9_-]{0,49}$/;

function collectFromMenus(menus: Menu[]) {
  const apps = new Set<string>();
  const mods = new Set<string>();
  const walk = (list: Menu[]) => {
    list.forEach((m) => {
      if (m.microApp && m.microApp !== 'main' && m.microApp !== 'platform') {
        apps.add(m.microApp);
      }
      if (m.moduleId) mods.add(m.moduleId);
      if (m.children?.length) walk(m.children);
    });
  };
  walk(menus);
  return { apps: Array.from(apps), mods: Array.from(mods) };
}

async function loadKnownMicroApps() {
  microAppLoading.value = true;
  try {
    const res: any = await listKnownMicroApps();
    const items = res?.data?.items ?? [];
    if (Array.isArray(items) && items.length > 0) {
      knownMicroApps.value = items as MicroAppOption[];
      return;
    }
    throw new Error('接口返回空');
  } catch (e) {
    console.warn('[MenuFormModal] /menus/micro-apps 失败，降级从 allMenus 推断', e);
    const builtin: MicroAppOption[] = [
      { id: 'main', name: '主应用', source: 'builtin' },
      { id: 'platform', name: '平台基础', source: 'builtin' },
    ];
    const { apps } = collectFromMenus(props.allMenus);
    const fromHistory: MicroAppOption[] = apps.map((id) => ({
      id,
      name: id,
      source: 'menu_history',
    }));
    knownMicroApps.value = [...builtin, ...fromHistory];
  } finally {
    microAppLoading.value = false;
  }
}

onMounted(loadKnownMicroApps);

const microAppOptions = computed(() =>
  knownMicroApps.value.map((m) => ({
    label:
      m.source === 'menu_history' ? `${m.id}（历史菜单）` : `${m.name} (${m.id})`,
    value: m.id,
  })),
);

const moduleIdOptions = computed(() => {
  const seen = new Set<string>();
  const out: Array<{ label: string; value: string }> = [];
  knownMicroApps.value.forEach((m) => {
    if (!seen.has(m.id)) {
      seen.add(m.id);
      out.push({ label: `${m.name} (${m.id})`, value: m.id });
    }
  });
  const { mods } = collectFromMenus(props.allMenus);
  mods.forEach((id) => {
    if (!seen.has(id)) {
      seen.add(id);
      out.push({ label: `${id}（历史模块）`, value: id });
    }
  });
  return out;
});

const microAppTip = computed(() => {
  const v = (form.value.microApp || '').trim();
  if (!v || form.value.menuType === 'F') return '';
  if (knownMicroApps.value.some((m) => m.id === v)) return '';
  return `「${v}」尚未在模块列表或历史菜单中出现。保存后菜单会保留该 id，等对应子应用部署并接入后即可正常跳转。`;
});

function handleMicroAppChange(v: string) {
  if (!form.value.moduleId && v && v !== 'main' && v !== 'platform') {
    form.value.moduleId = v;
  }
}

const showRef = computed({
  get: () => props.show,
  set: (v) => emit('update:show', v),
});

const editing = computed(() => !!(props.editing && props.editing.id));

function emptyForm(): MenuForm {
  return {
    parentId: 0,
    menuName: '',
    name: '',
    microApp: 'main',
    orderNum: 0,
    path: '',
    component: '',
    query: '',
    isFrame: 1,
    isCache: 0,
    menuType: 'C',
    visible: 0,
    status: 0,
    perms: '',
    icon: '#',
    moduleId: '',
    remark: '',
  };
}

const form = ref<MenuForm>(emptyForm());

const parentIsMenu = computed(() => {
  if (!form.value.parentId) return false;
  const find = (list: Menu[]): Menu | null => {
    for (const m of list) {
      if (m.id === form.value.parentId) return m;
      if (m.children?.length) {
        const r = find(m.children);
        if (r) return r;
      }
    }
    return null;
  };
  const p = find(props.allMenus);
  return p?.menuType === 'C';
});

watch(
  () => props.show,
  (v) => {
    if (!v) return;
    if (props.editing && props.editing.id) {
      form.value = {
        parentId: props.editing.parentId ?? 0,
        menuName: props.editing.menuName ?? '',
        name: props.editing.name ?? '',
        microApp: props.editing.microApp ?? 'main',
        orderNum: props.editing.orderNum ?? 0,
        path: props.editing.path ?? '',
        component: props.editing.component ?? '',
        query: props.editing.query ?? '',
        isFrame: props.editing.isFrame ?? 1,
        isCache: props.editing.isCache ?? 0,
        menuType: props.editing.menuType,
        visible: props.editing.visible ?? 0,
        status: props.editing.status ?? 0,
        perms: props.editing.perms ?? '',
        icon: props.editing.icon ?? '#',
        moduleId: props.editing.moduleId ?? '',
        remark: props.editing.remark ?? '',
      };
    } else {
      form.value = emptyForm();
      if (props.defaultMenuType) {
        form.value.menuType = props.defaultMenuType;
      }
      if (props.editing && props.editing.parentId) {
        form.value.parentId = props.editing.parentId;
      }
    }
  },
);

function handleParentChange(parentId: number | null) {
  if (!parentId) return;
  const find = (list: Menu[]): Menu | null => {
    for (const m of list) {
      if (m.id === parentId) return m;
      if (m.children?.length) {
        const r = find(m.children);
        if (r) return r;
      }
    }
    return null;
  };
  const p = find(props.allMenus);
  if (p?.menuType === 'C' && form.value.menuType !== 'F') {
    MessagePlugin.warning('上级是菜单类型，只能新增按钮，已自动切换为「按钮」');
    form.value.menuType = 'F';
  }
}

const rules = {
  menuName: [{ required: true, message: '请输入菜单名称', trigger: 'blur' }],
  menuType: [
    {
      required: true,
      trigger: 'change',
      validator: (_rule: unknown, value: string) => {
        if (!value) return { result: false, message: '请选择菜单类型' };
        if (parentIsMenu.value && value !== 'F') {
          return { result: false, message: '上级是菜单类型，只能新增按钮' };
        }
        return true;
      },
    },
  ],
  microApp: [
    {
      required: true,
      trigger: 'change',
      validator: (_rule: unknown, value: string) => {
        if (form.value.menuType === 'F') return true;
        const v = (value || '').trim();
        if (!v) return { result: false, message: '请选择或输入微应用 id' };
        if (!MICRO_APP_RE.test(v)) {
          return {
            result: false,
            message: 'microApp 格式非法：需小写字母开头，仅含小写字母/数字/下划线/连字符，长度 ≤50',
          };
        }
        return true;
      },
    },
  ],
  path: [
    {
      trigger: 'blur',
      validator: (_rule: unknown, value: string) => {
        if (form.value.menuType !== 'F' && !value) {
          return { result: false, message: '目录/菜单类型必须填写路由地址' };
        }
        return true;
      },
    },
  ],
  perms: [
    {
      trigger: 'blur',
      validator: (_rule: unknown, value: string) => {
        if (form.value.menuType === 'F' && !value) {
          return { result: false, message: '按钮类型必须填写权限标识' };
        }
        return true;
      },
    },
  ],
};

const menuTypeOptions = [
  { label: '目录 (M)', value: 'M' },
  { label: '菜单 (C)', value: 'C' },
  { label: '按钮 (F)', value: 'F' },
];

const iconOptions = [
  'Odometer', 'Dashboard', 'Setting', 'Settings', 'User', 'UserFilled',
  'People', 'Person', 'Key', 'Menu', 'Grid', 'List', 'Document',
  'DocumentText', 'File', 'Lock', 'Shield', 'Connection', 'Ribbon',
  'Apps', 'Search', 'Time', 'Home', 'Certificate', 'Cloud', 'Cube',
  'Globe', 'Layers', 'Notification', 'Server', 'Speedometer',
].map((i) => ({ label: i, value: i }));

const visibleOptions = [
  { label: '显示', value: 0 },
  { label: '隐藏', value: 1 },
];
const statusOptions = [
  { label: '正常', value: 0 },
  { label: '停用', value: 1 },
];
const frameOptions = [
  { label: '是（外链）', value: 0 },
  { label: '否（内部）', value: 1 },
];
const cacheOptions = [
  { label: '缓存', value: 0 },
  { label: '不缓存', value: 1 },
];

interface TreeOption {
  key: number;
  label: string;
  children?: TreeOption[];
  [key: string]: unknown;
}

const parentOptions = computed<TreeOption[]>(() => {
  const excludeId = props.editing?.id ?? 0;
  const build = (menus: Menu[]): TreeOption[] => {
    const out: TreeOption[] = [];
    for (const m of menus) {
      if (m.id === excludeId) continue;
      if (m.menuType === 'F') continue;
      out.push({
        key: m.id,
        label: m.menuType === 'C' ? `${m.menuName}（菜单，仅可加按钮）` : m.menuName,
        children:
          m.children && m.children.length > 0 ? build(m.children) : undefined,
      });
    }
    return out;
  };
  return [{ key: 0, label: '根目录', children: build(props.allMenus) }];
});

function close() {
  showRef.value = false;
}

async function submit() {
  const result = await formRef.value?.validate();
  if (result !== true) return;
  if (parentIsMenu.value && form.value.menuType !== 'F') {
    MessagePlugin.error('上级是菜单类型，只能新增按钮');
    return;
  }
  submitting.value = true;
  try {
    if (editing.value && props.editing?.id) {
      await updateMenu(props.editing.id, form.value);
      MessagePlugin.success('修改成功');
    } else {
      await createMenu(form.value);
      MessagePlugin.success('新增成功');
    }
    emit('success');
  } catch (e: any) {
    MessagePlugin.error(e?.message || '保存失败');
  } finally {
    submitting.value = false;
  }
}
</script>