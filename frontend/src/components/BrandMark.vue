<template>
  <svg
    class="brand-mark"
    :width="size"
    :height="size"
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    role="img"
    aria-label="多功能管理平台"
  >
    <defs>
      <linearGradient :id="gradientId" x1="6" y1="2" x2="42" y2="46" gradientUnits="userSpaceOnUse">
        <stop stop-color="#3B82F6" />
        <stop offset="0.55" stop-color="#2563EB" />
        <stop offset="1" stop-color="#06B6D4" />
      </linearGradient>
      <linearGradient :id="shineId" x1="8" y1="8" x2="38" y2="38" gradientUnits="userSpaceOnUse">
        <stop stop-color="#FFFFFF" stop-opacity="0.42" />
        <stop offset="1" stop-color="#FFFFFF" stop-opacity="0" />
      </linearGradient>
    </defs>

    <rect width="48" height="48" rx="14" :fill="`url(#${gradientId})`" />
    <rect x="1" y="1" width="46" height="46" rx="13" stroke="#FFFFFF" stroke-opacity="0.18" />

    <!-- 安全盾牌 -->
    <path
      d="M24 10.5 36 15v9c0 7.1-4.9 12.1-12 14.5C16.9 36.1 12 31.1 12 24v-9l12-4.5Z"
      fill="#FFFFFF"
      fill-opacity="0.14"
      stroke="#FFFFFF"
      stroke-width="1.5"
      stroke-linejoin="round"
    />
    <path
      d="m18.6 23.8 3.7 3.7 7.3-7.3"
      stroke="#FFFFFF"
      stroke-width="2.4"
      stroke-linecap="round"
      stroke-linejoin="round"
    />

    <!-- 高光与数据节点 -->
    <path d="M9 16.5c3.2-4.2 8.1-6.9 15-7.5" :stroke="`url(#${shineId})`" stroke-width="2" stroke-linecap="round" />
    <circle cx="36" cy="12.5" r="2.4" fill="#67E8F9" />
    <circle cx="36" cy="12.5" r="4.2" fill="#67E8F9" fill-opacity="0.18" />
  </svg>
</template>

<script lang="ts">
let uid = 0
</script>

<script setup lang="ts">
withDefaults(defineProps<{ size?: number }>(), { size: 36 })

uid += 1
const gradientId = `brand-mark-gradient-${uid}`
const shineId = `brand-mark-shine-${uid}`
</script>

<style scoped>
.brand-mark {
  display: block;
  flex-shrink: 0;
  filter: drop-shadow(0 6px 16px rgba(37, 99, 235, 0.28));
}
</style>
