# First Run — Nanobot Setup

Follow these steps to get Nanobot running.

## 1. Install Go

Download from [https://go.dev/dl/](https://go.dev/dl/) — need Go 1.21 or higher.

Verify:
```bash
go version
```

## 2. Clone & Install

```bash
git clone https://github.com/yourname/nanobot-go.git
cd nanobot-go
go mod tidy
```

## 3. Configure

```bash
cp config/config.json.example config/config.json
```

Open `config/config.json` and fill in at minimum:

```json
{
  "llm": {
    "default_provider": "zhipu",
    "providers": {
      "zhipu": {
        "api_key": "YOUR_ZHIPU_API_KEY"
      }
    }
  }
}
```

> **No API key?** Nanobot works with any OpenAI-compatible API.
> For free local models, set up [Ollama](https://ollama.com/) and use the `ollama` driver:
> ```json
> "providers": {
>   "ollama": {
>     "driver": "ollama",
>     "base_url": "http://localhost:11434",
>     "model": "qwen:7b"
>   }
> }
> ```

### Optional: Enable Feishu notifications

```json
"channels": {
  "feishu": {
    "enabled": true,
    "webhook_url": "https://open.feishu.cn/open-apis/bot/v2/hook/YOUR_WEBHOOK"
  }
}
```

To create a Feishu bot webhook: open Feishu → group settings → Bots → Custom Bot → copy the webhook URL.

## 4. Run

```bash
go run ./cmd/nanobot/ -mode cli
```

## 5. Customize (optional)

Edit `storage/AGENTS.md` to change Nanobot's system prompt and behavior.

---

**Having trouble?** Check `storage/logs/llm.log` for API errors, or open an issue on GitHub.
