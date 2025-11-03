# DevSmith Review App

**AI-powered code analysis for effective code comprehension and quality review.**

DevSmith teaches developers to read and supervise AI-generated code through five distinct reading modes, each optimized for different comprehension goals and cognitive load management.

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8.svg)](https://golang.org)
[![Docker](https://img.shields.io/badge/docker-required-2496ED.svg)](https://docker.com)

---

## � Quick Start (2 Commands)

```bash
# 1. Start Ollama (one-time setup)
ollama serve &
ollama pull mistral:7b-instruct

# 2. Start Review app
docker-compose -f docker-compose.review-only.yml up
```

**Visit: http://localhost:3000**

---

## 📑 Table of Contents

- [Overview](#overview)
- [Reading Modes](#reading-modes)
- [Prerequisites](#prerequisites)
- [Setup](#setup)
- [Usage](#usage)
- [Troubleshooting](#troubleshooting)
- [Development](#development)

---

## Overview

DevSmith addresses the critical challenge of the "Human-in-the-Loop" era: as AI generates more code, developers must shift from *writing* code to *reading, understanding, and validating* AI output.

### The Solution

DevSmith provides **5 AI-assisted reading modes**:
1. **👁️ Preview** (2-3 min) - Quick structural overview
2. **⚡ Skim** (5-7 min) - Understand abstractions and patterns
3. **🔎 Scan** (3-5 min) - Find specific code patterns
4. **🔬 Detailed** (10-15 min) - Deep algorithm analysis
5. **⚠️ Critical** (15-20 min) - Quality review and issue detection

Each mode is optimized for **cognitive load management** and builds transferable mental frameworks.

---

## Features

### Core Capabilities

✅ **5 Reading Modes** - AI-guided code analysis for different comprehension goals  
✅ **Local AI Models** - Ollama integration (no API costs, privacy-first)  
✅ **Multiple Models** - Mistral, CodeLlama, Llama2, DeepSeek support  
✅ **Circuit Breaker** - Graceful degradation when AI unavailable  
✅ **HTMX UI** - Fast, server-rendered interface  
✅ **Docker Ready** - Single-command deployment  
✅ **Health Checks** - Comprehensive monitoring of all components  
✅ **OpenTelemetry** - Distributed tracing to Jaeger  
✅ **Graceful Shutdown** - Zero data loss during restarts

### Quality & Reliability

✅ **Production-Ready** - Circuit breakers, health checks, graceful shutdown  
✅ **Error Handling** - User-friendly error templates with retry actions  
✅ **Observability** - Prometheus metrics + Jaeger tracing  
✅ **Test Coverage** - Unit, integration, and E2E tests  
✅ **API Documentation** - OpenAPI 3.0 specification  
✅ **Incident Runbooks** - Step-by-step troubleshooting guides

---

## Architecture

### Services

```
┌─────────────────────────────────────────────────────────────┐
│                    Nginx Gateway (port 3000)                 │
│                         /api/review/*                         │
└────────────────────┬────────────────────────────────────────┘
                     │
         ┌───────────┼───────────┐
         │           │           │
    ┌────▼────┐ ┌───▼────┐ ┌───▼────────┐
    │ Portal  │ │ Review │ │ Analytics  │
    │ :8080   │ │ :8081  │ │ :8083      │
    └────┬────┘ └───┬────┘ └───┬────────┘
         │          │           │
         └──────────┼───────────┘
---

## Reading Modes

### 👁️ Preview Mode (2-3 minutes)
**Purpose:** Quick structural assessment before deep dive

**What You Get:**
- File/folder tree with descriptions
- Identified bounded contexts
- Technology stack detection
- Architectural pattern analysis
- Entry points and dependencies

**Use Cases:**
- Evaluating GitHub repo before cloning
- Quick assessment of AI-generated code
- Determining project relevance

---

### ⚡ Skim Mode (5-7 minutes)
**Purpose:** Understand abstractions without implementation details

**What You Get:**
- Function/method signatures with descriptions
- Interface definitions
- Data models and entities
- Key workflows with diagrams
- API endpoint catalog

**Use Cases:**
- Understanding what a codebase does
- Preparing spec for AI implementation
- Architectural review

---

### 🔎 Scan Mode (3-5 minutes)
**Purpose:** Targeted information search

**What You Get:**
- Semantic search (not just string matching)
- Variable/function usage tracking
- Error source identification
- Pattern matching
- Related code discovery

**Use Cases:**
- "Where is auth validated?"
- "Find all database queries"
- "What calls this deprecated function?"

---

### 🔬 Detailed Mode (10-15 minutes)
**Purpose:** Deep understanding of algorithms

**What You Get:**
- Line-by-line explanation
- Variable state at each point
- Control flow analysis
- Algorithm identification
- Complexity analysis
- Edge case identification

**Use Cases:**
- Understanding complex algorithm before modifying
- Debugging subtle logic errors
- Learning from well-written code

---

### ⚠️ Critical Mode (15-20 minutes) - Human-in-the-Loop Review
**Purpose:** Evaluate quality and identify improvements

**What You Get:**
- **Architecture Issues**: Bounded context violations, layer mixing
- **Code Quality**: Error handling, scope problems, naming
- **Security**: SQL injection, unvalidated input, auth gaps
- **Performance**: N+1 queries, inefficient algorithms
- **Testing**: Untested code paths, missing coverage
- **Improvement Suggestions**: Specific refactoring with before/after

**Use Cases:**
- **PRIMARY**: Reviewing AI-generated code before merge
- Pre-commit quality checks
- Security audit
- Refactoring planning

---

## Prerequisites

### System Requirements

**Minimum (Recommended):**
- Docker 24.0+
- Docker Compose 2.0+
- 16GB RAM
- 50GB disk space
- Internet connection (for Ollama model download)

**For Best Experience:**
- 32GB RAM
- NVIDIA GPU with 8GB+ VRAM
- 100GB disk space

### Ollama Setup

```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Start Ollama service
ollama serve &

# Pull recommended model (16GB RAM)
ollama pull mistral:7b-instruct

# Alternative: Best quality model (32GB RAM)
ollama pull deepseek-coder-v2:16b

# Verify Ollama running
curl http://localhost:11434/api/tags
```

---

## Setup

### Option 1: Quick Start (Review App Only)

```bash
# Start minimal stack
docker-compose -f docker-compose.review-only.yml up

# Access Review app
open http://localhost:3000
```

**Services:**
- Nginx Gateway (port 3000)
- Review Service (port 8081)
- PostgreSQL (port 5432)
- Jaeger (port 16686, optional with --profile dev-tools)

---

### Option 2: Full Platform

```bash
# Start all services
docker-compose up -d

# Access services
open http://localhost:3000        # Portal (authentication, navigation)
open http://localhost:3000/review # Review app
open http://localhost:3000/logs   # Logs monitoring
open http://localhost:3000/analytics # Analytics dashboard
```

---

## Usage

### 1. Paste Code

Visit `http://localhost:3000/review/sessions/new`

Paste your code (or provide GitHub URL)

### 2. Select Reading Mode

Choose from dropdown:
- 👁️ Preview (quick overview)
- ⚡ Skim (abstractions)
- 🔎 Scan (search)
- 🔬 Detailed (deep dive)
- ⚠️ Critical (quality review)

### 3. Analyze

Click **"Analyze Code"** button

Wait 2-20 minutes depending on mode (see timings above)

### 4. Review Results

**Preview/Skim/Scan:** Interactive tree views, function lists, search results

**Detailed:** Line-by-line explanation with variable states

**Critical:** Issue list categorized by type (Architecture, Security, Quality, Performance, Testing)

---

## Troubleshooting

### "AI analysis service is unavailable"

**Cause:** Ollama not running or model not pulled

**Fix:**
```bash
# Check Ollama status
curl http://localhost:11434/api/tags

# If not running:
ollama serve &

# Pull model if missing:
ollama pull mistral:7b-instruct
```

---

### "Database connection failed"

**Cause:** PostgreSQL container not healthy

**Fix:**
```bash
# Check postgres status
docker-compose logs postgres

# Restart if unhealthy
docker-compose restart postgres
```

---

### "Service timeout" or slow responses

**Cause:** Model too large for available RAM

**Fix:**
```bash
# Check system RAM
free -h

# If < 16GB, use smaller model:
ollama pull mistral:7b-instruct

# Update docker-compose.review-only.yml:
# OLLAMA_MODEL=mistral:7b-instruct
```

---

### Review app shows blank page

## Development

See [ARCHITECTURE.md](ARCHITECTURE.md) for system design details.
See [DevsmithTDD.md](DevsmithTDD.md) for test-driven development guidelines.
See [Requirements.md](Requirements.md) for feature requirements.

---

## License

MIT License - see [LICENSE](LICENSE) for details.
