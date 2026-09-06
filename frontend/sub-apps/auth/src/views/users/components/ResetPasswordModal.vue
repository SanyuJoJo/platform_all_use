<template>
  <n-modal v-model:show="visible" title="重置密码" preset="dialog" @close="handleClose">
    <n-form ref="formRef" :model="formData" :rules="rules" label-placement="left" label-width="100">
      <n-form-item label="新密码" path="new_password">
        <n-input v-model:value="formData.new_password" type="password" placeholder="6-20位" />
      </n-form-item>
      <n-form-item label="确认密码" path="confirm_password">
        <n-input v-model:value="formData.confirm_password" type="password" placeholder="再次输入" />
      </n-form-item>
    </n-form>
    <template #action>
      <n-button @click="handleClose">取消</n-button>
      <n-button type="primary" :loading="submitting" @click="handleSubmit">确定</n-button>
    </template>
  </n-modal>
</template>
<script setup lang="ts">
import { ref, reactive, computed } from 'vue';
import { NModal, NForm, NFormItem, NInput, NButton, useMessage } from 'naive-ui';
import { userApi } from '@/api/user';
const props = defineProps<{
  visible: boolean;
  userId: number;
}>();
const emit = defineEmits(['update:visible', 'success']);
const message = useMessage();
const formRef = ref<any>(null);
const submitting = ref(false);
const formData = reactive({
  new_password: '',
  confirm_password: '',
});
const visible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val),
});
const rules = {
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度6-20位', trigger: 'blur' },
  ],
  confirm_password: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (rule: any, value: string) => value === formData.new_password,
      message: '两次密码不一致',
      trigger: 'blur',
    },
  ],
};
function handleClose() {
  visible.value = false;
}
async function handleSubmit() {
  try {
    await formRef.value?.validate();
    submitting.value = true;
    await userApi.resetPassword(props.userId, formData.new_password);
    message.success('密码重置成功');
    emit('success');
    handleClose();
  } catch (error: any) {
    message.error(error.message || '重置失败');
  } finally {
    submitting.value = false;
  }
}
</script>
