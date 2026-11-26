<template>
  <t-dialog
    v-model:visible="visible"
    :header="modalTitle"
    :confirm-btn="null"
    :cancel-btn="null"
    @close="handleClose"
  >
    <t-form ref="formRef" :data="formData" :rules="rules" label-align="left" :label-width="80">
      <t-form-item label="用户名" name="username" v-if="mode === 'create'">
        <t-input v-model="formData.username" placeholder="3-20位字母数字下划线" />
      </t-form-item>
      <t-form-item label="密码" name="password" v-if="mode === 'create'">
        <t-input v-model="formData.password" type="password" placeholder="6-20位" />
      </t-form-item>
      <t-form-item label="昵称" name="nickname">
        <t-input v-model="formData.nickname" placeholder="1-50位" />
      </t-form-item>
      <t-form-item label="邮箱" name="email">
        <t-input v-model="formData.email" placeholder="邮箱（可选）" />
      </t-form-item>
      <t-form-item label="状态" name="status">
        <t-radio-group v-model="formData.status">
          <t-radio :value="1">启用</t-radio>
          <t-radio :value="0">禁用</t-radio>
        </t-radio-group>
      </t-form-item>
      <t-form-item label="角色" name="role_ids">
        <t-select
          v-model="formData.role_ids"
          multiple
          :options="roleOptions"
          placeholder="选择角色（可多选）"
          style="width: 100%"
        />
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
import { ref, reactive, watch, computed, onMounted } from 'vue';
import { MessagePlugin, type FormInstanceFunctions } from 'tdesign-vue-next';
import { userApi } from '@/api/user';
import { roleApi } from '@/api/role';
import type { User, UserCreate, UserUpdate } from '@/types';
const props = defineProps<{
  visible: boolean;
  mode: 'create' | 'edit';
  initialData?: User | null;
}>();
const emit = defineEmits(['update:visible', 'success']);
const formRef = ref<FormInstanceFunctions>();
const submitting = ref(false);
const roleOptions = ref<{ label: string; value: number }[]>([]);
const formData = reactive({
  username: '',
  password: '',
  nickname: '',
  email: '',
  status: 1,
  role_ids: [] as number[],
});
const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_]{3,20}$/, message: '3-20位字母数字下划线', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度6-20位', trigger: 'blur' },
  ],
  nickname: [
    { required: true, message: '请输入昵称', trigger: 'blur' },
    { min: 1, max: 50, message: '昵称长度1-50位', trigger: 'blur' },
  ],
};
const visible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val),
});
const modalTitle = computed(() => (props.mode === 'create' ? '创建用户' : '编辑用户'));
async function loadRoles() {
  try {
    const res = await roleApi.getList({ page: 1, page_size: 100 });
    roleOptions.value = res.data.items.map((r) => ({ label: r.name, value: r.id }));
  } catch (error) {
    // ignore
  }
}
watch(
  () => props.visible,
  (val) => {
    if (val) {
      if (props.mode === 'edit' && props.initialData) {
        const user = props.initialData;
        formData.username = user.username;
        formData.nickname = user.nickname;
        formData.email = user.email || '';
        formData.status = user.status;
        formData.role_ids = user.roles.map((r) => r.id);
        formData.password = '';
      } else {
        formData.username = '';
        formData.password = '';
        formData.nickname = '';
        formData.email = '';
        formData.status = 1;
        formData.role_ids = [];
      }
    }
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
    submitting.value = true;
    if (props.mode === 'create') {
      const payload: UserCreate = {
        username: formData.username,
        password: formData.password,
        nickname: formData.nickname,
        email: formData.email || undefined,
        role_ids: formData.role_ids.length ? formData.role_ids : undefined,
        status: formData.status,
      };
      await userApi.create(payload);
      MessagePlugin.success('创建成功');
    } else {
      const payload: UserUpdate = {
        nickname: formData.nickname,
        email: formData.email || undefined,
        role_ids: formData.role_ids.length ? formData.role_ids : undefined,
        status: formData.status,
      };
      await userApi.update(props.initialData!.id, payload);
      MessagePlugin.success('更新成功');
    }
    emit('success');
    handleClose();
  } catch (error: any) {
    MessagePlugin.error(error.message || '操作失败');
  } finally {
    submitting.value = false;
  }
}
onMounted(() => {
  loadRoles();
});
</script>