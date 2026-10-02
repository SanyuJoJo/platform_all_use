<template>
  <div>
    <n-space vertical size="large">
      <n-card title="P10 管理">
        <template #header-extra>
          <n-space>
            <n-button
              v-permission="'crypto_console:csr:create'"
              type="primary"
              @click="openApply"
            >
              申请 P10
            </n-button>
            <n-button
              v-permission="'crypto_console:csr:create'"
              @click="openImport"
            >
              导入 P10
            </n-button>
            <n-button
              v-permission="'crypto_console:cert:sign'"
              type="success"
              @click="openSignCert"
            >
              签发证书
            </n-button>
            <n-button
              v-permission="'crypto_console:cert:view'"
              @click="openQueryEnvelope"
            >
              查看信封信息
            </n-button>
            <n-button @click="fetchList">刷新</n-button>
          </n-space>
        </template>

        <n-data-table
          :columns="columns"
          :data="list"
          :loading="loading"
          :pagination="pagination"
          remote
          @update:page="handlePageChange"
        />
      </n-card>
    </n-space>

    <!-- ================= 申请 P10 ================= -->
    <n-modal v-model:show="showApply" preset="card" title="申请 P10" style="width: 720px">
      <n-form :model="applyForm" label-placement="left" label-width="110">
        <n-form-item label="算法">
          <n-select v-model:value="applyForm.algorithm" :options="algorithmOptions" />
        </n-form-item>

        <n-form-item label="使用者 DN">
          <div class="dn-row">
            <n-input
              v-model:value="applyForm.dn"
              type="textarea"
              :autosize="{ minRows: 3, maxRows: 6 }"
              :placeholder="DN_HELP_EXAMPLE"
            />
            <n-popover trigger="click" placement="left-start" :width="440" style="max-width: 90vw">
              <template #trigger>
                <n-button circle size="small" quaternary type="info" class="dn-help-btn">?</n-button>
              </template>
              <div class="dn-help">
                <p class="dn-help__title">DN 项支持</p>
                <p class="dn-help__keys">{{ DN_HELP_KEYS }}</p>
                <p class="dn-help__title">示例</p>
                <p class="dn-help__example">{{ DN_HELP_EXAMPLE }}</p>
                <p class="dn-help__tip">
                  提示：如某字段值本身包含分隔符（例如 O 中出现逗号），
                  请改用其他分隔符（如 “;” 或 “\n” 表示换行）。
                </p>
              </div>
            </n-popover>
          </div>
        </n-form-item>

        <n-form-item label="DN 分隔符">
          <div class="dn-sep-row">
            <n-input
              v-model:value="applyForm.dnSeparator"
              placeholder="默认逗号 “,”，可输入 “;” 或 “\n” 表示换行"
              style="max-width: 320px"
            />
            <span class="dn-sep-hint">支持 “\n” 表示换行，“\t” 表示制表符</span>
          </div>
        </n-form-item>

        <n-form-item v-if="applyForm.dn.trim()" label="解析预览">
          <div class="dn-preview">
            <n-alert v-if="dnPreview.errors.length" type="error" :show-icon="true" style="width: 100%">
              <div v-for="(e, i) in dnPreview.errors" :key="i">{{ e }}</div>
            </n-alert>
            <template v-else>
              <n-tag
                v-for="(v, k) in dnPreview.subject"
                :key="k"
                size="small"
                type="success"
                style="margin: 2px 6px 2px 0"
              >
                {{ k }} = {{ v }}
              </n-tag>
            </template>
          </div>
        </n-form-item>

        <n-form-item label="SAN（逗号分隔）">
          <n-input v-model:value="applyForm.san" placeholder="例如 test.example.com,www.example.com" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showApply = false">取消</n-button>
          <n-button type="primary" :loading="submitting" :disabled="applyDisabled" @click="handleApply">
            提交
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- ================= 导入 P10 ================= -->
    <n-modal v-model:show="showImport" preset="card" title="导入 P10" style="width: 680px">
      <n-alert type="info" :show-icon="true" style="margin-bottom: 12px">
        上传 P10（CSR）文件；如有对应私钥可一并上传，加密私钥需填写密码。
      </n-alert>
      <n-form label-placement="top">
        <n-form-item label="P10 文件（PEM，必填）">
          <div class="file-row">
            <input ref="csrFileInputRef" type="file" accept=".csr,.pem,.p10" style="display: none" @change="onCsrFileChange" />
            <n-button @click="csrFileInputRef?.click()">选择文件</n-button>
            <span v-if="importForm.csr_file_name" class="file-name">{{ importForm.csr_file_name }}</span>
            <span v-else class="file-placeholder">未选择文件</span>
          </div>
        </n-form-item>

        <n-form-item label="私钥文件（PEM，可选）">
          <div class="file-row">
            <input ref="keyFileInputRef" type="file" accept=".pem,.key" style="display: none" @change="onKeyFileChange" />
            <n-button @click="keyFileInputRef?.click()">选择私钥</n-button>
            <span v-if="importForm.key_file_name" class="file-name">{{ importForm.key_file_name }}</span>
            <span v-else class="file-placeholder">未选择文件</span>
            <n-button v-if="importForm.key_file_name" size="tiny" quaternary @click="clearKeyFile">清除</n-button>
          </div>
        </n-form-item>

        <n-form-item label="私钥密码（仅加密私钥需要）">
          <n-input v-model:value="importForm.key_password" type="password" show-password-on="click" placeholder="未加密留空" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showImport = false">取消</n-button>
          <n-button type="primary" :loading="submitting" :disabled="!importForm.csr_pem.trim()" @click="handleImport">
            导入
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- ================= 签发证书 ================= -->
    <n-modal v-model:show="showSignCert" preset="card" title="签发证书" style="width: 900px">
      <n-form :model="signForm" label-placement="left" label-width="130">
        <n-form-item label="P10 来源">
          <n-radio-group v-model:value="signForm.csr_source">
            <n-radio value="existing">从本地选择 P10</n-radio>
            <n-radio value="upload">手动输入 CSR 和密钥</n-radio>
          </n-radio-group>
        </n-form-item>

        <template v-if="signForm.csr_source === 'existing'">
          <n-form-item label="选择 P10">
            <n-select v-model:value="signForm.csr_id" :options="csrOptions" filterable clearable placeholder="请选择 P10" style="width: 100%" />
          </n-form-item>
        </template>

        <template v-else>
          <n-form-item label="CSR PEM">
            <n-input v-model:value="signForm.csr_pem" type="textarea" :autosize="{ minRows: 4, maxRows: 8 }" placeholder="-----BEGIN CERTIFICATE REQUEST-----" />
          </n-form-item>
          <n-form-item label="私钥 PEM">
            <n-input v-model:value="signForm.csr_key_pem" type="textarea" :autosize="{ minRows: 4, maxRows: 8 }" placeholder="-----BEGIN PRIVATE KEY-----" />
          </n-form-item>
          <n-form-item label="私钥密码">
            <n-input v-model:value="signForm.csr_key_password" type="password" show-password-on="click" placeholder="私钥未加密可留空" />
          </n-form-item>
        </template>

        <!-- ★ 使用者 DN（可选，留空则用 P10 里的 DN） -->
        <n-form-item label="使用者 DN">
          <div class="dn-row">
            <n-input
              v-model:value="signForm.dn"
              type="textarea"
              :autosize="{ minRows: 3, maxRows: 6 }"
              :placeholder="'留空则沿用 P10 内的使用者 DN；填写则覆盖（例如 ' + DN_HELP_EXAMPLE + '）'"
            />
            <n-popover trigger="click" placement="left-start" :width="440" style="max-width: 90vw">
              <template #trigger>
                <n-button circle size="small" quaternary type="info" class="dn-help-btn">?</n-button>
              </template>
              <div class="dn-help">
                <p class="dn-help__title">DN 项支持</p>
                <p class="dn-help__keys">{{ DN_HELP_KEYS }}</p>
                <p class="dn-help__title">示例</p>
                <p class="dn-help__example">{{ DN_HELP_EXAMPLE }}</p>
                <p class="dn-help__tip">
                  提示：如某字段值本身包含分隔符（例如 O 中出现逗号），
                  请改用其他分隔符（如 “;” 或 “\n” 表示换行）。
                </p>
              </div>
            </n-popover>
          </div>
        </n-form-item>

        <n-form-item v-if="signForm.dn.trim()" label="DN 分隔符">
          <div class="dn-sep-row">
            <n-input
              v-model:value="signForm.dnSeparator"
              placeholder="默认逗号 “,”，可输入 “;” 或 “\n” 表示换行"
              style="max-width: 320px"
            />
            <span class="dn-sep-hint">支持 “\n” 表示换行，“\t” 表示制表符</span>
          </div>
        </n-form-item>

        <n-form-item v-if="signForm.dn.trim()" label="解析预览">
          <div class="dn-preview">
            <n-alert v-if="signDnPreview.errors.length" type="error" :show-icon="true" style="width: 100%">
              <div v-for="(e, i) in signDnPreview.errors" :key="i">{{ e }}</div>
            </n-alert>
            <template v-else>
              <n-tag
                v-for="(v, k) in signDnPreview.subject"
                :key="k"
                size="small"
                type="success"
                style="margin: 2px 6px 2px 0"
              >
                {{ k }} = {{ v }}
              </n-tag>
            </template>
          </div>
        </n-form-item>

        <n-form-item label="签发模式">
          <n-radio-group v-model:value="signForm.cert_mode">
            <n-radio value="normal">普通证书</n-radio>
            <n-radio value="dual">国密双证书</n-radio>
          </n-radio-group>
        </n-form-item>

        <template v-if="signForm.cert_mode === 'normal'">
          <n-form-item label="证书类型">
            <n-checkbox-group v-model:value="signForm.cert_types">
              <n-space>
                <n-checkbox value="server">服务端</n-checkbox>
                <n-checkbox value="client">客户端</n-checkbox>
                <n-checkbox value="signature">签名</n-checkbox>
                <n-checkbox value="encryption">加密</n-checkbox>
              </n-space>
            </n-checkbox-group>
          </n-form-item>
        </template>

        <n-form-item label="CA 来源">
          <n-radio-group v-model:value="signForm.ca_source">
            <n-radio value="local">本地 CA</n-radio>
            <n-radio value="manual">手动输入 CA</n-radio>
          </n-radio-group>
        </n-form-item>

        <template v-if="signForm.ca_source === 'local'">
          <n-form-item label="选择 CA">
            <n-select v-model:value="signForm.ca_id" :options="caOptions" filterable clearable placeholder="请选择 CA" style="width: 100%" />
          </n-form-item>
        </template>

        <template v-else>
          <n-form-item label="CA 证书 PEM">
            <n-input v-model:value="signForm.ca_cert_pem" type="textarea" :autosize="{ minRows: 3, maxRows: 6 }" />
          </n-form-item>
          <n-form-item label="CA 私钥 PEM">
            <n-input v-model:value="signForm.ca_key_pem" type="textarea" :autosize="{ minRows: 3, maxRows: 6 }" />
          </n-form-item>
          <n-form-item label="CA 私钥密码">
            <n-input v-model:value="signForm.ca_key_password" type="password" show-password-on="click" />
          </n-form-item>
        </template>

        <n-form-item label="有效期天数">
          <n-input-number v-model:value="signForm.validity_days" :min="1" />
        </n-form-item>
      </n-form>

      <template #footer>
        <n-space justify="end">
          <n-button @click="showSignCert = false">取消</n-button>
          <n-button type="primary" :loading="submitting" :disabled="signDisabled" @click="handleSignCert">
            签发
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- ================= 签发结果弹窗 ================= -->
    <n-modal v-model:show="showSignResult" preset="card" title="签发结果" style="width: 900px">
      <n-alert type="warning" :show-icon="true" style="margin-bottom: 12px">
        请立即复制并妥善保存以下内容。弹窗关闭后，内容将无法再次查看。
      </n-alert>

      <n-space vertical size="medium">
        <template v-if="signResult.cert_pem">
          <div class="result-block">
            <div class="result-block__header">
              <strong>证书（PEM）</strong>
              <n-button size="tiny" @click="copyText(signResult.cert_pem || '')">复制</n-button>
            </div>
            <n-input :value="signResult.cert_pem" type="textarea" :autosize="{ minRows: 6, maxRows: 12 }" readonly />
          </div>
        </template>

        <template v-if="signResult.key_pem">
          <div class="result-block">
            <div class="result-block__header">
              <strong>私钥（PEM）</strong>
              <n-button size="tiny" @click="copyText(signResult.key_pem || '')">复制</n-button>
            </div>
            <n-input :value="signResult.key_pem" type="textarea" :autosize="{ minRows: 6, maxRows: 12 }" readonly />
          </div>
        </template>

        <template v-if="signResult.sign_cert_pem">
          <div class="result-block">
            <div class="result-block__header">
              <strong>签名证书（PEM）</strong>
              <n-button size="tiny" @click="copyText(signResult.sign_cert_pem || '')">复制</n-button>
            </div>
            <n-input :value="signResult.sign_cert_pem" type="textarea" :autosize="{ minRows: 6, maxRows: 12 }" readonly />
          </div>
        </template>

        <template v-if="signResult.enc_cert_pem">
          <div class="result-block">
            <div class="result-block__header">
              <strong>加密证书（PEM）</strong>
              <n-button size="tiny" @click="copyText(signResult.enc_cert_pem || '')">复制</n-button>
            </div>
            <n-input :value="signResult.enc_cert_pem" type="textarea" :autosize="{ minRows: 6, maxRows: 12 }" readonly />
          </div>
        </template>

        <template v-if="signResult.encrypted_envelope">
          <div class="result-block result-block--envelope">
            <div class="result-block__header">
              <div>
                <strong>加密的数字信封</strong>
                <n-tag type="success" size="tiny" style="margin-left: 8px">请保存</n-tag>
              </div>
              <n-button size="tiny" @click="copyText(signResult.encrypted_envelope || '')">复制</n-button>
            </div>
            <n-alert type="info" :show-icon="true" style="margin-bottom: 8px">
              该字符串为加密的数字信封整体（base64），请完整保存。
              后续在"查看信封信息"中粘贴此字符串 + 上传您的 CSR 私钥，可解密出三件套。
            </n-alert>
            <n-input :value="signResult.encrypted_envelope" type="textarea" :autosize="{ minRows: 8, maxRows: 16 }" readonly />
          </div>
        </template>
      </n-space>

      <template #footer>
        <n-space justify="end">
          <n-button @click="closeSignResult">关闭</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- ================= 查看信封信息 ================= -->
    <n-modal v-model:show="showQueryEnvelope" preset="card" title="查看信封信息" style="width: 900px">
      <n-form :model="envelopeForm" label-placement="left" label-width="140">
        <n-form-item label="信封格式">
          <n-radio-group v-model:value="envelopeForm.format">
            <n-space>
              <n-radio value="pkcs10">PKCS10 格式信封</n-radio>
              <n-radio value="cfca" disabled>CFCA 证书格式（暂未实现）</n-radio>
            </n-space>
          </n-radio-group>
        </n-form-item>

        <n-form-item label="查询方式">
          <n-radio-group v-model:value="envelopeForm.mode">
            <n-space vertical>
              <n-radio value="cert">
                从签名证书解析
                <span class="hint-text" style="margin-left: 8px">
                  （若签名证书关联了私钥，自动使用其解密）
                </span>
              </n-radio>
              <n-radio value="manual">
                手动输入
                <span class="hint-text" style="margin-left: 8px">
                  （输入您的 CSR 私钥解密）
                </span>
              </n-radio>
            </n-space>
          </n-radio-group>
        </n-form-item>

        <template v-if="envelopeForm.mode === 'cert'">
          <n-form-item label="选择签名证书">
            <n-select
              v-model:value="envelopeForm.cert_id"
              :options="signCertOptions"
              :loading="signCertLoading"
              filterable
              clearable
              placeholder="请选择国密双证中的签名证书"
              style="width: 100%"
            />
          </n-form-item>
        </template>

        <template v-else>
          <n-alert type="warning" :show-icon="true" style="margin-bottom: 12px">
            注意：若您导入 P10 时未上传私钥，平台无法自动解密，必须使用本模式上传您的 CSR 私钥。
          </n-alert>

          <n-form-item label="签名私钥 PEM">
            <div class="file-row">
              <n-input
                v-model:value="envelopeForm.sign_key_pem"
                type="textarea"
                :autosize="{ minRows: 4, maxRows: 8 }"
                placeholder="-----BEGIN PRIVATE KEY-----"
                style="flex: 1"
              />
              <div class="file-col">
                <input
                  ref="envelopeKeyFileInputRef"
                  type="file"
                  accept=".pem,.key"
                  style="display: none"
                  @change="onEnvelopeKeyFileChange"
                />
                <n-button @click="envelopeKeyFileInputRef?.click()">选择私钥</n-button>
                <span v-if="envelopeForm.sign_key_file_name" class="file-name">
                  {{ envelopeForm.sign_key_file_name }}
                </span>
              </div>
            </div>
          </n-form-item>

          <n-form-item label="签名私钥密码">
            <n-input
              v-model:value="envelopeForm.sign_key_password"
              type="password"
              show-password-on="click"
              placeholder="私钥未加密可留空"
            />
          </n-form-item>
        </template>

        <n-form-item label="加密的数字信封">
          <div class="file-row">
            <n-input
              v-model:value="envelopeForm.encrypted_envelope"
              type="textarea"
              :autosize="{ minRows: 6, maxRows: 12 }"
              placeholder="粘贴签发时返回的加密数字信封字符串"
              style="flex: 1"
            />
            <div class="file-col">
              <input
                ref="envelopeFileInputRef"
                type="file"
                accept=".txt,.env,.b64"
                style="display: none"
                @change="onEnvelopeFileChange"
              />
              <n-button @click="envelopeFileInputRef?.click()">选择文件</n-button>
              <n-button
                v-if="envelopeForm.encrypted_envelope"
                size="small"
                quaternary
                @click="clearEnvelope"
              >
                清除
              </n-button>
              <span v-if="envelopeForm.envelope_file_name" class="file-name">
                {{ envelopeForm.envelope_file_name }}
              </span>
            </div>
          </div>
        </n-form-item>

        <n-form-item>
          <n-space>
            <n-button type="primary" :loading="querying" :disabled="queryDisabled" @click="handleQueryEnvelope">
              解密信封
            </n-button>
            <n-button @click="clearEnvelopeResult">清空结果</n-button>
          </n-space>
        </n-form-item>
      </n-form>

      <template v-if="envelopeResult">
        <n-divider />

        <n-descriptions bordered :column="1" label-placement="left" size="small" style="margin-bottom: 12px">
          <n-descriptions-item label="格式">{{ envelopeResult.format }}</n-descriptions-item>
          <n-descriptions-item label="查询方式">
            {{ envelopeResult.mode === 'manual' ? '手动输入' : '从签名证书解析' }}
          </n-descriptions-item>
        </n-descriptions>

        <div class="envelope-field">
          <div class="envelope-field__header">
            <span class="envelope-field__label">对称密钥密文（symmetric_key_cipher，base64）</span>
            <n-button size="tiny" @click="copyText(envelopeResult.symmetric_key_cipher)">复制</n-button>
          </div>
          <n-input :value="envelopeResult.symmetric_key_cipher" type="textarea" :autosize="{ minRows: 3, maxRows: 6 }" readonly />
        </div>

        <div class="envelope-field">
          <div class="envelope-field__header">
            <span class="envelope-field__label">IV（base64）</span>
            <n-button size="tiny" @click="copyText(envelopeResult.iv)">复制</n-button>
          </div>
          <n-input :value="envelopeResult.iv" readonly />
        </div>

        <div class="envelope-field">
          <div class="envelope-field__header">
            <span class="envelope-field__label">加密私钥密文（encrypted_private_key，base64）</span>
            <n-button size="tiny" @click="copyText(envelopeResult.encrypted_private_key)">复制</n-button>
          </div>
          <n-input :value="envelopeResult.encrypted_private_key" type="textarea" :autosize="{ minRows: 6, maxRows: 12 }" readonly />
        </div>

        <div v-if="envelopeResult.decrypted_key_pem" class="envelope-field result-block--envelope" style="padding: 10px 12px; border-radius: 6px">
          <div class="envelope-field__header">
            <span class="envelope-field__label">解密后的加密私钥（PEM）</span>
            <n-button size="tiny" @click="copyText(envelopeResult.decrypted_key_pem || '')">复制</n-button>
          </div>
          <n-input :value="envelopeResult.decrypted_key_pem" type="textarea" :autosize="{ minRows: 6, maxRows: 12 }" readonly />
        </div>
      </template>

      <template #footer>
        <n-space justify="end">
          <n-button @click="showQueryEnvelope = false">关闭</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- ================= 下载密钥口令输入 ================= -->
    <n-modal v-model:show="showDownloadKey" preset="card" title="下载 P10 私钥" style="width: 480px">
      <n-form>
        <n-form-item label="使用口令加密">
          <n-switch v-model:value="downloadKeyForm.encrypt" />
        </n-form-item>
        <n-form-item v-if="downloadKeyForm.encrypt" label="加密口令">
          <n-input v-model:value="downloadKeyForm.password" type="password" show-password-on="click" placeholder="请输入加密口令" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showDownloadKey = false">取消</n-button>
          <n-button type="primary" :disabled="downloadKeyForm.encrypt && !downloadKeyForm.password" @click="handleDownloadKey">
            下载
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, h } from 'vue';
import {
  NAlert, NButton, NCard, NCheckbox, NCheckboxGroup, NDataTable,
  NDescriptions, NDescriptionsItem, NDivider, NForm, NFormItem, NInput,
  NInputNumber, NModal, NPopconfirm, NPopover, NRadio, NRadioGroup,
  NSelect, NSpace, NSwitch, NTag, useMessage, type DataTableColumns,
} from 'naive-ui';
import {
  cryptoApi,
  type SignCertParams,
  type SignCertResult,
  type QueryEnvelopeParams,
  type QueryEnvelopeResult,
  type CertTypeItem,
  type ImportCsrParams,
} from '@/api/crypto';
import {
  DEFAULT_DN_SEPARATOR, DN_HELP_KEYS, DN_HELP_EXAMPLE,
  parseDN, validateDN,
} from '@/utils/dn';

