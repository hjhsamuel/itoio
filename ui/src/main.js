import { state, saveSession } from "./state.js";
import {
  icon, render, setRenderCallback, notice, handleAuthFailure, setAuthFailureCallback,
  scheduleRefresh, bindCommon, toastHost, clearAuthCookie,
  MAX_WS_RECONNECT_ATTEMPTS, WS_RECONNECT_MAX_DELAY, WS_RECONNECT_BASE_DELAY
} from "./utils.js";
import { authView, bindAuth, setLoginCallback } from "./auth.js";
import {
  roomsPage, settingsPage, loadInvites, createInvite, updatePassword,
  devicesPage, loadDevices, controlPage
} from "./management.js";
import {
  roomPage, updateRoomStatusUI, sendChat, pushChat, currentUser, peerName, sendInput,
  MSG_TYPE_CHAT, MSG_TYPE_INPUT, MSG_TYPE_CONTROL
} from "./room.js";
import { initCapture } from "./device.js";

const app = document.querySelector("#app");

setRenderCallback(() => {
  app.innerHTML = state.session ? shell() : authView();
  bindCommon();
  state.session ? bindApp() : bindAuth();
  attachVideoElements();
});

setAuthFailureCallback(() => {
  logout({ redirect: true, silent: true });
});

setLoginCallback(() => {
  setTimeout(loadInvites, 0);
  setTimeout(loadDevices, 0);
  setTimeout(connectSocket, 0);
});

function shell() {
  const inRoom = (state.page === "room" || state.page === "control") && state.room;
  return `
    <main class="app-shell ${inRoom ? "in-room" : ""}">
      <button class="sidebar-toggle" id="toggle-sidebar" aria-label="切换菜单">${icon("settings")}</button>
      <aside class="sidebar" id="app-sidebar">
        <div class="brand-mark">${icon("link")}<span>itoio</span></div>
        <nav>
          <button class="nav-item ${state.page === "rooms" ? "active" : ""}" data-page="rooms">${icon("video")}房间管理</button>
          <button class="nav-item ${state.page === "devices" ? "active" : ""}" data-page="devices">${icon("shield")}远程控制</button>
          <button class="nav-item ${state.page === "settings" ? "active" : ""}" data-page="settings">${icon("settings")}设置</button>
        </nav>
        <div class="account">
          <span>${state.session.nickname || state.session.id || "用户"}</span>
          <button class="icon-button" id="logout" title="登出" aria-label="登出">${icon("logOut")}</button>
        </div>
      </aside>
      <section class="workspace">
        ${pageView()}
      </section>
      ${toastHost()}
    </main>
  `;
}

function pageView() {
  if (state.page === "room" && state.room) return roomPage(webrtcStatusText, statusText);
  if (state.page === "control" && state.room) return controlPage(webrtcStatusText, statusText);
  if (state.page === "devices") return devicesPage(statusText);
  if (state.page === "settings") return settingsPage(statusText);
  return roomsPage(statusText);
}

