<template>
  <t-card :title="'P10 管理'" :bordered="false">
    <template #actions>
      <t-space>
        <t-button v-permission="'crypto_console:csr:create'" theme="primary" @click="openApply">申请 P10</t-button>
        <t-button v-permission="'crypto_console:csr:create'" @click="openImport">导入 P10</t-button>
        <t-button v-permission="'crypto_console:cert:sign'" theme="success" @click="openSignCert">签发证书</t-button>
        <t-button v-permission="'crypto_console:cert:view'" @click="openQueryEnvelope">查看信封信息</t-button>
        <t-button @click="fetchList">刷新</t-button>
      </t-space>
    </template>

    <t-table
      row-key="csr_id"
      :columns="columns"
      :data="list"
      :loading="loading"
      :pagination="pagination"
      :scroll="{ x: 1200 }"
      @page-change="handlePageChange"
    />
  </t-card>

  <!-- 申请 P10 -->
  <t-dialog :visible="showApply" :header="'申请 P10'" :confirm-btn="null" :cancel-btn="null" :width="720" @close="showApply = false">
    <t-form :data="applyForm" label-align="left" :label-width="110">
      <t-form-item label="算法">
        <t-select v-model="applyForm.algorithm" :options="algStore.options" :loading="algStore.loading" placeholder="选择算法" @change="syncApplyDigest" />
      </t-form-item>
      <t-form-item label="摘要算法">
        <t-select v-model="applyForm.digest_algorithm" :options="applyDigestOptions" placeholder="选择摘要算法" style="width: 260px" />
      </t-form-item>
      <t-form-item label="P10 类型">
        <t-radio-group v-model="applyForm.p10_format">
          <t-radio value="pkcs10">普通 P10（PKCS10）</t-radio>
          <t-radio value="cfca">CFCA P10（携带临时公钥）</t-radio>
        </t-radio-group>
      </t-form-item>
      <t-alert v-if="applyForm.p10_format === 'cfca'" theme="info" :message="'CFCA P10 会在 CSR 中携带平台内置临时公钥信息（作为 CSR attribute，OID 1.2.840.113549.1.9.63），用于后续签发 CFCA 格式双证时封装加密私钥。临时密钥对由平台统一维护。'" style="margin-bottom: 12px" />
      <t-form-item label="使用者 DN">
        <div class="dn-row">
          <t-textarea v-model="applyForm.dn" :autosize="{ minRows: 3, maxRows: 6 }" :placeholder="DN_HELP_EXAMPLE" />
          <t-tooltip placement="left" :content="DN_HELP_KEYS">
            <span class="dn-help-btn">?</span>
          </t-tooltip>
        </div>
      </t-form-item>
      <t-form-item label="DN 分隔符">
        <t-input v-model="applyForm.dnSeparator" placeholder="默认逗号 “,”，可输入 “;” 或 “\\n” 表示换行" style="max-width: 320px" />
      </t-form-item>
      <t-form-item v-if="applyForm.dn.trim()" label="解析预览">
        <div class="dn-preview">
          <t-alert v-if="dnPreview.errors.length" theme="error" :message="dnPreview.errors.join('；')" style="width: 100%" />
          <template v-else>
            <t-tag v-for="(v, k) in dnPreview.subject" :key="k" size="small" theme="success" variant="light" style="margin: 2px 6px 2px 0">
              {{ k }} = {{ v }}
            </t-tag>
          </template>
        </div>
      </t-form-item>
      <t-form-item label="SAN（逗号分隔）">
        <t-input v-model="applyForm.san" placeholder="例如 test.example.com,www.example.com" />
      </t-form-item>
    </t-form>
    <template #footer>
      <t-space style="display: flex; justify-content: flex-end">
        <t-button @click="showApply = false">取消</t-button>
        <t-button theme="primary" :loading="submitting" :disabled="applyDisabled" @click="handleApply">提交</t-button>
      </t-space>
    </template>
  </t-dialog>

  <!-- 申请 P10 结果（CFCA） -->
  <t-dialog :visible="showApplyResult" :header="'CFCA P10 申请结果'" :confirm-btn="null" :cancel-btn="null" :width="900" @close="closeApplyResult">
    <t-alert theme="success" :message="'CFCA P10 申请成功。临时密钥对由平台内置固定维护，无需另行保存；可在 P10 列表中下载该 CSR，用于后续签发 CFCA 格式双证。'" style="margin-bottom: 12px" />
    <t-space direction="vertical" size="medium">
      <template v-if="applyResult.csr_pem">
        <div class="result-block">
          <div class="result-block__header">
            <strong>CSR（PEM）</strong>
            <t-button size="small" @click="copyText(applyResult.csr_pem || '')">复制</t-button>
          </div>
          <t-textarea :value="applyResult.csr_pem" :autosize="{ minRows: 6, maxRows: 12 }" readonly />
        </div>
      </template>
      <template v-if="applyResult.sign_key_pem">
        <div class="result-block">
          <div class="result-block__header">
            <div><strong>签名私钥（sign_key_pem）</strong><t-tag theme="danger" size="small" style="margin-left: 8px">平台不保管</t-tag></div>
            <t-button size="small" @click="copyText(applyResult.sign_key_pem || '')">复制</t-button>
          </div>
          <t-textarea :value="applyResult.sign_key_pem" :autosize="{ minRows: 4, maxRows: 10 }" readonly />
        </div>
      </template>
    </t-space>
    <template #footer>
      <t-space style="display: flex; justify-content: flex-end">
        <t-button theme="primary" @click="closeApplyResult">关闭</t-button>
      </t-space>
    </template>
  </t-dialog>

  <!-- 导入 P10 -->
  <t-dialog :visible="showImport" :header="'导入 P10'" :confirm-btn="null" :cancel-btn="null" :width="680" @close="showImport = false">
    <t-alert theme="info" :message="'上传 P10（CSR）文件；如有对应私钥可一并上传，加密私钥需填写密码。'" style="margin-bottom: 12px" />
    <t-form label-align="top">
      <t-form-item label="P10 文件（PEM，必填）">
        <div class="file-row">
          <input ref="csrFileInputRef" type="file" accept=".csr,.pem,.p10" style="display: none" @change="onCsrFileChange" />
          <t-button @click="csrFileInputRef?.click()">选择文件</t-button>
          <span v-if="importForm.csr_file_name" class="file-name">{{ importForm.csr_file_name }}</span>
          <span v-else class="file-placeholder">未选择文件</span>
        </div>
      </t-form-item>
      <t-form-item label="私钥文件（PEM，可选）">
        <div class="file-row">
          <input ref="keyFileInputRef" type="file" accept=".pem,.key" style="display: none" @change="onKeyFileChange" />
          <t-button @click="keyFileInputRef?.click()">选择私钥</t-button>
          <span v-if="importForm.key_file_name" class="file-name">{{ importForm.key_file_name }}</span>
          <span v-else class="file-placeholder">未选择文件</span>
          <t-button v-if="importForm.key_file_name" size="small" variant="text" @click="clearKeyFile">清除</t-button>
        </div>
      </t-form-item>
      <t-form-item label="私钥密码（仅加密私钥需要）">
        <t-input v-model="importForm.key_password" type="password" placeholder="未加密留空" />
      </t-form-item>
    </t-form>
    <template #footer>
      <t-space style="display: flex; justify-content: flex-end">
        <t-button @click="showImport = false">取消</t-button>
        <t-button theme="primary" :loading="submitting" :disabled="!importForm.csr_pem.trim()" @click="handleImport">导入</t-button>
      </t-space>
    </template>
  </t-dialog>

  <!-- 签发证书 -->
  <t-dialog :visible="showSignCert" :header="'签发证书'" :confirm-btn="null" :cancel-btn="null" :width="900" @close="showSignCert = false">
    <t-form :data="signForm" label-align="left" :label-width="130">
      <t-form-item label="签发模式">
        <t-radio-group v-model="signForm.cert_mode">
          <t-radio value="normal">普通证书</t-radio>
          <t-radio value="dual">国密双证书</t-radio>
        </t-radio-group>
      </t-form-item>
      <t-form-item v-if="signForm.cert_mode === 'dual'" label="证书申请数据格式">
        <t-radio-group v-model="signForm.dual_format">
          <t-radio value="pkcs10">PKCS10</t-radio>
          <t-radio value="cfca">CFCA</t-radio>
        </t-radio-group>
      </t-form-item>
      <t-alert v-if="isCfcaDual" theme="info" :message="'CFCA 格式双证要求 CSR 携带临时公钥信息（OID 1.2.840.113549.1.9.63）。请选择本页「申请 P10 → CFCA P10」生成的 P10。'" style="margin-bottom: 12px" />
      <template v-if="signForm.cert_mode === 'normal'">
        <t-form-item label="证书类型">
          <t-checkbox-group v-model="signForm.cert_types">
            <t-checkbox value="signature">签名 signature</t-checkbox>
            <t-checkbox value="encryption">加密 encryption</t-checkbox>
          </t-checkbox-group>
        </t-form-item>
      </template>
      <t-form-item label="P10 来源">
        <t-radio-group v-model="signForm.csr_source">
          <t-radio value="existing">从本地选择 P10</t-radio>
          <t-radio value="upload">手动输入 CSR 和密钥</t-radio>
        </t-radio-group>
      </t-form-item>
      <template v-if="signForm.csr_source === 'existing'">
        <t-form-item label="选择 P10">
          <t-select v-model="signForm.csr_id" :options="csrOptions" filterable clearable placeholder="请选择 P10" style="width: 100%" />
        </t-form-item>
      </template>
      <template v-else>
        <t-form-item label="CSR PEM">
          <t-textarea v-model="signForm.csr_pem" :autosize="{ minRows: 4, maxRows: 8 }" placeholder="-----BEGIN CERTIFICATE REQUEST-----" />
        </t-form-item>
        <t-form-item label="提供私钥">
          <t-space align="center">
            <t-switch v-model="signForm.with_key" />
            <span class="hint-text">仅当需要平台代为解密数字信封、或从签名证书查询信封信息时，才需要提供</span>
          </t-space>
        </t-form-item>
        <template v-if="signForm.with_key">
          <t-form-item label="私钥 PEM">
            <t-textarea v-model="signForm.csr_key_pem" :autosize="{ minRows: 4, maxRows: 8 }" placeholder="-----BEGIN PRIVATE KEY-----" />
          </t-form-item>
          <t-form-item label="私钥密码">
            <t-input v-model="signForm.csr_key_password" type="password" placeholder="私钥未加密可留空" />
          </t-form-item>
        </template>
      </template>
      <t-form-item label="使用者 DN">
        <div class="dn-row">
          <t-textarea v-model="signForm.dn" :autosize="{ minRows: 3, maxRows: 6 }" :placeholder="`留空则沿用 P10 内的使用者 DN；填写则覆盖（例如 ${DN_HELP_EXAMPLE}）`" />
          <t-tooltip placement="left" :content="DN_HELP_KEYS">
            <span class="dn-help-btn">?</span>
          </t-tooltip>
        </div>
      </t-form-item>
      <t-form-item v-if="signForm.dn.trim()" label="DN 分隔符">
        <t-input v-model="signForm.dnSeparator" placeholder="默认逗号 “,”，可输入 “;” 或 “\\n” 表示换行" style="max-width: 320px" />
      </t-form-item>
      <t-form-item v-if="signForm.dn.trim()" label="解析预览">
        <div class="dn-preview">
          <t-alert v-if="signDnPreview.errors.length" theme="error" :message="signDnPreview.errors.join('；')" style="width: 100%" />
          <template v-else>
            <t-tag v-for="(v, k) in signDnPreview.subject" :key="k" size="small" theme="success" variant="light" style="margin: 2px 6px 2px 0">
              {{ k }} = {{ v }}
            </t-tag>
          </template>
        </div>
      </t-form-item>
      <t-form-item v-if="isCfcaDual" label="SAN（逗号分隔，可选）">
        <t-input v-model="signForm.cfca_san" placeholder="例如 test.example.com,www.example.com" />
      </t-form-item>
      <t-form-item label="CA 来源">
        <t-radio-group v-model="signForm.ca_source">
          <t-radio value="local">本地 CA</t-radio>
          <t-radio value="manual">手动输入 CA</t-radio>
        </t-radio-group>
      </t-form-item>
      <t-form-item v-if="signForm.ca_source === 'local'" label="选择 CA">
        <t-select v-model="signForm.ca_id" :options="caOptions" filterable clearable placeholder="请选择 CA" style="width: 100%" />
      </t-form-item>
      <template v-if="signForm.ca_source === 'manual'">
        <t-form-item label="CA 证书 PEM">
          <t-textarea v-model="signForm.ca_cert_pem" :autosize="{ minRows: 3, maxRows: 6 }" />
        </t-form-item>
        <t-form-item label="CA 私钥 PEM">
          <t-textarea v-model="signForm.ca_key_pem" :autosize="{ minRows: 3, maxRows: 6 }" />
        </t-form-item>
        <t-form-item label="CA 私钥密码">
          <t-input v-model="signForm.ca_key_password" type="password" />
        </t-form-item>
      </template>
      <t-form-item v-if="signForm.cert_mode === 'normal'" label="摘要算法">
        <t-select v-model="signForm.digest_algorithm" :options="signDigestOptions" placeholder="留空则按 CA 类型自动选择" style="width: 100%" />
      </t-form-item>
      <t-form-item label="有效期天数">
        <t-input-number v-model="signForm.validity_days" :min="1" />
      </t-form-item>
    </t-form>
    <template #footer>
      <t-space style="display: flex; justify-content: flex-end">
        <t-button @click="showSignCert = false">取消</t-button>
        <t-button theme="primary" :loading="submitting" :disabled="signDisabled" @click="handleSignCert">签发</t-button>
      </t-space>
    </template>
  </t-dialog>

  <!-- 签发结果弹窗 -->
  <t-dialog :visible="showSignResult" :header="'签发结果'" :confirm-btn="null" :cancel-btn="null" :width="900" @close="closeSignResult">
    <t-alert theme="warning" :message="'请立即复制并妥善保存以下内容。弹窗关闭后，内容将无法再次查看。'" style="margin-bottom: 12px" />
    <t-space direction="vertical" size="medium">
      <template v-if="signResult.cert_pem">
        <div class="result-block">
          <div class="result-block__header"><strong>证书（PEM）</strong><t-button size="small" @click="copyText(signResult.cert_pem || '')">复制</t-button></div>
          <t-textarea :value="signResult.cert_pem" :autosize="{ minRows: 6, maxRows: 12 }" readonly />
        </div>
      </template>
      <template v-if="signResult.key_pem">
        <div class="result-block">
          <div class="result-block__header"><strong>私钥（PEM）</strong><t-button size="small" @click="copyText(signResult.key_pem || '')">复制</t-button></div>
          <t-textarea :value="signResult.key_pem" :autosize="{ minRows: 6, maxRows: 12 }" readonly />
        </div>
      </template>
      <template v-if="signResult.sign_cert_pem">
        <div class="result-block">
          <div class="result-block__header"><strong>签名证书（PEM）</strong><t-button size="small" @click="copyText(signResult.sign_cert_pem || '')">复制</t-button></div>
          <t-textarea :value="signResult.sign_cert_pem" :autosize="{ minRows: 6, maxRows: 12 }" readonly />
        </div>
      </template>
      <template v-if="signResult.enc_cert_pem">
        <div class="result-block">
          <div class="result-block__header"><strong>加密证书（PEM）</strong><t-button size="small" @click="copyText(signResult.enc_cert_pem || '')">复制</t-button></div>
          <t-textarea :value="signResult.enc_cert_pem" :autosize="{ minRows: 6, maxRows: 12 }" readonly />
        </div>
      </template>
      <template v-if="signResult.encrypted_envelope">
        <div class="result-block result-block--envelope">
          <div class="result-block__header">
            <div><strong>加密的数字信封</strong><t-tag theme="success" size="small" style="margin-left: 8px">请保存</t-tag></div>
            <t-button size="small" @click="copyText(signResult.encrypted_envelope || '')">复制</t-button>
          </div>
          <t-alert theme="info" :message="'该字符串为 CFCA 5 段文本格式，请完整保存。可在本页「查看信封信息」中选择 CFCA 格式，粘贴信封即可由平台自动解密出加密私钥。'" style="margin-bottom: 8px" />
          <t-textarea :value="signResult.encrypted_envelope" :autosize="{ minRows: 8, maxRows: 16 }" readonly />
        </div>
      </template>
      <template v-if="signResult.sign_key_pem">
        <div class="result-block">
          <div class="result-block__header"><div><strong>签名私钥（sign_key_pem）</strong><t-tag theme="danger" size="small" style="margin-left: 8px">平台不保管</t-tag></div><t-button size="small" @click="copyText(signResult.sign_key_pem || '')">复制</t-button></div>
          <t-textarea :value="signResult.sign_key_pem" :autosize="{ minRows: 4, maxRows: 10 }" readonly />
        </div>
      </template>
    </t-space>
    <template #footer>
      <t-space style="display: flex; justify-content: space-between; width: 100%">
        <t-button v-if="signResult.encrypted_envelope" @click="decodeCfcaFromSignResult">去解信封</t-button>
        <t-button @click="closeSignResult">关闭</t-button>
      </t-space>
    </template>
  </t-dialog>

  <!-- 查看信封信息 -->
  <t-dialog :visible="showQueryEnvelope" :header="'查看信封信息'" :confirm-btn="null" :cancel-btn="null" :width="900" @close="showQueryEnvelope = false">
    <t-form :data="envelopeForm" label-align="left" :label-width="140">
      <t-form-item label="信封格式">
        <t-radio-group v-model="envelopeForm.format" @change="onEnvelopeFormatChange">
          <t-radio value="pkcs10">PKCS10 格式信封</t-radio>
          <t-radio value="cfca">CFCA 格式信封</t-radio>
        </t-radio-group>
      </t-form-item>
      <template v-if="envelopeForm.format === 'pkcs10'">
        <t-form-item label="查询方式">
          <t-radio-group v-model="envelopeForm.mode">
            <t-radio value="cert">从签名证书解析（若签名证书关联了私钥，自动使用其解密）</t-radio>
            <t-radio value="manual">手动输入（输入您的 CSR 私钥解密）</t-radio>
          </t-radio-group>
        </t-form-item>
        <template v-if="envelopeForm.mode === 'cert'">
          <t-form-item label="选择签名证书">
            <t-select v-model="envelopeForm.cert_id" :options="signCertOptions" :loading="signCertLoading" filterable clearable placeholder="请选择国密双证中的签名证书" style="width: 100%" />
          </t-form-item>
        </template>
        <template v-else>
          <t-alert theme="warning" :message="'注意：若您导入 P10 时未上传私钥，平台无法自动解密，必须使用本模式上传您的 CSR 私钥。'" style="margin-bottom: 12px" />
          <t-form-item label="签名私钥 PEM">
            <t-textarea v-model="envelopeForm.sign_key_pem" :autosize="{ minRows: 4, maxRows: 8 }" placeholder="-----BEGIN PRIVATE KEY-----" />
          </t-form-item>
          <t-form-item label="签名私钥密码">
            <t-input v-model="envelopeForm.sign_key_password" type="password" placeholder="私钥未加密可留空" />
          </t-form-item>
        </template>
      </template>
      <template v-else>
        <t-alert theme="info" :message="'平台使用内置固定临时密钥解密 CFCA 信封，您只需粘贴信封内容即可。'" style="margin-bottom: 12px" />
      </template>
      <t-form-item label="加密的数字信封">
        <t-textarea v-model="envelopeForm.encrypted_envelope" :autosize="{ minRows: 6, maxRows: 12 }" placeholder="粘贴签发时返回的加密数字信封字符串" />
      </t-form-item>
      <t-form-item>
        <t-space>
          <t-button theme="primary" :loading="querying" :disabled="queryDisabled" @click="handleQueryEnvelope">解密信封</t-button>
          <t-button @click="clearEnvelopeResult">清空结果</t-button>
        </t-space>
      </t-form-item>
    </t-form>

    <template v-if="envelopeResult">
      <t-divider />
      <template v-if="envelopeResult.kind === 'pkcs10'">
        <div class="envelope-field">
          <div class="envelope-field__header"><span class="envelope-field__label">对称密钥密文（symmetric_key_cipher）</span><t-button size="small" @click="copyText(envelopeResult.symmetric_key_cipher)">复制</t-button></div>
          <t-textarea :value="envelopeResult.symmetric_key_cipher" :autosize="{ minRows: 3, maxRows: 6 }" readonly />
        </div>
        <div class="envelope-field">
          <div class="envelope-field__header"><span class="envelope-field__label">IV（base64）</span><t-button size="small" @click="copyText(envelopeResult.iv)">复制</t-button></div>
          <t-input :value="envelopeResult.iv" readonly />
        </div>
        <div class="envelope-field">
          <div class="envelope-field__header"><span class="envelope-field__label">加密私钥密文（encrypted_private_key）</span><t-button size="small" @click="copyText(envelopeResult.encrypted_private_key)">复制</t-button></div>
          <t-textarea :value="envelopeResult.encrypted_private_key" :autosize="{ minRows: 6, maxRows: 12 }" readonly />
        </div>
        <div v-if="envelopeResult.decrypted_key_pem" class="envelope-field result-block--envelope">
          <div class="envelope-field__header"><span class="envelope-field__label">解密后的加密私钥（PEM）</span><t-button size="small" @click="copyText(envelopeResult.decrypted_key_pem || '')">复制</t-button></div>
          <t-textarea :value="envelopeResult.decrypted_key_pem" :autosize="{ minRows: 6, maxRows: 12 }" readonly />
        </div>
      </template>
      <template v-else>
        <div class="envelope-field result-block--envelope">
          <div class="envelope-field__header"><span class="envelope-field__label">解密后的加密私钥（PEM）</span><t-button size="small" @click="copyText(envelopeResult.decrypted_key_pem)">复制</t-button></div>
          <t-textarea :value="envelopeResult.decrypted_key_pem" :autosize="{ minRows: 6, maxRows: 12 }" readonly />
        </div>
      </template>
    </template>

    <template #footer>
      <t-space style="display: flex; justify-content: flex-end">
        <t-button @click="showQueryEnvelope = false">关闭</t-button>
      </t-space>
    </template>
  </t-dialog>

  <!-- 下载密钥口令输入 -->
  <t-dialog :visible="showDownloadKey" :header="'下载 P10 私钥'" :confirm-btn="null" :cancel-btn="null" :width="480" @close="showDownloadKey = false">
    <t-form>
      <t-form-item label="使用口令加密">
        <t-switch v-model="downloadKeyForm.encrypt" />
      </t-form-item>
      <t-form-item v-if="downloadKeyForm.encrypt" label="加密口令">
        <t-input v-model="downloadKeyForm.password" type="password" placeholder="请输入加密口令" />
      </t-form-item>
    </t-form>
    <template #footer>
      <t-space style="display: flex; justify-content: flex-end">
        <t-button @click="showDownloadKey = false">取消</t-button>
        <t-button theme="primary" :disabled="downloadKeyForm.encrypt && !downloadKeyForm.password" @click="handleDownloadKey">下载</t-button>
      </t-space>
    </template>
  </t-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, h } from 'vue';
