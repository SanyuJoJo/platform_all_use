<template>
  <n-modal v-model:show="visible" :title="modalTitle" preset="dialog" style="width: 600px;" @close="handleClose">
    <n-form ref="formRef" :model="formData" :rules="rules" label-placement="left" label-width="100">
      <n-form-item label="角色名称" path="name">
        <n-input v-model:value="formData.name" placeholder="如：访客" />
      </n-form-item>
      <n-form-item label="角色编码" path="code" v-if="mode === 'create'">
        <n-input v-model:value="formData.code" placeholder="如：guest" />
      </n-form-item>
      <n-form-item label="描述" path="description">
        <n-input v-model:value="formData.description" placeholder="可选" type="textarea" />
      </n-form-item>
      <n-form-item label="分配权限" path="permission_codes">
        <PermissionTree v-model:checked="formData.permission_codes" />
      </n-form-item>
    </n-form>
    <template #action>
      <n-button @click="handleClose">取消</n-button>
      <n-button type="primary" :loading="submitting" @click="handleSubmit">确定</n-button>
    </template>
  </n-modal>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue';
import { NModal, NForm, NFormItem, NInput, NButton, useMessage } from 'naive-ui';
import { roleApi } from '@/api/role';
import PermissionTree from '@/components/PermissionTree.vue';
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
// 表单数据
const formData = reactive({
  name: '',
  code: '',
  description: '',
  permission_codes: [] as string[],
});
const visible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val),
});
const modalTitle = computed(() => (props.mode === 'create' ? '创建角色' : '编辑角色'));
const rules = {
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入角色编码', trigger: 'blur' }],
};
watch(
  () => props.visible,
  (val) => {
    if (val) {
      if (props.mode === 'edit' && props.initialData) {
        const role = props.initialData;
        formData.name = role.name;
        formData.code = role.code;
        formData.description = role.description || '';
        formData.permission_codes = role.permission_codes || [];
      } else {
        formData.name = '';
        formData.code = '';
        formData.description = '';
        formData.permission_codes = [];
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
      await roleApi.create({
        name: formData.name,
        code: formData.code,
        description: formData.description || undefined,
        permission_codes: formData.permission_codes,
      });
      message.success('创建成功');
    } else {
      await roleApi.update(props.initialData!.id, {
        name: formData.name,
        description: formData.description || undefined,
        permission_codes: formData.permission_codes,
      });
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
</script>