function bindApp() {
  const sidebar = document.querySelector("#app-sidebar");
  const onFullscreenChange = () => {
    state.isFullscreen = !!(document.fullscreenElement || document.webkitFullscreenElement);
    if (state.isFullscreen && /Android|iPhone|iPad|iPod/i.test(navigator.userAgent)) {
      screen.orientation?.lock?.("landscape").catch(() => {});
    } else {
      try { screen.orientation?.unlock?.(); } catch (e) {}
    }

    document.querySelectorAll("[data-fullscreen]").forEach((btn) => {
      btn.title = state.isFullscreen ? "退出全屏" : "全屏";
      btn.innerHTML = icon(state.isFullscreen ? "minimize" : "maximize");
      bindCommon();
    });
  };
  document.addEventListener("fullscreenchange", onFullscreenChange);
  document.addEventListener("webkitfullscreenchange", onFullscreenChange);

  document.querySelector("#toggle-sidebar")?.addEventListener("click", () => {
    sidebar?.classList.toggle("open");
  });
  document.querySelector("#logout")?.addEventListener("click", logout);
  document.querySelectorAll("[data-page]").forEach((item) => {
    item.addEventListener("click", () => {
      const next = item.dataset.page;
      if (next === "room" && !state.room) {
        notice("请先创建或加入房间", "error");
        return;
      }
      if (state.room && next !== "room" && next !== "control" && next !== state.page) {
        leaveRoom();
      }
      state.page = next;
      sidebar?.classList.remove("open");
      render();
    });
  });
  document.querySelector("#connect-ws")?.addEventListener("click", connectSocket);
  document.querySelector("#create-room")?.addEventListener("submit", (event) => {
    event.preventDefault();
    sendRoomEvent("create", { secret: new FormData(event.currentTarget).get("secret") });
  });
  document.querySelector("#join-room")?.addEventListener("submit", (event) => {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const id = form.get("id");
    if (id && id.startsWith("rd-")) {
      notice("非法房间 ID", "error");
      return;
    }
    sendRoomEvent("join", { id: id, secret: form.get("secret") });
  });
  document.querySelector("#leave-room")?.addEventListener("click", leaveRoom);
  document.querySelector("#share-screen")?.addEventListener("click", () => {
    state.stream ? stopShare() : startShare();
  });
  document.querySelector("#create-invite")?.addEventListener("click", createInvite);
  document.querySelector("#refresh-invites")?.addEventListener("click", loadInvites);
  document.querySelectorAll("[data-control-device]").forEach((btn) => {
    btn.addEventListener("click", () => {
      const deviceId = btn.dataset.controlDevice;
      sendRoomEvent("cjoin", { id: "rd-" + deviceId, secret: "" });
    });
  });

  document.querySelector("#refresh-devices")?.addEventListener("click", () => loadDevices());
  document.querySelector("#prev-page")?.addEventListener("click", () => loadDevices(state.devicePagination.page - 1));
  document.querySelector("#next-page")?.addEventListener("click", () => loadDevices(state.devicePagination.page + 1));
  document.querySelector("#password-form")?.addEventListener("submit", updatePassword);
  document.querySelector("#chat-form")?.addEventListener("submit", sendChat);
  document.querySelectorAll("[data-fullscreen]").forEach((btn) => {
    btn.addEventListener("click", () => {
      const tile = btn.closest(".video-tile");
      if (!tile) return;
      const video = tile.querySelector("video");
      if (document.fullscreenElement || document.webkitFullscreenElement) {
        if (document.exitFullscreen) document.exitFullscreen();
        else if (document.webkitExitFullscreen) document.webkitExitFullscreen();
      } else {
        if (tile.requestFullscreen) {
          tile.requestFullscreen().catch(() => {});
        } else if (tile.webkitRequestFullscreen) {
          tile.webkitRequestFullscreen().catch(() => {});
        } else if (video && video.webkitEnterFullscreen) {
          video.webkitEnterFullscreen();
        }
      }
    });
  });
  document.querySelectorAll("[data-copy]").forEach((btn) => btn.addEventListener("click", async () => {
    await navigator.clipboard.writeText(btn.dataset.copy);
    const label = btn.getAttribute("aria-label") || "内容";
    notice(`已复制${label.includes("房间") ? "房间 ID" : "邀请码"}`, "success");
  }));
  const chatList = document.querySelector("#chat-list");
  if (chatList) chatList.scrollTop = chatList.scrollHeight;
}

function updateWsStatus(status) {
  state.wsStatus = status;
  if (state.page === "room" && state.room) {
    updateRoomStatusUI(webrtcStatusText, statusText);
  } else {
    render();
  }
}

function connectSocket() {
  if (state.socket?.readyState === WebSocket.OPEN) return;
  if (state.reconnectTimer) {
    clearTimeout(state.reconnectTimer);
    state.reconnectTimer = null;
  }
  const proto = location.protocol === "https:" ? "wss:" : "ws:";
  const url = `${proto}//${location.host}/ito/stream`;
  state.manualSocketClose = false;
  updateWsStatus("connecting");
  const ws = new WebSocket(url);
  state.socket = ws;
  ws.onopen = () => {
    state.reconnectAttempts = 0;
    updateWsStatus("online");
    notice("信令已连接", "success");
  };
  ws.onmessage = (event) => handleSignal(JSON.parse(event.data));
  ws.onclose = () => {
    if (state.socket === ws) state.socket = null;
    updateWsStatus("offline");
    if (!state.manualSocketClose && state.session) scheduleReconnect();
  };
  ws.onerror = () => {
    updateWsStatus("offline");
    ws.close();
  };
}