import { Tag, Space, Button, Popconfirm } from 'tdesign-vue-next';
import type { PrimaryTableCol, PageInfo } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import {
  cryptoApi,
  type SignCertParams,
  type SignCertResult,
  type QueryEnvelopeParams,
  type QueryEnvelopeResult,
  type CertTypeItem,
  type ImportCsrParams,
  type P10Format,
  type CSRCreateResult,
  type CFCASignParams,
  type CFCASignResult,
  type CFCADecodeEnvelopeParams,
  type CFCADecodeEnvelopeResult,
} from '@/api/crypto';
import {
  DEFAULT_DN_SEPARATOR,
  DN_HELP_KEYS,
  DN_HELP_EXAMPLE,
  parseDN,
  validateDN,
} from '@/utils/dn';
import { useAlgorithmStore } from '@/store/algorithm';
import {
  algorithmFamily,
  defaultDigestFor,
  defaultDigestForAlgorithm,
  digestOptionsFor,
  type DigestId,
} from '@/utils/digest';

const algStore = useAlgorithmStore();

const loading = ref(false);
const submitting = ref(false);
const querying = ref(false);

const showApply = ref(false);
const showApplyResult = ref(false);
const showImport = ref(false);
const showSignCert = ref(false);
const showDownloadKey = ref(false);
const showSignResult = ref(false);
const showQueryEnvelope = ref(false);

