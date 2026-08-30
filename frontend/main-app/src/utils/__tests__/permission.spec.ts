import { describe, expect, it, vi, afterEach } from 'vitest';
import { hasPermission } from '@/store/permission';
import { useUserStore } from '@/store/user';
vi.mock('@/store/user', () => ({
  useUserStore: vi.fn(),
}));
afterEach(() => {
  vi.clearAllMocks();
});
describe('hasPermission', () => {
  it('should return true if permission exists', () => {
    const mockStore = { permissions: { value: ['auth:user:view'] } };
    (useUserStore as any).mockReturnValue(mockStore);
    expect(hasPermission('auth:user:view')).toBe(true);
  });
  it('should return false if permission missing', () => {
    const mockStore = { permissions: { value: [] } };
    (useUserStore as any).mockReturnValue(mockStore);
    expect(hasPermission('auth:user:create')).toBe(false);
  });
});
