import { state } from "./state.js";
import { icon, escapeHtml, MAX_CHAT_MESSAGES, notice, render } from "./utils.js";

const MSG_TYPE_CHAT = 0x01;
const MSG_TYPE_INPUT = 0x02;

export function roomPage(webrtcStatusText, statusText) {
  const room = state.room;
  return `
    <div class="room-layout">
      <section class="room-stage">
        <header class="room-header">
          <div class="room-id">
            ${icon("video")}房间 ${escapeHtml(room.id)}
            <button class="icon-button" data-copy="${escapeHtml(room.id)}" title="复制房间 ID" aria-label="复制房间 ID">${icon("copy")}</button>
          </div>
          <div class="status-actions">
            <div class="status-pill ${state.webrtcStatus}">${icon("link")}${webrtcStatusText()}</div>
            <div class="status-pill ${state.wsStatus}">${icon("wifi")}${statusText()}</div>
          </div>
        </header>
        <div class="video-window">
          ${videoTiles()}
        </div>
        <div class="room-menu">
          <button class="ghost" data-page="rooms">${icon("video")}房间管理</button>
          <button class="secondary" id="leave-room">${icon("logOut")}离开房间</button>
          <button class="${state.stream ? "secondary" : "primary"}" id="share-screen" ${currentUser()?.owner ? "" : "disabled"}>
            ${state.stream ? icon("logOut") + "停止共享" : icon("screen") + "开始共享"}
          </button>
          <div class="room-count">${icon("users")}<span>${(room.users || []).length} 人在线</span></div>
        </div>
      </section>
      <aside class="chat-panel">
        <div class="chat-head">
          <div>
            <p class="eyebrow">Chat</p>
            <h2>房间聊天</h2>
          </div>
          <span>${state.chat.length}/${MAX_CHAT_MESSAGES}</span>
        </div>
        <div class="chat-list" id="chat-list">
          ${state.chat.map(chatMessage).join("") || '<p class="muted">暂无聊天记录。</p>'}
        </div>
        <form id="chat-form" class="chat-form">
          <input name="text" maxlength="500" placeholder="输入消息">
          <button class="icon-button" type="submit" title="发送" aria-label="发送">${icon("send")}</button>
        </form>
      </aside>
    </div>
  `;
}

function videoTiles() {
  const remoteEntries = Array.from(state.remoteStreams.entries());
  const remoteTiles = remoteEntries.slice(0, 1).map(([id]) => `
    <div class="video-tile">
      <video data-peer="${escapeHtml(id)}" autoplay playsinline></video>
      <div class="tile-label">${escapeHtml(peerName(id))}</div>
      <button class="fullscreen-btn" title="${state.isFullscreen ? "退出全屏" : "全屏"}" data-fullscreen>
        ${icon(state.isFullscreen ? "minimize" : "maximize")}
      </button>
    </div>
  `).join("");

  const localTile = state.stream ? `
    <div class="video-tile local">
      <video id="local-video" autoplay muted playsinline></video>
      <div class="tile-label">本地共享</div>
      <button class="fullscreen-btn" title="${state.isFullscreen ? "退出全屏" : "全屏"}" data-fullscreen>
        ${icon(state.isFullscreen ? "minimize" : "maximize")}
      </button>
    </div>
  ` : "";

  const hasVideo = localTile || remoteTiles;
  return `
    ${localTile}
    ${remoteTiles}
    ${hasVideo ? "" : '<div class="empty-video">等待屏幕共享或媒体连接。</div>'}
  `;
}

function chatMessage(item) {
  const mine = item.from === state.session?.id ? "mine" : "";
  return `
    <article class="chat-message ${mine}">
      <div>
        <strong>${escapeHtml(item.name || peerName(item.from) || "用户")}</strong>
        <time>${new Date(item.time || Date.now()).toLocaleTimeString()}</time>
      </div>
      <p>${escapeHtml(item.text)}</p>
    </article>
  `;
}

export function currentUser() {
  return state.room?.users?.find((user) => user.id === state.session?.id);
}

export function peerName(id) {
  return state.room?.users?.find((user) => user.id === id)?.nickname || id;
}

export function pushChat(data) {
  state.chat.push(data);
  state.chat = state.chat.slice(-MAX_CHAT_MESSAGES);
  updateChatUI();
}

export function updateChatUI() {
  const list = document.querySelector("#chat-list");
  if (list) {
    list.innerHTML = state.chat.map(chatMessage).join("") || '<p class="muted">暂无聊天记录。</p>';
    list.scrollTop = list.scrollHeight;
  }
  const counter = document.querySelector(".chat-head span");
  if (counter) {
    counter.textContent = `${state.chat.length}/${MAX_CHAT_MESSAGES}`;
  }
}

export function updateRoomStatusUI(webrtcStatusText, statusText) {
  const container = document.querySelector(".status-actions");
  if (container) {
    container.innerHTML = `
      <div class="status-pill ${state.webrtcStatus}">${icon("link")}${webrtcStatusText()}</div>
      <div class="status-pill ${state.wsStatus}">${icon("wifi")}${statusText()}</div>
    `;
  }
  const counter = document.querySelector(".room-count span");
  if (counter) {
    counter.textContent = `${(state.room?.users || []).length} 人在线`;
  }
}

export function sendChat(event) {
  event.preventDefault();
  const input = event.currentTarget.elements.text;
  const text = input.value.trim();
  if (!text) return;
  const data = {
    from: state.session.id,
    name: state.session.nickname || state.session.id,
    room: state.room?.id,
    text,
    time: Date.now()
  };
  pushChat(data);
  
  const encoded = new TextEncoder().encode(JSON.stringify(data));
  const packet = new Uint8Array(1 + encoded.length);
  packet[0] = MSG_TYPE_CHAT;
  packet.set(encoded, 1);

  let delivered = 0;
  state.dataChannels.forEach((channel) => {
    if (channel.readyState === "open") {
      channel.send(packet);
      delivered += 1;
    }
  });
  if (delivered === 0 && (state.room?.users || []).length > 1) {
    notice("WebRTC 数据通道尚未建立，消息仅保留在本地。", "error");
  }
  input.value = "";
}

export function sendInput(type, eventData) {
  const data = { type, ...eventData };
  const encoded = new TextEncoder().encode(JSON.stringify(data));
  const packet = new Uint8Array(1 + encoded.length);
  packet[0] = MSG_TYPE_INPUT;
  packet.set(encoded, 1);

  state.dataChannels.forEach((channel) => {
    if (channel.readyState === "open") {
      channel.send(packet);
    }
  });
}
