import { state } from "./state.js";
import { icon, notice, render } from "./utils.js";

let isCapturing = false;
let eventCallback = null;
const pressedKeys = new Set();

// Frame constants and state
const MAX_EVENT_PER_FRAME = 64;
const FPS = 60;
const FRAME_INTERVAL = 1000 / FPS;

const EventType = {
    KeyDown: 1,
    KeyUp: 2,
    MouseMove: 3,
    MouseLeftDown: 4,
    MouseLeftUp: 5,
    MouseRightDown: 6,
    MouseRightUp: 7,
    MouseMiddleDown: 8,
    MouseMiddleUp: 9,
    MouseWheel: 10,
};

// Map KeyboardEvent.code to our internal Key index (referencing KeyMap in Go)
const CodeToKeyIndex = {
    'KeyA': 0, 'KeyB': 1, 'KeyC': 2, 'KeyD': 3, 'KeyE': 4, 'KeyF': 5, 'KeyG': 6, 'KeyH': 7, 'KeyI': 8, 'KeyJ': 9,
    'KeyK': 10, 'KeyL': 11, 'KeyM': 12, 'KeyN': 13, 'KeyO': 14, 'KeyP': 15, 'KeyQ': 16, 'KeyR': 17, 'KeyS': 18, 'KeyT': 19,
    'KeyU': 20, 'KeyV': 21, 'KeyW': 22, 'KeyX': 23, 'KeyY': 24, 'KeyZ': 25,
    'Backspace': 26, 'Tab': 27, 'Enter': 28, 'ShiftLeft': 29, 'ShiftRight': 29, 'ControlLeft': 30, 'ControlRight': 30,
    'AltLeft': 31, 'AltRight': 31, 'Pause': 32, 'CapsLock': 33, 'Escape': 34, 'Space': 35, 'PageUp': 36, 'PageDown': 37,
    'End': 38, 'Home': 39, 'ArrowLeft': 40, 'ArrowUp': 41, 'ArrowRight': 42, 'ArrowDown': 43, 'PrintScreen': 44,
    'Insert': 45, 'Delete': 46,
    'Digit0': 47, 'Digit1': 48, 'Digit2': 49, 'Digit3': 50, 'Digit4': 51, 'Digit5': 52, 'Digit6': 53, 'Digit7': 54, 'Digit8': 55, 'Digit9': 56,
    'MetaLeft': 57, 'MetaRight': 58,
    'Numpad0': 59, 'Numpad1': 60, 'Numpad2': 61, 'Numpad3': 62, 'Numpad4': 63, 'Numpad5': 64, 'Numpad6': 65, 'Numpad7': 66, 'Numpad8': 67, 'Numpad9': 68,
    'NumpadMultiply': 69, 'NumpadAdd': 70, 'NumpadEnter': 71, 'NumpadSubtract': 72, 'NumpadDecimal': 73, 'NumpadDivide': 74,
    'F1': 75, 'F2': 76, 'F3': 77, 'F4': 78, 'F5': 79, 'F6': 80, 'F7': 81, 'F8': 82, 'F9': 83, 'F10': 84, 'F11': 85, 'F12': 86,
    'NumLock': 87, 'ScrollLock': 88,
    'Semicolon': 95, 'Equal': 96, 'Comma': 97, 'Minus': 98, 'Period': 99, 'Slash': 100, 'Backquote': 101, 'BracketLeft': 102, 'Backslash': 103, 'BracketRight': 104, 'Quote': 105
};

const MouseButtons = {
    Left: 1,
    Right: 2,
    Middle: 4,
};

let frameID = 0;
let lastFrameTime = performance.now();
let pendingEvents = [];
let currentKeyboard = [0n, 0n, 0n, 0n]; // 4 * uint64
let currentMouse = {
    Buttons: 0,
    X: 0,
    Y: 0
};

function updateKeyboardBitmap(keyIndex, isPress) {
    if (keyIndex === undefined || keyIndex < 0 || keyIndex >= 256) return;
    const arrayIdx = Math.floor(keyIndex / 64);
    const bitIdx = BigInt(keyIndex % 64);
    if (isPress) {
        currentKeyboard[arrayIdx] |= (1n << bitIdx);
    } else {
        currentKeyboard[arrayIdx] &= ~(1n << bitIdx);
    }
}

function getFrame() {
    const now = performance.now();
    const frame = {
        Seq: frameID, // We use frameID for Seq as well for now
        FrameID: frameID++,
        Keyboard: [...currentKeyboard],
        Mouse: { ...currentMouse },
        EventCnt: Math.min(pendingEvents.length, MAX_EVENT_PER_FRAME),
        Events: pendingEvents.slice(0, MAX_EVENT_PER_FRAME)
    };
    pendingEvents = pendingEvents.slice(MAX_EVENT_PER_FRAME);
    lastFrameTime = now;
    return frame;
}

function marshalFrame(frame) {
    const size = 58 + frame.EventCnt * 33;
    const buf = new ArrayBuffer(size);
    const view = new DataView(buf);

    // Header (58 bytes)
    view.setUint32(0, frame.Seq, true);
    view.setUint32(4, frame.FrameID, true);

    // Keyboard (32 bytes)
    for (let i = 0; i < 4; i++) {
        view.setBigUint64(8 + i * 8, frame.Keyboard[i], true);
    }

    // Mouse (17 bytes)
    view.setUint8(40, frame.Mouse.Buttons);
    view.setFloat64(41, frame.Mouse.X, true);
    view.setFloat64(49, frame.Mouse.Y, true);

    // EventCnt (1 byte)
    view.setUint8(57, frame.EventCnt);

    // Events (EventCnt * 33 bytes)
    for (let i = 0; i < frame.EventCnt; i++) {
        const e = frame.Events[i];
        const offset = 58 + i * 33;
        view.setUint8(offset + 0, e.Type);
        view.setUint16(offset + 1, e.Key, true);
        view.setInt32(offset + 3, e.DX, true);
        view.setInt32(offset + 7, e.DY, true);
        view.setFloat64(offset + 11, e.X, true);
        view.setFloat64(offset + 19, e.Y, true);
        view.setInt32(offset + 27, e.Wheel, true);
        view.setUint16(offset + 31, e.OffsetUs, true);
    }

    return new Uint8Array(buf);
}

