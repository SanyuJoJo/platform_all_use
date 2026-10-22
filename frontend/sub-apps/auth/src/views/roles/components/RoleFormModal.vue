<template>
  <n-modal
    v-model:show="visible"
    :title="modalTitle"
    preset="dialog"
    style="width: 780px;"
    @close="handleClose"
  >
    <n-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      label-placement="left"
      label-width="100"
    >
      <n-form-item label="角色名称" path="name">
        <n-input v-model:value="formData.name" placeholder="如：访客" />
      </n-form-item>
      <n-form-item label="角色编码" path="code" v-if="mode === 'create'">
        <n-input
          v-model:value="formData.code"
          placeholder="如：guest（只能小写字母和下划线）"
        />
      </n-form-item>
      <n-form-item label="描述" path="description">
        <n-input
          v-model:value="formData.description"
          placeholder="可选"
          type="textarea"
        />
      </n-form-item>

      <n-form-item label="菜单权限">
        <div class="menu-tree-wrapper">
          <div class="tree-toolbar">
            <n-space>
              <n-button size="tiny" @click="expandAll">展开全部</n-button>
              <n-button size="tiny" @click="collapseAll">折叠全部</n-button>
              <n-button size="tiny" @click="checkAll">全选</n-button>
              <n-button size="tiny" @click="uncheckAll">清空</n-button>
            </n-space>
            <span class="tip">
              各节点独立勾选。勾选菜单代表访问权限，勾选按钮代表操作权限。
            </span>
          </div>

          <n-spin :show="treeLoading">
            <!-- 
              ★ 核心修改：
              1. :cascade="false" 关闭级联，父子互不影响。
              2. check-on-click 可以不点框直接点文字勾选，体验更好。
            -->
            <n-tree
              :data="treeData"
              :checked-keys="checkedKeys"
              :expanded-keys="expandedKeys"
              :checkable="true"
              :cascade="false"
              :selectable="false"
              block-line
              key-field="key"
              label-field="label"
              children-field="children"
              :render-label="renderLabel"
              @update:checked-keys="handleCheckedChange"
              @update:expanded-keys="handleExpandedChange"
            />
          </n-spin>
        </div>
      </n-form-item>
    </n-form>
    <template #action>
      <n-button @click="handleClose">取消</n-button>
      <n-button type="primary" :loading="submitting" @click="handleSubmit">
        确定
      </n-button>
    </template>
  </n-modal>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted, h } from 'vue';
import {
  NModal,
  NForm,
  NFormItem,
  NInput,
  NButton,
  NSpace,
  NTree,
  NSpin,
  NTag,
  useMessage,
} from 'naive-ui';
import { roleApi } from '@/api/role';
import { getMenuTree, type Menu } from '@/api/menu';
import type { Role } from '@/types';

const props = defineProps<{
  visible: boolean;
  mode: 'create' | 'edit';
  initialData?: Role | null;
}>();

const emit = defineEmits(['update:visible', 'success']);

const message = useMessage();
const formRef = ref<any>(null);
const submitting = ref(false);
const treeLoading = ref(false);

/* ============================ 菜单树 ============================ */

interface MenuTreeNode {
  key: number;
  label: string;
  menuType: 'M' | 'C' | 'F';
  perms?: string;
  status?: number;
  children?: MenuTreeNode[];
  [key: string]: unknown;
}

const treeData = ref<MenuTreeNode[]>([]);
const checkedKeys = ref<number[]>([]);
const expandedKeys = ref<number[]>([]);

function toTreeNodes(menus: Menu[]): MenuTreeNode[] {
  const out: MenuTreeNode[] = [];
  for (const m of menus) {
    const node: MenuTreeNode = {
      key: m.id,
      label: m.menuName,
      menuType: m.menuType,
      perms: m.perms || undefined,
      status: m.status,
    };
    if (m.children && m.children.length > 0) {
      node.children = toTreeNodes(m.children);
    }
    out.push(node);
  }
  return out;
}

function collectAllKeys(nodes: MenuTreeNode[]): number[] {
  const keys: number[] = [];
  const walk = (list: MenuTreeNode[]) => {
    list.forEach((n) => {
      keys.push(n.key);
      if (n.children?.length) walk(n.children);
    });
  };
  walk(nodes);
  return keys;
}

// 用于展开勾选节点的祖先链
function buildParentMap(nodes: MenuTreeNode[]): Map<number, number | null> {
  const map = new Map<number, number | null>();
  const walk = (list: MenuTreeNode[], parent: number | null) => {
    list.forEach((n) => {
      map.set(n.key, parent);
      if (n.children?.length) walk(n.children, n.key);
    });
  };
  walk(nodes, null);
  return map;
}

function expandAncestorsOf(keys: number[]) {
  if (keys.length === 0) return;
  const parentMap = buildParentMap(treeData.value);
  const expandedSet = new Set<number>(expandedKeys.value);
  keys.forEach((k) => {
    let cur: number | null | undefined = parentMap.get(k);
    while (cur !== null && cur !== undefined) {
      expandedSet.add(cur);
      cur = parentMap.get(cur);
    }
  });
  expandedKeys.value = Array.from(expandedSet);
}

