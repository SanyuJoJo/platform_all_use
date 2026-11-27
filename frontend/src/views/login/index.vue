<template>
  <div class="login-page">
    <div class="login-shell">
      <!-- 左侧品牌区 -->
      <section class="brand-panel">
        <div class="brand-orbit" aria-hidden="true">
          <svg viewBox="0 0 640 640" fill="none" xmlns="http://www.w3.org/2000/svg">
            <defs>
              <radialGradient id="loginHalo" cx="0" cy="0" r="1" gradientUnits="userSpaceOnUse" gradientTransform="translate(320 320) rotate(90) scale(260)">
                <stop stop-color="#3B82F6" stop-opacity="0.32" />
                <stop offset="1" stop-color="#3B82F6" stop-opacity="0" />
              </radialGradient>
              <linearGradient id="loginShield" x1="230" y1="220" x2="420" y2="430" gradientUnits="userSpaceOnUse">
                <stop stop-color="#60A5FA" stop-opacity="0.42" />
                <stop offset="1" stop-color="#22D3EE" stop-opacity="0.12" />
              </linearGradient>
            </defs>

            <circle cx="320" cy="320" r="252" fill="url(#loginHalo)" />
            <circle cx="320" cy="320" r="238" stroke="#93C5FD" stroke-opacity="0.16" />
            <circle cx="320" cy="320" r="184" stroke="#93C5FD" stroke-opacity="0.22" stroke-dasharray="3 10" class="orbit-dashed" />
            <circle cx="320" cy="320" r="126" stroke="#93C5FD" stroke-opacity="0.14" />

            <g stroke="#60A5FA" stroke-opacity="0.32">
              <path d="M320 82 500 186 500 454 320 558 140 454 140 186Z" />
              <path d="M320 82v476M140 186l360 268M500 186 140 454" stroke-opacity="0.5" />
            </g>

            <path
              d="M320 196 414 232v86c0 64-42 110-94 133-52-23-94-69-94-133v-86l94-36Z"
              fill="url(#loginShield)"
              stroke="#7DD3FC"
              stroke-width="2"
              stroke-linejoin="round"
            />
            <path d="m282 316 27 27 54-54" stroke="#E0F2FE" stroke-width="8" stroke-linecap="round" stroke-linejoin="round" />

            <g fill="#67E8F9">
              <circle cx="320" cy="82" r="5" />
              <circle cx="500" cy="186" r="4" />
              <circle cx="500" cy="454" r="5" />
              <circle cx="320" cy="558" r="4" />
              <circle cx="140" cy="454" r="5" />
              <circle cx="140" cy="186" r="4" />
            </g>
            <g fill="#67E8F9" fill-opacity="0.2">
              <circle cx="320" cy="82" r="13" />
              <circle cx="500" cy="454" r="13" />
              <circle cx="140" cy="454" r="13" />
            </g>
          </svg>
        </div>

        <header class="brand-header">
          <BrandMark :size="44" />
          <div class="brand-name">
            <span class="brand-cn">多功能管理平台</span>
            <span class="brand-en">ENTERPRISE MANAGEMENT PLATFORM</span>
          </div>
        </header>

        <div class="brand-copy">
          <div class="eyebrow">
            <span class="eyebrow-dot"></span>
            企业级安全运维控制台
          </div>
          <h1>
            让每一次连接<br />
            <span class="gradient-text">安全、可控、可信</span>
          </h1>
          <p class="lead">
            面向企业关键业务的统一管理平台，集中管理 SSL 代理、证书、加密服务与音视频资源，
            为业务连续性提供坚实底座。
          </p>

          <ul class="feature-list">
            <li v-for="item in features" :key="item.title">
              <span class="feature-icon">
                <t-icon :name="item.icon" />
              </span>
              <div class="feature-text">
                <strong>{{ item.title }}</strong>
                <span>{{ item.desc }}</span>
              </div>
            </li>
          </ul>
        </div>

        <div class="float-chip chip-a">
          <span class="chip-icon"><t-icon name="secured" /></span>
          <div>
            <strong>安全合规</strong>
            <span>国密 · 后量子加密</span>
          </div>
        </div>
        <div class="float-chip chip-b">
          <span class="chip-icon"><t-icon name="internet" /></span>
          <div>
            <strong>统一入口</strong>
            <span>SSL · 证书 · 代理</span>
          </div>
        </div>
        <div class="float-chip chip-c">
          <span class="chip-icon"><t-icon name="chart" /></span>
          <div>
            <strong>7 × 24</strong>
            <span>运行监控与审计</span>
          </div>
        </div>

        <footer class="brand-footer">
          <span>© 2026 多功能管理平台</span>
          <span class="dot-sep"></span>
          <span>安全 · 稳定 · 可观测</span>
        </footer>
      </section>

      <!-- 右侧登录区 -->
      <section class="form-panel">
        <div class="form-inner">
          <div class="mobile-brand">
            <BrandMark :size="40" />
            <div class="mobile-brand-text">
              <strong>多功能管理平台</strong>
              <span>ENTERPRISE CONSOLE</span>
            </div>
          </div>

          <div class="form-head">
            <div class="form-head-badge">控制台登录</div>
            <h2>欢迎回来</h2>
            <p>请使用您的账号登录，继续管理企业资源</p>
          </div>

          <t-form
            ref="formRef"
            :data="form"
            :rules="rules"
            label-align="top"
            class="login-form"
            @submit="handleLogin"
          >
            <t-form-item label="账号" name="username">
              <t-input
                v-model="form.username"
                size="large"
                placeholder="请输入用户名"
                autocomplete="username"
                clearable
              >
                <template #prefix-icon>
                  <t-icon name="user" />
                </template>
              </t-input>
            </t-form-item>

            <t-form-item label="密码" name="password">
              <t-input
                v-model="form.password"
                size="large"
                :type="showPassword ? 'text' : 'password'"
                placeholder="请输入密码"
                autocomplete="current-password"
                @keyup="handlePasswordKeyup"
              >
                <template #prefix-icon>
                  <t-icon name="lock-on" />
                </template>
                <template #suffix-icon>
                  <span class="password-toggle" @click="showPassword = !showPassword">
                    <t-icon :name="showPassword ? 'browse' : 'browse-off'" />
                  </span>
                </template>
              </t-input>
            </t-form-item>

            <div class="form-options">
              <t-checkbox v-model="remember">记住账号</t-checkbox>
              <button type="button" class="link-btn" @click="handleForgot">忘记密码？</button>
            </div>

            <t-button
              theme="primary"
              type="submit"
              size="large"
              block
              :loading="loading"
              class="login-btn"
            >
              {{ loading ? '正在登录…' : '登 录' }}
            </t-button>
          </t-form>

          <div class="form-tip">
            <t-icon name="info-circle" />
            <span>建议使用 Chrome / Edge 最新版本访问，以获得最佳体验</span>
          </div>

          <div class="form-footer">
            <span>版本 v1.0.0</span>
            <span class="dot-sep"></span>
            <span>安全合规 · 全程审计</span>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  MessagePlugin,
  type FormInstanceFunctions,
  type SubmitContext,
} from 'tdesign-vue-next'
import { useUserStore } from '@/store/user'
import BrandMark from '@/components/BrandMark.vue'

