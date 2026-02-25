/**
 * Telegram Bot extension for Sypher-mini
 * Connects via node-telegram-bot-api and relays messages to/from Go core.
 *
 * Protocol: HTTP callback for inbound, HTTP POST for outbound (same as Baileys)
 * Config: TELEGRAM_BOT_TOKEN env, SYPHER_CORE_CALLBACK (e.g. http://localhost:18790/inbound)
 */

import TelegramBot from 'node-telegram-bot-api';
import * as http from 'http';

const BOT_TOKEN = process.env.TELEGRAM_BOT_TOKEN || '';
const PORT = parseInt(process.env.PORT || '3003', 10);
const CORE_CALLBACK = process.env.SYPHER_CORE_CALLBACK || 'http://localhost:18790/inbound';

interface InboundPayload {
  type: string;
  channel: string;
  from: string;
  content: string;
  chat_id: string;
}

let bot: TelegramBot | null = null;

async function sendToCore(payload: InboundPayload) {
  const opts = {
    method: 'POST' as const,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  };
  const maxRetries = 3;
  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      const res = await fetch(CORE_CALLBACK, opts);
      if (res.ok) return;
    } catch (e) {
      const err = e as Error & { cause?: unknown };
      const errMsg = [err?.message, String(err?.cause ?? ''), String(e)].join(' ');
      const isRefused = errMsg.includes('ECONNREFUSED');
      if (attempt === maxRetries - 1) {
        console.error('Failed to send to core:', e);
        if (isRefused) {
          console.error('Hint: Start the gateway first: sypher gateway');
        }
      } else {
        await new Promise((r) => setTimeout(r, 1000 * (attempt + 1)));
      }
    }
  }
}

function connect() {
  if (!BOT_TOKEN) {
    console.error('TELEGRAM_BOT_TOKEN is required. Get one from @BotFather on Telegram.');
    process.exit(1);
  }
  bot = new TelegramBot(BOT_TOKEN, { polling: true });

  bot.on('message', async (msg) => {
    const chatId = msg.chat?.id;
    const fromId = msg.from?.id;
    const text = msg.text;
    if (!chatId || !fromId || !text) return;

    const from = String(fromId);
    await sendToCore({
      type: 'inbound',
      channel: 'telegram',
      from,
      content: text,
      chat_id: String(chatId),
    });
  });

  console.log('Telegram bot connected. Send messages to your bot.');
}

// HTTP server for outbound (core -> extension)
const server = http.createServer(async (req, res) => {
  if (req.method === 'POST' && req.url === '/send') {
    let body = '';
    req.on('data', (chunk) => { body += chunk; });
    req.on('end', async () => {
      try {
        const { to, content } = JSON.parse(body);
        if (bot && to && content) {
          await bot.sendMessage(to, content);
          res.writeHead(200, { 'Content-Type': 'application/json' });
          res.end(JSON.stringify({ ok: true }));
        } else {
          res.writeHead(400);
          res.end(JSON.stringify({ error: 'Missing to or content' }));
        }
      } catch (e) {
        res.writeHead(500);
        res.end(JSON.stringify({ error: String(e) }));
      }
    });
  } else if (req.method === 'GET' && req.url === '/health') {
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({ ok: true, bot: !!bot }));
  } else {
    res.writeHead(404);
    res.end();
  }
});

connect();
server.listen(PORT, () => {
  console.log(`Telegram extension listening on port ${PORT}`);
});
