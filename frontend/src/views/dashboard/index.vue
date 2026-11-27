<template>
  <div class="dashboard">
    <!-- 欢迎横幅 -->
    <section class="hero">
      <div class="hero-content">
        <div class="hero-badge">
          <span class="hero-dot"></span>
          控制台已就绪
        </div>
        <h1>{{ greeting }}，{{ displayName }}</h1>
        <p class="hero-sub">{{ today }} · 欢迎回到多功能管理平台控制台</p>
        <div class="hero-actions">
          <t-button theme="primary" size="large" @click="goWorkspace">
            进入工作台
          </t-button>
          <t-button variant="outline" size="large" class="hero-ghost" @click="showHelp">
            查看使用说明
          </t-button>
        </div>
      </div>

      <div class="hero-visual" aria-hidden="true">
        <span class="hero-ring ring-1"></span>
        <span class="hero-ring ring-2"></span>
        <span class="hero-ring ring-3"></span>
        <span class="hero-shield">
          <t-icon name="secured" />
        </span>
        <span class="hero-node node-1"></span>
        <span class="hero-node node-2"></span>
        <span class="hero-node node-3"></span>
      </div>
    </section>

    <!-- 关键指标 -->
    <section class="stat-grid">
      <article v-for="stat in stats" :key="stat.label" class="stat-card">
        <span class="stat-icon" :class="`tone-${stat.tone}`">
          <t-icon :name="stat.icon" />
        </span>
        <div class="stat-content">
          <span class="stat-label">{{ stat.label }}</span>
          <strong class="stat-value">{{ stat.value }}</strong>
          <span class="stat-sub">{{ stat.sub }}</span>
        </div>
      </article>
    </section>

    <div class="main-grid">
      <!-- 快捷入口 -->
      <section class="panel quick-panel">
        <header class="panel-head">
          <div>
            <h3>快捷入口</h3>
            <p>根据当前账号权限自动生成</p>
          </div>
          <t-tag v-if="quickLinks.length" theme="primary" variant="light">
            {{ quickLinks.length }} 个功能
          </t-tag>
        </header>

        <div class="quick-grid">
          <button
            v-for="item in quickLinks"
            :key="item.path"
            type="button"
            class="quick-item"
            @click="router.push(item.path)"
          >
            <span class="quick-icon">
              <t-icon :name="tdesignIcon(item.icon)" />
            </span>
            <span class="quick-text">
              <strong>{{ item.label }}</strong>
              <small>{{ item.group || '常用功能' }}</small>
            </span>
            <t-icon name="chevron-right" class="quick-arrow" />
          </button>

          <div v-if="!quickLinks.length" class="empty-state">
            <t-icon name="app" />
            <span>暂无可访问的功能菜单</span>
          </div>
        </div>
      </section>

      <!-- 平台能力 -->
      <section class="panel capability-panel">
        <header class="panel-head">
          <div>
            <h3>平台能力</h3>
            <p>统一安全运维底座</p>
          </div>
        </header>

        <ul class="capability-list">
          <li v-for="item in capabilities" :key="item.title">
            <span class="cap-icon">
              <t-icon :name="item.icon" />
            </span>
            <div>
              <strong>{{ item.title }}</strong>
              <span>{{ item.desc }}</span>
            </div>
          </li>
        </ul>
      </section>
    </div>

    <!-- 账号与系统 -->
    <section class="panel info-panel">
      <header class="panel-head">
        <div>
          <h3>账号与系统</h3>
          <p>当前登录会话信息</p>
        </div>
        <t-tag theme="success" variant="light">安全连接</t-tag>
      </header>

      <div class="info-grid">
        <div class="info-item">
          <span>当前账号</span>
          <strong>{{ userStore.user?.username || '-' }}</strong>
        </div>
        <div class="info-item">
          <span>所属角色</span>
          <strong>{{ roleLabel }}</strong>
        </div>
        <div class="info-item">
          <span>权限项</span>
          <strong>{{ permissionCount }} 项</strong>
        </div>
        <div class="info-item">
          <span>平台版本</span>
          <strong>v1.0.0</strong>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { MessagePlugin } from 'tdesign-vue-next'
