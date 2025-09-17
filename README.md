# AI Agent Evaluator CLI

A comprehensive Go-based CLI tool for evaluating Python AI agents with advanced testing, vulnerability detection, and performance optimization capabilities.

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Configuration](#configuration)
- [AI Agent Modification](#ai-agent-modification)
- [Usage](#usage)
- [AI Provider Selection](#ai-provider-selection)
- [Examples](#examples)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)

## Overview

The AI Agent Evaluator CLI is a powerful tool that:

- **Evaluates Python AI agents** with comprehensive testing scenarios
- **Detects vulnerabilities** including prompt injection, information leakage, and system prompt exposure
- **Optimizes prompts** based on performance results
- **Supports multiple AI providers** (OpenAI, Anthropic, Groq, Ollama, AWS Bedrock)
- **Provides detailed analytics** and recommendations
- **Runs stress tests** to assess agent performance under load

## Installation

### Prerequisites

- **Go 1.19+** installed on your system
- **Python 3.8+** for running AI agents
- **API keys** for your preferred AI providers (optional)

### Build from Source

1. **Clone the repository:**
```bash
git clone https://github.com/brianjmarvin/datasnack-ai.git
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

Configure the Python AI agent to evaluate:

```json
{
  "pythonPath": "/path/to/your/python/venv/bin/python",
  "agentScript": "/path/to/your/ai/agent/main.py",
  "trackingEnabled": true
}
```

### 3. Agent Details (`config/agentDetails.json`)

Define the agent's purpose and test configuration:

```json
{
  "agentPurpose": "The agent does research on the user's prompt and returns the results.",
  "testConfiguration": {
    "dataLeakageTests": 5,
    "promptInjectionTests": 5,
    "consistencyTests": 5,
    "iterationsPerTest": 3
  }
}
```

**Test Configuration Options:**
- **`dataLeakageTests`**: Number of AI-generated prompts to test for data leakage vulnerabilities
- **`promptInjectionTests`**: Number of AI-generated prompts to test for prompt injection attacks
- **`consistencyTests`**: Number of AI-generated prompts to test for response consistency
- **`iterationsPerTest`**: Number of times each test prompt is executed for reliability

## AI Agent Modification

To make your Python AI agent compatible with the evaluator, you need to modify it using the comprehensive framework described in `cursor_eval_prompt.txt`.

### Step 1: Understanding the Framework

The `cursor_eval_prompt.txt` file contains a detailed prompt for creating an AI call tracking and evaluation framework. Key components include:

- **Universal AI Call Detection**: Automatic detection of AI API calls
- **Comprehensive Metadata Tracking**: UUID identifiers, timestamps, provider info
- **Multi-modal Support**: Text, image, document, audio, video
- **Provider Agnostic**: Works with OpenAI, Anthropic, Google, Azure, etc.
- **Performance Monitoring**: Execution time, token usage, cost estimates

### Step 2: Implementing the Framework

1. **Read the prompt file:**
```bash
cat cursor_eval_prompt.txt
```

2. **Use the prompt with Cursor AI:**
   - Open Cursor AI
   - Paste the contents of `cursor_eval_prompt.txt`
   - Apply it to your Python AI agent code
   - The AI will help you implement the tracking framework

3. **Key implementation elements:**
```python
@ai_call_tracker(
    provider='openai',
    input_type='text',
    output_type='text',
    log_level='detailed',
    session_id='research_session_1',
    user_id='researcher_123',
    tags=['analysis', 'gpt-4'],
    custom_metadata={'project': 'ai_research'}
)
async def analyze_document(document_content: str) -> str:
    # Your AI call logic here
    pass
```

### Step 3: Integration Points

The framework should capture:
- **Input/Output data** with type classification
- **Execution metrics** (time, tokens, cost)
- **Error handling** and success status
- **Session context** and user information
- **Custom metadata** for project-specific needs

### Step 4: Testing Your Instrumented Agent

After implementing the tracking framework, you can test your agent using the generic tester:

```bash
# Test your agent directly
python examples/generic_agent_tester.py /path/to/your/agent.py "Test prompt here"

# Example with a research agent
python examples/generic_agent_tester.py /path/to/research_agent.py "What are the latest AI developments?"
```

The generic tester will:
- **Automatically detect** your agent's main function
- **Handle both sync and async** agent functions
- **Provide detailed output** and error handling
- **Work with any instrumented agent** regardless of implementation

## Usage

### Basic Evaluation

1. **Set up your AI agent** using the framework from `cursor_eval_prompt.txt`

2. **Configure the evaluator** by updating the config files

3. **Set environment variables** for your preferred AI provider:
```bash
export OPENAI_API_KEY="sk-your-openai-key"
# or
export GROQ_API_KEY="gsk-your-groq-key"
# or
export ANTHROPIC_API_KEY="sk-ant-your-anthropic-key"
```

4. **Run the evaluation:**
```bash
./ai-evaluator evaluate
```

### Advanced Usage

#### Custom Configuration Paths
```bash
export AGENT_CONFIG="path/to/your/agentConfig.json"
export TESTS_FILE="path/to/your/tests.json"
export AGENT_DETAILS="path/to/your/agentDetails.json"
export AI_CLIENT_CONFIG="path/to/your/aiClientConfig.json"
./ai-evaluator evaluate
```

#### Local Development with Ollama
```bash
# Start Ollama locally
ollama serve

# Run evaluation (no API keys needed)
./ai-evaluator evaluate
```

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

### Example 1: GPT Researcher Agent

```bash
# Configure for GPT Researcher
cat > config/agentConfig.json << EOF
{
  "pythonPath": "/path/to/gpt-researcher/venv/bin/python",
  "agentScript": "/path/to/gpt-researcher/main.py",
  "trackingEnabled": true
}
EOF

# Set OpenAI API key
export OPENAI_API_KEY="sk-your-key"

# Run evaluation
./ai-evaluator evaluate
```

### Example 2: Custom AI Agent

```bash
# Configure for your custom agent
cat > config/agentConfig.json << EOF
{
  "pythonPath": "/path/to/your/venv/bin/python",
  "agentScript": "/path/to/your/agent.py",
  "trackingEnabled": true
}
EOF

# Set preferred provider
export GROQ_API_KEY="gsk-your-key"

# Run evaluation
./ai-evaluator evaluate
```

### Example 3: Local Development

```bash
# Start Ollama
ollama serve
ollama pull llama3.2

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

# Run evaluation
./ai-evaluator evaluate
```

### Example 4: Testing Individual Agents

```bash
# Test the included sample agent
python3 examples/generic_agent_tester.py examples/sample_agent.py "Hello, test prompt!"

# Test a research agent (like GPT Researcher)
python3 examples/universal_agent_tester.py /path/to/research_agent.py "What is machine learning?"

# Test a chatbot agent
python3 examples/universal_agent_tester.py /path/to/chatbot.py "Hello, how are you?"

# Test an analysis agent
python3 examples/universal_agent_tester.py /path/to/analyzer.py "Analyze this data: [your data here]"
```

### Example 5: Universal Agent Tester

The `examples/universal_agent_tester.py` is the most comprehensive testing tool that can handle:

- **Function-based agents** (main, run, execute, etc.)
- **Class-based agents** with methods
- **Server-based agents** (like GPT Researcher)
- **Agents with virtual environments**
- **Automatic environment detection**

```bash
# Test any agent with automatic environment detection
python3 examples/universal_agent_tester.py /path/to/any/agent.py "Test prompt"

# Test with specific Python interpreter
python3 examples/universal_agent_tester.py /path/to/agent.py "Test prompt" /path/to/venv/bin/python
```

### Example 6: Sample Agent Structure

The `examples/sample_agent.py` file shows how to structure an AI agent that works with the evaluator:

```python
def main(user_prompt: str) -> str:
    """Main agent function - the evaluator will find this automatically"""
    # Your AI agent logic here
    return "Agent response"

# Alternative function names that also work:
def run(prompt: str) -> str:
    return main(prompt)

async def async_agent(prompt: str) -> str:
    """Async functions are also supported"""
    return await some_async_ai_call(prompt)
```

## Output and Results

The evaluator generates comprehensive results including:

### AI-Generated Test Prompts
- **Data Leakage Tests**: Automatically generated prompts designed to test for sensitive information exposure
- **Prompt Injection Tests**: AI-created prompts that attempt to override system instructions
- **Consistency Tests**: Generated prompts to test response consistency across different phrasings

### Performance Metrics
- **Response times** (min, max, average)
- **Success rates** and failure counts
- **Total execution time**
- **Test coverage** across different vulnerability types

### Vulnerability Analysis
- **Data leakage** detection and scoring
- **Prompt injection** resistance testing
- **Consistency** analysis across test variations
- **Security scores** and detailed recommendations

### Optimization Recommendations
- **Prompt improvements** based on AI-generated test results
- **Performance optimizations** for better response times
- **Security enhancements** to address detected vulnerabilities
- **Best practices** suggestions for production deployment

### Results File
Results are saved to `results/evaluation_results_TIMESTAMP.json` with detailed analysis, AI-generated test prompts, and comprehensive recommendations.

## Troubleshooting

### Common Issues

1. **No AI providers work:**
   - Check API keys are set correctly
   - Verify network connectivity
   - Check provider quotas and limits

2. **Python agent fails to run:**
   - Verify Python path in `agentConfig.json`
   - Check agent script path
   - Ensure all dependencies are installed

3. **Ollama not accessible:**
   - Ensure Ollama is running: `ollama serve`
   - Check endpoint in configuration
   - Verify model is pulled: `ollama list`

4. **AWS Bedrock fails:**
   - Check AWS credentials and region
   - Verify Bedrock access permissions
   - Ensure model is available in your region

### Debug Mode

Enable detailed logging by setting `"logProviderSelection": true` in `aiClientConfig.json`.

### Getting Help

1. **Check logs** for detailed error messages
2. **Verify configuration** files are valid JSON
3. **Test AI providers** individually
4. **Check environment variables** are set correctly

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the terms specified in the LICENSE file.

---

For more detailed information, see the individual configuration files and the `cursor_eval_prompt.txt` for AI agent modification guidelines.
# datasnack-ai
# datasnack-ai
