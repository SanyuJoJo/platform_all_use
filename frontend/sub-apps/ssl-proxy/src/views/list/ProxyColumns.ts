import { h, type VNode } from 'vue';
import {
  NButton,
  NPopconfirm,
  NSpace,
  NTag,
} from 'naive-ui';
import type { DataTableColumns } from 'naive-ui';
import { STATUS_TAG_TYPE } from '@/constants/proxyOptions';
import type { ProxyListItem, ProxyStatus } from '@/types';

export type ProxyActionKind = 'start' | 'stop' | 'restart';

export interface ColumnHandlers {
  onDetail: (row: ProxyListItem) => void;
  onStart: (row: ProxyListItem) => void | Promise<void>;
  onStop: (row: ProxyListItem) => void | Promise<void>;
  onDelete: (row: ProxyListItem) => void | Promise<void>;
  isRunning: (status: ProxyStatus) => boolean;
  isTransitioning: (status: ProxyStatus) => boolean;
  isPending: (serviceId: string) => boolean;
  pendingAction: (serviceId: string) => ProxyActionKind | null;
  certName: (certId?: string) => string;
}

interface StartStopState {
  label: string;
  loading: boolean;
  disabled: boolean;
  type: 'primary' | 'warning' | 'default';
  isStopAction: boolean;
}

function resolveStartStopState(
  row: ProxyListItem,
  handlers: ColumnHandlers
): StartStopState {
  const act = handlers.pendingAction(row.service_id);

  if (act === 'stop') {
    return { label: '停止中…', loading: true, disabled: true, type: 'warning', isStopAction: true };
  }
  if (act === 'start') {
    return { label: '启动中…', loading: true, disabled: true, type: 'primary', isStopAction: false };
  }
  if (act === 'restart') {
    return { label: '重启中…', loading: true, disabled: true, type: 'default', isStopAction: false };
  }

  if (row.status === 'STOPPING') {
    return { label: '停止中…', loading: true, disabled: true, type: 'warning', isStopAction: true };
  }
  if (row.status === 'STARTING') {
    return { label: '启动中…', loading: true, disabled: true, type: 'primary', isStopAction: false };
  }

  if (row.status === 'RUNNING') {
    return { label: '停止', loading: false, disabled: false, type: 'warning', isStopAction: true };
  }
  return { label: '启动', loading: false, disabled: false, type: 'primary', isStopAction: false };
}

export function buildColumns(
  handlers: ColumnHandlers
): DataTableColumns<ProxyListItem> {
  const { onDetail, onStart, onStop, onDelete, isPending, certName } = handlers;

  function signCertId(row: ProxyListItem): string {
    if (row.mode === 'client') {
      return row.client_cert_id || row.cert_id || '';
    }
    return row.cert_id || row.client_cert_id || '';
  }

  function encCertId(row: ProxyListItem): string {
    if (row.mode === 'client') {
      return row.client_enc_cert_id || '';
    }
    return row.enc_cert_id || '';
  }

  return [
    {
      type: 'selection',
      width: 40,
      disabled(row) {
        return resolveStartStopState(row, handlers).disabled;
      },
    },
    {
      title: '服务名称',
      key: 'service_name',
      minWidth: 160,
      ellipsis: { tooltip: true },
    },
    {
      title: '本地端口',
      key: 'listen_port',
      width: 110,
      render(row) {
        const addr = row.listen_address || '';
        const port = addr.includes(':') ? addr.split(':').pop() : addr;
        return h('span', { class: 'mono' }, port || '—');
      },
    },
    {
      title: '后端服务地址及端口',
      key: 'backend_address',
      minWidth: 200,
      ellipsis: { tooltip: true },
      render(row) {
        const addr = (row.backend_address || '').replace(/^https?:\/\//, '');
        return h('span', { class: 'mono' }, addr || '—');
      },
    },
    {
      title: '签名证书',
      key: 'sign_cert',
      minWidth: 160,
      ellipsis: { tooltip: true },
      render(row) {
        const id = signCertId(row);
        return id ? certName(id) : '—';
      },
    },
    {
      title: '加密证书',
      key: 'enc_cert',
      minWidth: 160,
      ellipsis: { tooltip: true },
      render(row) {
        const id = encCertId(row);
        return id ? certName(id) : '—';
      },
    },
    {
      title: '服务状态',
      key: 'status',
      width: 100,
      render(row) {
        return h(
          NTag,
          { size: 'small', type: STATUS_TAG_TYPE[row.status] || 'default' },
          { default: () => row.status }
        );
      },
    },
    {
      title: '操作',
      key: 'actions',
      width: 200,
      fixed: 'right',
      render(row) {
        const st = resolveStartStopState(row, handlers);
        const pend = isPending(row.service_id);
        const children: VNode[] = [];

        children.push(
          h(
            NButton,
            {
              size: 'small',
              type: st.type,
              loading: st.loading,
              disabled: st.disabled,
              onClick: () => {
                if (st.disabled) return;
                if (st.isStopAction) onStop(row);
                else onStart(row);
              },
            },
            { default: () => st.label }
          )
        );

        children.push(
          h(
            NButton,
            { size: 'small', onClick: () => onDetail(row) },
            { default: () => '详情' }
          )
        );

        children.push(
          h(
            NPopconfirm,
            {
              positiveText: '确认删除',
              negativeText: '取消',
              onPositiveClick: () => onDelete(row),
            },
            {
              trigger: () =>
                h(
                  NButton,
                  { size: 'small', type: 'error', disabled: pend || st.disabled },
                  { default: () => '删除' }
                ),
              default: () =>
                `确认删除「${row.service_name}」？该操作不可恢复，且会写审计日志。`,
            }
          )
        );

        return h(NSpace, { size: 4 }, { default: () => children });
      },
    },
  ];
}