const message = useMessage();

const loading = ref(false);
const submitting = ref(false);
const querying = ref(false);

const showApply = ref(false);
const showImport = ref(false);
const showSignCert = ref(false);
const showDownloadKey = ref(false);
const showSignResult = ref(false);
const showQueryEnvelope = ref(false);

const list = ref<any[]>([]);
const caList = ref<any[]>([]);
const signResult = ref<SignCertResult>({});
const envelopeResult = ref<QueryEnvelopeResult | null>(null);

const pagination = reactive({
  page: 1, pageSize: 20, itemCount: 0,
  showSizePicker: true, pageSizes: [10, 20, 50],
});

// ---------------------------------------------------------------------------
// 申请 P10
// ---------------------------------------------------------------------------
const applyForm = reactive({
  algorithm: 'SM2', dn: '', dnSeparator: DEFAULT_DN_SEPARATOR, san: '',
});

const dnPreview = computed(() => parseDN(applyForm.dn, applyForm.dnSeparator));
const applyDisabled = computed(() => !applyForm.dn.trim() || dnPreview.value.errors.length > 0);

// ---------------------------------------------------------------------------
// 导入 P10
// ---------------------------------------------------------------------------
const csrFileInputRef = ref<HTMLInputElement | null>(null);
const keyFileInputRef = ref<HTMLInputElement | null>(null);
const importForm = reactive({
  csr_pem: '', csr_file_name: '',
  key_pem: '', key_file_name: '', key_password: '',
});