const list = ref<any[]>([]);
const caList = ref<any[]>([]);

const signResult = ref<SignCertResult & Partial<CFCASignResult>>({});

const applyResult = ref<{ csr_id: string; csr_pem: string; sign_key_pem: string }>({
  csr_id: '',
  csr_pem: '',
  sign_key_pem: '',
});

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showJumper: true,
});

const applyForm = reactive({
  algorithm: 'SM2',
  p10_format: 'pkcs10' as P10Format,
  /** 签名摘要算法（SM3 / SHA256 / SHA384 / SHA512） */
  digest_algorithm: 'SM3' as DigestId | '',
  dn: '',
  dnSeparator: DEFAULT_DN_SEPARATOR,
  san: '',
});

/** 申请 P10 表单：算法族 / 摘要选项 / 提示 */
const applyFamily = computed(() => algorithmFamily(applyForm.algorithm, algStore.find(applyForm.algorithm)));
const applyDigestOptions = computed(() =>
  digestOptionsFor(applyFamily.value, applyForm.algorithm, algStore.find(applyForm.algorithm)).map((o) => ({
    label: o.label,
    value: o.value,
  })),
);

/** 切换公钥算法时同步摘要默认值 */
function syncApplyDigest() {
  applyForm.digest_algorithm = defaultDigestForAlgorithm(applyForm.algorithm, algStore.find(applyForm.algorithm));
}

