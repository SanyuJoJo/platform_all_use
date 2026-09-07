import { useModuleStore } from '@/store/module';
export function hasPermission(code: string): boolean {
  const store = useModuleStore();
  return store.hasPermission(code);
}
