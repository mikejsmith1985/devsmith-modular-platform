# Continue Extension Setup Guide

**Purpose:** Use DeepSeek-Coder with Ollama and Continue extension in VS Code for AI-assisted coding.

This setup provides local AI coding assistance with good effort-to-results ratio, no API costs, and complete privacy.

---

## Prerequisites

- **OS**: macOS, Linux, or Windows
- **RAM**: 16GB+ recommended (8GB minimum with smaller models)
- **Storage**: 10GB free for models
- **VS Code**: Installed and updated

---

## Installation Steps

### 1. Install Ollama

**macOS:**
```bash
brew install ollama
```

**Linux:**
```bash
curl -fsSL https://ollama.ai/install.sh | sh
```

**Windows:**
Download from [ollama.com](https://ollama.com)

**Verify Installation:**
```bash
ollama --version
```

---

### 2. Start Ollama Service

```bash
ollama serve
```

**Verify Service Running:**
```bash
curl http://localhost:11434
# Should return: "Ollama is running"
```

**Note:** Ollama runs as a background service. You can start it once and leave it running.

---

### 3. Pull DeepSeek-Coder Model

Choose a model based on your available RAM:

**For 16GB RAM (Recommended):**
```bash
ollama pull deepseek-coder:6.7b
```

**For 8GB RAM (Minimal):**
```bash
ollama pull deepseek-coder:1.5b
```

**For 32GB RAM (Best Quality):**
```bash
ollama pull deepseek-coder-v2:16b
```

**Alternative (16GB RAM):**
```bash
ollama pull qwen2.5-coder:7b
```

**Verify Model Downloaded:**
```bash
ollama list
# Should show your selected model
```

---

### 4. Install Continue Extension in VS Code

1. Open VS Code
2. Open Extensions panel (`Ctrl+Shift+X` or `Cmd+Shift+X`)
3. Search for **"Continue"**
4. Click **Install**
5. Restart VS Code if prompted

---

### 5. Configure Continue

#### Open Configuration

1. Click the Continue icon in the Activity Bar (left sidebar)
2. Click the gear icon ⚙️ at the bottom of the Continue panel
3. Select **"Open Config"**

This opens `~/.continue/config.json` (or `config.yaml` depending on version)

#### Add DeepSeek-Coder Configuration

**For JSON config:**
```json
{
  "models": [
    {
      "name": "DeepSeek Coder",
      "provider": "ollama",
      "model": "deepseek-coder:6.7b",
      "roles": ["chat", "edit", "apply", "autocomplete"],
      "capabilities": ["tool_use"],
      "completionOptions": {
        "temperature": 0.7,
        "top_p": 0.9,
        "maxTokens": 4096
      }
    }
  ],
  "tabAutocompleteModel": {
    "provider": "ollama",
    "model": "deepseek-coder:6.7b"
  }
}
```

**For YAML config:**
```yaml
models:
  - name: DeepSeek Coder
    provider: ollama
    model: deepseek-coder:6.7b
    roles:
      - chat
      - edit
      - apply
      - autocomplete
    capabilities:
      - tool_use
    completionOptions:
      temperature: 0.7
      top_p: 0.9
      maxTokens: 4096

tabAutocompleteModel:
  provider: ollama
  model: deepseek-coder:6.7b
```

**Adjust model name if using different version** (e.g., `:1.5b`, `-v2:16b`)

---

### 6. Use Continue

#### Chat Mode
1. Open Continue sidebar (icon in Activity Bar)
2. Select "DeepSeek Coder" from model dropdown
3. Ask coding questions or request code generation

#### Inline Editing
1. Highlight code in editor
2. Press `Ctrl+I` (or `Cmd+I` on macOS)
3. Describe what you want to change
4. Continue will suggest edits

#### Autocomplete
- Type code naturally
- Continue will suggest completions inline
- Press `Tab` to accept suggestions

---

## Common Problems & Solutions

### ❌ Model not found (404 error)

**Solution:**
```bash
# Pull the model again
ollama pull deepseek-coder:6.7b

# Verify it's installed
ollama list

# Check exact tag name matches config
```

---

### ❌ Connection errors

**Causes:**
- Ollama not running
- Firewall blocking port 11434

**Solutions:**
```bash
# Ensure Ollama is running
ollama serve

# Verify connection
curl http://localhost:11434

# Check if port is open
netstat -an | grep 11434  # Linux/macOS
netstat -ano | findstr 11434  # Windows

# Check firewall settings (allow port 11434)
```

---

### ❌ Agent mode not supported

**Error:** `Agent mode requires tool_use capability`

**Solution:**

Add `capabilities` to config:
```json
{
  "capabilities": ["tool_use"]
}
```

**Alternative:** Use a different model like `qwen2.5-coder:7b` which has better tool support.

---

### ❌ Performance is slow

**Solutions:**

**1. Use a smaller model:**
```bash
# Switch to 1.5b for faster responses
ollama pull deepseek-coder:1.5b
```

Update config:
```json
{
  "model": "deepseek-coder:1.5b"
}
```

**2. Adjust completion settings:**
```json
{
  "completionOptions": {
    "temperature": 0.7,
    "top_p": 0.9,
    "maxTokens": 2048
  }
}
```

**3. Monitor Ollama resource usage:**
```bash
ollama ps  # Shows running models and memory usage
```

**4. Close other memory-intensive applications**

---

### ❌ Tag mismatch errors

**Error:** `Model tag not found`

**Solution:**

Always pull the exact tag:
```bash
# NOT this:
ollama pull deepseek-coder

# Use this:
ollama pull deepseek-coder:6.7b
```

List available tags:
```bash
ollama list
```

---

## Model Comparison for Continue

| Model | RAM | Speed | Quality | Best Use Case |
|-------|-----|-------|---------|---------------|
| `deepseek-coder:1.5b` | 8GB | ⚡⚡⚡ | ⭐⭐ | Quick autocomplete, simple edits |
| `deepseek-coder:6.7b` | 16GB | ⚡⚡ | ⭐⭐⭐ | **Balanced - recommended** |
| `deepseek-coder-v2:16b` | 32GB | ⚡ | ⭐⭐⭐⭐ | Complex refactoring, architecture questions |
| `qwen2.5-coder:7b` | 16GB | ⚡⚡ | ⭐⭐⭐ | Alternative to 6.7b, better tool use |

---

## Tips for Best Results

### 1. Be Specific in Prompts
```
❌ "Fix this code"
✅ "Add error handling for network timeouts in this function"
```

### 2. Provide Context
```
❌ "Create a user model"
✅ "Create a Go struct for User with fields: ID (uuid), Email (string), CreatedAt (time.Time)"
```

### 3. Use Temperature Settings
- **0.2-0.3**: Deterministic, factual code
- **0.7-0.8**: Balanced creativity
- **0.9-1.0**: Experimental, diverse suggestions

### 4. Monitor Memory Usage
```bash
# Check Ollama memory usage
ollama ps

# If using too much RAM, switch to smaller model
```

### 5. Restart Ollama if Issues Persist
```bash
# Kill Ollama
pkill ollama

# Restart
ollama serve
```

---

## Integration with DevSmith Platform

Continue is complementary to the DevSmith platform:

- **Continue**: IDE assistance (autocomplete, inline edits, chat)
- **DevSmith Review**: Structured code analysis with 5 reading modes
- **DevSmith Analytics**: Development pattern analysis

All three use the same Ollama instance and models, so setup is shared!

---

## Advanced Configuration

### Custom System Prompt
```json
{
  "systemMessage": "You are an expert Go developer. Always follow Go best practices and write idiomatic Go code."
}
```

### Multiple Models
```json
{
  "models": [
    {
      "name": "Fast (1.5B)",
      "model": "deepseek-coder:1.5b",
      "roles": ["autocomplete"]
    },
    {
      "name": "Quality (6.7B)",
      "model": "deepseek-coder:6.7b",
      "roles": ["chat", "edit"]
    }
  ]
}
```

### Context Length
```json
{
  "completionOptions": {
    "maxTokens": 8192  // Increase for larger context
  }
}
```

---

## Troubleshooting Checklist

- [ ] Ollama service is running (`curl http://localhost:11434`)
- [ ] Model is downloaded (`ollama list`)
- [ ] Model name in config matches exactly (including tag)
- [ ] VS Code restarted after config changes
- [ ] Sufficient RAM available (check with `htop` or Task Manager)
- [ ] Firewall allows port 11434
- [ ] Continue extension is latest version

---

## Getting Help

**Continue Documentation:** https://continue.dev/docs
**Ollama Documentation:** https://ollama.ai/docs
**DevSmith Issues:** https://github.com/your-repo/issues

---

**Created:** 2025-10-20
**For:** VS Code users of DevSmith Platform
**Tested with:** Continue v0.8+, Ollama v0.1.20+
