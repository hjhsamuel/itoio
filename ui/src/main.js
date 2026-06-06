const state = {
  session: readSession(),
  authRoute: location.pathname === "/register" ? "register" : "login",
  page: "rooms",
  invites: [],
  room: null,
  socket: null,
  reconnectTimer: null,
  refreshTimer: null,
  manualSocketClose: false,
  reconnectAttempts: 0,
  iceServers: [],
  iceMode: "stun",
  relayFallbackPeers: new Set(),
  peerStates: new Map(),
  stream: null,
  peers: new Map(),
  dataChannels: new Map(),
  remoteStreams: new Map(),
  chat: [],
  notices: [],
  wsStatus: "offline",
  webrtcStatus: "offline"
};

const app = document.querySelector("#app");
const AUTH_COOKIE = "ito";
const MAX_CHAT_MESSAGES = 100;
const MAX_WS_RECONNECT_ATTEMPTS = 5;
const WS_RECONNECT_BASE_DELAY = 1000;
const WS_RECONNECT_MAX_DELAY = 16000;
const REFRESH_SAFETY_WINDOW = 30000;

const icons = {
  copy: '<svg viewBox="0 0 24 24"><rect x="9" y="9" width="11" height="11" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>',
  key: '<svg viewBox="0 0 24 24"><circle cx="7.5" cy="15.5" r="5.5"/><path d="m21 2-9.6 9.6M15 8h5v5"/></svg>',
  link: '<svg viewBox="0 0 24 24"><path d="M10 13a5 5 0 0 0 7.1 0l2-2a5 5 0 0 0-7.1-7.1l-1.2 1.2"/><path d="M14 11a5 5 0 0 0-7.1 0l-2 2a5 5 0 0 0 7.1 7.1l1.2-1.2"/></svg>',
  lock: '<svg viewBox="0 0 24 24"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>',
  logOut: '<svg viewBox="0 0 24 24"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><path d="m16 17 5-5-5-5M21 12H9"/></svg>',
  plus: '<svg viewBox="0 0 24 24"><path d="M12 5v14M5 12h14"/></svg>',
  screen: '<svg viewBox="0 0 24 24"><rect x="3" y="4" width="18" height="13" rx="2"/><path d="M8 21h8M12 17v4"/></svg>',
  send: '<svg viewBox="0 0 24 24"><path d="m22 2-7 20-4-9-9-4Z"/><path d="M22 2 11 13"/></svg>',
  settings: '<svg viewBox="0 0 24 24"><path d="M12 15.5a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7Z"/><path d="M19.4 15a1.7 1.7 0 0 0 .34 1.88l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06A1.7 1.7 0 0 0 15 19.4a1.7 1.7 0 0 0-1 .6 1.7 1.7 0 0 0-.4 1.1V21a2 2 0 1 1-4 0v-.09A1.7 1.7 0 0 0 8.6 19.4a1.7 1.7 0 0 0-1.88.34l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06A1.7 1.7 0 0 0 4.6 15a1.7 1.7 0 0 0-.6-1 1.7 1.7 0 0 0-1.1-.4H3a2 2 0 1 1 0-4h.09A1.7 1.7 0 0 0 4.6 8.6a1.7 1.7 0 0 0-.34-1.88l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06A1.7 1.7 0 0 0 9 4.6a1.7 1.7 0 0 0 1-.6 1.7 1.7 0 0 0 .4-1.1V3a2 2 0 1 1 4 0v.09A1.7 1.7 0 0 0 15.4 4.6a1.7 1.7 0 0 0 1.88-.34l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06A1.7 1.7 0 0 0 19.4 9c.36.11.7.31 1 .6.3.3.49.64.6 1H21a2 2 0 1 1 0 4h-.09A1.7 1.7 0 0 0 19.4 15Z"/></svg>',
  shield: '<svg viewBox="0 0 24 24"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10Z"/></svg>',
  users: '<svg viewBox="0 0 24 24"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.9M16 3.1a4 4 0 0 1 0 7.8"/></svg>',
  video: '<svg viewBox="0 0 24 24"><path d="m22 8-6 4 6 4V8Z"/><rect x="2" y="6" width="14" height="12" rx="2"/></svg>',
  wifi: '<svg viewBox="0 0 24 24"><path d="M5 13a10 10 0 0 1 14 0M8.5 16.5a5 5 0 0 1 7 0"/><path d="M12 20h.01"/></svg>',
  maximize: '<svg viewBox="0 0 24 24"><path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3"/></svg>'
};

