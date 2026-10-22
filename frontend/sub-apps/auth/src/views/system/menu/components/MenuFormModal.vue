<template>
  <n-modal
    v-model:show="showRef"
    preset="card"
    :title="editing ? '编辑菜单' : '新增菜单'"
    style="width: 760px;"
    :mask-closable="false"
  >
    <n-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-placement="left"
      label-width="100"
    >
      <n-grid :cols="2" :x-gap="16">
        <n-form-item-gi label="上级菜单" path="parentId">
          <n-tree-select
            v-model:value="form.parentId"
            :options="parentOptions"
            key-field="key"
            label-field="label"
            children-field="children"
            placeholder="选择上级菜单"
            clearable
            @update:value="handleParentChange"
          />
        </n-form-item-gi>
        <n-form-item-gi label="菜单类型" path="menuType">
          <n-select
            v-model:value="form.menuType"
            :options="menuTypeOptions"
            :disabled="lockMenuType"
          />
        </n-form-item-gi>
      </n-grid>

      <n-alert
        v-if="parentIsMenu && form.menuType !== 'F'"
        type="warning"
        :bordered="false"
        style="margin-bottom: 12px;"
      >
        上级是「菜单」类型，只能新增「按钮」，已将类型自动切换为按钮。
      </n-alert>

      <n-alert
        v-if="form.menuType === 'F'"
        type="info"
        :bordered="false"
        style="margin-bottom: 12px;"
      >
        按钮不生成路由、不出现在侧边栏。它只承载权限标识，用于前端
        <code>v-permission</code> 控制按钮显隐，以及后端接口权限校验。
      </n-alert>

      <n-grid :cols="2" :x-gap="16">
        <n-form-item-gi label="菜单名称" path="menuName">
          <n-input
            v-model:value="form.menuName"
            :placeholder="form.menuType === 'F' ? '如：新增用户' : '如：用户管理'"
          />
        </n-form-item-gi>
        <n-form-item-gi label="路由名称" path="name">
          <n-input
            v-model:value="form.name"
            :disabled="form.menuType === 'F'"
            placeholder="英文 PascalCase（留空则按路由地址派生）"
          />
        </n-form-item-gi>
      </n-grid>

      <!-- 按钮类型不需要路由地址 / 组件路径 -->
      <n-grid v-if="form.menuType !== 'F'" :cols="2" :x-gap="16">
        <n-form-item-gi label="路由地址" path="path">
          <n-input v-model:value="form.path" placeholder="如：/system/users" />
        </n-form-item-gi>
        <n-form-item-gi label="组件路径" path="component">
          <n-input v-model:value="form.component" placeholder="如：system/user/index" />
        </n-form-item-gi>
      </n-grid>

      <n-grid v-if="form.menuType !== 'F'" :cols="2" :x-gap="16">
        <n-form-item-gi label="微应用" path="microApp">
          <n-select
            v-model:value="form.microApp"
            :options="microAppOptions"
            :loading="microAppLoading"
            filterable
            tag
            clearable
            placeholder="选择或输入 microApp id"
            @update:value="handleMicroAppChange"
          />
        </n-form-item-gi>
        <n-form-item-gi label="模块 ID" path="moduleId">
          <n-select
            v-model:value="form.moduleId"
            :options="moduleIdOptions"
            :loading="microAppLoading"
            filterable
            tag
            clearable
            placeholder="选择或输入模块 ID"
          />
        </n-form-item-gi>
      </n-grid>

      <n-alert
        v-if="microAppTip && form.menuType !== 'F'"
        type="info"
        :bordered="false"
        style="margin-bottom: 12px;"
      >
        {{ microAppTip }}
      </n-alert>

      <n-grid :cols="2" :x-gap="16">
        <n-form-item-gi label="权限标识" path="perms">
          <n-input
            v-model:value="form.perms"
            :placeholder="
              form.menuType === 'F'
                ? '如：auth:user:create'
                : '如：auth:user:list（进入页面所需权限码）'
            "
          />
        </n-form-item-gi>
        <n-form-item-gi v-if="form.menuType !== 'F'" label="图标" path="icon">
          <n-select
            v-model:value="form.icon"
            :options="iconOptions"
            filterable
            tag
            clearable
            placeholder="选择或输入图标名"
          />
        </n-form-item-gi>
      </n-grid>

      <n-grid :cols="3" :x-gap="16">
        <n-form-item-gi label="显示排序" path="orderNum">
          <n-input-number v-model:value="form.orderNum" :min="0" style="width: 100%;" />
        </n-form-item-gi>
        <n-form-item-gi v-if="form.menuType !== 'F'" label="显示状态" path="visible">
          <n-select v-model:value="form.visible" :options="visibleOptions" />
        </n-form-item-gi>
        <n-form-item-gi label="菜单状态" path="status">
          <n-select v-model:value="form.status" :options="statusOptions" />
        </n-form-item-gi>
      </n-grid>

      <n-grid v-if="form.menuType !== 'F'" :cols="2" :x-gap="16">
        <n-form-item-gi label="是否外链" path="isFrame">
          <n-select v-model:value="form.isFrame" :options="frameOptions" />
        </n-form-item-gi>
        <n-form-item-gi label="是否缓存" path="isCache">
          <n-select v-model:value="form.isCache" :options="cacheOptions" />
        </n-form-item-gi>
      </n-grid>

      <n-form-item label="备注" path="remark">
        <n-input
          v-model:value="form.remark"
          type="textarea"
          :rows="2"
          placeholder="备注"
        />
      </n-form-item>
    </n-form>
    <template #footer>
      <div style="text-align: right;">
        <n-button style="margin-right: 8px;" @click="close">取消</n-button>
        <n-button type="primary" :loading="submitting" @click="submit">保存</n-button>
      </div>
    </template>
  </n-modal>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, type PropType } from 'vue';
