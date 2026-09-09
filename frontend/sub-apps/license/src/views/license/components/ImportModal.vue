<template>
  <n-modal v-model:show="visible" title="导入 License" preset="dialog" style="width: 500px;" @close="handleClose">
    <n-form ref="formRef" :model="formData" :rules="rules" label-placement="left" label-width="120">
      <n-form-item label="导入方式" path="method">
        <n-radio-group v-model:value="formData.method" @update:value="handleMethodChange">
          <n-radio value="file">上传文件</n-radio>
          <n-radio value="code">激活码</n-radio>
        </n-radio-group>
      </n-form-item>
      <n-form-item v-if="formData.method === 'file'" label="License文件" path="file">
        <n-upload
          ref="uploadRef"
          :max="1"
          :show-file-list="true"
          @change="handleFileChange"
          accept=".lic,.json"
        >
          <n-button>选择文件</n-button>
        </n-upload>
        <div style="font-size:12px; color:#999;">支持 .lic 或 .json 格式</div>
      </n-form-item>
      <n-form-item v-if="formData.method === 'code'" label="激活码" path="activation_code">
        <n-input v-model:value="formData.activation_code" placeholder="请输入激活码" />
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
import { NModal, NForm, NFormItem, NRadioGroup, NRadio, NUpload, NButton, NInput, useMessage } from 'naive-ui';
import { licenseApi } from '@/api/license';
import type { UploadFileInfo, UploadInst } from 'naive-ui';
const props = defineProps<{ visible: boolean }>();
const emit = defineEmits(['update:visible', 'success']);
const message = useMessage();
const formRef = ref<any>(null);
const uploadRef = ref<UploadInst | null>(null);
const submitting = ref(false);
const formData = reactive({
  method: 'file' as 'file' | 'code',
  file: null as File | null,
  activation_code: '',
});
const visible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val),
});
const rules = {
  file: [
    {
      validator: (rule: any, value: any) => {
        if (formData.method === 'file' && !formData.file) {
          return new Error('请选择 License 文件');
        }
        return true;
      },
      trigger: 'change',
    },
  ],
  activation_code: [
    {
      validator: (rule: any, value: string) => {
        if (formData.method === 'code' && !value.trim()) {
          return new Error('请输入激活码');
        }
        return true;
      },
      trigger: 'blur',
    },
  ],
};
// 切换方式：清空表单数据及上传列表
function handleMethodChange() {
  formData.file = null;
  formData.activation_code = '';
  // 清空上传组件文件列表（兼容 clear 和 clearFiles）
  if (uploadRef.value) {
    if (typeof uploadRef.value.clear === 'function') {
      uploadRef.value.clear();
    } else if (typeof (uploadRef.value as any).clearFiles === 'function') {
      (uploadRef.value as any).clearFiles();
    }
  }
}
// 处理文件变化
function handleFileChange(data: { file: UploadFileInfo; fileList: UploadFileInfo[] }) {
  const fileInfo = data.fileList[0];
  if (fileInfo && fileInfo.file) {
    formData.file = fileInfo.file as File;
  } else {
    formData.file = null;
  }
}
function handleClose() {
  visible.value = false;
}
async function handleSubmit() {
  try {
    await formRef.value?.validate();
    submitting.value = true;
    if (formData.method === 'file' && formData.file) {
      await licenseApi.importLicenseFile(formData.file);
    } else if (formData.method === 'code' && formData.activation_code.trim()) {
      await licenseApi.importLicenseCode(formData.activation_code.trim());
    } else {
      message.warning('请选择一种方式并填写完整');
      return;
    }
    message.success('导入成功');
    emit('success');
    handleClose();
  } catch (error: any) {
    message.error(error.message || '导入失败');
  } finally {
    submitting.value = false;
  }
}
</script>