function scheduleReconnect() {
  if (state.reconnectTimer) return;
  state.reconnectAttempts += 1;
  if (state.reconnectAttempts > MAX_WS_RECONNECT_ATTEMPTS) {
    notice("信令重连失败，请重新登录", "error");
    handleAuthFailure();
    return;
  }
  const delay = Math.min(WS_RECONNECT_MAX_DELAY, WS_RECONNECT_BASE_DELAY * (2 ** (state.reconnectAttempts - 1)));
  notice(`信令连接已断开，${Math.round(delay / 1000)} 秒后自动重连（${state.reconnectAttempts}/${MAX_WS_RECONNECT_ATTEMPTS}）。`, "error");
  render();
  state.reconnectTimer = setTimeout(() => {
    state.reconnectTimer = null;
    connectSocket();
  }, delay);
}

function closeSocketManually() {
  state.manualSocketClose = true;
  if (state.reconnectTimer) {
    clearTimeout(state.reconnectTimer);
    state.reconnectTimer = null;
  }
  state.socket?.close();
}

function sendRoomEvent(type, data) {
  if (!state.socket || state.socket.readyState !== WebSocket.OPEN) {
    notice("请先连接信令", "error");
    return;
  }
  state.socket.send(JSON.stringify({ type, data }));
}

function leaveRoom() {
  if (state.socket?.readyState === WebSocket.OPEN) {
    sendRoomEvent("leave", {});
  } else {
    closeSocketManually();
  }
  cleanupRoom();
  notice("已离开房间", "success");
}

function cleanupRoom() {
  stopShare();
  state.peers.forEach((peer, id) => closePeer(id));
  state.peers.clear();
  state.dataChannels.clear();
  state.remoteStreams.clear();
  state.room = null;
  state.chat = [];
  state.peerStates.clear();
  state.relayFallbackPeers.clear();
  state.webrtcStatus = "offline";
  state.page = "rooms";
  render();
}

async function handleSignal(message) {
  const { type, data } = message;
  if (type === "ice_config") {
    state.iceServers = Array.isArray(data?.ice_servers) ? data.ice_servers : [];
    state.iceMode = data?.mode || "stun";
    return;
  }
  if (type === "enter") {
    if (data.ok) notice(`已进入房间 ${data.data}`, "success");
    else notice(data.data || "进入房间失败", "error");
    return;
  }
  if (type === "room") {
    const isNewRoom = state.room?.id !== data.id;
    state.room = data;
    if (data.id.startsWith("rd-")) {
      state.page = "control";
    } else {
      state.page = "room";
    }
    if (isNewRoom) {
      state.chat = [];
      render();
    } else {
      updateRoomStatusUI(webrtcStatusText, statusText);
    }
    if (currentUser()?.owner) {
      for (const user of data.users.filter((u) => !u.owner && u.id !== state.session.id)) {
        await ensurePeer(user.id, true);
      }
    }
    return;
  }
  if (type === "room_closed") {
    cleanupRoom();
    notice(data?.reason || "房间已关闭", "error");
    return;
  }
  if (["offer", "answer", "candidate", "stop_share"].includes(type)) {
    await receivePeerSignal(type, data);
  }
}

async function startShare() {
  try {
    const display = await navigator.mediaDevices.getDisplayMedia({
      video: {
        width: { ideal: 1920 },
        height: { ideal: 1080 },
        frameRate: { ideal: 60 }
      },
      audio: true
    });
    const tracks = [...display.getTracks()];
    try {
      const mic = await navigator.mediaDevices.getUserMedia({ audio: true, video: false });
      tracks.push(...mic.getAudioTracks());
    } catch (e) {
      console.warn("Could not acquire microphone", e);
    }
    state.stream = new MediaStream(tracks);
    display.getVideoTracks()[0]?.addEventListener("ended", () => stopShare());
    render();
    const targets = (state.room?.users || []).filter((u) => !u.owner && u.id !== state.session.id);
    for (const user of targets) {
      await ensurePeer(user.id, true);
    }
  } catch (err) {
    if (err.name !== "NotAllowedError") {
      notice(err.message || "无法开始屏幕共享", "error");
    }
    stopShare();
  }
}

