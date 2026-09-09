<template>
  <n-modal v-model:show="visible" title="在线激活" preset="dialog" style="width: 500px;" @close="handleClose">
    <n-form ref="formRef" :model="formData" :rules="rules" label-placement="left" label-width="120">
      <n-form-item label="激活码" path="activation_code">
        <n-input v-model:value="formData.activation_code" placeholder="请输入激活码" />
      </n-form-item>
      <n-form-item label="机器码" path="machine_code">
        <n-input v-model:value="formData.machine_code" placeholder="自动获取中..." />
        <n-button size="small" style="margin-left:8px;" @click="refreshMachineCode">重新获取</n-button>
      </n-form-item>
    </n-form>
    <template #action>
      <n-button @click="handleClose">取消</n-button>
      <n-button type="primary" :loading="submitting" @click="handleSubmit">确定</n-button>
    </template>
  </n-modal>
</template>
<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue';
import { NModal, NForm, NFormItem, NInput, NButton, useMessage } from 'naive-ui';
import { licenseApi } from '@/api/license';
import { getMachineCode, refreshMachineCode as refreshCode } from '@/utils/machineCode';
const props = defineProps<{ visible: boolean }>();
const emit = defineEmits(['update:visible', 'success']);
const message = useMessage();
const formRef = ref<any>(null);
const submitting = ref(false);
const formData = reactive({
  activation_code: '',
  machine_code: '',
});
const visible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val),
});
const rules = {
  activation_code: [
    { required: true, message: '请输入激活码', trigger: 'blur' },
  ],
  machine_code: [
    { required: true, message: '请获取机器码', trigger: 'blur' },
  ],
};
async function refreshMachineCode() {
  try {
    const code = await refreshCode();
    formData.machine_code = code;
  } catch (error) {
    message.error('获取机器码失败，请手动输入');
    // 不设置 machineCodeReady，让用户可手动编辑
  }
}
onMounted(() => {
  refreshMachineCode();
});
function handleClose() {
  visible.value = false;
}
async function handleSubmit() {
  try {
    await formRef.value?.validate();
    submitting.value = true;
    await licenseApi.activate(formData.activation_code, formData.machine_code);
    message.success('激活成功');
    emit('success');
    handleClose();
  } catch (error: any) {
    message.error(error.message || '激活失败');
  } finally {
    submitting.value = false;
  }
}
</script>
