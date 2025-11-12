# DevSmith Modular Platform

**AI-Powered Development Platform for Code Review, Health Monitoring, and Centralized Logging**

DevSmith is a self-hosted platform that helps development teams leverage AI for code review, system monitoring, and cross-repository logging. Built with modern microservices architecture and powered by local AI models (Ollama) or cloud AI providers (OpenAI, Anthropic).

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8.svg)](https://golang.org)
[![React](https://img.shields.io/badge/react-18.0+-61DAFB.svg)](https://reactjs.org)
[![Docker](https://img.shields.io/badge/docker-required-2496ED.svg)](https://docker.com)

---

## 🚀 Quick Start (15 Minutes)

```bash
# 1. Install Ollama (AI model host)
curl -fsSL https://ollama.com/install.sh | sh

# 2. Pull an AI model (choose based on your RAM)
ollama pull qwen2.5-coder:7b  # For 12-16GB RAM (recommended)

# 3. Start Ollama
ollama serve

# 4. Clone repository
git clone https://github.com/mikejsmith1985/devsmith-modular-platform.git
cd devsmith-modular-platform

# 5. Configure environment
cp .env.example .env
nano .env  # Add your GitHub OAuth credentials

# 6. Start platform
docker-compose up -d

# 7. Open browser
open http://localhost:3000
```

**Full setup guide:** [DEPLOYMENT.md](./DEPLOYMENT.md)

---

## ✨ Features

- **🔍 Code Review:** AI-powered analysis with 5 reading modes (Preview, Skim, Scan, Detailed, Strategic)
- **🏥 Health Monitoring:** Real-time system metrics, log aggregation, AI diagnostics
- **🔧 Projects API:** Cross-repository logging with batch ingestion (100x faster)
- **🤖 AI Factory:** Configure multiple AI providers (Ollama, OpenAI, Anthropic)
- **📊 Analytics:** *(Coming Soon)* Trend analysis, cost tracking, custom dashboards

---

## 📚 Documentation

- **[DEPLOYMENT.md](./DEPLOYMENT.md)** - Installation & setup (15 min)
- **[API_INTEGRATION.md](./API_INTEGRATION.md)** - Integrate external apps (5 min)
- **[ARCHITECTURE.md](./ARCHITECTURE.md)** - System architecture
- **[TROUBLESHOOTING_GUIDE.md](./docs/TROUBLESHOOTING_GUIDE.md)** - Common issues

---

## 🏗️ Architecture

**Stack:** React 18 + Go + PostgreSQL + Redis + Traefik + Jaeger  
**Services:** Portal, Review, Logs, Analytics  
**AI Providers:** Ollama (local), OpenAI, Anthropic

See [ARCHITECTURE.md](./ARCHITECTURE.md) for details.

---

## 📋 Prerequisites

| Software | Version | Purpose |
|----------|---------|---------|
| Docker | 24.0+ | Container runtime |
| Docker Compose | 2.20+ | Multi-container orchestration |
| Ollama | 0.1.0+ | AI model hosting |
| Git | 2.30+ | Source control |

**System Requirements:** 12GB RAM, 4 CPU cores, 20GB storage (minimum)

**GitHub OAuth:** Register at https://github.com/settings/developers

---

## 💻 Usage

### Code Review
1. Navigate to http://localhost:3000/review
2. Paste code or connect GitHub repository
3. Select reading mode (Preview / Skim / Scan / Detailed / Strategic)
4. Review AI insights

### Send Logs from Your App
```bash
curl -X POST http://localhost:3000/api/logs/batch \
  -H "X-API-Key: dsk_your_api_key_here" \
  -H "Content-Type: application/json" \
  -d '{"project_slug":"my-app","logs":[...]}'
```

See [API_INTEGRATION.md](./API_INTEGRATION.md) for Node.js, Go, Python, Java examples.

---

## 🐛 Troubleshooting

**Services won't start?**
```bash
docker-compose ps  # Check status
docker-compose logs portal  # View logs
```

**Ollama connection failed?**
```bash
curl http://localhost:11434/api/tags  # Test Ollama
ollama serve  # Start if not running
```

**More help:** [TROUBLESHOOTING_GUIDE.md](./docs/TROUBLESHOOTING_GUIDE.md)

---

## 🤝 Contributing

1. Fork the repository
2. Create feature branch: `git checkout -b feature/amazing-feature`
3. Write tests (see [DevsmithTDD.md](./DevsmithTDD.md))
4. Commit: `git commit -m "feat: add amazing feature"`
5. Open Pull Request

**Commit convention:** [Conventional Commits](https://www.conventionalcommits.org/)

---

## 📄 License

MIT License - see [LICENSE](./LICENSE) file

---

## 📞 Support

- **Issues:** https://github.com/mikejsmith1985/devsmith-modular-platform/issues
- **Discussions:** https://github.com/mikejsmith1985/devsmith-modular-platform/discussions
- **Beta Feedback:** Open an issue with label `beta-feedback`

---

**⭐ Star us on GitHub** if you find this project useful!