function stopShare() {
  if (state.stream) {
    state.stream.getTracks().forEach((track) => track.stop());
    state.stream = null;
  }
  const targets = (state.room?.users || []).filter((u) => !u.owner && u.id !== state.session.id);
  for (const user of targets) {
    sendSignal(user.id, "stop_share", {});
  }
  state.peers.forEach((peer) => {
    peer.getSenders().forEach(sender => {
      if (sender.track) {
        sender.track.stop();
        sender.replaceTrack(null).catch(() => {});
      }
    });
  });
  render();
}

async function ensurePeer(id, initiator = false, forceRelay = false) {
  let peer = state.peers.get(id);
  let existed = !!peer;
  if (peer && peer.__itoioRelay !== forceRelay) {
    closePeer(id);
    peer = null;
    existed = false;
  }
  if (!peer) {
    const options = {
      iceServers: state.iceServers,
      bundlePolicy: "max-bundle",
      rtcpMuxPolicy: "require"
    };
    if (forceRelay) options.iceTransportPolicy = "relay";
    peer = new RTCPeerConnection(options);
    peer.__itoioRelay = forceRelay;
    state.peers.set(id, peer);
    state.peerStates.set(id, "connecting");
    updateWebrtcStatus();
    if (initiator) {
      registerDataChannel(id, peer.createDataChannel("chat"));
    }
    peer.ondatachannel = (event) => {
      if (event.channel.label === "chat") registerDataChannel(id, event.channel);
    };
    peer.ontrack = (event) => {
      const stream = event.streams[0] || new MediaStream([event.track]);
      if (event.track.kind === "video" && "playoutDelayHint" in event.track) {
        event.track.playoutDelayHint = 0;
      }
      state.remoteStreams.set(id, stream);
      render();
      event.track.addEventListener("ended", () => {
        if (state.remoteStreams.get(id) === stream) {
          state.remoteStreams.delete(id);
          render();
        }
      });
    };
    peer.onicecandidate = (event) => {
      if (event.candidate) sendSignal(id, "candidate", event.candidate, { relay: peer.__itoioRelay });
    };
    peer.onconnectionstatechange = () => {
      if (["failed", "disconnected", "closed"].includes(peer.connectionState)) {
        handlePeerFailure(id, peer);
      }
    };
    peer.oniceconnectionstatechange = () => {
      updatePeerHolePunchStatus(id, peer);
    };
  }

  let tracksAdded = false;
  let tracksReplaced = false;
  if (state.stream) {
    const senders = peer.getSenders();
    state.stream.getTracks().forEach((track) => {
      const existingSender = senders.find((s) => s.track?.kind === track.kind);
      if (existingSender) {
        if (existingSender.track !== track) {
          existingSender.replaceTrack(track).catch(e => console.error("replaceTrack failed", e));
          tracksReplaced = true;
        }
      } else {
        const sender = peer.addTrack(track, state.stream);
        if (track.kind === "video") {
          track.contentHint = "motion";
          const params = sender.getParameters();
          if (!params.encodings) params.encodings = [{}];
          params.encodings[0].maxBitrate = 2000000;
          params.encodings[0].networkPriority = "high";
          sender.setParameters(params).catch((e) => console.error("Could not set video parameters", e));
        }
        tracksAdded = true;
      }
    });
  }

  if (initiator && (!existed || tracksAdded || (tracksReplaced && peer.signalingState === "stable"))) {
    let offer = await peer.createOffer();
    if (offer.sdp) {
      offer = {
        type: offer.type,
        sdp: optimizeSDP(offer.sdp)
      };
    }
    await peer.setLocalDescription(offer);
    sendSignal(id, "offer", offer, { relay: peer.__itoioRelay });
  }
  return peer;
}