// ---------------------------------------------------------------------------
// 签发证书
// ---------------------------------------------------------------------------
const signForm = reactive({
  csr_source: 'existing' as 'existing' | 'upload',
  csr_id: '', csr_pem: '', csr_key_pem: '', csr_key_password: '',
  // ★ 使用者 DN（可选）
  dn: '', dnSeparator: DEFAULT_DN_SEPARATOR,
  cert_mode: 'normal' as 'normal' | 'dual',
  cert_types: ['server'] as CertTypeItem[],
  ca_source: 'local' as 'local' | 'manual',
  ca_id: '', ca_cert_pem: '', ca_key_pem: '', ca_key_password: '',
  validity_days: 365,
});

const signDnPreview = computed(() => parseDN(signForm.dn, signForm.dnSeparator));

// ---------------------------------------------------------------------------
// 查询信封信息
// ---------------------------------------------------------------------------
const envelopeKeyFileInputRef = ref<HTMLInputElement | null>(null);
const envelopeFileInputRef = ref<HTMLInputElement | null>(null);

const envelopeForm = reactive({
  format: 'pkcs10' as 'pkcs10' | 'cfca',
  mode: 'cert' as 'cert' | 'manual',
  cert_id: '',
  sign_key_pem: '',
  sign_key_file_name: '',
  sign_key_password: '',
  encrypted_envelope: '',
  envelope_file_name: '',
});