const dnPreview = computed(() => parseDN(applyForm.dn, applyForm.dnSeparator));
const applyDisabled = computed(
  () => !applyForm.dn.trim() || dnPreview.value.errors.length > 0 || !applyForm.algorithm,
);

const csrFileInputRef = ref<HTMLInputElement | null>(null);
const keyFileInputRef = ref<HTMLInputElement | null>(null);
const importForm = reactive({
  csr_pem: '',
  csr_file_name: '',
  key_pem: '',
  key_file_name: '',
  key_password: '',
});

const signForm = reactive({
  csr_source: 'existing' as 'existing' | 'upload',
  csr_id: '',
  csr_pem: '',
  with_key: false,
  csr_key_pem: '',
  csr_key_password: '',
  dn: '',
  dnSeparator: DEFAULT_DN_SEPARATOR,
  cert_mode: 'normal' as 'normal' | 'dual',
  dual_format: 'pkcs10' as 'pkcs10' | 'cfca',
  cfca_san: '',
  cert_types: ['signature'] as CertTypeItem[],
  ca_source: 'local' as 'local' | 'manual',
  ca_id: '',
  ca_cert_pem: '',
  ca_key_pem: '',
  ca_key_password: '',
  /** 签名摘要算法（普通证书生效）：空则由后端按 CA 公钥算法自动选择 */
  digest_algorithm: '' as DigestId | '',
  validity_days: 365,
});