function optimizeSDP(sdp) {
  if (!sdp) return sdp;
  const lines = sdp.split("\r\n");
  let videoMLineIndex = -1;
  for (let i = 0; i < lines.length; i++) {
    if (lines[i].startsWith("m=video")) {
      videoMLineIndex = i;
      break;
    }
  }
  if (videoMLineIndex !== -1) {
    const mLine = lines[videoMLineIndex];
    const parts = mLine.split(" ");
    const payloads = parts.slice(3);
    const h264Payloads = [];
    const otherPayloads = [];
    for (const pt of payloads) {
      const rtpmap = lines.find(l => l.startsWith(`a=rtpmap:${pt} H264/90000`));
      if (rtpmap) h264Payloads.push(pt);
      else otherPayloads.push(pt);
    }
    if (h264Payloads.length > 0) {
      parts.splice(3, parts.length - 3, ...h264Payloads, ...otherPayloads);
      lines[videoMLineIndex] = parts.join(" ");
    }
  }
  return lines.join("\r\n")
    .replace(/a=fmtp:(\d+) (.*)/g, (match, pt, params) => {
      if (params.indexOf("x-google-min-bitrate") === -1) {
        return `a=fmtp:${pt} ${params};x-google-min-bitrate=500;x-google-max-bitrate=2000;x-google-start-bitrate=1000`;
      }
      return match;
    })
    .replace(/a=mid:(.*)/g, (match) => {
      return `${match}\r\na=extmap-allow-mixed`;
    })
    .replace(/a=rtcp-fb:(\d+) nack\r\n/g, "")
    .replace(/a=rtcp-fb:(\d+) nack pli\r\n/g, "a=rtcp-fb:$1 nack pli\r\na=rtcp-fb:$1 goog-remb\r\n");
}

async function receivePeerSignal(type, payload) {
  const from = payload.from;
  const wantsRelay = payload.relay === true;
  if (type === "stop_share") {
    state.remoteStreams.delete(from);
    render();
    return;
  }
  const peer = await ensurePeer(from, false, wantsRelay);
  if (type === "offer") {
    if (peer.signalingState !== "stable") {
      closePeer(from);
      return receivePeerSignal(type, payload);
    }
    await peer.setRemoteDescription(payload.data);
    let answer = await peer.createAnswer();
    if (answer.sdp) {
      answer = {
        type: answer.type,
        sdp: optimizeSDP(answer.sdp)
      };
    }
    await peer.setLocalDescription(answer);
    sendSignal(from, "answer", answer, { relay: peer.__itoioRelay });
  } else if (type === "answer") {
    if ((payload.relay === true) !== peer.__itoioRelay) return;
    await peer.setRemoteDescription(payload.data);
  } else if (type === "candidate") {
    if ((payload.relay === true) !== peer.__itoioRelay) return;
    await peer.addIceCandidate(payload.data);
  }
}

function closePeer(id) {
  state.peers.get(id)?.close();
  state.peers.delete(id);
  state.dataChannels.get(id)?.close();
  state.dataChannels.delete(id);
}

function handlePeerFailure(id, peer) {
  if (state.peers.get(id) !== peer) return;
  if (state.iceMode === "turn" && !peer.__itoioRelay && !state.relayFallbackPeers.has(id) && currentUser()?.owner) {
    state.relayFallbackPeers.add(id);
    state.peerStates.set(id, "connecting");
    updateWebrtcStatus();
    closePeer(id);
    ensurePeer(id, true, true).catch((err) => {
      console.error("Could not switch WebRTC peer to TURN relay", err);
      state.peerStates.set(id, "failed");
      updateWebrtcStatus();
    });
    return;
  }
  state.peerStates.set(id, "failed");
  updateWebrtcStatus();
}

async function updatePeerHolePunchStatus(id, peer) {
  if (!["connected", "completed"].includes(peer.iceConnectionState)) {
    if (["failed", "disconnected", "closed"].includes(peer.iceConnectionState)) {
      handlePeerFailure(id, peer);
    } else {
      state.peerStates.set(id, "connecting");
    }
    updateWebrtcStatus();
    return;
  }
  try {
    const stats = await peer.getStats();
    let selectedPair = null;
    stats.forEach((report) => {
      if (report.type === "candidate-pair" && report.selected) selectedPair = report;
    });
    if (!selectedPair) {
      stats.forEach((report) => {
        if (report.type === "transport" && report.selectedCandidatePairId) selectedPair = stats.get(report.selectedCandidatePairId);
      });
    }
    const local = selectedPair ? stats.get(selectedPair.localCandidateId) : null;
    const remote = selectedPair ? stats.get(selectedPair.remoteCandidateId) : null;
    const direct = local?.candidateType !== "relay" && remote?.candidateType !== "relay";
    const oldStatus = state.peerStates.get(id);
    const newStatus = direct ? "direct" : "relay";
    state.peerStates.set(id, newStatus);
    
    if (newStatus !== oldStatus && (newStatus === "direct" || newStatus === "relay")) {
      if (state.page === "control") {
        sendInput(MSG_TYPE_CONTROL, { type: "start" });
      }
    }
  } catch {
    state.peerStates.set(id, "connecting");
  }
  updateWebrtcStatus();
}

