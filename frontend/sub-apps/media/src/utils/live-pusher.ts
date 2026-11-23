import { reactive } from 'vue';
import { mediaApi } from '@/api/media';

export interface ActivePush {
  streamId: string;
  name: string;
  sourceType: 'camera' | 'file';
  pc: RTCPeerConnection;
  stream: MediaStream;
  hiddenVideo?: HTMLVideoElement;
  status: 'connecting' | 'connected' | 'failed';
  /** 是否仍在推流（后台推流为 true；停止后 false）。 */
  pushing: boolean;
}

class LivePusherManager {
  public activePushes = reactive(new Map<string, ActivePush>());
  /** 每个流的连接超时计时器（key: streamId）。 */
  private connTimers = new Map<string, ReturnType<typeof setTimeout>>();

  async startPush(
    streamId: string,
    name: string,
    sourceType: 'camera' | 'file',
    mediaStream: MediaStream,
    hiddenVideo?: HTMLVideoElement,
  ): Promise<ActivePush> {
    this.stopPush(streamId, false);

    if (hiddenVideo) {
      hiddenVideo.style.position = 'fixed';
      hiddenVideo.style.top = '-9999px';
      hiddenVideo.style.left = '-9999px';
      hiddenVideo.style.width = '1px';
      hiddenVideo.style.height = '1px';
      hiddenVideo.style.opacity = '0';
      hiddenVideo.style.pointerEvents = 'none';
      if (!hiddenVideo.parentNode) {
        document.body.appendChild(hiddenVideo);
      }
    }

    const pc = new RTCPeerConnection({ iceServers: [] });
    mediaStream.getTracks().forEach(track => pc.addTrack(track, mediaStream));

    const offer = await pc.createOffer();
    await pc.setLocalDescription(offer);

    // 注：保留 SDP 原样（含 opus）。mpegts.js 不能解 opus FLV，
    // 但 opus 是 WebRTC 推流的唯一可靠音频（captureStream/getUserMedia 均如此）。
    // 去掉 opus 会导致 captureStream(文件推流) 的 m=audio 无 codec -> 推流失败。
    // 拉流侧应使用 WebRTC 播放（浏览器原生支持 opus）或转码。
    let modifiedSdp = offer.sdp!;

    const resp = await mediaApi.pushWebRTC(streamId, modifiedSdp);
    await pc.setRemoteDescription(new RTCSessionDescription({
      type: 'answer',
      sdp: resp.data.sdp,
    }));

    const pushItem: ActivePush = {
      streamId,
      name,
      sourceType,
      pc,
      stream: mediaStream,
      hiddenVideo,
      status: 'connecting',
      pushing: true,
    };
    this.activePushes.set(streamId, pushItem);

    // ★ 连接超时：10s 内 ICE 未 connected 则视为推流失败，清理干净。
    //   避免"本地画面还在播放、ZLM 实际没流"的状态不一致。
    const timer = setTimeout(() => {
      const cur = this.activePushes.get(streamId);
      if (cur && cur.status !== 'connected') {
        console.warn(`[LivePusher] ${streamId} 连接超时（10s 未 connected），停止推流`);
        this.stopPush(streamId, true);
      }
    }, 10000);
    this.connTimers.set(streamId, timer);

    pc.onconnectionstatechange = () => {
      console.log(`[LivePusher] ${streamId} state =`, pc.connectionState);
      if (pc.connectionState === 'connected') {
        pushItem.status = 'connected';
        this.clearConnTimer(streamId);
      } else if (pc.connectionState === 'failed' || pc.connectionState === 'closed') {
        // ★ 彻底失败才清理；disconnected 短暂断开让它自动重连，不主动杀
        console.error(`[LivePusher] ${streamId} ${pc.connectionState}, stopping.`);
        pushItem.status = 'failed';
        this.stopPush(streamId, true);
      } else if (pc.connectionState === 'disconnected') {
        console.warn(`[LivePusher] ${streamId} temporarily disconnected, waiting for reconnect...`);
        pushItem.status = 'connecting';
        // 重新起连接超时（若已清）
        if (!this.connTimers.has(streamId)) {
          const t2 = setTimeout(() => {
            const cur = this.activePushes.get(streamId);
            if (cur && cur.status !== 'connected') {
              console.warn(`[LivePusher] ${streamId} 重连超时，停止推流`);
              this.stopPush(streamId, true);
            }
          }, 15000);
          this.connTimers.set(streamId, t2);
        }
      }
    };

    return pushItem;
  }

  /** 后台推流：关闭弹窗但保持推送。返回当前 push（可能已清理则 undefined）。 */
  backgroundKeep(streamId: string): ActivePush | undefined {
    const push = this.activePushes.get(streamId);
    if (push) push.pushing = true;
    return push;
  }

  /** 停止推流（释放摄像头/音频/ICE，并通知 ZLM）。 */
  stopPush(streamId: string, notify = true) {
    const push = this.activePushes.get(streamId);
    if (!push) {
      if (notify) mediaApi.stopPush(streamId).catch(() => {});
      return;
    }
    this.clearConnTimer(streamId);

    try { push.pc.close(); } catch (_) { /* noop */ }
    push.stream.getTracks().forEach(t => { try { t.stop(); } catch (_) { /* noop */ } });
    push.pushing = false;
    push.status = 'failed';

    if (push.hiddenVideo) {
      try {
        push.hiddenVideo.pause();
        push.hiddenVideo.removeAttribute('src');
        push.hiddenVideo.load();
        if (push.hiddenVideo.parentNode) {
          push.hiddenVideo.parentNode.removeChild(push.hiddenVideo);
        }
      } catch (_) { /* noop */ }
    }

    this.activePushes.delete(streamId);
    if (notify) {
      mediaApi.stopPush(streamId).catch(() => {});
    }
  }

  private clearConnTimer(streamId: string) {
    const t = this.connTimers.get(streamId);
    if (t) {
      clearTimeout(t);
      this.connTimers.delete(streamId);
    }
  }

  stopAll() {
    Array.from(this.activePushes.keys()).forEach(id => this.stopPush(id, true));
  }
}

export const livePusher = new LivePusherManager();