/** 签发表单：按所选本地 CA 的公钥算法给出可选摘要 */
const signFamily = computed(() => {
  const ca = caList.value.find((c: any) => c.ca_id === signForm.ca_id) as any;
  const alg = ca?.algorithm || ca?.public_key_algorithm || '';
  return algorithmFamily(alg, algStore.find(alg));
});

/** 所选 CA 的算法 ID（用于展示后量子算法自带的摘要名） */
const signCaAlgorithm = computed(() => {
  const ca = caList.value.find((c: any) => c.ca_id === signForm.ca_id) as any;
  return (ca?.algorithm || ca?.public_key_algorithm || '') as string;
});

const signDigestOptions = computed(() => [
  { label: '自动（按 CA 公钥算法选择）', value: '' },
  ...digestOptionsFor(signFamily.value, signCaAlgorithm.value).map((o) => ({
    label: o.label,
    value: o.value,
  })),
]);

const signDnPreview = computed(() => parseDN(signForm.dn, signForm.dnSeparator));

const isCfcaDual = computed(
  () => signForm.cert_mode === 'dual' && signForm.dual_format === 'cfca',
);

const envelopeForm = reactive({
  format: 'pkcs10' as 'pkcs10' | 'cfca',
  mode: 'cert' as 'cert' | 'manual',
  cert_id: '',
  sign_key_pem: '',
  sign_key_password: '',
  encrypted_envelope: '',
});

const signCertList = ref<any[]>([]);
const signCertLoading = ref(false);

const signCertOptions = computed(() =>
  signCertList.value.map((c) => ({
    label: `${c.subject_cn || c.cert_id} (${c.cert_id})`,
    value: c.cert_id,
  })),
);

async function fetchSignCertOptions() {
  signCertLoading.value = true;
  try {
    const res = await cryptoApi.listCerts({ page: 1, page_size: 200, cert_type: 'dual_sign' });
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
  if (envelopeForm.format === 'cfca') return false;
  if (envelopeForm.mode === 'cert') return !envelopeForm.cert_id;
  return !envelopeForm.sign_key_pem.trim();
});