const REMEMBER_KEY = 'login:remembered-username'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const formRef = ref<FormInstanceFunctions>()
const loading = ref(false)
const showPassword = ref(false)
const remember = ref(false)

const form = reactive({
  username: 'admin',
  password: '',
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' as const }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' as const }],
}

const features = [
  {
    icon: 'secured',
    title: '统一安全入口',
    desc: 'SSL 代理、证书与密钥集中管理，降低运维复杂度',
  },
  {
    icon: 'key',
    title: '国密与后量子加密',
    desc: '覆盖国密算法与后量子密码能力，满足合规要求',
  },
  {
    icon: 'chart',
    title: '全链路可观测',
    desc: '运行状态、操作审计与风险事件一站式追踪',
  },
]

onMounted(() => {
  const remembered = localStorage.getItem(REMEMBER_KEY)
  if (remembered) {
    form.username = remembered
    remember.value = true
  }
})

function triggerSubmit() {
  if (loading.value) return
  formRef.value?.submit()
}

function handlePasswordKeyup(_value: string, context: { e: KeyboardEvent }) {
  if (context?.e?.key === 'Enter') triggerSubmit()
}

async function handleLogin({ validateResult }: SubmitContext) {
  if (validateResult !== true || loading.value) return

  loading.value = true
  try {
    const username = form.username.trim()
    await userStore.login(username, form.password)

    if (remember.value) {
      localStorage.setItem(REMEMBER_KEY, username)
    } else {
      localStorage.removeItem(REMEMBER_KEY)
    }

    MessagePlugin.success('登录成功，正在进入控制台')
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/'
    await router.push(redirect)
  } catch (e: any) {
    MessagePlugin.error(e?.message || '登录失败，请检查账号和密码')
  } finally {
    loading.value = false
  }
}

