import { h, type VNode } from 'vue';
import { Button, Popconfirm, Space, Tag } from 'tdesign-vue-next';
import type { PrimaryTableCol } from 'tdesign-vue-next';
import { TCPS_STATUS_TAG_TYPE } from '@/constants/tcpsOptions';
import type { TcpsListItem, TcpsStatus } from '@/types';

export type TcpsActionKind = 'start' | 'stop' | 'restart';

export interface TcpsColumnHandlers {
  onDetail: (row: TcpsListItem) => void;
  onStart: (row: TcpsListItem) => void | Promise<void>;
  onStop: (row: TcpsListItem) => void | Promise<void>;
  onDelete: (row: TcpsListItem) => void | Promise<void>;
  isRunning: (status: TcpsStatus) => boolean;
  isTransitioning: (status: TcpsStatus) => boolean;
  isPending: (serviceId: string) => boolean;
  pendingAction: (serviceId: string) => TcpsActionKind | null;
  certName: (certId?: string) => string;
  hasPerm: (code: string) => boolean;
}

interface StartStopState {
  label: string;
  loading: boolean;
  disabled: boolean;
  theme: 'primary' | 'warning' | 'default';
  isStopAction: boolean;
}

function resolveStartStopState(
  row: TcpsListItem,
  handlers: TcpsColumnHandlers,
): StartStopState {
  const act = handlers.pendingAction(row.service_id);

  if (act === 'stop') {
    return { label: '停止中…', loading: true, disabled: true, theme: 'warning', isStopAction: true };
  }
  if (act === 'start') {
    return { label: '启动中…', loading: true, disabled: true, theme: 'primary', isStopAction: false };
  }
  if (act === 'restart') {
    return { label: '重启中…', loading: true, disabled: true, theme: 'default', isStopAction: false };
  }

  if (row.status === 'STOPPING') {
    return { label: '停止中…', loading: true, disabled: true, theme: 'warning', isStopAction: true };
  }
  if (row.status === 'STARTING') {
    return { label: '启动中…', loading: true, disabled: true, theme: 'primary', isStopAction: false };
  }

  if (row.status === 'RUNNING') {
    return { label: '停止', loading: false, disabled: false, theme: 'warning', isStopAction: true };
  }
  return { label: '启动', loading: false, disabled: false, theme: 'primary', isStopAction: false };
}

export function buildTcpsColumns(
  handlers: TcpsColumnHandlers,
): PrimaryTableCol<any>[] {
  const { onDetail, onStart, onStop, onDelete, isPending, certName, hasPerm } = handlers;

  function signCertId(row: TcpsListItem): string {
    if (row.mode === 'client') {
      return row.client_cert_id || row.cert_id || '';
    }
    return row.cert_id || row.client_cert_id || '';
  }

  function encCertId(row: TcpsListItem): string {
    if (row.mode === 'client') {
      return row.client_enc_cert_id || '';
    }
    return row.enc_cert_id || '';
  }

  return [
    {
      type: 'multiple',
      width: 40,
      disabled: (row) => resolveStartStopState(row as any, handlers).disabled,
    },
    {
      colKey: 'service_name',
      title: '服务名称',
      minWidth: 160,
      ellipsis: true,
    },
    {
      colKey: 'listen_port',
      title: '本地端口',
      width: 110,
      cell: (_h, { row }) => {
        const addr = row.listen_address || '';
        const port = addr.includes(':') ? addr.split(':').pop() : addr;
        return h('span', { class: 'mono' }, port || '—');
      },
    },
    {
      colKey: 'backend_address',
      title: '后端服务地址及端口',
      minWidth: 200,
      ellipsis: true,
      cell: (_h, { row }) => {
        const addr = (row.backend_address || '').replace(/^[a-zA-Z]+:\/\//, '');
        return h('span', { class: 'mono' }, addr || '—');
      },
    },
    {
      colKey: 'sign_cert',
      title: '签名证书',
      minWidth: 160,
      ellipsis: true,
      cell: (_h, { row }) => {
        const id = signCertId(row);
        return id ? certName(id) : '—';
      },
    },
    {
      colKey: 'enc_cert',
      title: '加密证书',
      minWidth: 160,
      ellipsis: true,
      cell: (_h, { row }) => {
        const id = encCertId(row);
        return id ? certName(id) : '—';
      },
    },
    {
      colKey: 'status',
      title: '服务状态',
      width: 100,
      cell: (_h, { row }) =>
        h(
          Tag,
          { theme: TCPS_STATUS_TAG_TYPE[row.status] || 'default', variant: 'light' },
          () => row.status,
        ),
    },
    {
      colKey: 'actions',
      title: '操作',
      width: 200,
      fixed: 'right',
      cell: (_h, { row }) => {
        const st = resolveStartStopState(row, handlers);
        const pend = isPending(row.service_id);
        const children: VNode[] = [];

        const canViewConfig = hasPerm('ssl_proxy:tcps:config');

        const canStartStop = st.isStopAction
          ? hasPerm('ssl_proxy:tcps:stop')
          : hasPerm('ssl_proxy:tcps:start');

        if (canStartStop) {
          children.push(
            h(
              Button,
              {
                size: 'small',
                theme: st.theme,
                loading: st.loading,
                disabled: st.disabled,
                onClick: () => {
                  if (st.disabled) return;
                  if (st.isStopAction) onStop(row);
                  else onStart(row);
                },
              },
              () => st.label,
            ),
          );
        }

        if (canViewConfig) {
          children.push(
            h(Button, { size: 'small', variant: 'text', theme: 'primary', onClick: () => onDetail(row) }, () => '详情'),
          );
        }

        if (hasPerm('ssl_proxy:tcps:delete')) {
          children.push(
            h(
              Popconfirm,
              {
                content: `确认删除「${row.service_name}」？该操作不可恢复，且会写审计日志。`,
                confirmBtn: '确认删除',
                cancelBtn: '取消',
                onConfirm: () => onDelete(row),
              },
              () =>
                h(
                  Button,
                  { size: 'small', variant: 'text', theme: 'danger', disabled: pend || st.disabled },
                  () => '删除',
                ),
            ),
          );
        }

        if (children.length === 0) {
          return h('span', { style: 'color: #999; font-size: 12px' }, '无操作权限');
        }

        return h(Space, { size: 4 }, () => children);
      },
    },
  ];
}