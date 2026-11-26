<template>
  <t-dialog
    v-model:visible="visible"
    :header="'重置密码'"
    :confirm-btn="null"
    :cancel-btn="null"
    @close="handleClose"
  >
    <t-form ref="formRef" :data="formData" :rules="rules" label-align="left" :label-width="100">
      <t-form-item label="新密码" name="new_password">
        <t-input v-model="formData.new_password" type="password" placeholder="6-20位" />
      </t-form-item>
      <t-form-item label="确认密码" name="confirm_password">
        <t-input v-model="formData.confirm_password" type="password" placeholder="再次输入" />
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
import { ref, reactive, computed } from 'vue';
import { MessagePlugin, type FormInstanceFunctions } from 'tdesign-vue-next';
import { userApi } from '@/api/user';
const props = defineProps<{
  visible: boolean;
  userId: number;
}>();
const emit = defineEmits(['update:visible', 'success']);
const formRef = ref<FormInstanceFunctions>();
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
      validator: (value: string) =>
        value === formData.new_password ? true : { result: false, message: '两次密码不一致' },
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
    const result = await formRef.value?.validate();
    if (result !== true) return;
    submitting.value = true;
    await userApi.resetPassword(props.userId, formData.new_password);
    MessagePlugin.success('密码重置成功');
    emit('success');
    handleClose();
  } catch (error: any) {
    MessagePlugin.error(error.message || '重置失败');
  } finally {
    submitting.value = false;
  }
}
</script>