import { useUserStore } from '@/store/user'
import { useMenuStore } from '@/store/menu'
import { tdesignIcon } from '@/utils/icon-map'
import { flattenMenuLeaves } from '@/utils/menu'
import { formatUserRoles } from '@/utils/user'

const router = useRouter()
const userStore = useUserStore()
const menuStore = useMenuStore()

const displayName = computed(
  () => userStore.user?.nickname || userStore.user?.username || '管理员',
)

const roleLabel = computed(() => formatUserRoles(userStore.user?.roles))

const permissionCount = computed(() => userStore.permissions?.length || 0)

const menuLeaves = computed(() => flattenMenuLeaves(menuStore.menuTree))

const quickLinks = computed(() => menuLeaves.value.slice(0, 8))

const firstMenuPath = computed(() => {
  const leaves = menuLeaves.value
  return (leaves.find((item) => item.label !== '仪表盘') || leaves[0])?.path
})

const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 6) return '夜深了'
  if (hour < 12) return '上午好'
  if (hour < 14) return '中午好'
  if (hour < 18) return '下午好'
  return '晚上好'
})

const today = computed(() =>
  new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    weekday: 'long',
  }).format(new Date()),
)

const stats = computed(() => [
  {
    label: '可访问功能',
    value: menuLeaves.value.length,
    sub: '已按权限授权',
    icon: 'app',
    tone: 'blue',
  },
  {
    label: '我的角色',
    value: userStore.user?.roles?.length || 0,
    sub: roleLabel.value,
    icon: 'usergroup',
    tone: 'violet',
  },
  {
    label: '权限项',
    value: permissionCount.value,
    sub: '细粒度权限控制',
    icon: 'secured',
    tone: 'cyan',
  },
  {
    label: '当前会话',
    value: '在线',
    sub: '已通过身份认证',
    icon: 'check-circle',
    tone: 'green',
  },
])

const capabilities = [
  {
    icon: 'secured',
    title: '统一安全入口',
    desc: 'SSL 代理、证书与密钥集中管理',
  },
  {
    icon: 'certificate',
    title: '证书全生命周期',
    desc: '签发、续期、吊销与到期提醒',
  },
  {
    icon: 'key',
    title: '国密与后量子',
    desc: '覆盖国密算法与后量子密码能力',
  },
  {
    icon: 'file',
    title: '全链路审计',
    desc: '关键操作留痕，风险事件可追溯',
  },
]

function goWorkspace() {
  if (firstMenuPath.value) {
    router.push(firstMenuPath.value)
    return
  }
  MessagePlugin.info('暂无可访问的功能菜单，请联系管理员分配权限')
}

function showHelp() {
  MessagePlugin.info('如需帮助，请联系系统管理员获取使用手册')
}
</script>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 18px;
  color: #1e293b;
}

/* ============================ 欢迎横幅 ============================ */
.hero {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 210px;
  padding: 34px 40px;
  overflow: hidden;
  border-radius: 18px;
  background:
    radial-gradient(720px 380px at 88% 8%, rgba(34, 211, 238, 0.24), transparent 62%),
    radial-gradient(520px 340px at 4% 100%, rgba(37, 99, 235, 0.32), transparent 66%),
    linear-gradient(120deg, #0b1220 0%, #12224a 52%, #0d2a46 100%);
  box-shadow: 0 18px 42px rgba(15, 23, 42, 0.16);
  color: #fff;
}

.hero::before {
  content: '';
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(148, 197, 253, 0.06) 1px, transparent 1px),
    linear-gradient(90deg, rgba(148, 197, 253, 0.06) 1px, transparent 1px);
  background-size: 38px 38px;
  mask-image: radial-gradient(circle at 20% 50%, #000 0%, transparent 80%);
  pointer-events: none;
}

.hero-content {
  position: relative;
  z-index: 2;
  max-width: 620px;
}

.hero-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 28px;
  padding: 0 12px;
  margin-bottom: 16px;
  border: 1px solid rgba(52, 211, 153, 0.24);
  border-radius: 999px;
  background: rgba(16, 185, 129, 0.1);
  color: #a7f3d0;
  font-size: 12px;
}

