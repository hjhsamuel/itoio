import { state } from "./state.js";
import { icon, field, toastHost, api, persistLogin, notice, render } from "./utils.js";

let loginCallback = () => {};
export function setLoginCallback(cb) { loginCallback = cb; }

export function authView() {
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
          <p>服务端负责身份、房间和信令协调，音视频数据优先通过浏览器端到端连接传输</p>
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

export function bindAuth() {
  document.querySelector("[data-auth-route]")?.addEventListener("click", (event) => {
    state.authRoute = event.currentTarget.dataset.authRoute;
    history.replaceState(null, "", state.authRoute === "register" ? "/register" : "/login");
    render();
  });
  document.querySelector("#auth-form")?.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const data = Object.fromEntries(new FormData(form));
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
      loginCallback();
      render();
    } catch (err) {
      notice(err.message, "error");
    }
  });
}
