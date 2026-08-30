export function formatDate(date: string | Date): string {
  const d = new Date(date);
  return d.toLocaleString('zh-CN');
}
