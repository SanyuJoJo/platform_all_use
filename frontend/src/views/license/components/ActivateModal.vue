<template>
  <t-dialog
    v-model:visible="visible"
    :header="'在线激活'"
    :confirm-btn="null"
    :cancel-btn="null"
    :width="500"
    @close="handleClose"
  >
    <t-form ref="formRef" :data="formData" :rules="rules" label-align="left" :label-width="120">
      <t-form-item label="激活码" name="activation_code">
        <t-input v-model="formData.activation_code" placeholder="请输入激活码" />
      </t-form-item>
      <t-form-item label="机器码" name="machine_code">
        <t-input v-model="formData.machine_code" placeholder="自动获取中..." />
        <t-button size="small" style="margin-left: 8px" @click="refreshMachineCode">重新获取</t-button>
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
import { ref, reactive, computed, onMounted } from 'vue';
import { MessagePlugin, type FormInstanceFunctions } from 'tdesign-vue-next';
import { licenseApi } from '@/api/license';
import { getMachineCode, refreshMachineCode as refreshCode } from '@/utils/machineCode';
const props = defineProps<{ visible: boolean }>();
const emit = defineEmits(['update:visible', 'success']);
const formRef = ref<FormInstanceFunctions>();
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
  activation_code: [{ required: true, message: '请输入激活码', trigger: 'blur' }],
  machine_code: [{ required: true, message: '请获取机器码', trigger: 'blur' }],
};
async function refreshMachineCode() {
  try {
    const code = await refreshCode();
    formData.machine_code = code;
  } catch (error) {
    MessagePlugin.error('获取机器码失败，请手动输入');
  }
}
onMounted(() => {
  refreshMachineCode();
});
function handleClose() {
  visible.value = false;
}
async function handleSubmit() {
  const result = await formRef.value?.validate();
  if (result !== true) return;
  submitting.value = true;
  try {
    await licenseApi.activate(formData.activation_code, formData.machine_code);
    MessagePlugin.success('激活成功');
    emit('success');
    handleClose();
  } catch (error: any) {
    MessagePlugin.error(error.message || '激活失败');
  } finally {
    submitting.value = false;
  }
}
</script>