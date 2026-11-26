<template>
  <t-dialog
    v-model:visible="visible"
    :header="'安装模块'"
    :confirm-btn="null"
    :cancel-btn="null"
    :width="500"
    @close="handleClose"
  >
    <t-form ref="formRef" :data="formData" :rules="rules" label-align="left" :label-width="120">
      <t-form-item label="安装方式" name="install_type">
        <t-radio-group v-model="formData.install_type">
          <t-radio value="zip">ZIP 包路径</t-radio>
          <t-radio value="path">源码目录路径</t-radio>
        </t-radio-group>
      </t-form-item>
      <t-form-item
        :label="formData.install_type === 'zip' ? 'ZIP 文件路径' : '源码目录路径'"
        :name="formData.install_type === 'zip' ? 'file_path' : 'source_path'"
      >
        <t-input
          v-model="formData[formData.install_type === 'zip' ? 'file_path' : 'source_path']"
          :placeholder="formData.install_type === 'zip' ? '例如 /tmp/module.zip' : '例如 /opt/modules/my-module'"
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
import { ref, reactive, computed, watch } from 'vue';
import { MessagePlugin, type FormInstanceFunctions } from 'tdesign-vue-next';
import { moduleApi } from '@/api/modules';
const props = defineProps<{
  visible: boolean;
}>();
const emit = defineEmits(['update:visible', 'success']);
const formRef = ref<FormInstanceFunctions>();
const submitting = ref(false);
const formData = reactive({
  install_type: 'zip' as 'zip' | 'path',
  file_path: '',
  source_path: '',
});
watch(
  () => formData.install_type,
  (newType) => {
    if (newType === 'zip') {
      formData.source_path = '';
    } else {
      formData.file_path = '';
    }
  },
);
const rules = {
  file_path: [
    {
      validator: (_rule: unknown, value: string) => {
        if (formData.install_type === 'zip' && !value) {
          return { result: false, message: '请输入 ZIP 文件路径' };
        }
        return true;
      },
      trigger: 'blur',
    },
  ],
  source_path: [
    {
      validator: (_rule: unknown, value: string) => {
        if (formData.install_type === 'path' && !value) {
          return { result: false, message: '请输入源码目录路径' };
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
  const result = await formRef.value?.validate();
  if (result !== true) return;
  submitting.value = true;
  try {
    const payload: any = {
      install_type: formData.install_type,
    };
    if (formData.install_type === 'zip') {
      payload.file_path = formData.file_path;
    } else {
      payload.source_path = formData.source_path;
    }
    await moduleApi.install(payload);
    MessagePlugin.success('安装成功');
    emit('success');
    handleClose();
  } catch (err: any) {
    MessagePlugin.error(err.message || '安装失败');
  } finally {
    submitting.value = false;
  }
}
</script>