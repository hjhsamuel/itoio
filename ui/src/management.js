import { state } from "./state.js";
import { icon, field, escapeHtml, xhrJson, notice, api, render } from "./utils.js";

export function controlPage(webrtcStatusText, statusText) {
  const room = state.room;
  if (!room) return `<div>未连接到设备</div>`;
  
  return `
    <div class="room-layout">
      <section class="room-stage">
        <header class="room-header">
          <div class="room-id">
            ${icon("shield")}控制设备 ${escapeHtml(room.id.replace("rd-", ""))}
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
          <div class="room-count">${icon("users")}<span class="room-count-text">${(room.users || []).length} 人在线</span></div>
        </div>
      </section>
      <aside class="chat-panel">
        <div class="chat-head">
          <div>
            <p class="eyebrow">Stats</p>
            <h2>连接统计</h2>
          </div>
        </div>
        <div class="chat-list" style="padding: 1rem; font-size: 0.875rem; color: var(--text-muted);">
            <p>设备 ID: ${escapeHtml(room.id.replace("rd-", ""))}</p>
            <p>信令状态: ${statusText()}</p>
            <p>WebRTC: ${webrtcStatusText()}</p>
        </div>
      </aside>
    </div>
  `;
}

function videoTiles() {
  const remoteEntries = Array.from(state.remoteStreams.entries());
  const remoteTiles = remoteEntries.slice(0, 1).map(([id]) => `
    <div class="video-tile">
      <video data-peer="${escapeHtml(id)}" autoplay playsinline tabindex="0"></video>
      <div class="tile-label">${escapeHtml(peerName(id))}</div>
      <button class="fullscreen-btn" title="${state.isFullscreen ? "退出全屏" : "全屏"}" data-fullscreen>
        ${icon(state.isFullscreen ? "minimize" : "maximize")}
      </button>
    </div>
  `).join("");

  const hasVideo = remoteTiles;
  return `
    ${remoteTiles}
    ${hasVideo ? "" : '<div class="empty-video">等待设备视频流连接。</div>'}
  `;
}

function peerName(id) {
  return state.room?.users?.find((user) => user.id === id)?.nickname || id;
}

export function devicesPage(statusText) {
  return `
    <header class="topbar">
      <div>
        <p class="eyebrow">Remote Control</p>
        <h1>远程控制</h1>
      </div>
      <div class="status-pill ${state.wsStatus}">${icon("wifi")}${statusText()}</div>
    </header>
    <section class="surface device-control">
      <div class="section-head">
        <div>
          <p class="eyebrow">Devices</p>
          <h2>受控设备列表</h2>
          <p>选择并开始远程控制您的设备</p>
        </div>
        <button class="ghost" id="refresh-devices">${icon("link")}刷新</button>
      </div>
    <div class="device-list">
      ${state.devices.map((item) => `
        <div class="device-row">
          <div class="device-info">
            ${icon("screen")}
            <strong>${escapeHtml(item.name)}</strong>
            <span class="badge ${item.state === 1 ? "offline" : "online"}">${item.state === 1 ? "离线" : "在线"}</span>
          </div>
          <div class="device-actions">
            ${item.state === 1 ? `<button class="primary small" data-control-device="${escapeHtml(item.id)}">${icon("monitor")}控制</button>` : ""}
          </div>
        </div>
      `).join("") || '<p class="muted">还没有绑定的设备。</p>'}
    </div>
    ${state.devicePagination.total > state.devicePagination.limit ? `
      <div class="pagination">
        <button class="ghost small" id="prev-page" ${state.devicePagination.page <= 1 ? "disabled" : ""}>${icon("chevronLeft")}上一页</button>
        <span>第 ${state.devicePagination.page} / ${Math.ceil(state.devicePagination.total / state.devicePagination.limit)} 页</span>
        <button class="ghost small" id="next-page" ${state.devicePagination.page >= Math.ceil(state.devicePagination.total / state.devicePagination.limit) ? "disabled" : ""}>下一页${icon("chevronRight")}</button>
      </div>
    ` : ""}
  </section>
  `;
}

export async function loadDevices(page) {
  if (page) state.devicePagination.page = page;
  try {
    const res = await xhrJson("GET", `/ito/device?page=${state.devicePagination.page}&limit=${state.devicePagination.limit}`);
    state.devices = res?.data || [];
    state.devicePagination.total = res?.total || 0;
    render();
  } catch (err) {
    notice(err.message, "error");
  }
}