type EnvelopeResult =
  | (QueryEnvelopeResult & { kind: 'pkcs10' })
  | (CFCADecodeEnvelopeResult & { kind: 'cfca' });

const envelopeResult = ref<EnvelopeResult | null>(null);

const downloadKeyForm = reactive({ csrId: '', encrypt: false, password: '' });

const columns: PrimaryTableCol<any>[] = [
  {
    colKey: 'subject',
    title: 'DN',
    width: 280,
    ellipsis: true,
    cell: (_h, { row }) => row.subject || row.subject_cn || '-',
  },
  {
    colKey: 'public_key_algorithm',
    title: '密钥算法',
    width: 140,
    cell: (_h, { row }) => algStore.label(row.algorithm || row.public_key_algorithm),
  },
  {
    colKey: 'is_cfca',
    title: '是否为CFCA',
    width: 130,
    cell: (_h, { row }) => {
      const isCfca = !!row.is_cfca;
      return h(Tag, { theme: isCfca ? 'success' : 'default', variant: 'light' }, () => (isCfca ? '是' : '否'));
    },
  },
  {
    colKey: 'cert_config_status',
    title: '证书配置状态',
    width: 150,
    cell: () => h(Tag, { theme: 'warning', variant: 'light' }, () => '未配置'),
  },
  {
    colKey: 'actions',
    title: '操作',
    width: 280,
    fixed: 'right',
    cell: (_h, { row }) =>
      h(Space, { size: 4 }, () => [
        h(Button, { size: 'small', variant: 'text', theme: 'primary', onClick: () => handleDownloadCsr(row) }, () => '下载P10'),
        h(Button, { size: 'small', variant: 'text', theme: 'primary', onClick: () => openDownloadKey(row) }, () => '下载密钥'),
        h(
          Popconfirm,
          { content: `确认删除 P10「${row.subject || row.csr_id}」？`, confirmBtn: '确认删除', cancelBtn: '取消', onConfirm: () => handleDelete(row) },
          () => h(Button, { size: 'small', variant: 'text', theme: 'danger' }, () => '删除'),
        ),
      ]),
  },
];

const csrOptions = computed(() =>
  list.value.map((c) => ({
    label: `${c.subject || c.subject_cn || c.csr_id} (${c.csr_id})`,
    value: c.csr_id,
  })),
);

const caOptions = computed(() =>
  caList.value.map((c) => ({
    label: `${c.subject_cn || c.ca_id} (${c.ca_id})`,
    value: c.ca_id,
  })),
);

const signDisabled = computed(() => {
  if (signForm.dn.trim() && signDnPreview.value.errors.length > 0) return true;
  if (signForm.ca_source === 'local' && !signForm.ca_id) return true;
  if (signForm.ca_source === 'manual' && (!signForm.ca_cert_pem.trim() || !signForm.ca_key_pem.trim())) return true;

  if (signForm.csr_source === 'existing' && !signForm.csr_id) return true;
  if (signForm.csr_source === 'upload') {
    if (!signForm.csr_pem.trim()) return true;
    if (signForm.with_key && !signForm.csr_key_pem.trim()) return true;
  }

  if (signForm.cert_mode === 'normal' && signForm.cert_types.length === 0) return true;
  return false;
});

async function fetchList() {
  loading.value = true;
  try {
    const res = await cryptoApi.listCsrs({ page: pagination.current, page_size: pagination.pageSize });
    const data: any = res.data || {};
    list.value = data.items || data || [];
    pagination.total = data.total || list.value.length;
  } catch (e: any) {
    MessagePlugin.error(e.message || '加载失败');
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

function handlePageChange(info: PageInfo) {
  pagination.current = info.current;
  pagination.pageSize = info.pageSize;
  fetchList();
}

function openApply() {
  applyForm.algorithm = 'SM2';
  applyForm.p10_format = 'pkcs10';
  applyForm.digest_algorithm = 'SM3';
  applyForm.dn = '';
  applyForm.dnSeparator = DEFAULT_DN_SEPARATOR;
  applyForm.san = '';
  showApply.value = true;
}

async function handleApply() {
  const r = validateDN(applyForm.dn, applyForm.dnSeparator);
  if (r.errors.length) {
    MessagePlugin.error(r.errors[0]);
    return;
  }
  submitting.value = true;
  try {
    const params: Record<string, unknown> = {
      algorithm: applyForm.algorithm,
      subject: r.subject,
      key_source: 'generate',
      p10_format: applyForm.p10_format,
    };
    // 签名摘要算法：PQC 无需摘要；CFCA 格式固定 SM3
    if (applyFamily.value !== 'PQC' && applyForm.digest_algorithm) {
      params.digest_algorithm = applyForm.digest_algorithm;
    }
    const sanList = applyForm.san.split(',').map((s) => s.trim()).filter(Boolean);
    if (sanList.length) params.san = sanList;

    const res: any = await cryptoApi.execute<CSRCreateResult>('csr.create', params);
    const data: CSRCreateResult = (res?.data as CSRCreateResult) || {};

    if (applyForm.p10_format === 'cfca') {
      applyResult.value = {
        csr_id: data.csr_id || '',
        csr_pem: data.csr_pem || '',
        sign_key_pem: data.sign_key_pem || '',
      };
      showApply.value = false;
      showApplyResult.value = true;
      await fetchList();
      return;
    }

    MessagePlugin.success('P10 申请成功');
    showApply.value = false;
    await fetchList();
  } catch (e: any) {
    MessagePlugin.error(e.message || '申请失败');
  } finally {
    submitting.value = false;
  }
}

function closeApplyResult() {
  showApplyResult.value = false;
  applyResult.value = { csr_id: '', csr_pem: '', sign_key_pem: '' };
}

function openImport() {
  importForm.csr_pem = '';
  importForm.csr_file_name = '';
  importForm.key_pem = '';
  importForm.key_file_name = '';
  importForm.key_password = '';
  if (csrFileInputRef.value) csrFileInputRef.value.value = '';
  if (keyFileInputRef.value) keyFileInputRef.value.value = '';
  showImport.value = true;
}

async function onCsrFileChange(e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) {
    importForm.csr_pem = '';
    importForm.csr_file_name = '';
    return;
  }
  try {
    importForm.csr_pem = await file.text();
    importForm.csr_file_name = file.name;
  } catch (err: any) {
    MessagePlugin.error('读取文件失败: ' + (err?.message || ''));
  }
}

async function onKeyFileChange(e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) {
    importForm.key_pem = '';
    importForm.key_file_name = '';
    return;
  }
  try {
    importForm.key_pem = await file.text();
    importForm.key_file_name = file.name;
  } catch (err: any) {
    MessagePlugin.error('读取文件失败: ' + (err?.message || ''));
  }
}