.hero-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #34d399;
  box-shadow: 0 0 0 4px rgba(52, 211, 153, 0.16);
}

.hero h1 {
  margin: 0;
  font-size: clamp(24px, 2.2vw, 32px);
  line-height: 1.3;
  font-weight: 700;
  letter-spacing: -0.01em;
}

.hero-sub {
  margin: 10px 0 0;
  color: rgba(226, 232, 240, 0.64);
  font-size: 13.5px;
}

.hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 24px;
}

.hero-ghost.t-button {
  border-color: rgba(191, 219, 254, 0.3);
  background: rgba(255, 255, 255, 0.06);
  color: #e2e8f0;
}

.hero-ghost.t-button:hover {
  border-color: rgba(191, 219, 254, 0.5);
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
}

.hero-visual {
  position: relative;
  flex-shrink: 0;
  width: 260px;
  height: 190px;
  margin-right: 12px;
}

.hero-ring {
  position: absolute;
  top: 50%;
  left: 50%;
  border: 1px solid rgba(147, 197, 253, 0.18);
  border-radius: 50%;
  transform: translate(-50%, -50%);
}

.ring-1 {
  width: 168px;
  height: 168px;
}

.ring-2 {
  width: 122px;
  height: 122px;
  border-style: dashed;
  border-color: rgba(103, 232, 249, 0.28);
  animation: ring-spin 28s linear infinite;
}

.ring-3 {
  width: 78px;
  height: 78px;
  background: radial-gradient(circle, rgba(37, 99, 235, 0.35), transparent 70%);
}

.hero-shield {
  position: absolute;
  top: 50%;
  left: 50%;
  display: grid;
  place-items: center;
  width: 72px;
  height: 72px;
  border: 1px solid rgba(125, 211, 252, 0.35);
  border-radius: 22px;
  background: linear-gradient(135deg, rgba(37, 99, 235, 0.5), rgba(6, 182, 212, 0.28));
  box-shadow: 0 16px 36px rgba(2, 6, 23, 0.35);
  color: #e0f2fe;
  font-size: 34px;
  transform: translate(-50%, -50%) rotate(-4deg);
  backdrop-filter: blur(6px);
}

.hero-node {
  position: absolute;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #67e8f9;
  box-shadow: 0 0 0 5px rgba(103, 232, 249, 0.12);
}

.node-1 {
  top: 18px;
  right: 34px;
}

.node-2 {
  bottom: 24px;
  left: 18px;
}

.node-3 {
  right: 8px;
  bottom: 46px;
  width: 5px;
  height: 5px;
}

@keyframes ring-spin {
  to {
    transform: translate(-50%, -50%) rotate(360deg);
  }
}

/* ============================ 指标卡 ============================ */
.stat-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
}

.stat-card {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  padding: 20px;
  border: 1px solid #eaeef5;
  border-radius: 14px;
  background: #fff;
  box-shadow: 0 4px 16px rgba(15, 23, 42, 0.03);
  transition: transform 0.18s ease, box-shadow 0.18s ease, border-color 0.18s ease;
}

.stat-card:hover {
  border-color: #dbeafe;
  box-shadow: 0 12px 28px rgba(15, 23, 42, 0.07);
  transform: translateY(-2px);
}

.stat-icon {
  display: grid;
  place-items: center;
  flex-shrink: 0;
  width: 44px;
  height: 44px;
  border-radius: 12px;
  font-size: 21px;
}

.tone-blue {
  background: #eff6ff;
  color: #2563eb;
}

