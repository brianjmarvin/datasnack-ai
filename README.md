# 🛡️ DataSnack AI Security Tester

A super easy CLI tool that tests any AI endpoint for security problems! 🚀

## 🎯 What This Tool Does

Think of this as a security guard for your AI endpoints. It automatically:
- 🔍 **Finds security holes** in your AI endpoints
- 🧪 **Tests with tricky prompts** to see if your AI can be tricked
- 📊 **Checks for data leaks** - making sure your AI doesn't spill secrets
- 📄 **Tests file uploads** - CSV, PDF, images, you name it!
- 🤖 **Uses AI to test AI** - pretty cool, right?

## 🚀 Quick Start (5 minutes!)

### 1. Build the tool
```bash
git clone https://github.com/brianjmarvin/DataSnackOS-RISK.git
cd code-check-cli
go build -o datasnack
```

### 2. Create a simple config file
Create `config/my-endpoint.json`:
```json
{
  "service": {
    "name": "My AI Service",
    "base_url": "http://localhost:8080",
    "description": "My awesome AI endpoint"
  },
  "endpoints": {
    "health": {
      "path": "/health",
      "method": "GET",
      "description": "Health check"
    },
    "single_evaluation": {
      "path": "/api/chat",
      "method": "POST",
      "description": "Main chat endpoint",
      "request_schema": {
        "type": "object",
        "properties": {
          "message": {
            "type": "string",
            "description": "User message"
          }
        },
        "required": ["message"]
      }
    }
  }
}
```

### 3. Run the test! 🎉
```bash
./datasnack endpointEval config/my-endpoint.json
```

That's it! The tool will test your endpoint and save results to `results/` folder.

## 📋 Main Commands

### 🔍 `analyze` - Smart Endpoint Discovery
Automatically figures out what your AI endpoint needs:

```bash
./datasnack analyze /path/to/your/ai/code
```

**What it does:**
- 🔎 **Scans your code** (Python, JavaScript, Go)
- 🧠 **Uses AI to understand** what your endpoint expects
- 📝 **Creates a perfect config** file for testing
- 🎯 **No manual work needed!**

**Example:**
```bash
./datasnack analyze /Users/me/my-ai-project
# Creates: config/endpoint_config_20250123_143022.json
```

### 🧪 `endpointEval` - Security Testing
Tests your endpoint for security problems:

```bash
./datasnack endpointEval config/endpoint_config_20250123_143022.json
```

**What it tests:**
- 🕵️ **Data Leakage** - Can your AI be tricked into revealing secrets?
- 💉 **Prompt Injection** - Can malicious prompts break your AI?
- 🔄 **Consistency** - Does your AI give the same answer to similar questions?
- 📄 **File Uploads** - Can your AI handle documents safely?

## ⚙️ Configuration Made Easy

### 🤖 AI Provider Setup (`config/agentConfig.json`)
Tell the tool which AI to use for testing:

```json
{
  "baseURL": "http://localhost:8080",
  "endpoint": "/api/v1/chatHandler",
  "agentRootFolder": "/path/to/your/agent/code",
  "trackingEnabled": true,
  "agentPurpose": "This API provides AI-powered chat functionality for customer support.",
  "testConfiguration": {
    "dataLeakageTests": 5,
    "promptInjectionTests": 5,
    "consistencyTests": 5,
    "iterationsPerTest": 3
  }
}
```

### 🔑 AI Client Setup (`config/aiClientConfig.json`)
Choose which AI provider to use for generating test prompts:

```json
{
  "preferredOrder": [
    {
      "provider": "gollm",
      "type": "openai",
      "model": "gpt-4o-mini",
      "envKey": "OPENAI_API_KEY",
      "description": "OpenAI GPT-4o-mini - Fast and cost-effective"
    },
    {
      "provider": "gollm",
      "type": "ollama",
      "model": "llama3.2",
      "envKey": "OLLAMA_ENDPOINT",
      "endpoint": "http://localhost:11434",
      "description": "Ollama Local - Complete privacy"
    }
  ],
  "fallbackToBedrock": true,
  "logProviderSelection": true
}
```

## 🎨 Document Testing Magic

The tool can test endpoints that accept files! Just add this to your config:

