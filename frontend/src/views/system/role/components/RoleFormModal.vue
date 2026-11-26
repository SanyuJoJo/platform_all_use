<template>
  <t-dialog
    v-model:visible="visible"
    :header="modalTitle"
    :confirm-btn="null"
    :cancel-btn="null"
    :width="780"
    @close="handleClose"
  >
    <t-form ref="formRef" :data="formData" :rules="rules" label-align="left" :label-width="100">
      <t-form-item label="角色名称" name="name">
        <t-input v-model="formData.name" placeholder="如：访客" />
      </t-form-item>
      <t-form-item label="角色编码" name="code" v-if="mode === 'create'">
        <t-input v-model="formData.code" placeholder="如：guest（只能小写字母和下划线）" />
      </t-form-item>
      <t-form-item label="描述" name="description">
        <t-textarea v-model="formData.description" placeholder="可选" />
      </t-form-item>

      <t-form-item label="菜单权限">
        <div class="menu-tree-wrapper">
          <div class="tree-toolbar">
            <t-space>
              <t-button size="small" @click="expandAll">展开全部</t-button>
              <t-button size="small" @click="collapseAll">折叠全部</t-button>
              <t-button size="small" @click="checkAll">全选</t-button>
              <t-button size="small" @click="uncheckAll">清空</t-button>
            </t-space>
            <span class="tip">各节点独立勾选。勾选菜单代表访问权限，勾选按钮代表操作权限。</span>
          </div>

          <t-loading :loading="treeLoading">
            <t-tree
              :data="treeData"
              :keys="{ value: 'key', label: 'label', children: 'children' }"
              v-model="checkedKeys"
              :expanded="expandedKeys"
              checkable
              :check-strictly="true"
              :label="true"
              @expand="handleExpandedChange"
            >
              <template #label="{ node }">
                <div class="tree-node">
                  <span class="node-label">{{ node.label }}</span>
                  <t-tag size="small" :theme="tagTheme(node.menuType)" variant="light" style="margin-left: 6px">
                    {{ tagText(node.menuType) }}
                  </t-tag>
                  <t-tag v-if="node.status === 1" size="small" theme="default" variant="light" style="margin-left: 6px">
                    已停用
                  </t-tag>
                  <code v-if="node.perms" class="node-perms">{{ node.perms }}</code>
                </div>
              </template>
            </t-tree>
          </t-loading>
        </div>
      </t-form-item>
    </t-form>
    <template #footer>
      <t-space>
        <t-button @click="handleClose">取消</t-button>
        <t-button theme="primary" :loading="submitting" @click="handleSubmit">确定</t-button>
      </t-space>
    </template>
  </t-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted } from 'vue';
import { MessagePlugin, type FormInstanceFunctions } from 'tdesign-vue-next';
import { roleApi } from '@/api/role';
import { getMenuTree, type Menu } from '@/api/system-menu';
import type { Role } from '@/types';

const props = defineProps<{
  visible: boolean;
  mode: 'create' | 'edit';
  initialData?: Role | null;
}>();

const emit = defineEmits(['update:visible', 'success']);

const formRef = ref<FormInstanceFunctions>();
const submitting = ref(false);
const treeLoading = ref(false);

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
    MessagePlugin.error(e?.message || '加载菜单树失败');
  } finally {
    treeLoading.value = false;
  }
}

onMounted(loadMenuTree);

function tagTheme(mt: string): 'primary' | 'success' | 'warning' | 'default' {
  return mt === 'M' ? 'primary' : mt === 'C' ? 'success' : 'warning';
}
function tagText(mt: string): string {
  return mt === 'M' ? '目录' : mt === 'C' ? '菜单' : '按钮';
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

function handleExpandedChange(keys: Array<string | number>) {
  expandedKeys.value = keys.map((k) => Number(k));
}

const visible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val),
});
const modalTitle = computed(() =>
  props.mode === 'create' ? '创建角色' : '编辑角色',
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

      checkedKeys.value = rawMenuIds;
    } else {
      formData.name = '';
      formData.code = '';
      formData.description = '';
      checkedKeys.value = [];
    }

    expandAncestorsOf(checkedKeys.value);
  },
  { immediate: true },
);

function handleClose() {
  visible.value = false;
}

async function handleSubmit() {
  try {
    const result = await formRef.value?.validate();
    if (result !== true) return;

    const menuIdsToSave = checkedKeys.value;

    submitting.value = true;
    if (props.mode === 'create') {
      await roleApi.create({
        name: formData.name,
        code: formData.code,
        description: formData.description || undefined,
        menuIds: menuIdsToSave,
      });
      MessagePlugin.success('创建成功');
    } else {
      await roleApi.update(props.initialData!.id, {
        name: formData.name,
        description: formData.description || undefined,
        menuIds: menuIdsToSave,
      });
      MessagePlugin.success('更新成功');
    }
    emit('success');
    handleClose();
  } catch (error: any) {
    if (error.code === 20001) {
      MessagePlugin.error('角色编码已存在，请使用其他编码');
    } else if (error.code === 20004) {
      MessagePlugin.error('角色名称已存在，请使用其他名称');
    } else {
      MessagePlugin.error(error.message || '操作失败');
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