.tone-violet {
  background: #f5f3ff;
  color: #7c3aed;
}

.tone-cyan {
  background: #ecfeff;
  color: #0891b2;
}

.tone-green {
  background: #ecfdf5;
  color: #059669;
}

.stat-content {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.stat-label {
  color: #64748b;
  font-size: 12.5px;
}

.stat-value {
  margin: 3px 0 2px;
  color: #0f172a;
  font-size: 26px;
  font-weight: 700;
  line-height: 1.15;
}

.stat-sub {
  overflow: hidden;
  color: #94a3b8;
  font-size: 11.5px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* ============================ 主内容区 ============================ */
.main-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.65fr) minmax(300px, 1fr);
  gap: 16px;
}

.panel {
  padding: 22px;
  border: 1px solid #eaeef5;
  border-radius: 14px;
  background: #fff;
  box-shadow: 0 4px 16px rgba(15, 23, 42, 0.03);
}

.panel-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 18px;
}

.panel-head h3 {
  margin: 0;
  color: #0f172a;
  font-size: 15.5px;
  font-weight: 600;
}

.panel-head p {
  margin: 4px 0 0;
  color: #94a3b8;
  font-size: 12px;
}

.quick-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.quick-item {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
  padding: 13px 14px;
  border: 1px solid #eef2f7;
  border-radius: 11px;
  background: #fbfcfe;
  text-align: left;
  cursor: pointer;
  transition: border-color 0.16s ease, background 0.16s ease, transform 0.16s ease;
}

.quick-item:hover {
  border-color: #bfdbfe;
  background: #f5f9ff;
  transform: translateY(-1px);
}

.quick-icon {
  display: grid;
  place-items: center;
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: linear-gradient(135deg, #eff6ff, #ecfeff);
  color: #2563eb;
  font-size: 18px;
}

.quick-text {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
}

.quick-text strong {
  overflow: hidden;
  color: #1e293b;
  font-size: 13px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.quick-text small {
  overflow: hidden;
  margin-top: 2px;
  color: #94a3b8;
  font-size: 11px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.quick-arrow {
  flex-shrink: 0;
  color: #cbd5e1;
  font-size: 14px;
}

.empty-state {
  display: flex;
  grid-column: 1 / -1;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 36px 0;
  color: #94a3b8;
  font-size: 12.5px;
}

.empty-state .t-icon {
  font-size: 26px;
}

.capability-list {
  display: grid;
  gap: 14px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.capability-list li {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.cap-icon {
  display: grid;
  place-items: center;
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: #f1f6ff;
  color: #2563eb;
  font-size: 18px;
}

.capability-list strong {
  display: block;
  margin-bottom: 2px;
  color: #1e293b;
  font-size: 13px;
  font-weight: 600;
}

.capability-list span {
  color: #94a3b8;
  font-size: 11.5px;
  line-height: 1.5;
}

/* ============================ 账号与系统 ============================ */
.info-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 5px;
  padding: 15px 16px;
  border-radius: 11px;
  background: #f8fafc;
}

.info-item span {
  color: #94a3b8;
  font-size: 11.5px;
}

.info-item strong {
  overflow: hidden;
  color: #1e293b;
  font-size: 13.5px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* ============================ 响应式 ============================ */
@media (max-width: 1280px) {
  .stat-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .info-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 1100px) {
  .main-grid {
    grid-template-columns: 1fr;
  }

  .hero-visual {
    width: 200px;
    transform: scale(0.88);
  }
}

@media (max-width: 820px) {
  .hero {
    padding: 28px 24px;
  }

  .hero-visual {
    display: none;
  }

  .stat-grid {
    grid-template-columns: 1fr;
  }

  .quick-grid {
    grid-template-columns: 1fr;
  }

  .info-grid {
    grid-template-columns: 1fr;
  }
}

@media (prefers-reduced-motion: reduce) {
  .ring-2 {
    animation: none;
  }

  .stat-card,
  .quick-item {
    transition: none;
  }
}
</style>
