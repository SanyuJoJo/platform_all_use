<template>
  <t-dialog
    v-model:visible="visible"
    :header="'模块配置'"
    :confirm-btn="null"
    :cancel-btn="null"
    :width="600"
    @close="handleClose"
  >
    <t-form ref="formRef" label-align="left" :label-width="100">
      <t-form-item label="配置（JSON）">
        <t-textarea v-model="configText" :autosize="{ minRows: 10 }" placeholder="请输入 JSON 格式的配置" />
      </t-form-item>
    </t-form>
    <template #footer>
      <t-space>
        <t-button @click="handleClose">取消</t-button>
        <t-button theme="primary" :loading="submitting" @click="handleSubmit">保存</t-button>
      </t-space>
    </template>
  </t-dialog>
</template>
<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { moduleApi } from '@/api/modules';
const props = defineProps<{
  visible: boolean;
  moduleId: string;
  initialConfig: any;
}>();
const emit = defineEmits(['update:visible', 'success']);
const submitting = ref(false);
const configText = ref('');
const visible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val),
});
watch(
  () => props.visible,
  (val) => {
    if (val && props.initialConfig) {
      configText.value = JSON.stringify(props.initialConfig, null, 2);
    }
  },
  { immediate: true },
);
function handleClose() {
  visible.value = false;
}
async function handleSubmit() {
  let configObj;
  try {
    configObj = JSON.parse(configText.value);
  } catch (e) {
    MessagePlugin.error('JSON 格式错误，请检查');
    return;
  }
  submitting.value = true;
  try {
    await moduleApi.updateConfig(props.moduleId, configObj);
    MessagePlugin.success('配置更新成功');
    emit('success');
    handleClose();
  } catch (err: any) {
    MessagePlugin.error(err.message || '保存配置失败');
  } finally {
    submitting.value = false;
  }
}
</script>