function clearKeyFile() {
  importForm.key_pem = '';
  importForm.key_file_name = '';
  importForm.key_password = '';
  if (keyFileInputRef.value) keyFileInputRef.value.value = '';
}

async function handleImport() {
  if (!importForm.csr_pem.trim()) {
    MessagePlugin.warning('请选择 P10 文件');
    return;
  }
  submitting.value = true;
  try {
    const payload: ImportCsrParams = { csr_pem: importForm.csr_pem.trim() };
    if (importForm.key_pem.trim()) {
      payload.key_pem = importForm.key_pem.trim();
      if (importForm.key_password) payload.key_password = importForm.key_password;
    }
    await cryptoApi.importCsr(payload);
    MessagePlugin.success('导入成功');
    showImport.value = false;
    await fetchList();
  } catch (e: any) {
    MessagePlugin.error(e.message || '导入失败');
  } finally {
    submitting.value = false;
  }
}

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
    MessagePlugin.success('下载成功');
  } catch (e: any) {
    MessagePlugin.error(e.message || '下载失败');
  }
}

function openDownloadKey(row: any) {
  downloadKeyForm.csrId = row.csr_id;
  downloadKeyForm.encrypt = false;
  downloadKeyForm.password = '';
  showDownloadKey.value = true;
}

async function handleDownloadKey() {
  if (downloadKeyForm.encrypt && !downloadKeyForm.password) {
    MessagePlugin.warning('请输入加密口令');
    return;
  }
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
    MessagePlugin.success('下载成功');
  } catch (e: any) {
    MessagePlugin.error(e.message || '下载失败');
  }
}

function openSignCert() {
  signForm.csr_source = 'existing';
  signForm.csr_id = '';
  signForm.csr_pem = '';
  signForm.with_key = false;
  signForm.csr_key_pem = '';
  signForm.csr_key_password = '';
  signForm.dn = '';
  signForm.dnSeparator = DEFAULT_DN_SEPARATOR;
  signForm.cert_mode = 'normal';
  signForm.dual_format = 'pkcs10';
  signForm.cfca_san = '';
  signForm.cert_types = ['signature'];
  signForm.ca_source = 'local';
  signForm.ca_id = '';
  signForm.ca_cert_pem = '';
  signForm.ca_key_pem = '';
  signForm.ca_key_password = '';
  signForm.digest_algorithm = '';
  signForm.validity_days = 365;
  if (caList.value.length === 0) fetchCAOptions();
  showSignCert.value = true;
}

async function handleSignCert() {
  let subjectObj: Record<string, string> | undefined = undefined;
  if (signForm.dn.trim()) {
    const r = validateDN(signForm.dn, signForm.dnSeparator);
    if (r.errors.length) {
      MessagePlugin.error(r.errors[0]);
      return;
    }
    subjectObj = r.subject;
  }

  if (isCfcaDual.value) {
    submitting.value = true;
    try {
      const payload: CFCASignParams = {
        ca_source: signForm.ca_source,
        csr_source: signForm.csr_source,
        validity_days: signForm.validity_days,
        return_sign_key: true,
      };
      if (signForm.ca_source === 'local') {
        payload.ca_id = signForm.ca_id;
      } else {
        payload.ca_cert_pem = signForm.ca_cert_pem;
        payload.ca_key_pem = signForm.ca_key_pem;
        if (signForm.ca_key_password) payload.ca_key_password = signForm.ca_key_password;
      }
      if (signForm.csr_source === 'existing') {
        payload.csr_id = signForm.csr_id;
      } else if (signForm.csr_source === 'upload') {
        payload.csr_pem = signForm.csr_pem;
        if (signForm.with_key && signForm.csr_key_pem.trim()) {
          payload.csr_key_pem = signForm.csr_key_pem;
          if (signForm.csr_key_password) payload.csr_key_password = signForm.csr_key_password;
        }
      }
      if (subjectObj) payload.subject = subjectObj;
      if (signForm.cfca_san.trim()) {
        const sanList = signForm.cfca_san.split(',').map((s) => s.trim()).filter(Boolean);
        if (sanList.length) payload.san = sanList;
      }

      const res: any = await cryptoApi.cfcaSign(payload);
      signResult.value = (res?.data as SignCertResult & CFCASignResult) || {};
      showSignCert.value = false;
      showSignResult.value = true;
      await fetchList();
    } catch (e: any) {
      MessagePlugin.error(e.message || 'CFCA 双证签发失败');
    } finally {
      submitting.value = false;
    }
    return;
  }

  submitting.value = true;
  try {
    const payload: SignCertParams = {
      ca_source: signForm.ca_source,
      validity_days: signForm.validity_days,
      csr_source: signForm.csr_source,
      cert_mode: signForm.cert_mode,
    };
    if (subjectObj) payload.subject = subjectObj;
    if (signForm.csr_source === 'existing') {
      payload.csr_id = signForm.csr_id;
    } else {
      payload.csr_pem = signForm.csr_pem;
      if (signForm.with_key && signForm.csr_key_pem.trim()) {
        payload.csr_key_pem = signForm.csr_key_pem;
        if (signForm.csr_key_password) payload.csr_key_password = signForm.csr_key_password;
      }
    }
    if (signForm.cert_mode === 'normal') {
      payload.cert_types = [...signForm.cert_types];
    } else {
      payload.algorithm = 'SM2';
    }
    // ★ 签名摘要算法：仅普通证书可指定；国密双证固定 SM3，不传
    if (signForm.ca_source === 'local' && signForm.cert_mode === 'normal') {
      // PQC 的摘要由子算法自带，后端不接受显式值，这里不传
      if (signFamily.value !== 'PQC' && signForm.digest_algorithm) {
        payload.digest_algorithm = signForm.digest_algorithm;
      }
    }
    if (signForm.ca_source === 'local') {
      payload.ca_id = signForm.ca_id;
    } else {
      payload.ca_cert_pem = signForm.ca_cert_pem;
      payload.ca_key_pem = signForm.ca_key_pem;
      if (signForm.ca_key_password) payload.ca_key_password = signForm.ca_key_password;
    }

    const res: any = await cryptoApi.signCert(payload);
    signResult.value = (res?.data as SignCertResult) || {};
    showSignCert.value = false;
    showSignResult.value = true;
    await fetchList();
  } catch (e: any) {
    MessagePlugin.error(e.message || '签发失败');
  } finally {
    submitting.value = false;
  }
}