function readSession() {
  try {
    return JSON.parse(localStorage.getItem("itoio.session") || "null");
  } catch {
    return null;
  }
}

function saveSession(session) {
  state.session = session;
  if (session) localStorage.setItem("itoio.session", JSON.stringify(session));
  else localStorage.removeItem("itoio.session");
}

function persistLogin(data) {
  saveSession({
    id: data.id,
    nickname: data.nickname,
    expire: Date.now() + (Number(data.expire || 0) * 1000)
  });
  scheduleRefresh();
}

function setAuthCookie(token, maxAge) {
  const secure = location.protocol === "https:" ? "; Secure" : "";
  document.cookie = `${AUTH_COOKIE}=${encodeURIComponent(token)}; Path=/; Max-Age=${Number(maxAge || 600)}; SameSite=Lax${secure}`;
}

function clearAuthCookie() {
  document.cookie = `${AUTH_COOKIE}=; Path=/; Max-Age=0; SameSite=Lax`;
}

async function api(path, options = {}) {
  const headers = { "Content-Type": "application/json", ...(options.headers || {}) };
  const res = await fetch(path, { ...options, headers, credentials: "same-origin" });
  const payload = await res.json().catch(() => ({}));
  if (!res.ok || (payload.code && payload.code >= 400)) {
    if (isAuthFailure(res.status, payload)) handleAuthFailure();
    throw new Error(payload.message || `请求失败：${res.status}`);
  }
  return ('data' in payload) ? payload.data : '';
}

function xhrJson(method, path, body) {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open(method, path);
    xhr.withCredentials = true;
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.onload = () => {
      let payload = {};
      try { payload = JSON.parse(xhr.responseText || "{}"); } catch {}
      if (xhr.status >= 200 && xhr.status < 300 && (!payload.code || payload.code < 400)) resolve(payload.data);
      else {
        if (isAuthFailure(xhr.status, payload)) handleAuthFailure();
        reject(new Error(payload.message || `请求失败：${xhr.status}`));
      }
    };
    xhr.onerror = () => reject(new Error("网络不可用"));
    xhr.send(body ? JSON.stringify(body) : undefined);
  });
}

function isAuthFailure(status, payload = {}) {
  const message = String(payload.message || "").toLowerCase();
  return status === 401 || payload.code === 401 || message.includes("unauthorized") || message.includes("认证失败");
}

function handleAuthFailure() {
  if (!state.session) return;
  logout({ redirect: true, silent: true });
}

function scheduleRefresh() {
  if (state.refreshTimer) {
    clearTimeout(state.refreshTimer);
    state.refreshTimer = null;
  }
  if (!state.session?.expire) return;
  const delay = Math.max(0, state.session.expire - Date.now() - REFRESH_SAFETY_WINDOW);
  state.refreshTimer = setTimeout(refreshSession, delay);
}

async function refreshSession() {
  if (!state.session) return;
  try {
    const data = await api("/refresh", { method: "PUT" });
    state.session.expire = Date.now() + (Number(data?.expire || 0) * 1000);
    saveSession(state.session);
    scheduleRefresh();
  } catch (err) {
    if (state.session) notice(err.message, "error");
  }
}

function notice(text, tone = "info") {
  state.notices.unshift({ id: crypto.randomUUID(), text, tone });
  state.notices = state.notices.slice(0, 4);
  updateToasts();
  setTimeout(() => {
    state.notices = state.notices.filter((item) => item.text !== text);
    updateToasts();
  }, 4200);
}

