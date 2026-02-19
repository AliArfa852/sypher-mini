/**
 * WhatsApp Baileys extension for Sypher-mini
 * Connects via @whiskeysockets/baileys and relays messages to/from Go core.
 *
 * Protocol: HTTP callback for inbound, HTTP POST for outbound (JSON-RPC style)
 * Config: Core sets extensions.whatsapp_baileys.url (e.g. http://localhost:3002)
 */
import makeWASocket, { useMultiFileAuthState, fetchLatestBaileysVersion } from '@whiskeysockets/baileys';
import pino from 'pino';
import * as path from 'path';
import * as http from 'http';
import qrcode from 'qrcode-terminal';
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
    const { state, saveCreds } = await useMultiFileAuthState(AUTH_DIR);
    const { version } = await fetchLatestBaileysVersion();
    sock = makeWASocket({
        auth: state,
        version,
        logger: pino({ level: 'silent' }),
        syncFullHistory: false,
    });
    sock.ev.on('creds.update', saveCreds);
    sock.ev.on('connection.update', (update) => {
        if (update.qr) {
            console.log('\nScan QR with WhatsApp (Settings → Linked Devices):\n');
            qrcode.generate(update.qr, { small: true });
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