function handleForgot() {
  MessagePlugin.info('请联系系统管理员重置密码')
}
</script>

<style scoped>
.login-page {
  position: relative;
  min-height: 100vh;
  background: #0b1220;
  overflow: hidden;
}

.login-shell {
  display: flex;
  width: 100%;
  min-height: 100vh;
}

/* ============================ 左侧品牌区 ============================ */
.brand-panel {
  position: relative;
  flex: 0 0 56%;
  display: flex;
  flex-direction: column;
  padding: 44px 56px 36px;
  overflow: hidden;
  background:
    radial-gradient(900px 600px at 8% 0%, rgba(37, 99, 235, 0.34), transparent 62%),
    radial-gradient(760px 520px at 100% 100%, rgba(6, 182, 212, 0.2), transparent 60%),
    linear-gradient(145deg, #0b1220 0%, #0e1b35 52%, #0a1428 100%);
  color: #fff;
}

.brand-panel::before {
  content: '';
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(148, 197, 253, 0.055) 1px, transparent 1px),
    linear-gradient(90deg, rgba(148, 197, 253, 0.055) 1px, transparent 1px);
  background-size: 44px 44px;
  mask-image: radial-gradient(circle at 30% 40%, #000 0%, transparent 78%);
  pointer-events: none;
}

.brand-panel::after {
  content: '';
  position: absolute;
  width: 560px;
  height: 560px;
  right: -180px;
  top: -160px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(37, 99, 235, 0.28), transparent 68%);
  filter: blur(10px);
  pointer-events: none;
}

.brand-orbit {
  position: absolute;
  right: -120px;
  bottom: -90px;
  width: 640px;
  height: 640px;
  opacity: 0.62;
  pointer-events: none;
  animation: orbit-breathe 9s ease-in-out infinite;
}

.brand-orbit svg {
  width: 100%;
  height: 100%;
}

.orbit-dashed {
  transform-origin: 320px 320px;
  animation: orbit-spin 42s linear infinite;
}

.brand-header {
  position: relative;
  z-index: 3;
  display: flex;
  align-items: center;
  gap: 14px;
}

.brand-name {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.brand-cn {
  font-size: 17px;
  font-weight: 600;
  letter-spacing: 0.04em;
}

.brand-en {
  font-size: 9px;
  letter-spacing: 0.2em;
  color: rgba(191, 219, 254, 0.5);
}

.brand-copy {
  position: relative;
  z-index: 3;
  max-width: 560px;
  margin-top: auto;
  margin-bottom: auto;
  padding: 56px 0 40px;
}

.eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 30px;
  padding: 0 14px;
  margin-bottom: 26px;
  border: 1px solid rgba(147, 197, 253, 0.24);
  border-radius: 999px;
  background: rgba(37, 99, 235, 0.12);
  color: #bfdbfe;
  font-size: 12px;
  letter-spacing: 0.08em;
  backdrop-filter: blur(8px);
}

.eyebrow-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #22d3ee;
  box-shadow: 0 0 0 4px rgba(34, 211, 238, 0.16);
}

.brand-copy h1 {
  margin: 0;
  font-size: clamp(36px, 3.4vw, 52px);
  line-height: 1.22;
  font-weight: 700;
  letter-spacing: -0.01em;
}

