// src/micro-frontend/event-bus.ts
import mitt from 'mitt';

console.log('[DEBUG] event-bus.ts: 创建事件总线');

export const eventBus = mitt();

// 添加事件监听日志（可选）
eventBus.on('*', (type, payload) => {
  console.log(`[DEBUG] eventBus: 事件触发 ${type}`, payload);
});