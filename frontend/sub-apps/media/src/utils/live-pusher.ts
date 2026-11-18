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
}

class LivePusherManager {
  public activePushes = reactive(new Map<string, ActivePush>());

  async startPush(
    streamId: string,
    name: string,
    sourceType: 'camera' | 'file',
    mediaStream: MediaStream,
    hiddenVideo?: HTMLVideoElement,
  ): Promise<ActivePush> {
    this.stopPush(streamId);

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

    // ★ 核心修复 1：修剪 SDP，只保留 OPUS 音频，剔除 PCMA，避免 FLV 播放失败
    let modifiedSdp = offer.sdp!;
    if (modifiedSdp.includes('PCMA')) {
      // 移除 PCMA (PT=8) 的支持
      modifiedSdp = modifiedSdp.replace(/a=rtpmap:8 PCMA\/8000\r\n/g, '');
      modifiedSdp = modifiedSdp.replace(/ 8 /g, ' '); // 从 m=audio 列表里去掉 8
    }

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
    };
    this.activePushes.set(streamId, pushItem);

    pc.onconnectionstatechange = () => {
      console.log(`[LivePusher] ${streamId} state =`, pc.connectionState);
      if (pc.connectionState === 'connected') {
        pushItem.status = 'connected';
      } else if (pc.connectionState === 'failed') {
        // ★ 核心修复 2：只有彻底 failed 才清理。短暂的 disconnected 让它自动重连，绝不主动断开
        console.error(`[LivePusher] ${streamId} failed, stopping.`);
        pushItem.status = 'failed';
        this.stopPush(streamId);
      } else if (pc.connectionState === 'disconnected') {
        console.warn(`[LivePusher] ${streamId} temporarily disconnected, waiting for reconnect...`);
        pushItem.status = 'connecting'; // 临时断开，不杀
      }
    };

    return pushItem;
  }

  stopPush(streamId: string) {
    const push = this.activePushes.get(streamId);
    if (!push) return;

    push.pc.close();
    push.stream.getTracks().forEach(t => t.stop());
    
    if (push.hiddenVideo) {
      push.hiddenVideo.pause();
      push.hiddenVideo.removeAttribute('src');
      push.hiddenVideo.load();
      if (push.hiddenVideo.parentNode) {
        push.hiddenVideo.parentNode.removeChild(push.hiddenVideo);
      }
    }
    
    this.activePushes.delete(streamId);
    mediaApi.stopPush(streamId).catch(() => {});
  }

  stopAll() {
    Array.from(this.activePushes.keys()).forEach(id => this.stopPush(id));
  }
}

export const livePusher = new LivePusherManager();