const signCertList = ref<any[]>([]);
const signCertLoading = ref(false);

const signCertOptions = computed(() =>
  signCertList.value.map((c) => ({
    label: `${c.subject_cn || c.cert_id} (${c.cert_id})`,
    value: c.cert_id,
  }))
);

async function fetchSignCertOptions() {
  signCertLoading.value = true;
  try {
    const res = await cryptoApi.listCerts({
      page: 1,
      page_size: 200,
      cert_type: 'dual_sign',
    });
    const data: any = res.data || {};
    signCertList.value = data.items || data || [];
  } catch (e: any) {
    console.warn('[csr] fetch sign cert options failed', e);
    signCertList.value = [];
  } finally {
    signCertLoading.value = false;
  }
}

const queryDisabled = computed(() => {
  if (!envelopeForm.encrypted_envelope.trim()) return true;
  if (envelopeForm.mode === 'cert') return !envelopeForm.cert_id;
  return !envelopeForm.sign_key_pem.trim();
});

// ---------------------------------------------------------------------------
// 下载私钥
// ---------------------------------------------------------------------------
const downloadKeyForm = reactive({ csrId: '', encrypt: false, password: '' });

// ---------------------------------------------------------------------------
// 表格列
// ---------------------------------------------------------------------------
const columns: DataTableColumns<any> = [
  { title: 'P10 ID', key: 'csr_id', ellipsis: { tooltip: true } },
  { title: '使用者', key: 'subject_cn', ellipsis: { tooltip: true } },
  { title: '算法', key: 'algorithm' },
  { title: '状态', key: 'status' },
  { title: '创建时间', key: 'created_at' },
  {
    title: '操作',
    key: 'actions',
    width: 280,
    render(row) {
      return h(NSpace, null, {
        default: () => [
          h(NButton, { size: 'small', onClick: () => handleDownloadCsr(row) }, { default: () => '下载P10' }),
          h(NButton, { size: 'small', onClick: () => openDownloadKey(row) }, { default: () => '下载密钥' }),
          h(NPopconfirm, {
            onPositiveClick: () => handleDelete(row),
            positiveText: '确认删除',
            negativeText: '取消',
          }, {
            trigger: () => h(NButton, { size: 'small', type: 'error' }, { default: () => '删除' }),
            default: () => `确认删除 P10「${row.csr_id}」？`,
          }),
        ],
      });
    },
  },
];

