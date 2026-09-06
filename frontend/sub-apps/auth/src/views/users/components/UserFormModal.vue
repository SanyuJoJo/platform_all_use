<template>
  <n-modal v-model:show="visible" :title="modalTitle" preset="dialog" @close="handleClose">
    <n-form ref="formRef" :model="formData" :rules="rules" label-placement="left" label-width="80">
      <n-form-item label="用户名" path="username" v-if="mode === 'create'">
        <n-input v-model:value="formData.username" placeholder="3-20位字母数字下划线" />
      </n-form-item>
      <n-form-item label="密码" path="password" v-if="mode === 'create'">
        <n-input v-model:value="formData.password" type="password" placeholder="6-20位" />
      </n-form-item>
      <n-form-item label="昵称" path="nickname">
        <n-input v-model:value="formData.nickname" placeholder="1-50位" />
      </n-form-item>
      <n-form-item label="邮箱" path="email">
        <n-input v-model:value="formData.email" placeholder="邮箱（可选）" />
      </n-form-item>
      <n-form-item label="状态" path="status">
        <n-radio-group v-model:value="formData.status">
          <n-radio :value="1">启用</n-radio>
          <n-radio :value="0">禁用</n-radio>
        </n-radio-group>
      </n-form-item>
      <n-form-item label="角色" path="role_ids">
        <n-select
          v-model:value="formData.role_ids"
          multiple
          :options="roleOptions"
          placeholder="选择角色（可多选）"
        />
      </n-form-item>
    </n-form>
    <template #action>
      <n-button @click="handleClose">取消</n-button>
      <n-button type="primary" :loading="submitting" @click="handleSubmit">确定</n-button>
    </template>
  </n-modal>
</template>
<script setup lang="ts">
import { ref, reactive, watch, computed, onMounted } from 'vue';
import { NModal, NForm, NFormItem, NInput, NRadioGroup, NRadio, NSelect, NButton, useMessage } from 'naive-ui';
import { userApi } from '@/api/user';
import { roleApi } from '@/api/role';
import type { User, UserCreate, UserUpdate } from '@/types';
const props = defineProps<{
  visible: boolean;
  mode: 'create' | 'edit';
  initialData?: User | null;
}>();
const emit = defineEmits(['update:visible', 'success']);
const message = useMessage();
const formRef = ref<any>(null);
const submitting = ref(false);
// 角色下拉选项
const roleOptions = ref<{ label: string; value: number }[]>([]);
// 表单数据
const formData = reactive({
  username: '',
  password: '',
  nickname: '',
  email: '',
  status: 1,
  role_ids: [] as number[],
});
// 验证规则
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
// 加载角色列表
async function loadRoles() {
  try {
    const res = await roleApi.getList({ page: 1, page_size: 100 });
    roleOptions.value = res.data.items.map(r => ({ label: r.name, value: r.id }));
  } catch (error) {
    // ignore
  }
}
// 当弹窗打开时，填充数据
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
        formData.role_ids = user.roles.map(r => r.id);
        formData.password = ''; // 编辑时不显示密码
      } else {
        // 创建模式重置
        formData.username = '';
        formData.password = '';
        formData.nickname = '';
        formData.email = '';
        formData.status = 1;
        formData.role_ids = [];
      }
    }
  },
  { immediate: true }
);
function handleClose() {
  visible.value = false;
}
async function handleSubmit() {
  try {
    await formRef.value?.validate();
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
      message.success('创建成功');
    } else {
      const payload: UserUpdate = {
        nickname: formData.nickname,
        email: formData.email || undefined,
        role_ids: formData.role_ids.length ? formData.role_ids : undefined,
        status: formData.status,
      };
      await userApi.update(props.initialData!.id, payload);
      message.success('更新成功');
    }
    emit('success');
    handleClose();
  } catch (error: any) {
    message.error(error.message || '操作失败');
  } finally {
    submitting.value = false;
  }
}
onMounted(() => {
  loadRoles();
});
</script>