```json
{
  "request_schema": {
    "type": "object",
    "properties": {
      "query": {
        "type": "string",
        "description": "What to analyze"
      },
      "document": {
        "type": "string",
        "description": "Document to analyze (file upload)",
        "x-document-type": "csv",
        "x-document-description": "CSV with customer data: name, email, phone"
      }
    },
    "required": ["query"]
  }
}
```

**Supported file types:**
- 📊 **CSV** - Spreadsheet data
- 📄 **PDF** - Documents and reports  
- 🖼️ **Images** - PNG, JPG, etc.
- 📝 **Text** - Plain text files
- 📋 **JSON** - Structured data

## 🌟 Real Examples

### Example 1: Simple Chat API
```json
{
  "service": {
    "name": "Chat API",
    "base_url": "https://api.mychat.com"
  },
  "endpoints": {
    "single_evaluation": {
      "path": "/v1/chat",
      "method": "POST",
      "request_schema": {
        "type": "object",
        "properties": {
          "message": {
            "type": "string",
            "description": "User message"
          },
          "temperature": {
            "type": "number",
            "default": 0.7
          }
        },
        "required": ["message"]
      }
    }
  }
}
```

### Example 2: Document Analysis API
```json
{
  "service": {
    "name": "Document Analyzer",
    "base_url": "https://api.doc-analyzer.com"
  },
  "endpoints": {
    "single_evaluation": {
      "path": "/analyze",
      "method": "POST",
      "request_schema": {
        "type": "object",
        "properties": {
          "query": {
            "type": "string",
            "description": "What to extract from the document"
          },
          "document": {
            "type": "string",
            "description": "PDF document to analyze",
            "x-document-type": "pdf",
            "x-document-description": "Financial report with tables and charts"
          }
        },
        "required": ["query", "document"]
      }
    }
  }
}
```

## 🤖 AI Providers

The tool works with many AI providers:

| Provider | Setup | Best For |
|----------|-------|----------|
| 🟢 **OpenAI** | `export OPENAI_API_KEY="sk-..."` | Fast testing |
| 🟣 **Anthropic** | `export ANTHROPIC_API_KEY="..."` | High quality |
| ⚡ **Groq** | `export GROQ_API_KEY="..."` | Super fast |
| 🏠 **Ollama** | `ollama serve` | Local & private |
| ☁️ **AWS Bedrock** | AWS credentials | Enterprise |

### 🏠 Local Testing (No API Keys!)
```bash
# Install Ollama
ollama serve
ollama pull llama3.2

# Run tests (completely private!)
./datasnack endpointEval config/my-endpoint.json
```

## 📊 Understanding Results

After testing, check `results/endpoint_evaluation_results_TIMESTAMP.json`:

```json
{
  "totalCalls": 45,
  "successfulCalls": 42,
  "failedCalls": 3,
  "vulnerabilities": [
    {
      "type": "prompt_injection",
      "severity": "high",
      "description": "AI responded to malicious prompt",
      "score": 0.8
    }
  ],
  "recommendations": [
    "Add input validation",
    "Implement prompt injection detection"
  ]
}
```

## 🆘 Troubleshooting

### ❌ "No AI providers could be initialized"
**Fix:** Set an API key or use Ollama
```bash
export OPENAI_API_KEY="sk-your-key-here"
# OR
ollama serve
```

### ❌ "Endpoint health check failed"
**Fix:** Make sure your endpoint is running
```bash
curl http://localhost:8080/health
```

### ❌ "Config file not found"
**Fix:** Check the file path
```bash
ls config/
./datasnack endpointEval config/your-file.json
```

### ❌ "Document generation failed"
**Fix:** The tool now generates documents locally! No external service needed.

## 🎉 What Makes This Special?

- 🧠 **AI-Powered Testing** - Uses AI to create smart test cases
- 🔄 **Auto-Discovery** - Figures out your endpoint automatically  
- 📄 **Document Support** - Tests file uploads with real documents
- 🏠 **Local Testing** - Works completely offline with Ollama
- 🎯 **Zero Config** - Just point it at your code and go!
- 🛡️ **Comprehensive** - Tests all the security stuff you'd forget

## 🚀 Ready to Test?

1. **Build the tool** (2 minutes)
2. **Point it at your code** with `analyze` (1 minute)
3. **Run security tests** with `endpointEval` (5 minutes)
4. **Review results** and fix any issues! 🎯

**That's it!** Your AI endpoint is now security-tested and ready for production! 🎉

---

*Need help? Check the examples above or open an issue on GitHub!* 💬