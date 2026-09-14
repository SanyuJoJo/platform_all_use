import mitt from 'mitt';
import type { Module } from '@/types/module';

/**
 * 平台事件总线。
 *
 * 通过为 mitt 声明 Events 泛型，保证 on/emit 的参数类型安全。
 * 新增事件时在此处补充。
 */
export type EventBusEvents = {
  'modules:changed': Module[];
  'modules:refreshed': Module[];
  'platform:auth-expired': void;
  'platform:logout': void;
};

export const eventBus = mitt<EventBusEvents>();

// 全局调试日志（可选，仅在开发模式输出）
if (import.meta.env.DEV) {
  eventBus.on('*', (type, payload) => {
    // type 在 mitt 中为 keyof Events | '*'
    console.log(`[DEBUG] eventBus: 事件触发 ${String(type)}`, payload);
  });
}