function updateWebrtcStatus() {
  const remoteUserCount = (state.room?.users || []).filter((user) => user.id !== state.session?.id).length;
  const peerStatuses = Array.from(state.peerStates.values());
  if (!state.room || remoteUserCount === 0) {
    state.webrtcStatus = "offline";
  } else if (peerStatuses.filter((status) => status === "direct" || status === "relay").length === remoteUserCount) {
    state.webrtcStatus = "online";
  } else if (peerStatuses.includes("connecting")) {
    state.webrtcStatus = "connecting";
  } else {
    state.webrtcStatus = "offline";
  }
  if (state.page === "room" && state.room) {
    updateRoomStatusUI(webrtcStatusText, statusText);
  } else {
    render();
  }
}

function sendSignal(to, typ, data, options = {}) {
  state.socket?.send(JSON.stringify({
    type: typ,
    data: { from: state.session.id, to, typ, data, ...options }
  }));
}

function registerDataChannel(id, channel) {
  state.dataChannels.set(id, channel);
  channel.binaryType = "arraybuffer";
  channel.onmessage = (event) => {
    if (!(event.data instanceof ArrayBuffer)) return;
    const view = new Uint8Array(event.data);
    if (view.length === 0) return;
    const type = view[0];
    const payload = view.slice(1);
    if (type === MSG_TYPE_CHAT) {
      try {
        const text = new TextDecoder().decode(payload);
        const data = JSON.parse(text);
        pushChat(data);
      } catch (e) {
        console.error("Failed to decode chat message", e);
      }
    }
  };
  channel.onclose = () => {
    if (state.dataChannels.get(id) === channel) state.dataChannels.delete(id);
  };
}

function attachVideoElements() {
  const local = document.querySelector("#local-video");
  if (local && local.srcObject !== state.stream) {
    local.srcObject = state.stream;
  }
  document.querySelectorAll("video[data-peer]").forEach((video) => {
    const stream = state.remoteStreams.get(video.dataset.peer);
    if (stream && video.srcObject !== stream) {
      video.srcObject = stream;
      
      if (state.page === "control") {
        initCapture(video, (frame) => {
          sendInput(MSG_TYPE_INPUT, frame);
        });
      }
      
      video.addEventListener('mouseenter', () => {
        video.focus();
      });
    }
  });
}

function statusText() {
  return { online: "WebRTC 信令在线", connecting: "正在连接", offline: "信令离线" }[state.wsStatus];
}

function webrtcStatusText() {
  return { online: "WebRTC 直连成功", connecting: "正在打洞", offline: "WebRTC 未直连" }[state.webrtcStatus];
}

function logout(options = {}) {
  if (state.room && state.socket?.readyState === WebSocket.OPEN) {
    sendRoomEvent("leave", {});
  }
  stopShare();
  closeSocketManually();
  clearAuthCookie();
  if (state.refreshTimer) {
    clearTimeout(state.refreshTimer);
    state.refreshTimer = null;
  }
  saveSession(null);
  state.room = null;
  state.invites = [];
  state.chat = [];
  state.peerStates.clear();
  state.iceServers = [];
  state.iceMode = "stun";
  state.relayFallbackPeers.clear();
  state.reconnectAttempts = 0;
  state.page = "rooms";
  state.wsStatus = "offline";
  state.webrtcStatus = "offline";
  state.authRoute = "login";
  if (options.redirect) history.replaceState(null, "", "/login");
  render();
}

if (state.session) {
  if (state.session.expire && state.session.expire <= Date.now()) {
    handleAuthFailure();
  } else {
    if (location.pathname === "/login" || location.pathname === "/register") {
      history.replaceState(null, "", "/");
    }
    scheduleRefresh();
    setTimeout(loadInvites, 0);
    setTimeout(connectSocket, 0);
  }
}
render();