function updateToasts() {
  const container = document.querySelector(".toasts");
  if (container) {
    container.innerHTML = state.notices.map((n) => `<div class="toast ${n.tone}">${escapeHtml(n.text)}</div>`).join("");
  }
}

function icon(name) {
  return `<span class="icon" aria-hidden="true">${icons[name] || ""}</span>`;
}

function render() {
  app.innerHTML = state.session ? shell() : authView();
  bindCommon();
  state.session ? bindApp() : bindAuth();
  attachVideoElements();
}

function authView() {
  const isRegister = state.authRoute === "register";
  return `
    <main class="auth-shell">
      <section class="auth-art" aria-label="itoio">
        <div class="brand-mark">${icon("link")}<span>itoio</span></div>
        <div class="signal-map">
          <span class="node n1"></span><span class="node n2"></span><span class="node n3"></span>
          <span class="beam b1"></span><span class="beam b2"></span><span class="beam b3"></span>
        </div>
        <div class="auth-copy">
          <p class="eyebrow">P2P WebRTC Control Plane</p>
          <h1>私有房间里的低延迟屏幕协同</h1>
          <p>服务端只负责身份、房间和信令协调，音视频数据通过浏览器端到端连接传输</p>
        </div>
      </section>
      <section class="auth-panel">
        <div class="panel-head">
          <p class="eyebrow">${isRegister ? "Create Identity" : "Welcome Back"}</p>
          <h2>${isRegister ? "注册账号" : "登录 itoio"}</h2>
        </div>
        <form id="auth-form" class="stack">
          ${isRegister ? field("code", "邀请码", "text", "输入邀请码") : ""}
          ${field("name", "用户名", "text", "账号")}
          ${isRegister ? field("nickname", "昵称", "text", "用于房间展示") : ""}
          ${field("password", "密码", "password", "至少 8 位，包含大小写、数字和特殊字符")}
          <button class="primary" type="submit">${isRegister ? icon("plus") + "创建账号" : icon("shield") + "进入控制台"}</button>
        </form>
        <button class="text-button" data-auth-route="${isRegister ? "login" : "register"}">
          ${isRegister ? "已有账号，去登录" : "使用邀请码注册新账号"}
        </button>
      </section>
      ${toastHost()}
    </main>
  `;
}

function field(name, label, type, placeholder, value = "") {
  return `
    <label class="field">
      <span>${label}</span>
      <input name="${name}" type="${type}" placeholder="${placeholder}" value="${escapeHtml(value)}" autocomplete="${type === "password" ? "current-password" : "off"}">
    </label>
  `;
}

function shell() {
  const inRoom = state.page === "room" && state.room;
  return `
    <main class="app-shell ${inRoom ? "in-room" : ""}">
      <aside class="sidebar">
        <div class="brand-mark">${icon("link")}<span>itoio</span></div>
        <nav>
          <button class="nav-item ${state.page === "rooms" ? "active" : ""}" data-page="rooms">${icon("video")}房间管理</button>
          <button class="nav-item ${state.page === "settings" ? "active" : ""}" data-page="settings">${icon("settings")}设置</button>
        </nav>
        <div class="account">
          <span>${escapeHtml(state.session.nickname || state.session.id || "用户")}</span>
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
  if (state.page === "room" && state.room) return roomPage();
  if (state.page === "settings") return settingsPage();
  return roomsPage();
}

function roomsPage() {
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
          <h2>${state.room ? `当前房间 ${escapeHtml(state.room.id)}` : "创建或加入房间"}</h2>
        </div>
        ${state.room ? `<button class="secondary" data-page="room">${icon("video")}进入房间</button>` : ""}
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

function settingsPage() {
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

function roomPage() {
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
      <button class="fullscreen-btn" title="全屏" onclick="this.closest('.video-tile').requestFullscreen()">${icon("maximize")}</button>
    </div>
  `).join("");

  const localTile = state.stream ? `
    <div class="video-tile local">
      <video id="local-video" autoplay muted playsinline></video>
      <div class="tile-label">本地共享</div>
      <button class="fullscreen-btn" title="全屏" onclick="this.closest('.video-tile').requestFullscreen()">${icon("maximize")}</button>
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
        <button class="primary" type="submit">${icon("lock")}更新密码</button>
      </form>
    </section>
  `;
}