// ---------------------------------------------------------------------------
// 下拉选项
// ---------------------------------------------------------------------------
const algorithmOptions = [
  { label: 'SM2', value: 'SM2' },
  { label: 'RSA', value: 'RSA' },
  { label: 'ECC', value: 'ECC' },
  { label: 'ML-DSA', value: 'ML-DSA' },
];

const csrOptions = computed(() =>
  list.value.map((c) => ({
    label: `${c.subject_cn || c.csr_id} (${c.csr_id})`,
    value: c.csr_id,
  }))
);

const caOptions = computed(() =>
  caList.value.map((c) => ({
    label: `${c.subject_cn || c.ca_id} (${c.ca_id})`,
    value: c.ca_id,
  }))
);

const signDisabled = computed(() => {
  if (signForm.csr_source === 'existing' && !signForm.csr_id) return true;
  if (signForm.csr_source === 'upload' && (!signForm.csr_pem.trim() || !signForm.csr_key_pem.trim())) return true;
  if (signForm.dn.trim() && signDnPreview.value.errors.length > 0) return true;
  if (signForm.cert_mode === 'normal' && signForm.cert_types.length === 0) return true;
  if (signForm.ca_source === 'local' && !signForm.ca_id) return true;
  if (signForm.ca_source === 'manual' && (!signForm.ca_cert_pem.trim() || !signForm.ca_key_pem.trim())) return true;
  return false;
});

