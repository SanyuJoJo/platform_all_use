<template>
  <n-modal v-model:show="visible" title="模块配置" preset="dialog" style="width: 600px;" @close="handleClose">
    <n-form ref="formRef" label-placement="left" label-width="100">
      <n-form-item label="配置（JSON）">
        <n-input
          v-model:value="configText"
          type="textarea"
          :rows="10"
          placeholder="请输入 JSON 格式的配置"
        />
      </n-form-item>
    </n-form>
    <template #action>
      <n-button @click="handleClose">取消</n-button>
      <n-button type="primary" :loading="submitting" @click="handleSubmit">保存</n-button>
    </template>
  </n-modal>
</template>
<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { NModal, NForm, NFormItem, NInput, NButton, useMessage } from 'naive-ui';
import { moduleApi } from '@/api/modules';
const props = defineProps<{
  visible: boolean;
  moduleId: string;
  initialConfig: any;
}>();
const emit = defineEmits(['update:visible', 'success']);
const message = useMessage();
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
  { immediate: true }
);
function handleClose() {
  visible.value = false;
}
async function handleSubmit() {
  try {
    let configObj;
    try {
      configObj = JSON.parse(configText.value);
    } catch (e) {
      message.error('JSON 格式错误，请检查');
      return;
    }
    submitting.value = true;
    await moduleApi.updateConfig(props.moduleId, configObj);
    message.success('配置更新成功');
    emit('success');
    handleClose();
  } catch (err: any) {
    message.error(err.message || '保存配置失败');
  } finally {
    submitting.value = false;
  }
}
</script>
