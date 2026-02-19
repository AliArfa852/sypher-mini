# Docker Deployment

Deploy Sypher-mini with Docker or docker-compose.

---

## Quick Start

```bash
# Copy env template and set API keys
cp .env.example .env
# Edit .env: CEREBRAS_API_KEY=xxx or OPENAI_API_KEY=xxx

# Build and run
docker-compose up -d

# Check health
curl http://localhost:18790/health
```

---

## Configuration

### Environment variables

| Variable | Description |
|----------|-------------|
| `CEREBRAS_API_KEY` | Cerebras API key |
| `OPENAI_API_KEY` | OpenAI API key |
| `ANTHROPIC_API_KEY` | Anthropic API key |
| `GEMINI_API_KEY` | Google Gemini API key |
| `TZ` | Timezone (default: UTC) |

### Persistent data

Config and workspace are stored in the `sypher-mini-data` volume. To inspect or edit config:

```bash
docker run --rm -v sypher-mini-data:/data alpine cat /data/config.json
```

---

## Build only

```bash
docker build -t sypher-mini:latest .
```

## Clean & rebuild

```bash
# Rebuild image (no cache)
docker-compose build --no-cache

# Stop, rebuild, start
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

Or with Make: `make docker-down` then `make docker` then `make docker-run`.

## Run without compose

```bash
docker run -d \
  -p 18790:18790 \
  -e CEREBRAS_API_KEY=your-key \
  -v sypher-mini-data:/home/sypher/.sypher-mini \
  --name sypher-mini \
  sypher-mini:latest
```

---

## WhatsApp (Baileys)

For WhatsApp via Baileys, the extension runs as a separate Node process. In Docker, you would need a multi-container setup or run the extension outside the container. Use the WebSocket bridge for containerized deployments.