// ---------------------------------------------------------------------------
// 数据加载
// ---------------------------------------------------------------------------
async function fetchList() {
  loading.value = true;
  try {
    const res = await cryptoApi.listCsrs({ page: pagination.page, page_size: pagination.pageSize });
    const data: any = res.data || {};
    list.value = data.items || data || [];
    pagination.itemCount = data.total || list.value.length;
  } catch (e: any) {
    message.error(e.message || '加载失败');
  } finally {
    loading.value = false;
  }
}

async function fetchCAOptions() {
  try {
    const res = await cryptoApi.listCas({ page: 1, page_size: 200 });
    const data: any = res.data || {};
    caList.value = data.items || data || [];
  } catch (e: any) {
    console.warn('[csr] fetch CA options failed', e);
    caList.value = [];
  }
}

function handlePageChange(page: number) {
  pagination.page = page;
  fetchList();
}

// ---------------------------------------------------------------------------
// 申请 P10
// ---------------------------------------------------------------------------
function openApply() {
  applyForm.algorithm = 'SM2';
  applyForm.dn = '';
  applyForm.dnSeparator = DEFAULT_DN_SEPARATOR;
  applyForm.san = '';
  showApply.value = true;
}

async function handleApply() {
  const r = validateDN(applyForm.dn, applyForm.dnSeparator);
  if (r.errors.length) { message.error(r.errors[0]); return; }
  submitting.value = true;
  try {
    const params: Record<string, unknown> = {
      algorithm: applyForm.algorithm,
      subject: r.subject,
      key_source: 'generate',
    };
    const sanList = applyForm.san.split(',').map((s) => s.trim()).filter(Boolean);
    if (sanList.length) params.san = sanList;

    await cryptoApi.execute('csr.create', params);
    message.success('P10 申请成功');
    showApply.value = false;
    await fetchList();
  } catch (e: any) {
    message.error(e.message || '申请失败');
  } finally {
    submitting.value = false;
  }
}

// ---------------------------------------------------------------------------
// 导入 P10
// ---------------------------------------------------------------------------
function openImport() {
  importForm.csr_pem = ''; importForm.csr_file_name = '';
  importForm.key_pem = ''; importForm.key_file_name = ''; importForm.key_password = '';
  if (csrFileInputRef.value) csrFileInputRef.value.value = '';
  if (keyFileInputRef.value) keyFileInputRef.value.value = '';
  showImport.value = true;
}

async function onCsrFileChange(e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) { importForm.csr_pem = ''; importForm.csr_file_name = ''; return; }
  try {
    importForm.csr_pem = await file.text();
    importForm.csr_file_name = file.name;
  } catch (err: any) {
    message.error('读取文件失败: ' + (err?.message || ''));
  }
}

async function onKeyFileChange(e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) { importForm.key_pem = ''; importForm.key_file_name = ''; return; }
  try {
    importForm.key_pem = await file.text();
    importForm.key_file_name = file.name;
  } catch (err: any) {
    message.error('读取文件失败: ' + (err?.message || ''));
  }
}

function clearKeyFile() {
  importForm.key_pem = ''; importForm.key_file_name = ''; importForm.key_password = '';
  if (keyFileInputRef.value) keyFileInputRef.value.value = '';
}