function toastHost() {
  return `<div class="toasts">${state.notices.map((n) => `<div class="toast ${n.tone}">${escapeHtml(n.text)}</div>`).join("")}</div>`;
}

function bindCommon() {
  document.querySelectorAll(".icon svg").forEach((svg) => {
    svg.setAttribute("fill", "none");
    svg.setAttribute("stroke", "currentColor");
    svg.setAttribute("stroke-width", "2");
    svg.setAttribute("stroke-linecap", "round");
    svg.setAttribute("stroke-linejoin", "round");
  });
}

function bindAuth() {
  document.querySelector("[data-auth-route]")?.addEventListener("click", (event) => {
    state.authRoute = event.currentTarget.dataset.authRoute;
    history.replaceState(null, "", state.authRoute === "register" ? "/register" : "/login");
    render();
  });
  document.querySelector("#auth-form")?.addEventListener("submit", async (event) => {
    event.preventDefault();
    const data = Object.fromEntries(new FormData(event.currentTarget));
    try {
      if (state.authRoute === "register") {
        await api("/register", { method: "POST", body: JSON.stringify(data) });
        state.authRoute = "login";
        notice("注册成功，请登录", "success");
        render();
        return;
      }
      const session = await api("/login", { method: "POST", body: JSON.stringify(data) });
      persistLogin(session);
      notice("登录成功", "success");
      history.replaceState(null, "", "/");
      render();
      setTimeout(loadInvites, 0);
      setTimeout(connectSocket, 0);
    } catch (err) {
      notice(err.message, "error");
    }
  });
}

