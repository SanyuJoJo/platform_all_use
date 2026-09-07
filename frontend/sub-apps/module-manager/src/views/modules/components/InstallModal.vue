<template>
  <n-modal v-model:show="visible" title="安装模块" preset="dialog" style="width: 500px;" @close="handleClose">
    <n-form ref="formRef" :model="formData" :rules="rules" label-placement="left" label-width="120">
      <n-form-item label="安装方式" path="install_type">
        <n-radio-group v-model:value="formData.install_type">
          <n-radio value="zip">ZIP 包路径</n-radio>
          <n-radio value="path">源码目录路径</n-radio>
        </n-radio-group>
      </n-form-item>
      <n-form-item
        :label="formData.install_type === 'zip' ? 'ZIP 文件路径' : '源码目录路径'"
        :path="formData.install_type === 'zip' ? 'file_path' : 'source_path'"
      >
        <n-input
          v-model:value="formData[formData.install_type === 'zip' ? 'file_path' : 'source_path']"
          :placeholder="formData.install_type === 'zip' ? '例如 /tmp/module.zip' : '例如 /opt/modules/my-module'"
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
import { ref, reactive, computed, watch } from 'vue';
import { NModal, NForm, NFormItem, NRadioGroup, NRadio, NInput, NButton, useMessage } from 'naive-ui';
import { moduleApi } from '@/api/modules';
const props = defineProps<{
  visible: boolean;
}>();
const emit = defineEmits(['update:visible', 'success']);
const message = useMessage();
const formRef = ref<any>(null);
const submitting = ref(false);
const formData = reactive({
  install_type: 'zip' as 'zip' | 'path',
  file_path: '',
  source_path: '',
});
// 监听安装方式切换，清空另一个路径字段
watch(() => formData.install_type, (newType) => {
  if (newType === 'zip') {
    formData.source_path = '';
  } else {
    formData.file_path = '';
  }
});
const rules = {
  file_path: [
    {
      validator: (rule: any, value: string) => {
        if (formData.install_type === 'zip' && !value) {
          return new Error('请输入 ZIP 文件路径');
        }
        return true;
      },
      trigger: 'blur',
    },
  ],
  source_path: [
    {
      validator: (rule: any, value: string) => {
        if (formData.install_type === 'path' && !value) {
          return new Error('请输入源码目录路径');
        }
        return true;
      },
      trigger: 'blur',
    },
  ],
};
const visible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val),
});
function handleClose() {
  visible.value = false;
}
async function handleSubmit() {
  try {
    await formRef.value?.validate();
    submitting.value = true;
    const payload: any = {
      install_type: formData.install_type,
    };
    if (formData.install_type === 'zip') {
      payload.file_path = formData.file_path;
    } else {
      payload.source_path = formData.source_path;
    }
    await moduleApi.install(payload);
    message.success('安装成功');
    emit('success');
    handleClose();
  } catch (err: any) {
    message.error(err.message || '安装失败');
  } finally {
    submitting.value = false;
  }
}
</script>