function uint8ArrayToHex(arr) {
    return Array.from(arr).map(b => b.toString(16).padStart(2, '0')).join('');
}


let captureTimer = null;

function sendFrame() {
    if (!isCapturing) return;
    const frame = getFrame();
    const binary = marshalFrame(frame);
    if (eventCallback) {
        eventCallback(binary);
    } else {
        console.log(`[Frame]`, uint8ArrayToHex(binary));
    }
}

export function initCapture(videoElement, callback) {
    if (!videoElement) return;

    if (eventCallback === callback && isCapturing) {
        // Already initialized for this callback, but maybe different video element
        // However, the previous listeners are on the old video element.
        // For simplicity in this project, we assume one video element at a time for capture.
    }

    eventCallback = callback;
    isCapturing = true;

    const startTimer = () => {
        if (!captureTimer) {
            captureTimer = setInterval(sendFrame, FRAME_INTERVAL);
        }
    };

    const stopTimer = () => {
        if (captureTimer) {
            clearInterval(captureTimer);
            captureTimer = null;
        }
    };

    // Focus handling to ensure keyboard events are captured
    videoElement.addEventListener('focus', () => {
        console.log("Video focused - starting capture");
        startTimer();
    });

    videoElement.addEventListener('blur', () => {
        console.log("Video blurred - stopping capture after reset");
        
        // Reset state
        if (pressedKeys.size > 0) {
            console.log(`Clearing ${pressedKeys.size} pressed keys on blur`);
            pressedKeys.forEach(code => {
                const keyIndex = CodeToKeyIndex[code];
                updateKeyboardBitmap(keyIndex, false);
                logEvent('KeyUp', { keyIndex: keyIndex });
            });
            pressedKeys.clear();
        }
        currentMouse.Buttons = 0;
        
        // Send one last frame with reset state
        sendFrame();
        stopTimer();
    });

    videoElement.addEventListener('mouseleave', () => {
        videoElement.blur();
    });

    // Mouse Events
    videoElement.addEventListener('mousemove', (e) => {
        const rect = videoElement.getBoundingClientRect();
        const x = (e.clientX - rect.left) / rect.width;
        const y = (e.clientY - rect.top) / rect.height;

        const nx = Math.max(0, Math.min(1, x));
        const ny = Math.max(0, Math.min(1, y));

        currentMouse.X = nx;
        currentMouse.Y = ny;

        logEvent('MouseMove', {
            nx: nx,
            ny: ny,
            dx: e.movementX,
            dy: e.movementY
        });
    });

    videoElement.addEventListener('mousedown', (e) => {
        let type = '';
        if (e.button === 0) {
            type = 'MouseLeftDown';
            currentMouse.Buttons |= MouseButtons.Left;
        } else if (e.button === 1) {
            type = 'MouseMiddleDown';
            currentMouse.Buttons |= MouseButtons.Middle;
        } else if (e.button === 2) {
            type = 'MouseRightDown';
            currentMouse.Buttons |= MouseButtons.Right;
        }

        if (type) {
            logEvent(type, {});
        }
    });

    videoElement.addEventListener('mouseup', (e) => {
        let type = '';
        if (e.button === 0) {
            type = 'MouseLeftUp';
            currentMouse.Buttons &= ~MouseButtons.Left;
        } else if (e.button === 1) {
            type = 'MouseMiddleUp';
            currentMouse.Buttons &= ~MouseButtons.Middle;
        } else if (e.button === 2) {
            type = 'MouseRightUp';
            currentMouse.Buttons &= ~MouseButtons.Right;
        }

        if (type) {
            logEvent(type, {});
        }
    });

    videoElement.addEventListener('wheel', (e) => {
        logEvent('MouseWheel', { deltaY: Math.round(e.deltaY) });
    }, { passive: true });

    // Keyboard Events
    videoElement.addEventListener('keydown', (e) => {
        // Only capture if video is focused
        if (!pressedKeys.has(e.code)) {
            pressedKeys.add(e.code);
            const keyIndex = CodeToKeyIndex[e.code];
            updateKeyboardBitmap(keyIndex, true);
            logEvent('KeyDown', { keyIndex: keyIndex });
        }
        // Prevent default scrolling for arrow keys etc if focused
        if (["ArrowUp", "ArrowDown", "ArrowLeft", "ArrowRight", " "].includes(e.key)) {
            e.preventDefault();
        }
    });

    videoElement.addEventListener('keyup', (e) => {
        pressedKeys.delete(e.code);
        const keyIndex = CodeToKeyIndex[e.code];
        updateKeyboardBitmap(keyIndex, false);
        logEvent('KeyUp', { keyIndex: keyIndex });
    });
}

function logEvent(type, data) {
    if (!isCapturing) return;

    const now = performance.now();
    const offsetUs = Math.round((now - lastFrameTime) * 1000);

    const event = {
        Type: EventType[type] || 0,
        Key: data.keyIndex || 0,
        DX: data.dx || 0,
        DY: data.dy || 0,
        X: data.nx || 0,
        Y: data.ny || 0,
        Wheel: data.deltaY || 0,
        OffsetUs: Math.min(65535, offsetUs)
    };

    pendingEvents.push(event);
}