function bindApp() {
  document.querySelector("#logout")?.addEventListener("click", logout);
  document.querySelectorAll("[data-page]").forEach((item) => {
    item.addEventListener("click", () => {
      const next = item.dataset.page;
      if (next === "room" && !state.room) {
        notice("请先创建或加入房间", "error");
        return;
      }
      state.page = next;
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
    sendRoomEvent("join", { id: form.get("id"), secret: form.get("secret") });
  });
  document.querySelector("#leave-room")?.addEventListener("click", leaveRoom);
  document.querySelector("#share-screen")?.addEventListener("click", () => {
    state.stream ? stopShare() : startShare();
  });
  document.querySelector("#create-invite")?.addEventListener("click", createInvite);
  document.querySelector("#refresh-invites")?.addEventListener("click", loadInvites);
  document.querySelector("#password-form")?.addEventListener("submit", updatePassword);
  document.querySelector("#chat-form")?.addEventListener("submit", sendChat);
  document.querySelectorAll("[data-copy]").forEach((btn) => btn.addEventListener("click", async () => {
    await navigator.clipboard.writeText(btn.dataset.copy);
    const label = btn.getAttribute("aria-label") || "内容";
    notice(`已复制${label.includes("房间") ? "房间 ID" : "邀请码"}`, "success");
  }));
  const chatList = document.querySelector("#chat-list");
  if (chatList) chatList.scrollTop = chatList.scrollHeight;
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
  state.wsStatus = "connecting";
  render();
  const ws = new WebSocket(url);
  state.socket = ws;
  ws.onopen = () => {
    state.reconnectAttempts = 0;
    state.wsStatus = "online";
    notice("信令已连接", "success");
    render();
  };
  ws.onmessage = (event) => handleSignal(JSON.parse(event.data));
  ws.onclose = () => {
    if (state.socket === ws) state.socket = null;
    state.wsStatus = "offline";
    if (!state.manualSocketClose && state.session) scheduleReconnect();
    else render();
  };
  ws.onerror = () => {
    state.wsStatus = "offline";
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
    state.page = "room";
    if (isNewRoom) state.chat = [];
    render();
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
  if (["offer", "answer", "candidate"].includes(type)) {
    await receivePeerSignal(type, data);
  }
}

async function startShare() {
  try {
    const display = await navigator.mediaDevices.getDisplayMedia({
      video: {
        width: { max: 1920 },
        height: { max: 1080 },
        frameRate: { max: 60 }
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
  state.peers.forEach((peer) => peer.close());
  state.peers.clear();
  state.dataChannels.clear();
  state.remoteStreams.clear();
  state.peerStates.clear();
  state.relayFallbackPeers.clear();
  state.webrtcStatus = "offline";
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
    const options = { iceServers: state.iceServers };
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
      state.remoteStreams.set(id, stream);
      render();
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
  if (state.stream) {
    const senders = peer.getSenders();
    state.stream.getTracks().forEach((track) => {
      if (!senders.some((s) => s.track === track)) {
        const sender = peer.addTrack(track, state.stream);
        if (track.kind === "video") {
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

  if (initiator && (!existed || tracksAdded)) {
    const offer = await peer.createOffer();
    await peer.setLocalDescription(offer);
    sendSignal(id, "offer", offer, { relay: peer.__itoioRelay });
  }
  return peer;
}

async function receivePeerSignal(type, payload) {
  const from = payload.from;
  const wantsRelay = payload.relay === true;
  const peer = await ensurePeer(from, false, wantsRelay);
  if (type === "offer") {
    await peer.setRemoteDescription(payload.data);
    const answer = await peer.createAnswer();
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
    state.peerStates.set(id, direct ? "direct" : "relay");
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
  render();
}

function sendSignal(to, typ, data, options = {}) {
  state.socket?.send(JSON.stringify({
    type: typ,
    data: { from: state.session.id, to, typ, data, ...options }
  }));
}

function registerDataChannel(id, channel) {
  state.dataChannels.set(id, channel);
  channel.onmessage = (event) => {
    try {
      const payload = JSON.parse(event.data);
      if (payload.type === "chat") pushChat(payload.data);
    } catch {
      pushChat({
        from: id,
        name: peerName(id),
        room: state.room?.id,
        text: event.data,
        time: Date.now()
      });
    }
  };
  channel.onclose = () => {
    if (state.dataChannels.get(id) === channel) state.dataChannels.delete(id);
  };
}

function sendChat(event) {
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
  const payload = JSON.stringify({ type: "chat", data });
  let delivered = 0;
  state.dataChannels.forEach((channel) => {
    if (channel.readyState === "open") {
      channel.send(payload);
      delivered += 1;
    }
  });
  if (delivered === 0 && (state.room?.users || []).length > 1) {
    notice("WebRTC 数据通道尚未建立，消息仅保留在本地。", "error");
  }
  input.value = "";
}

function pushChat(data) {
  state.chat.push(data);
  state.chat = state.chat.slice(-MAX_CHAT_MESSAGES);
  render();
}

async function createInvite() {
  try {
    const data = await api("/ito/admin/code", { method: "POST" });
    state.invites.unshift(data);
    notice("邀请码已生成", "success");
    render();
  } catch (err) {
    notice(err.message, "error");
  }
}

async function loadInvites() {
  try {
    state.invites = await xhrJson("GET", "/ito/admin/code?page=1&limit=20") || [];
    notice("邀请码已刷新", "success");
    render();
  } catch (err) {
    notice(err.message, "error");
  }
}

async function updatePassword(event) {
  event.preventDefault();
  const data = Object.fromEntries(new FormData(event.currentTarget));
  try {
    await api("/ito/admin/passwd", { method: "PUT", body: JSON.stringify(data) });
    event.currentTarget.reset();
    notice("密码已更新", "success");
  } catch (err) {
    notice(err.message, "error");
  }
}

function currentUser() {
  return state.room?.users?.find((user) => user.id === state.session?.id);
}

function peerName(id) {
  return state.room?.users?.find((user) => user.id === id)?.nickname || id;
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

function formatExpire(expire) {
  if (!expire) return "有效期未返回";
  return new Date(expire * 1000).toLocaleString();
}

function escapeHtml(value) {
  return String(value ?? "").replace(/[&<>"']/g, (char) => ({
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    '"': "&quot;",
    "'": "&#039;"
  }[char]));
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