async function handleImport() {
  if (!importForm.csr_pem.trim()) { message.warning('请选择 P10 文件'); return; }
  submitting.value = true;
  try {
    const payload: ImportCsrParams = { csr_pem: importForm.csr_pem.trim() };
    if (importForm.key_pem.trim()) {
      payload.key_pem = importForm.key_pem.trim();
      if (importForm.key_password) payload.key_password = importForm.key_password;
    }
    await cryptoApi.importCsr(payload);
    message.success('导入成功');
    showImport.value = false;
    await fetchList();
  } catch (e: any) {
    message.error(e.message || '导入失败');
  } finally {
    submitting.value = false;
  }
}

// ---------------------------------------------------------------------------
// 下载 P10
// ---------------------------------------------------------------------------
async function handleDownloadCsr(row: any) {
  try {
    const res: any = await cryptoApi.downloadCsr(row.csr_id);
    const blob: Blob = res.data instanceof Blob ? res.data : new Blob([res.data]);
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${row.csr_id}.csr`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    message.success('下载成功');
  } catch (e: any) {
    message.error(e.message || '下载失败');
  }
}

// ---------------------------------------------------------------------------
// 下载密钥
// ---------------------------------------------------------------------------
function openDownloadKey(row: any) {
  downloadKeyForm.csrId = row.csr_id;
  downloadKeyForm.encrypt = false;
  downloadKeyForm.password = '';
  showDownloadKey.value = true;
}

async function handleDownloadKey() {
  if (downloadKeyForm.encrypt && !downloadKeyForm.password) { message.warning('请输入加密口令'); return; }
  try {
    const res: any = await cryptoApi.downloadCsrKey(downloadKeyForm.csrId, downloadKeyForm.encrypt, downloadKeyForm.password);
    const blob: Blob = res.data instanceof Blob ? res.data : new Blob([res.data]);
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${downloadKeyForm.csrId}.key.pem`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    showDownloadKey.value = false;
    message.success('下载成功');
  } catch (e: any) {
    message.error(e.message || '下载失败');
  }
}

// ---------------------------------------------------------------------------
// 签发证书
// ---------------------------------------------------------------------------
function openSignCert() {
  signForm.csr_source = 'existing';
  signForm.csr_id = '';
  signForm.csr_pem = '';
  signForm.csr_key_pem = '';
  signForm.csr_key_password = '';
  signForm.dn = '';
  signForm.dnSeparator = DEFAULT_DN_SEPARATOR;
  signForm.cert_mode = 'normal';
  signForm.cert_types = ['server'];
  signForm.ca_source = 'local';
  signForm.ca_id = '';
  signForm.ca_cert_pem = '';
  signForm.ca_key_pem = '';
  signForm.ca_key_password = '';
  signForm.validity_days = 365;
  if (caList.value.length === 0) fetchCAOptions();
  showSignCert.value = true;
}

async function handleSignCert() {
  // 若填了 DN，先校验
  let subjectObj: Record<string, string> | undefined = undefined;
  if (signForm.dn.trim()) {
    const r = validateDN(signForm.dn, signForm.dnSeparator);
    if (r.errors.length) {
      message.error(r.errors[0]);
      return;
    }
    subjectObj = r.subject;
  }

  submitting.value = true;
  try {
    const payload: SignCertParams = {
      ca_source: signForm.ca_source,
      validity_days: signForm.validity_days,
      csr_source: signForm.csr_source,
      cert_mode: signForm.cert_mode,
    };

    // ★ 若填写了 DN，作为 subject 传给后端（覆盖 P10 里的默认 DN）
    if (subjectObj) {
      payload.subject = subjectObj;
    }

    if (signForm.csr_source === 'existing') {
      payload.csr_id = signForm.csr_id;
    } else {
      payload.csr_pem = signForm.csr_pem;
      payload.csr_key_pem = signForm.csr_key_pem;
      if (signForm.csr_key_password) {
        payload.csr_key_password = signForm.csr_key_password;
      }
    }

    if (signForm.cert_mode === 'normal') {
      payload.cert_types = [...signForm.cert_types];
    } else {
      payload.algorithm = 'SM2';
    }

    if (signForm.ca_source === 'local') {
      payload.ca_id = signForm.ca_id;
    } else {
      payload.ca_cert_pem = signForm.ca_cert_pem;
      payload.ca_key_pem = signForm.ca_key_pem;
      if (signForm.ca_key_password) {
        payload.ca_key_password = signForm.ca_key_password;
      }
    }

    const res: any = await cryptoApi.signCert(payload);
    signResult.value = (res?.data as SignCertResult) || {};

    showSignCert.value = false;
    showSignResult.value = true;
    await fetchList();
  } catch (e: any) {
    message.error(e.message || '签发失败');
  } finally {
    submitting.value = false;
  }
}

function closeSignResult() {
  showSignResult.value = false;
  signResult.value = {};
}

// ---------------------------------------------------------------------------
// 查看信封信息
// ---------------------------------------------------------------------------
function openQueryEnvelope() {
  envelopeForm.format = 'pkcs10';
  envelopeForm.mode = 'cert';
  envelopeForm.cert_id = '';
  envelopeForm.sign_key_pem = '';
  envelopeForm.sign_key_file_name = '';
  envelopeForm.sign_key_password = '';
  envelopeForm.encrypted_envelope = '';
  envelopeForm.envelope_file_name = '';
  if (envelopeKeyFileInputRef.value) envelopeKeyFileInputRef.value.value = '';
  if (envelopeFileInputRef.value) envelopeFileInputRef.value.value = '';
  if (signCertList.value.length === 0) fetchSignCertOptions();
  envelopeResult.value = null;
  showQueryEnvelope.value = true;
}