function closeSignResult() {
  showSignResult.value = false;
  signResult.value = {};
}

function decodeCfcaFromSignResult() {
  const env = signResult.value.encrypted_envelope || '';
  closeSignResult();
  openQueryEnvelope();
  envelopeForm.format = 'cfca';
  envelopeForm.encrypted_envelope = env;
  envelopeResult.value = null;
}

function openQueryEnvelope() {
  envelopeForm.format = 'pkcs10';
  envelopeForm.mode = 'cert';
  envelopeForm.cert_id = '';
  envelopeForm.sign_key_pem = '';
  envelopeForm.sign_key_password = '';
  envelopeForm.encrypted_envelope = '';
  if (signCertList.value.length === 0) fetchSignCertOptions();
  envelopeResult.value = null;
  showQueryEnvelope.value = true;
}

function onEnvelopeFormatChange(v: 'pkcs10' | 'cfca') {
  envelopeResult.value = null;
  if (v === 'cfca') {
    envelopeForm.mode = 'cert';
    envelopeForm.cert_id = '';
    envelopeForm.sign_key_pem = '';
    envelopeForm.sign_key_password = '';
  }
}

async function handleQueryEnvelope() {
  if (!envelopeForm.encrypted_envelope.trim()) {
    MessagePlugin.warning('请输入加密的数字信封');
    return;
  }
  if (envelopeForm.format === 'cfca') {
    querying.value = true;
    try {
      const payload: CFCADecodeEnvelopeParams = { encrypted_envelope: envelopeForm.encrypted_envelope.trim() };
      const res: any = await cryptoApi.cfcaDecodeEnvelope(payload);
      envelopeResult.value = { kind: 'cfca', ...((res?.data as CFCADecodeEnvelopeResult) || {}) } as EnvelopeResult;
      MessagePlugin.success('解密成功');
    } catch (e: any) {
      envelopeResult.value = null;
      MessagePlugin.error(e.message || '解密失败');
    } finally {
      querying.value = false;
    }
    return;
  }
  if (envelopeForm.mode === 'cert' && !envelopeForm.cert_id) {
    MessagePlugin.warning('请选择签名证书');
    return;
  }
  if (envelopeForm.mode === 'manual' && !envelopeForm.sign_key_pem.trim()) {
    MessagePlugin.warning('请输入签名私钥');
    return;
  }
  querying.value = true;
  try {
    const payload: QueryEnvelopeParams = {
      format: 'pkcs10',
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
    envelopeResult.value = { kind: 'pkcs10', ...((res?.data as QueryEnvelopeResult) || {}) } as EnvelopeResult;
    MessagePlugin.success('解密成功');
  } catch (e: any) {
    envelopeResult.value = null;
    MessagePlugin.error(e.message || '解密失败');
  } finally {
    querying.value = false;
  }
}

function clearEnvelopeResult() {
  envelopeResult.value = null;
}

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
    MessagePlugin.success('已复制');
  } catch (e: any) {
    MessagePlugin.error('复制失败: ' + (e?.message || ''));
  }
}

async function handleDelete(row: any) {
  try {
    await cryptoApi.deleteCsr(row.csr_id);
    MessagePlugin.success('删除成功');
    await fetchList();
  } catch (e: any) {
    MessagePlugin.error(e.message || '删除失败');
  }
}

onMounted(async () => {
  await Promise.all([fetchList(), fetchCAOptions(), algStore.ensureLoaded()]);
});
</script>

<style scoped>
.dn-row {
  display: flex;
  gap: 8px;
  width: 100%;
  align-items: flex-start;
}
.dn-help-btn {
  margin-top: 6px;
  flex-shrink: 0;
  font-weight: 700;
  cursor: help;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  border: 1px solid #909399;
  color: #909399;
  font-size: 11px;
}
.dn-sep-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.dn-sep-hint {
  color: #6b7280;
  font-size: 12px;
}
.dn-preview {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  width: 100%;
}
.file-row {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  flex-wrap: wrap;
  width: 100%;
}
.file-name {
  color: #2563eb;
  font-size: 13px;
  word-break: break-all;
}
.file-placeholder {
  color: #9ca3af;
  font-size: 13px;
}
.hint-text {
  color: #6b7280;
  font-size: 12px;
}
.result-block {
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  padding: 10px 12px;
  background: #fafafa;
  width: 100%;
}
.result-block--envelope {
  background: #f0fdf4;
  border-color: #86efac;
}
.result-block__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-size: 13px;
  color: #111827;
}
.envelope-field {
  margin-top: 10px;
}
.envelope-field__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}
.envelope-field__label {
  font-size: 12px;
  color: #374151;
  font-weight: 500;
}
</style>