export function roomsPage(statusText) {
  return `
    <header class="topbar">
      <div>
        <p class="eyebrow">Room Management</p>
        <h1>房间管理</h1>
      </div>
      <div class="status-actions">
        <div class="status-pill ${state.wsStatus}">${icon("wifi")}${statusText()}</div>
        <button class="ghost" id="connect-ws">${icon("wifi")}连接信令</button>
      </div>
    </header>
    <section class="surface room-control">
      <div class="section-head">
        <div>
          <p class="eyebrow">Create or Join</p>
          <h2>${state.room && !state.room.id.startsWith("rd-") ? `当前房间 ${escapeHtml(state.room.id)}` : "创建或加入房间"}</h2>
        </div>
        ${state.room && !state.room.id.startsWith("rd-") ? `<button class="secondary" data-page="room">${icon("video")}进入房间</button>` : ""}
      </div>
      <div class="split">
        <form id="create-room" class="stack">
          <h3>创建房间</h3>
          ${field("secret", "房间密钥", "password", "用于成员加入")}
          <button class="primary" type="submit">${icon("plus")}创建</button>
        </form>
        <form id="join-room" class="stack">
          <h3>加入房间</h3>
          ${field("id", "房间 ID", "text", "输入创建者分享的 ID")}
          ${field("secret", "房间密钥", "password", "输入房间密钥")}
          <button class="secondary" type="submit">${icon("users")}加入</button>
        </form>
      </div>
    </section>
  `;
}

export function settingsPage(statusText) {
  return `
    <header class="topbar">
      <div>
        <p class="eyebrow">Settings</p>
        <h1>设置</h1>
      </div>
      <div class="status-pill ${state.wsStatus}">${icon("wifi")}${statusText()}</div>
    </header>
    <div class="settings-grid">
      ${invitePanel()}
      ${accountPanel()}
    </div>
  `;
}

function invitePanel() {
  return `
    <section class="surface compact" id="invite">
      <div class="section-head">
        <div>
          <p class="eyebrow">Invites</p>
          <h2>邀请码管理</h2>
          <p>邀请码有效期为7天，最多同时生成5个</p>
        </div>
        <button class="ghost" id="refresh-invites">${icon("link")}刷新</button>
      </div>
      <button class="secondary wide" id="create-invite">${icon("key")}生成邀请码</button>
      <div class="invite-list">
        ${state.invites.map((item) => `
          <div class="invite-row">
            <code>${escapeHtml(item.code || item)}</code>
            <span>${formatExpire(item.expire)}</span>
            <button class="icon-button" data-copy="${escapeHtml(item.code || item)}" title="复制" aria-label="复制">${icon("copy")}</button>
          </div>
        `).join("") || '<p class="muted">还没有加载邀请码。</p>'}
      </div>
    </section>
  `;
}

function accountPanel() {
  return `
    <section class="surface compact" id="account">
      <div class="section-head">
        <div>
          <p class="eyebrow">Accounts</p>
          <h2>用户账号管理</h2>
        </div>
      </div>
      <div class="account-card">
        <span>当前账号</span>
        <strong>${escapeHtml(state.session.nickname || state.session.id || "用户")}</strong>
      </div>
      <form id="password-form" class="stack">
        ${field("old", "当前密码", "password", "输入当前密码")}
        ${field("new", "新密码", "password", "包含大小写、数字和特殊字符，长度不低于8位")}
        ${field("confirm", "确认新密码", "password", "再次输入新密码")}
        <button class="primary" type="submit">${icon("lock")}更新密码</button>
      </form>
    </section>
  `;
}

function formatExpire(expire) {
  if (!expire) return "有效期未返回";
  return new Date(expire * 1000).toLocaleString();
}

export async function createInvite() {
  try {
    const data = await api("/ito/admin/code", { method: "POST" });
    state.invites.unshift(data);
    notice("邀请码已生成", "success");
    render();
  } catch (err) {
    notice(err.message, "error");
  }
}

export async function loadInvites() {
  try {
    state.invites = await xhrJson("GET", "/ito/admin/code?page=1&limit=20") || [];
    notice("邀请码已刷新", "success");
    render();
  } catch (err) {
    notice(err.message, "error");
  }
}

export async function updatePassword(event) {
  event.preventDefault();
  const form = event.currentTarget;
  const data = Object.fromEntries(new FormData(form));
  if (data.new !== data.confirm) {
    notice("两次输入的密码不一致", "error");
    return;
  }
  try {
    await api("/ito/admin/passwd", { method: "PUT", body: JSON.stringify(data) });
    form.reset();
    notice("密码已更新", "success");
  } catch (err) {
    notice(err.message, "error");
  }
}
