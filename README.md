# DataSnack AI Agent Evaluator CLI

A comprehensive Go-based CLI tool for evaluating Python AI agents and n8n workflows with advanced testing, vulnerability detection, and intelligent prompt optimization capabilities.

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Configuration](#configuration)
- [Commands](#commands)
- [AI Provider Selection](#ai-provider-selection)
- [Examples](#examples)
- [Output and Results](#output-and-results)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)

## Overview

The DataSnack AI Agent Evaluator CLI is a powerful tool that:

- **Evaluates Python AI agents** with comprehensive HTTP endpoint testing
- **Tests n8n workflows** with automated webhook integration
- **Detects vulnerabilities** including prompt injection, data leakage, and security issues
- **Generates intelligent prompt suggestions** based on evaluation results
- **Supports multiple AI providers** (OpenAI, Anthropic, Groq, Ollama, AWS Bedrock)
- **Provides detailed analytics** and actionable recommendations
- **Uses dynamic schema-based payloads** for flexible agent integration

## Installation

### Prerequisites

- **Go 1.19+** installed on your system
- **Python 3.8+** for running AI agents
- **API keys** for your preferred AI providers (optional)
- **n8n instance** (for n8n workflow testing)

### Build from Source

1. **Clone the repository:**
```bash
git clone https://github.com/brianjmarvin/DataSnackOS-RISK.git
cd code-check-cli
```

2. **Install dependencies:**
```bash
go mod tidy
```

3. **Build the CLI:**
```bash
go build -o ai-evaluator
```

4. **Verify installation:**
```bash
./ai-evaluator --help
```

## Configuration

The CLI uses several configuration files to customize behavior:

### 1. AI Client Configuration (`config/aiClientConfig.json`)

This file defines which AI providers to use and in what order:

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
      "type": "groq",
      "model": "llama-3.1-70b-versatile",
      "envKey": "GROQ_API_KEY",
      "description": "Groq Llama - Ultra-fast inference"
    },
    {
      "provider": "awsbedrock",
      "type": "bedrock",
      "model": "anthropic.claude-3-5-sonnet-20240620-v2:0",
      "envKey": "AWS_REGION",
      "description": "AWS Bedrock Claude - Enterprise grade"
    }
  ],
  "fallbackToBedrock": true,
  "logProviderSelection": true
}
```

### 2. Agent Configuration (`config/agentConfig.json`)

Configure the Python AI agent to evaluate (includes both agent settings and test configuration):

```json
{
  "pythonPath": "/path/to/your/python/venv/bin/python",
  "agentScript": "/path/to/your/ai/agent/main.py",
  "agentRootFolder": "/path/to/your/ai/agent/root",
  "trackingEnabled": true,
  "agentPurpose": "The agent does research on the user's prompt and returns the results.",
  "testConfiguration": {
    "dataLeakageTests": 5,
    "promptInjectionTests": 5,
    "consistencyTests": 5,
    "iterationsPerTest": 3
  }
}
```

**Key Fields:**
- **`agentRootFolder`**: Path to the AI agent's root directory (used for finding evaluation configs)
- **`pythonPath`**: Python interpreter path for the agent
- **`agentScript`**: Main script file for the agent
- **`trackingEnabled`**: Enable/disable tracking features
- **`agentPurpose`**: Description of what the agent does (used for AI-generated test prompts)
- **`testConfiguration`**: Test parameters for evaluation

**Test Configuration Options:**
- **`dataLeakageTests`**: Number of AI-generated prompts to test for data leakage vulnerabilities
- **`promptInjectionTests`**: Number of AI-generated prompts to test for prompt injection attacks
- **`consistencyTests`**: Number of AI-generated prompts to test for response consistency
- **`iterationsPerTest`**: Number of times each test prompt is executed for reliability

## Commands

The CLI provides several commands for different evaluation scenarios:

### 1. `evaluate` - Python AI Agent Evaluation

Evaluates Python AI agents using HTTP endpoints with dynamic schema-based payloads.

```bash
./ai-evaluator evaluate
```

**Features:**
- **Dynamic Config Loading**: Automatically finds evaluation config in `{agentRootFolder}/backend/evaluation/config/evaluation_config.yaml`
- **Schema-Based Payloads**: Generates request payloads based on YAML schema definitions
- **Comprehensive Testing**: Data leakage, prompt injection, and consistency tests
- **AI-Powered Test Generation**: Uses AI to create sophisticated test prompts
- **Detailed Results**: Saves results to `results/evaluation_results_TIMESTAMP.json`

### 2. `evaluaten8n` - N8N Workflow Evaluation

Evaluates n8n workflows by adding webhook nodes and testing them programmatically.

```bash
./ai-evaluator evaluaten8n path/to/workflow.json
```

**Features:**
- **Automatic Webhook Integration**: Adds webhook trigger and response nodes
- **Workflow Testing**: Tests workflows with standardized request/response format
- **Vulnerability Detection**: Identifies security issues in workflow responses
- **Results Analysis**: Comprehensive evaluation of workflow behavior

### 3. `convert` - N8N Workflow Conversion

Converts n8n workflows to include webhook nodes for programmatic evaluation.

```bash
./ai-evaluator convert path/to/workflow.json
```

**Features:**
- **AI-Powered Conversion**: Uses AI to intelligently add webhook nodes
- **Manual Fallback**: Falls back to manual conversion if AI fails
- **Smart Node Detection**: Identifies final nodes and creates proper connections
- **Validation**: Ensures converted workflows have proper webhook integration

### 4. `suggestions` - Prompt Improvement Suggestions

Analyzes evaluation results and generates intelligent suggestions for improving AI agent prompts.

```bash
./ai-evaluator suggestions
```

**Features:**
- **Automatic Analysis**: Finds the most recent evaluation results
- **AI-Powered Suggestions**: Uses AI to generate specific prompt improvements
- **Vulnerability Mapping**: Maps vulnerabilities to specific prompts
- **Confidence Scoring**: Provides confidence levels for each suggestion
- **Detailed Reports**: Saves suggestions to `results/prompt_suggestions_TIMESTAMP.json`

## AI Provider Selection

The CLI automatically selects the best available AI provider based on:

1. **Configuration order** in `aiClientConfig.json`
2. **Available API keys** in environment variables
3. **Provider functionality** (tested before selection)
4. **Fallback mechanisms** (AWS Bedrock if enabled)

### Supported Providers

| Provider | Type | Environment Variable | Description |
|----------|------|---------------------|-------------|
| OpenAI | `openai` | `OPENAI_API_KEY` | Fast and cost-effective |
| Anthropic | `anthropic` | `ANTHROPIC_API_KEY` | High quality responses |
| Groq | `groq` | `GROQ_API_KEY` | Ultra-fast inference |
| Ollama | `ollama` | `OLLAMA_ENDPOINT` | Local models, complete privacy |
| AWS Bedrock | `awsbedrock` | `AWS_REGION` | Enterprise grade |

### Provider Selection Logging

When `logProviderSelection` is enabled, you'll see:
```
Trying AI provider 1/5: OpenAI GPT-4o-mini - Fast and cost-effective (openai)
Successfully initialized AI client: OpenAI GPT-4o-mini - Fast and cost-effective
```

## Examples

### Example 1: Python AI Agent Evaluation

```bash
# Configure for your AI agent
cat > config/agentConfig.json << EOF
{
  "pythonPath": "/path/to/your/venv/bin/python",
  "agentScript": "/path/to/your/agent/main.py",
  "agentRootFolder": "/path/to/your/agent/root",
  "trackingEnabled": true,
  "agentPurpose": "The agent does research on the user's prompt and returns the results.",
  "testConfiguration": {
    "dataLeakageTests": 5,
    "promptInjectionTests": 5,
    "consistencyTests": 5,
    "iterationsPerTest": 3
  }
}
EOF

# Set OpenAI API key
export OPENAI_API_KEY="sk-your-key"

# Run evaluation
./ai-evaluator evaluate
```

### Example 2: N8N Workflow Testing

```bash
# Convert workflow to include webhooks
./ai-evaluator convert n8n/my-workflow.json

# Evaluate the converted workflow
./ai-evaluator evaluaten8n n8n/my-workflow_eval.json
```

### Example 3: Generate Prompt Suggestions

```bash
# First run an evaluation
./ai-evaluator evaluate

# Then generate suggestions based on results
./ai-evaluator suggestions
```

### Example 4: Local Development with Ollama

```bash
# Start Ollama locally
ollama serve

# Configure for local model
cat > config/aiClientConfig.json << EOF
{
  "preferredOrder": [
    {
      "provider": "gollm",
      "type": "ollama",
      "model": "llama3.2",
      "envKey": "OLLAMA_ENDPOINT",
      "endpoint": "http://localhost:11434",
      "description": "Ollama Local - Complete privacy"
    }
  ],
  "fallbackToBedrock": false,
  "logProviderSelection": true
}
EOF

# Run evaluation (no API keys needed)
./ai-evaluator evaluate
```

## Output and Results

The evaluator generates comprehensive results including:

### Evaluation Results (`evaluation_results_TIMESTAMP.json`)

- **Test Summary**: Total calls, success/failure rates, average response time
- **Vulnerabilities**: Detailed analysis of security issues found
- **Performance Metrics**: Response times, execution statistics
- **Recommendations**: High-level guidance for improvements

### Prompt Suggestions (`prompt_suggestions_TIMESTAMP.json`)

- **Individual Suggestions**: Specific improvements for each prompt
- **Vulnerability Mapping**: Which vulnerabilities each suggestion addresses
- **Confidence Scores**: AI confidence in each suggestion
- **Impact Assessment**: Expected improvement from each change
- **Overall Recommendations**: Strategic guidance for implementation

### N8N Workflow Results

- **Webhook Integration**: Status of webhook node addition
- **Response Analysis**: Evaluation of workflow responses
- **Security Assessment**: Identification of potential vulnerabilities
- **Performance Metrics**: Response times and reliability

## AI Agent Integration

### Required Agent Structure

Your Python AI agent should implement HTTP endpoints as defined in the evaluation config:

**Example endpoint structure:**
```python
@app.post("/api/evaluation/evaluate")
async def evaluate_single(request: EvaluationRequest):
    # Process the request
    response = await your_ai_agent.process(request.query)
    
    # Return standardized response
    return {
        "success": True,
        "query": request.query,
        "response": response,
        "metrics": {
            "response_time": 1.5,
            "total_time": 2.0,
            "response_length": len(response),
            "word_count": len(response.split()),
            "has_content": bool(response),
            "source_count": 0,
            "has_citations": False
        },
        "agent_info": {
            "agent_type": "research",
            "report_type": "research_report",
            "report_source": "web"
        },
        "error": None
    }
```

### Configuration Files

The agent should provide two configuration files:

1. **`backend/evaluation/config/evaluation_config.yaml`**: API endpoint schemas
2. **`backend/evaluation/config/prompt_config.yaml`**: Prompt catalog and metadata

## Troubleshooting

### Common Issues

1. **No AI providers work:**
   - Check API keys are set correctly
   - Verify network connectivity
   - Check provider quotas and limits

2. **Python agent fails to run:**
   - Verify Python path in `agentConfig.json`
   - Check agent script path and agentRootFolder
   - Ensure all dependencies are installed
   - Verify the agent is running on the expected port
   - Check that agentPurpose and testConfiguration are properly set

3. **Evaluation config not found:**
   - Check that `agentRootFolder` points to the correct directory
   - Verify `backend/evaluation/config/evaluation_config.yaml` exists
   - The CLI will fall back to local config if not found

4. **N8N workflow issues:**
   - Ensure n8n server is running on localhost:5678
   - Import and activate the workflow in n8n interface
   - Check webhook paths are correctly configured

5. **Ollama not accessible:**
   - Ensure Ollama is running: `ollama serve`
   - Check endpoint in configuration
   - Verify model is pulled: `ollama list`

6. **AWS Bedrock fails:**
   - Check AWS credentials and region
   - Verify Bedrock access permissions
   - Ensure model is available in your region

### Debug Mode

Enable detailed logging by setting `"logProviderSelection": true` in `aiClientConfig.json`.

### Getting Help

1. **Check logs** for detailed error messages
2. **Verify configuration** files are valid JSON/YAML
3. **Test AI providers** individually
4. **Check environment variables** are set correctly
5. **Verify agent endpoints** are accessible

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the terms specified in the LICENSE file.

---

For more detailed information, see the individual configuration files and the `config/datasnack-instrumentation.md` for AI agent integration guidelines.