import {
  NModal,
  NForm,
  NFormItem,
  NFormItemGi,
  NGrid,
  NInput,
  NInputNumber,
  NSelect,
  NTreeSelect,
  NButton,
  NAlert,
  useMessage,
  type FormInst,
  type FormRules,
} from 'naive-ui';
import {
  createMenu,
  updateMenu,
  listKnownMicroApps,
  type Menu,
  type MenuForm,
  type MicroAppOption,
} from '../../../../api/menu';

const props = defineProps({
  show: { type: Boolean, default: false },
  editing: { type: Object as PropType<Menu | null>, default: null },
  allMenus: { type: Array as PropType<Menu[]>, default: () => [] },
  /** 新建时的默认 menuType（由父级类型决定） */
  defaultMenuType: {
    type: String as PropType<'M' | 'C' | 'F' | undefined>,
    default: undefined,
  },
  /** 是否锁定 menuType（C 下新增按钮时锁定为 F） */
  lockMenuType: { type: Boolean, default: false },
});

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void;
  (e: 'success'): void;
}>();

const message = useMessage();
const formRef = ref<FormInst | null>(null);
const submitting = ref(false);

/* ============================ 微应用候选 ============================ */

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
  }))
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
  return `「${v}」尚未在模块列表或历史菜单中出现。保存后菜单会保留该 id，等对应子应用部署并接入 qiankun 后即可正常跳转。`;
});

function handleMicroAppChange(v: string) {
  if (!form.value.moduleId && v && v !== 'main' && v !== 'platform') {
    form.value.moduleId = v;
  }
}

/* ============================ 表单 ============================ */

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

/** 当前父节点是否为菜单类型（C） */
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
      // 编辑
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
      // 新建：应用 defaultMenuType
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

/** 用户切换父节点时，若新父节点是 C，强制 menuType=F */
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
    message.warning('上级是菜单类型，只能新增按钮，已自动切换为「按钮」');
    form.value.menuType = 'F';
  }
}

const rules: FormRules = {
  menuName: { required: true, message: '请输入菜单名称', trigger: 'blur' },
  menuType: {
    required: true,
    trigger: 'change',
    validator: (_rule, value: string) => {
      if (!value) return new Error('请选择菜单类型');
      if (parentIsMenu.value && value !== 'F') {
        return new Error('上级是菜单类型，只能新增按钮');
      }
      return true;
    },
  },
  microApp: {
    required: true,
    trigger: ['blur', 'change'],
    validator: (_rule, value: string) => {
      if (form.value.menuType === 'F') return true; // F 不校验 microApp
      const v = (value || '').trim();
      if (!v) return new Error('请选择或输入微应用 id');
      if (!MICRO_APP_RE.test(v)) {
        return new Error(
          'microApp 格式非法：需小写字母开头，仅含小写字母/数字/下划线/连字符，长度 ≤50'
        );
      }
      return true;
    },
  },
  path: {
    trigger: 'blur',
    validator: (_rule, value: string) => {
      if (form.value.menuType !== 'F' && !value) {
        return new Error('目录/菜单类型必须填写路由地址');
      }
      return true;
    },
  },
  perms: {
    trigger: 'blur',
    validator: (_rule, value: string) => {
      if (form.value.menuType === 'F' && !value) {
        return new Error('按钮类型必须填写权限标识');
      }
      return true;
    },
  },
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
      if (m.menuType === 'F') continue; // 按钮不能做父节点
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
  try {
    await formRef.value?.validate();
  } catch {
    return;
  }
  // 提交前再次校验父节点约束
  if (parentIsMenu.value && form.value.menuType !== 'F') {
    message.error('上级是菜单类型，只能新增按钮');
    return;
  }
  submitting.value = true;
  try {
    if (editing.value && props.editing?.id) {
      await updateMenu(props.editing.id, form.value);
      message.success('修改成功');
    } else {
      await createMenu(form.value);
      message.success('新增成功');
    }
    emit('success');
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  } finally {
    submitting.value = false;
  }
}
</script>