async function loadMenuTree() {
  treeLoading.value = true;
  try {
    const res = await getMenuTree();
    const list = (res?.data ?? []) as Menu[];
    treeData.value = toTreeNodes(list);
    expandedKeys.value = treeData.value.map((n) => n.key);
  } catch (e: any) {
    message.error(e?.message || '加载菜单树失败');
  } finally {
    treeLoading.value = false;
  }
}

onMounted(loadMenuTree);

function renderLabel({ option }: { option: any }) {
  const node = option as MenuTreeNode;
  const mt = node.menuType;
  const tagType: 'info' | 'success' | 'warning' =
    mt === 'M' ? 'info' : mt === 'C' ? 'success' : 'warning';
  const tagText = mt === 'M' ? '目录' : mt === 'C' ? '菜单' : '按钮';

  const children: any[] = [
    h('span', { class: 'node-label' }, node.label),
    h(
      NTag,
      { size: 'tiny', type: tagType, style: 'margin-left: 6px;' },
      { default: () => tagText }
    ),
  ];

  if (node.status === 1) {
    children.push(
      h(
        NTag,
        { size: 'tiny', type: 'default', style: 'margin-left: 6px;' },
        { default: () => '已停用' }
      )
    );
  }

  if (node.perms) {
    children.push(h('code', { class: 'node-perms' }, node.perms));
  }
  return h('div', { class: 'tree-node' }, children);
}

function expandAll() {
  expandedKeys.value = collectAllKeys(treeData.value);
}
function collapseAll() {
  expandedKeys.value = [];
}
function checkAll() {
  checkedKeys.value = collectAllKeys(treeData.value);
}
function uncheckAll() {
  checkedKeys.value = [];
}

// 因为关闭了级联，返回的 keys 就是用户真实点击的节点
function handleCheckedChange(keys: Array<string | number>) {
  checkedKeys.value = keys.map((k) => Number(k));
}

function handleExpandedChange(keys: Array<string | number>) {
  expandedKeys.value = keys.map((k) => Number(k));
}

/* ============================ 表单 ============================ */

const visible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val),
});
const modalTitle = computed(() =>
  props.mode === 'create' ? '创建角色' : '编辑角色'
);

const formData = reactive({
  name: '',
  code: '',
  description: '',
});

const rules = {
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  code: [
    { required: true, message: '请输入角色编码', trigger: 'blur' },
    {
      pattern: /^[a-z_]+$/,
      message: '编码只能包含小写字母和下划线',
      trigger: 'blur',
    },
  ],
};

watch(
  () => props.visible,
  async (val) => {
    if (!val) return;

    if (treeData.value.length === 0) {
      await loadMenuTree();
    }

    if (props.mode === 'edit' && props.initialData) {
      const role = props.initialData;
      formData.name = role.name;
      formData.code = role.code;
      formData.description = role.description || '';

      let rawMenuIds: number[] = [];
      if (Array.isArray(role.menuIds)) {
        rawMenuIds = role.menuIds;
      } else {
        try {
          const res = await roleApi.getDetail(role.id);
          const detail = res?.data as Role | undefined;
          rawMenuIds = Array.isArray(detail?.menuIds) ? detail!.menuIds! : [];
        } catch {
          rawMenuIds = [];
        }
      }

      // 直接回显所有 ID（包括父节点和子节点）
      checkedKeys.value = rawMenuIds;
    } else {
      formData.name = '';
      formData.code = '';
      formData.description = '';
      checkedKeys.value = [];
    }

    expandAncestorsOf(checkedKeys.value);
  },
  { immediate: true }
);

function handleClose() {
  visible.value = false;
}

async function handleSubmit() {
  try {
    await formRef.value?.validate();

    // 关闭级联后，用户勾了哪些就传哪些（后端能正确处理任何 ID）
    const menuIdsToSave = checkedKeys.value;

    submitting.value = true;
    if (props.mode === 'create') {
      await roleApi.create({
        name: formData.name,
        code: formData.code,
        description: formData.description || undefined,
        menuIds: menuIdsToSave,
      });
      message.success('创建成功');
    } else {
      await roleApi.update(props.initialData!.id, {
        name: formData.name,
        description: formData.description || undefined,
        menuIds: menuIdsToSave,
      });
      message.success('更新成功');
    }
    emit('success');
    handleClose();
  } catch (error: any) {
    if (error.code === 20001) {
      message.error('角色编码已存在，请使用其他编码');
    } else if (error.code === 20004) {
      message.error('角色名称已存在，请使用其他名称');
    } else {
      message.error(error.message || '操作失败');
    }
  } finally {
    submitting.value = false;
  }
}
</script>

<style scoped>
.menu-tree-wrapper {
  width: 100%;
  max-height: 420px;
  overflow-y: auto;
  border: 1px solid #e0e0e6;
  border-radius: 4px;
  padding: 8px 12px;
  background: #fafafc;
}
.tree-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
  padding-bottom: 8px;
  border-bottom: 1px dashed #e0e0e6;
}
.tree-toolbar .tip {
  font-size: 12px;
  color: #999;
}
.tree-node {
  display: flex;
  align-items: center;
  gap: 4px;
}
.node-label {
  font-size: 13px;
}
.node-perms {
  font-size: 11px;
  color: #999;
  background: #f0f0f5;
  padding: 1px 4px;
  border-radius: 2px;
  margin-left: auto;
}
</style>