.gradient-text {
  background: linear-gradient(92deg, #93c5fd 0%, #67e8f9 48%, #a7f3d0 100%);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.lead {
  max-width: 500px;
  margin: 22px 0 0;
  font-size: 14px;
  line-height: 1.8;
  color: rgba(226, 232, 240, 0.66);
}

.feature-list {
  display: grid;
  gap: 18px;
  max-width: 520px;
  margin: 36px 0 0;
  padding: 0;
  list-style: none;
}

.feature-list li {
  display: flex;
  align-items: flex-start;
  gap: 14px;
}

.feature-icon {
  display: grid;
  place-items: center;
  flex-shrink: 0;
  width: 38px;
  height: 38px;
  border: 1px solid rgba(96, 165, 250, 0.24);
  border-radius: 11px;
  background: rgba(96, 165, 250, 0.12);
  color: #93c5fd;
  font-size: 18px;
}

.feature-text strong {
  display: block;
  margin-bottom: 3px;
  font-size: 14px;
  font-weight: 600;
  color: #f1f5f9;
}

.feature-text span {
  font-size: 12.5px;
  line-height: 1.55;
  color: rgba(203, 213, 225, 0.58);
}

/* 悬浮信息卡 */
.float-chip {
  position: absolute;
  z-index: 4;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 11px 15px;
  border: 1px solid rgba(147, 197, 253, 0.18);
  border-radius: 14px;
  background: rgba(12, 24, 45, 0.62);
  box-shadow: 0 18px 44px rgba(2, 6, 23, 0.38);
  backdrop-filter: blur(14px);
  animation: chip-float 7s ease-in-out infinite;
}

.float-chip strong {
  display: block;
  margin-bottom: 2px;
  font-size: 12.5px;
  font-weight: 600;
  color: #f8fafc;
}

.float-chip span {
  font-size: 11px;
  color: rgba(191, 219, 254, 0.56);
}

.chip-icon {
  display: grid;
  place-items: center;
  width: 30px;
  height: 30px;
  border-radius: 9px;
  background: rgba(37, 99, 235, 0.2);
  color: #7dd3fc;
  font-size: 15px;
}

.chip-a {
  top: 26%;
  right: 7%;
}

.chip-b {
  right: 27%;
  bottom: 19%;
  animation-delay: -2.4s;
}

.chip-c {
  right: 5%;
  bottom: 34%;
  animation-delay: -4.2s;
}

.brand-footer {
  position: relative;
  z-index: 3;
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: auto;
  font-size: 12px;
  color: rgba(148, 163, 184, 0.5);
}

.dot-sep {
  width: 3px;
  height: 3px;
  border-radius: 50%;
  background: currentColor;
  opacity: 0.6;
}

/* ============================ 右侧登录区 ============================ */
.form-panel {
  position: relative;
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 48px 56px;
  background: #fff;
}

.form-panel::before {
  content: '';
  position: absolute;
  inset: 0;
  background:
    radial-gradient(620px 420px at 100% 0%, rgba(37, 99, 235, 0.06), transparent 68%),
    radial-gradient(520px 420px at 0% 100%, rgba(6, 182, 212, 0.05), transparent 68%);
  pointer-events: none;
}

.form-inner {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 400px;
}

.mobile-brand {
  display: none;
  align-items: center;
  gap: 12px;
  margin-bottom: 34px;
}

.mobile-brand-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.mobile-brand-text strong {
  font-size: 16px;
  color: #0f172a;
}

.mobile-brand-text span {
  font-size: 9px;
  letter-spacing: 0.18em;
  color: #94a3b8;
}

.form-head {
  margin-bottom: 30px;
}

.form-head-badge {
  display: inline-flex;
  align-items: center;
  height: 24px;
  padding: 0 10px;
  margin-bottom: 14px;
  border-radius: 6px;
  background: #eff6ff;
  color: #2563eb;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.08em;
}

.form-head h2 {
  margin: 0;
  font-size: 30px;
  line-height: 1.2;
  font-weight: 700;
  color: #0f172a;
  letter-spacing: -0.01em;
}

.form-head p {
  margin: 10px 0 0;
  font-size: 13.5px;
  color: #64748b;
}

.login-form :deep(.t-form__label) {
  padding-bottom: 7px;
  font-size: 13px;
  font-weight: 600;
  color: #334155;
}

.login-form :deep(.t-form__item) {
  margin-bottom: 20px;
}

.login-form :deep(.t-input) {
  height: 46px;
  border-radius: 10px;
  background: #f8fafc;
  border-color: #e2e8f0;
  transition: border-color 0.18s ease, background 0.18s ease, box-shadow 0.18s ease;
}

.login-form :deep(.t-input:hover) {
  background: #fff;
  border-color: #bfdbfe;
}

.login-form :deep(.t-input.t-is-focused) {
  background: #fff;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
}

.login-form :deep(.t-input__prefix .t-icon) {
  color: #94a3b8;
  font-size: 17px;
}

.password-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border-radius: 6px;
  color: #94a3b8;
  cursor: pointer;
  transition: color 0.15s ease, background 0.15s ease;
}

