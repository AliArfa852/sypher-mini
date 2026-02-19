"use strict";
/**
 * WhatsApp Baileys extension for Sypher-mini
 * Connects via @whiskeysockets/baileys and relays messages to/from Go core.
 *
 * Protocol: HTTP callback for inbound, HTTP POST for outbound (JSON-RPC style)
 * Config: Core sets extensions.whatsapp_baileys.url (e.g. http://localhost:3002)
 */
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const baileys_1 = __importStar(require("@whiskeysockets/baileys"));
const pino_1 = __importDefault(require("pino"));
const path = __importStar(require("path"));
const http = __importStar(require("http"));
const qrcode_terminal_1 = __importDefault(require("qrcode-terminal"));
const AUTH_DIR = process.env.SYPHER_WHATSAPP_AUTH || path.join(process.env.HOME || process.env.USERPROFILE || '.', '.sypher-mini', 'whatsapp-auth');
const PORT = parseInt(process.env.PORT || '3002', 10);
const CORE_CALLBACK = process.env.SYPHER_CORE_CALLBACK || 'http://localhost:18790/inbound';
let sock = null;
async function sendToCore(payload) {
    try {
        const url = new URL(CORE_CALLBACK);
        await fetch(url.toString(), {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
        });
    }
    catch (e) {
        console.error('Failed to send to core:', e);
    }
}
async function connect() {
    const { state, saveCreds } = await (0, baileys_1.useMultiFileAuthState)(AUTH_DIR);
    const { version } = await (0, baileys_1.fetchLatestBaileysVersion)();
    sock = (0, baileys_1.default)({
        auth: state,
        version,
        logger: (0, pino_1.default)({ level: 'silent' }),
        syncFullHistory: false,
    });
    sock.ev.on('creds.update', saveCreds);
    sock.ev.on('connection.update', (update) => {
        if (update.qr) {
            console.log('\nScan QR with WhatsApp (Settings → Linked Devices):\n');
            qrcode_terminal_1.default.generate(update.qr, { small: true });
        }
        const status = update.connection;
        const err = update.lastDisconnect?.error;
        if (status === 'close' && err?.output?.statusCode !== 401) {
            console.log('Reconnecting...');
            setTimeout(connect, 3000);
        }
        else if (status === 'open') {
            console.log('WhatsApp connected.');
        }
    });
    sock.ev.on('messages.upsert', async ({ messages }) => {
        for (const m of messages) {
            if (m.key.fromMe)
                continue;
            if (m.message?.conversation || m.message?.extendedTextMessage?.text) {
                const text = m.message?.conversation || m.message?.extendedTextMessage?.text || '';
                const from = m.key.remoteJid || '';
                await sendToCore({
                    type: 'inbound',
                    from,
                    content: text,
                    chat_id: from,
                });
            }
        }
    });
}
// HTTP server for outbound (core -> extension)
const server = http.createServer(async (req, res) => {
    if (req.method === 'POST' && req.url === '/send') {
        let body = '';
        req.on('data', (chunk) => { body += chunk; });
        req.on('end', async () => {
            try {
                const { to, content } = JSON.parse(body);
                if (sock && to && content) {
                    await sock.sendMessage(to, { text: content });
                    res.writeHead(200, { 'Content-Type': 'application/json' });
                    res.end(JSON.stringify({ ok: true }));
                }
                else {
                    res.writeHead(400);
                    res.end(JSON.stringify({ error: 'Missing to or content' }));
                }
            }
            catch (e) {
                res.writeHead(500);
                res.end(JSON.stringify({ error: String(e) }));
            }
        });
    }
    else {
        res.writeHead(404);
        res.end();
    }
});
server.listen(PORT, () => {
    console.log(`WhatsApp Baileys extension on port ${PORT}`);
    connect();
});
