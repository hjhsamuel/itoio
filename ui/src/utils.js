import { state, saveSession } from "./state.js";

export const AUTH_COOKIE = "ito";
export const MAX_CHAT_MESSAGES = 100;
export const MAX_WS_RECONNECT_ATTEMPTS = 5;
export const WS_RECONNECT_BASE_DELAY = 1000;
export const WS_RECONNECT_MAX_DELAY = 16000;
export const REFRESH_SAFETY_WINDOW = 30000;

export const icons = {
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
  maximize: '<svg viewBox="0 0 24 24"><path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3"/></svg>',
  minimize: '<svg viewBox="0 0 24 24"><path d="M8 3v3a2 2 0 0 1-2 2H3m18 0h-3a2 2 0 0 1-2-2V3m0 18v-3a2 2 0 0 1 2-2h3M3 16h3a2 2 0 0 1 2 2v3"/></svg>'
};

let renderCallback = () => {};
export function setRenderCallback(cb) { renderCallback = cb; }
export function render() { renderCallback(); }

export function persistLogin(data) {
  saveSession({
    id: data.id,
    nickname: data.nickname,
    expire: Date.now() + (Number(data.expire || 0) * 1000)
  });
  scheduleRefresh();
}

export function setAuthCookie(token, maxAge) {
  const secure = location.protocol === "https:" ? "; Secure" : "";
  document.cookie = `${AUTH_COOKIE}=${encodeURIComponent(token)}; Path=/; Max-Age=${Number(maxAge || 600)}; SameSite=Lax${secure}`;
}

export function clearAuthCookie() {
  document.cookie = `${AUTH_COOKIE}=; Path=/; Max-Age=0; SameSite=Lax`;
}

export async function api(path, options = {}) {
  const headers = { "Content-Type": "application/json", ...(options.headers || {}) };
  const res = await fetch(path, { ...options, headers, credentials: "same-origin" });
  const payload = await res.json().catch(() => ({}));
  if (!res.ok || (payload.code && payload.code >= 400)) {
    if (isAuthFailure(res.status, payload)) handleAuthFailure();
    throw new Error(payload.message || `请求失败：${res.status}`);
  }
  return ('data' in payload) ? payload.data : '';
}

export function xhrJson(method, path, body) {
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

export function isAuthFailure(status, payload = {}) {
  const message = String(payload.message || "").toLowerCase();
  return status === 401 || payload.code === 401 || message.includes("unauthorized") || message.includes("认证失败");
}

let authFailureCallback = () => {};
export function setAuthFailureCallback(cb) { authFailureCallback = cb; }
export function handleAuthFailure() {
  if (!state.session) return;
  authFailureCallback();
}

export function scheduleRefresh() {
  if (state.refreshTimer) {
    clearTimeout(state.refreshTimer);
    state.refreshTimer = null;
  }
  if (!state.session?.expire) return;
  const delay = Math.max(0, state.session.expire - Date.now() - REFRESH_SAFETY_WINDOW);
  state.refreshTimer = setTimeout(refreshSession, delay);
}

export async function refreshSession() {
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

export function notice(text, tone = "info") {
  state.notices.unshift({ id: crypto.randomUUID(), text, tone });
  state.notices = state.notices.slice(0, 4);
  updateToasts();
  setTimeout(() => {
    state.notices = state.notices.filter((item) => item.text !== text);
    updateToasts();
  }, 4200);
}

export function updateToasts() {
  const container = document.querySelector(".toasts");
  if (container) {
    container.innerHTML = state.notices.map((n) => `<div class="toast ${n.tone}">${escapeHtml(n.text)}</div>`).join("");
  }
}

export function icon(name) {
  return `<span class="icon" aria-hidden="true">${icons[name] || ""}</span>`;
}

export function escapeHtml(value) {
  return String(value ?? "").replace(/[&<>"']/g, (char) => ({
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    '"': "&quot;",
    "'": "&#039;"
  }[char]));
}

export function bindCommon() {
  document.querySelectorAll(".icon svg").forEach((svg) => {
    svg.setAttribute("fill", "none");
    svg.setAttribute("stroke", "currentColor");
    svg.setAttribute("stroke-width", "2");
    svg.setAttribute("stroke-linecap", "round");
    svg.setAttribute("stroke-linejoin", "round");
  });
}

export function toastHost() {
  return `<div class="toasts">${state.notices.map((n) => `<div class="toast ${n.tone}">${escapeHtml(n.text)}</div>`).join("")}</div>`;
}

export function field(name, label, type, placeholder, value = "") {
  return `
    <label class="field">
      <span>${label}</span>
      <input name="${name}" type="${type}" placeholder="${placeholder}" value="${escapeHtml(value)}" autocomplete="${type === "password" ? "current-password" : "off"}">
    </label>
  `;
}