.password-toggle:hover {
  color: #2563eb;
  background: #eff6ff;
}

.form-options {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: -4px 0 22px;
}

.link-btn {
  padding: 0;
  border: none;
  background: none;
  color: #2563eb;
  font-size: 13px;
  cursor: pointer;
  transition: color 0.15s ease;
}

.link-btn:hover {
  color: #1d4ed8;
  text-decoration: underline;
}

.login-btn.t-button {
  height: 46px;
  border: none;
  border-radius: 10px;
  background: linear-gradient(92deg, #2563eb 0%, #0891b2 100%);
  box-shadow: 0 12px 26px rgba(37, 99, 235, 0.28);
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 0.12em;
  transition: transform 0.16s ease, box-shadow 0.16s ease, filter 0.16s ease;
}

.login-btn.t-button:hover {
  background: linear-gradient(92deg, #1d4ed8 0%, #0e7490 100%);
  box-shadow: 0 16px 32px rgba(37, 99, 235, 0.34);
  transform: translateY(-1px);
}

.login-btn.t-button:active {
  transform: translateY(0);
  filter: brightness(0.98);
}

.form-tip {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-top: 22px;
  padding: 11px 13px;
  border: 1px solid #eef2f7;
  border-radius: 10px;
  background: #f8fafc;
  color: #64748b;
  font-size: 12px;
  line-height: 1.55;
}

.form-tip .t-icon {
  flex-shrink: 0;
  margin-top: 1px;
  color: #94a3b8;
}

.form-footer {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  margin-top: 30px;
  font-size: 12px;
  color: #94a3b8;
}

/* ============================ 动画 ============================ */
@keyframes orbit-spin {
  to {
    transform: rotate(360deg);
  }
}

@keyframes orbit-breathe {
  0%,
  100% {
    opacity: 0.55;
  }
  50% {
    opacity: 0.72;
  }
}

@keyframes chip-float {
  0%,
  100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-8px);
  }
}

/* ============================ 响应式 ============================ */
@media (max-width: 1280px) {
  .brand-panel {
    padding: 40px 44px 32px;
  }

  .chip-b {
    display: none;
  }
}

@media (max-width: 1100px) {
  .brand-panel {
    flex-basis: 50%;
    padding: 36px 36px 28px;
  }

  .brand-orbit {
    right: -220px;
    opacity: 0.42;
  }

  .float-chip {
    display: none;
  }

  .brand-copy h1 {
    font-size: 34px;
  }

  .form-panel {
    padding: 40px 36px;
  }
}

@media (max-width: 900px) {
  .login-shell {
    display: block;
  }

  .brand-panel {
    display: none;
  }

  .login-page {
    background:
      radial-gradient(720px 520px at 10% 0%, rgba(37, 99, 235, 0.16), transparent 64%),
      radial-gradient(620px 480px at 100% 100%, rgba(6, 182, 212, 0.12), transparent 62%),
      linear-gradient(160deg, #f0f6ff 0%, #ffffff 46%, #f7fbff 100%);
  }

  .form-panel {
    min-height: 100vh;
    padding: 40px 24px;
    background: transparent;
  }

  .form-panel::before {
    display: none;
  }

  .mobile-brand {
    display: flex;
  }

  .form-inner {
    max-width: 420px;
    padding: 32px 28px 26px;
    border: 1px solid rgba(226, 232, 240, 0.9);
    border-radius: 18px;
    background: rgba(255, 255, 255, 0.92);
    box-shadow: 0 24px 60px rgba(15, 23, 42, 0.1);
    backdrop-filter: blur(12px);
  }

  .form-head h2 {
    font-size: 26px;
  }
}

@media (max-width: 480px) {
  .form-panel {
    padding: 20px 14px;
  }

  .form-inner {
    padding: 26px 20px 22px;
    border-radius: 16px;
  }

  .form-head {
    margin-bottom: 24px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .brand-orbit,
  .orbit-dashed,
  .float-chip {
    animation: none;
  }
}
</style>
