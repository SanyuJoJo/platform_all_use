<template>
  <t-dialog
    v-model:visible="visible"
    :header="'导入 License'"
    :confirm-btn="null"
    :cancel-btn="null"
    :width="500"
    @close="handleClose"
  >
    <t-form ref="formRef" :data="formData" :rules="rules" label-align="left" :label-width="120">
      <t-form-item label="导入方式" name="method">
        <t-radio-group v-model="formData.method" @change="handleMethodChange">
          <t-radio value="file">上传文件</t-radio>
          <t-radio value="code">激活码</t-radio>
        </t-radio-group>
      </t-form-item>
      <t-form-item v-if="formData.method === 'file'" label="License文件" name="file">
        <t-upload
          :auto-upload="false"
          :files="uploadFiles"
          :max="1"
          accept=".lic,.json"
          @change="handleFileChange"
          @remove="handleFileRemove"
        >
          <t-button>选择文件</t-button>
        </t-upload>
        <div style="font-size: 12px; color: #999">支持 .lic 或 .json 格式</div>
      </t-form-item>
      <t-form-item v-if="formData.method === 'code'" label="激活码" name="activation_code">
        <t-input v-model="formData.activation_code" placeholder="请输入激活码" />
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
import { MessagePlugin, type FormInstanceFunctions, type UploadFile } from 'tdesign-vue-next';
import { licenseApi } from '@/api/license';
const props = defineProps<{ visible: boolean }>();
const emit = defineEmits(['update:visible', 'success']);
const formRef = ref<FormInstanceFunctions>();
const submitting = ref(false);
const uploadFiles = ref<UploadFile[]>([]);
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
      validator: (_rule: unknown, _value: unknown) => {
        if (formData.method === 'file' && !formData.file) {
          return { result: false, message: '请选择 License 文件' };
        }
        return true;
      },
      trigger: 'change',
    },
  ],
  activation_code: [
    {
      validator: (_rule: unknown, value: string) => {
        if (formData.method === 'code' && !value.trim()) {
          return { result: false, message: '请输入激活码' };
        }
        return true;
      },
      trigger: 'blur',
    },
  ],
};
function handleMethodChange() {
  formData.file = null;
  formData.activation_code = '';
  uploadFiles.value = [];
}
function handleFileChange(files: UploadFile[]) {
  const file = files?.[0]?.raw;
  formData.file = (file as File) || null;
  uploadFiles.value = files;
}
function handleFileRemove() {
  formData.file = null;
  uploadFiles.value = [];
}
function handleClose() {
  visible.value = false;
}
async function handleSubmit() {
  const result = await formRef.value?.validate();
  if (result !== true) return;
  submitting.value = true;
  try {
    if (formData.method === 'file' && formData.file) {
      await licenseApi.importLicenseFile(formData.file);
    } else if (formData.method === 'code' && formData.activation_code.trim()) {
      await licenseApi.importLicenseCode(formData.activation_code.trim());
    } else {
      MessagePlugin.warning('请选择一种方式并填写完整');
      return;
    }
    MessagePlugin.success('导入成功');
    emit('success');
    handleClose();
  } catch (error: any) {
    MessagePlugin.error(error.message || '导入失败');
  } finally {
    submitting.value = false;
  }
}
</script>