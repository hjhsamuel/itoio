export const state = {
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
  webrtcStatus: "offline",
  isFullscreen: false
};

function readSession() {
  try {
    return JSON.parse(localStorage.getItem("itoio.session") || "null");
  } catch {
    return null;
  }
}

export function saveSession(session) {
  state.session = session;
  if (session) localStorage.setItem("itoio.session", JSON.stringify(session));
  else localStorage.removeItem("itoio.session");
}