async function onEnvelopeKeyFileChange(e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) { envelopeForm.sign_key_pem = ''; envelopeForm.sign_key_file_name = ''; return; }
  try {
    envelopeForm.sign_key_pem = await file.text();
    envelopeForm.sign_key_file_name = file.name;
  } catch (err: any) {
    message.error('读取私钥文件失败: ' + (err?.message || ''));
  }
}

async function onEnvelopeFileChange(e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) { envelopeForm.encrypted_envelope = ''; envelopeForm.envelope_file_name = ''; return; }
  try {
    envelopeForm.encrypted_envelope = (await file.text()).trim();
    envelopeForm.envelope_file_name = file.name;
  } catch (err: any) {
    message.error('读取信封文件失败: ' + (err?.message || ''));
  }
}

function clearEnvelope() {
  envelopeForm.encrypted_envelope = '';
  envelopeForm.envelope_file_name = '';
  if (envelopeFileInputRef.value) envelopeFileInputRef.value.value = '';
}

async function handleQueryEnvelope() {
  if (!envelopeForm.encrypted_envelope.trim()) { message.warning('请输入加密的数字信封'); return; }
  if (envelopeForm.mode === 'cert' && !envelopeForm.cert_id) { message.warning('请选择签名证书'); return; }
  if (envelopeForm.mode === 'manual' && !envelopeForm.sign_key_pem.trim()) { message.warning('请输入签名私钥'); return; }

  querying.value = true;
  try {
    const payload: QueryEnvelopeParams = {
      format: envelopeForm.format,
      mode: envelopeForm.mode,
      encrypted_envelope: envelopeForm.encrypted_envelope.trim(),
    };
    if (envelopeForm.mode === 'cert') {
      payload.cert_id = envelopeForm.cert_id;
    } else {
      payload.sign_key_pem = envelopeForm.sign_key_pem.trim();
      if (envelopeForm.sign_key_password) payload.sign_key_password = envelopeForm.sign_key_password;
    }

    const res: any = await cryptoApi.queryEnvelope(payload);
    envelopeResult.value = (res?.data as QueryEnvelopeResult) || null;
    message.success('解密成功');
  } catch (e: any) {
    envelopeResult.value = null;
    message.error(e.message || '解密失败');
  } finally {
    querying.value = false;
  }
}

function clearEnvelopeResult() {
  envelopeResult.value = null;
}

// ---------------------------------------------------------------------------
// 复制
// ---------------------------------------------------------------------------
async function copyText(text: string) {
  if (!text) return;
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text);
    } else {
      const ta = document.createElement('textarea');
      ta.value = text;
      ta.style.position = 'fixed';
      ta.style.opacity = '0';
      document.body.appendChild(ta);
      ta.select();
      document.execCommand('copy');
      document.body.removeChild(ta);
    }
    message.success('已复制');
  } catch (e: any) {
    message.error('复制失败: ' + (e?.message || ''));
  }
}

// ---------------------------------------------------------------------------
// 删除
// ---------------------------------------------------------------------------
async function handleDelete(row: any) {
  try {
    await cryptoApi.deleteCsr(row.csr_id);
    message.success('删除成功');
    await fetchList();
  } catch (e: any) {
    message.error(e.message || '删除失败');
  }
}

onMounted(async () => {
  await Promise.all([fetchList(), fetchCAOptions()]);
});
</script>

<style scoped>
.dn-row { display: flex; gap: 8px; width: 100%; align-items: flex-start; }
.dn-row .n-input { flex: 1; min-width: 0; }
.dn-help-btn { margin-top: 6px; flex-shrink: 0; font-weight: 700; }
.dn-help { line-height: 1.7; font-size: 13px; }
.dn-help__title { margin: 0 0 4px; font-weight: 600; color: #111827; }
.dn-help__keys { margin: 0 0 10px; color: #2563eb; word-break: break-all; }
.dn-help__example { margin: 0 0 10px; color: #374151; background: #f3f4f6; padding: 6px 8px; border-radius: 4px; word-break: break-all; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 12px; }
.dn-help__tip { margin: 0; color: #6b7280; font-size: 12px; }
.dn-sep-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.dn-sep-hint { color: #6b7280; font-size: 12px; }
.dn-preview { display: flex; flex-wrap: wrap; align-items: center; width: 100%; }
.file-row { display: flex; align-items: flex-start; gap: 12px; flex-wrap: wrap; width: 100%; }
.file-col { display: flex; flex-direction: column; gap: 6px; min-width: 180px; }
.file-name { color: #2563eb; font-size: 13px; word-break: break-all; }
.file-placeholder { color: #9ca3af; font-size: 13px; }
.hint-text { color: #6b7280; font-size: 12px; }
.result-block { border: 1px solid #e5e7eb; border-radius: 6px; padding: 10px 12px; background: #fafafa; }
.result-block--envelope { background: #f0fdf4; border-color: #86efac; }
.result-block__header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; font-size: 13px; color: #111827; }
.envelope-field { margin-top: 10px; }
.envelope-field__header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 4px; }
.envelope-field__label { font-size: 12px; color: #374151; font-weight: 500; }
</style>
