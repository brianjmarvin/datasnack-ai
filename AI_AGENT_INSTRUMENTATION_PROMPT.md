# Generic AI Agent Instrumentation Prompt

## Overview
This prompt provides a complete blueprint for implementing lightweight, provider-agnostic instrumentation endpoints for any Python AI agent or framework. The goal is to create endpoints that allow evaluation CLIs to programmatically test AI agent workflows without including any evaluation logic in the agent itself.

## Core Requirements

### 1. Endpoint Design Principles
- **Minimal Response Payload**: Return only essential metrics and response data
- **Simplified Integration**: Easy for CLI tools to consume
- **Comprehensive Error Handling**: Graceful failure modes
- **Provider Agnostic**: Support multiple AI models (OpenAI, Anthropic, Llama, etc.)
- **No Evaluation Logic**: Pure instrumentation - evaluation handled by calling CLI

### 2. Essential Response Payload
**Single Query Response:**
```json
{
  "success": boolean,
  "query": string,
  "response": string,
  "metrics": {
    "response_time": float,
    "total_time": float,
    "response_length": int,
    "word_count": int,
    "character_count": int,
    "has_content": boolean,
    "timestamp": string
  },
  "provider_info": {
    "provider": string,
    "model": string,
    "temperature": string,
    "reasoning_effort": string
  },
  "timing": {
    "response_time": float,
    "total_time": float
  },
  "error": string | null
}
```

**Batch Query Response:**
```json
{
  "success": boolean,
  "total_queries": int,
  "successful_queries": int,
  "failed_queries": int,
  "success_rate": float,
  "results": [EvaluationResponse],
  "aggregate_metrics": {
    "average_response_time": float,
    "average_response_length": float,
    "average_word_count": float,
    "average_character_count": float,
    "total_response_length": int,
    "total_word_count": int,
    "total_character_count": int
  },
  "timing": {
    "total_time": float,
    "average_time_per_query": float
  },
  "error": string | null
}
```

### 3. Required FastAPI Endpoints
Create the following endpoints in `backend/evaluation/endpoints.py`:

```python
# Health check endpoint
@router.get("/health", response_model=HealthCheckResponse)
async def health_check()

# Single query instrumentation
@router.post("/evaluate", response_model=EvaluationResponse)
async def evaluate_single(request: EvaluationRequest)

# Batch query instrumentation
@router.post("/evaluate/batch", response_model=BatchEvaluationResponse)
async def evaluate_batch(request: BatchEvaluationRequest)

# List supported providers
@router.get("/providers")
async def list_providers()

# Get metrics schema
@router.get("/metrics/schema")
async def get_metrics_schema()
```

### 4. Request Models
```python
class EvaluationRequest(BaseModel):
    query: str
    provider: str = "openai"
    model: str = "gpt-4o-mini"
    temperature: float = 0.0
    reasoning_effort: Optional[str] = "medium"
    report_type: Optional[str] = "research_report"
    report_source: Optional[str] = "web"
    tone: Optional[str] = "objective"
    headers: Optional[Dict[str, str]] = None
    config_path: Optional[str] = None

class BatchEvaluationRequest(BaseModel):
    queries: List[str]
    provider: str = "openai"
    model: str = "gpt-4o-mini"
    temperature: float = 0.0
    reasoning_effort: Optional[str] = "medium"
    report_type: Optional[str] = "research_report"
    report_source: Optional[str] = "web"
    tone: Optional[str] = "objective"
    headers: Optional[Dict[str, str]] = None
    config_path: Optional[str] = None
    parallel: bool = True
    max_concurrent: int = 5
```

## Implementation Steps

### Step 1: Create Directory Structure
```
backend/
├── evaluation/
│   ├── __init__.py
│   ├── models.py          # Pydantic models
│   ├── service.py         # Core instrumentation logic
│   ├── endpoints.py       # FastAPI endpoints
│   ├── config/
│   │   └── evaluation_config.yaml
│   └── README.md
```

### Step 2: Implement Core Service
Create `backend/evaluation/service.py` with the following key methods:

```python
class AIEvaluationService:
    async def evaluate_single(self, request: EvaluationRequest) -> EvaluationResponse:
        """
        Instrument a single AI query and return response with basic metrics.
        CRITICAL: This should ONLY call the existing AI agent workflow and collect metrics.
        NO evaluation logic should be included.
        """
        
    async def evaluate_batch(self, request: BatchEvaluationRequest) -> BatchEvaluationResponse:
        """
        Instrument multiple queries and return aggregated results.
        """
        
    def _calculate_basic_metrics(self, query: str, response: str, total_time: float, **kwargs) -> Dict[str, Union[float, int, str]]:
        """
        Calculate only basic metrics: response_length, word_count, character_count, has_content
        """
        
    def _calculate_aggregate_metrics(self, results: List[EvaluationResponse]) -> Dict[str, Union[float, int]]:
        """
        Calculate aggregate metrics for batch results
        """
```

### Step 3: Key Implementation Guidelines

#### 3.1 Environment Variable Management
```python
# Set environment variables for the specific model and provider
import os
os.environ["SMART_LLM"] = f"{request.provider}:{request.model}"
os.environ["FAST_LLM"] = f"{request.provider}:{request.model}"
```

#### 3.2 Agent Workflow Integration
```python
# Call the existing AI agent workflow - DO NOT create new logic
from backend.server.websocket_manager import run_agent  # or equivalent

result = await run_agent(
    task=request.query,
    report_type=request.report_type,
    report_source=request.report_source,
    # ... other parameters
    return_researcher=True  # If available for additional metrics
)
```

#### 3.3 Error Handling
```python
try:
    # Agent execution
    result = await run_agent(...)
    return EvaluationResponse(success=True, ...)
except Exception as e:
    logger.error(f"Agent execution failed: {str(e)}", exc_info=True)
    return EvaluationResponse(
        success=False,
        error=str(e),
        # ... other fields with default values
    )
```

### Step 4: Integration with Main Application
```python
# In backend/server/server.py or main.py
from backend.evaluation.endpoints import router as evaluation_router
app.include_router(evaluation_router, prefix="/api/evaluation")
```

### Step 5: Configuration File
Create `backend/evaluation/config/evaluation_config.yaml`:
```yaml
service:
  name: "AI Agent Instrumentation"
  version: "1.0.0"
  description: "Lightweight instrumentation for AI agent evaluation"

endpoints:
  health: "/api/evaluation/health"
  evaluate: "/api/evaluation/evaluate"
  batch_evaluate: "/api/evaluation/evaluate/batch"
  providers: "/api/evaluation/providers"
  metrics_schema: "/api/evaluation/metrics/schema"

cli_examples:
  health_check: "curl -X GET http://localhost:8000/api/evaluation/health"
  single_evaluation: "curl -X POST http://localhost:8000/api/evaluation/evaluate -H 'Content-Type: application/json' -d '{\"query\": \"test query\", \"provider\": \"openai\", \"model\": \"gpt-4o-mini\"}'"
  batch_evaluation: "curl -X POST http://localhost:8000/api/evaluation/evaluate/batch -H 'Content-Type: application/json' -d '{\"queries\": [\"query1\", \"query2\"], \"provider\": \"openai\", \"model\": \"gpt-4o-mini\"}'"
```

## Critical Success Factors

### 1. Pure Instrumentation
- **NEVER include evaluation logic** in the agent endpoints
- **ONLY collect basic metrics** (timing, content length, etc.)
- **ONLY call existing agent workflows** - don't create new logic
- **ONLY return response data** - let CLI handle evaluation

### 2. Model Configuration
- **Always use valid, available models** (e.g., gpt-4o-mini, not gpt-5)
- **Set environment variables dynamically** based on request parameters
- **Handle model fallbacks gracefully** if primary model fails

### 3. Error Handling
- **Catch all exceptions** and return structured error responses
- **Log errors appropriately** for debugging
- **Never crash the service** - always return a response

### 4. Testing Strategy
```bash
# Test simple queries first
curl -X POST "http://localhost:8000/api/evaluation/evaluate" \
  -H "Content-Type: application/json" \
  -d '{"query": "What is 2+2?", "provider": "openai", "model": "gpt-4o-mini"}'

# Test complex queries
curl -X POST "http://localhost:8000/api/evaluation/evaluate" \
  -H "Content-Type: application/json" \
  -d '{"query": "How does AI work?", "provider": "openai", "model": "gpt-4o-mini", "report_type": "research_report"}'
```

## Common Pitfalls to Avoid

### 1. Mock Responses
- **Check model configuration** - ensure valid models are used
- **Verify API keys** are loaded correctly
- **Test with simple queries first** to isolate issues
- **Check server logs** for error messages

### 2. Environment Issues
- **Load environment variables** in the main application entry point
- **Use absolute paths** when possible
- **Check Python path** and module imports
- **Verify dependencies** are installed

### 3. Agent Integration
- **Use existing agent workflows** - don't reinvent the wheel
- **Pass parameters correctly** to the agent
- **Handle different response formats** from the agent
- **Extract metrics from agent objects** when available

## Documentation Requirements

### 1. API Documentation
- Complete endpoint documentation with examples
- Request/response schemas
- Error codes and messages
- Rate limiting information

### 2. CLI Integration Guide
- How to call endpoints from CLI tools
- Authentication requirements
- Batch processing examples
- Error handling strategies

### 3. Provider Setup Guide
- How to configure different AI providers
- API key management
- Model selection guidelines
- Cost optimization tips

### 4. Metrics Explanation
- What each metric represents
- How to interpret results
- Performance benchmarks
- Troubleshooting guide

## Example Usage

### Health Check
```bash
curl -X GET "http://localhost:8000/api/evaluation/health"
```

### Single Evaluation
```bash
curl -X POST "http://localhost:8000/api/evaluation/evaluate" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "Explain quantum computing",
    "provider": "openai",
    "model": "gpt-4o-mini",
    "temperature": 0.0,
    "report_type": "research_report",
    "report_source": "web",
    "tone": "objective"
  }'
```

### Batch Evaluation
```bash
curl -X POST "http://localhost:8000/api/evaluation/evaluate/batch" \
  -H "Content-Type: application/json" \
  -d '{
    "queries": [
      "What is machine learning?",
      "How does neural networks work?",
      "Explain deep learning"
    ],
    "provider": "openai",
    "model": "gpt-4o-mini",
    "temperature": 0.0,
    "parallel": true,
    "max_concurrent": 3
  }'
```

## Success Criteria

The instrumentation is successful when:
1. ✅ **Simple queries return real AI responses** (not mock responses)
2. ✅ **Complex queries return real research reports** (not mock responses)
3. ✅ **All metrics are collected accurately** (timing, content length, etc.)
4. ✅ **Error handling works gracefully** (returns structured error responses)
5. ✅ **Batch processing works efficiently** (parallel execution when requested)
6. ✅ **CLI integration is straightforward** (clear API, good documentation)
7. ✅ **Provider switching works** (different models/providers can be used)
8. ✅ **Performance is acceptable** (reasonable response times)

## Final Notes

This prompt is designed to be **generic and reusable** for any Python AI agent. The key is to:
- **Focus on instrumentation, not evaluation**
- **Use existing agent workflows**
- **Handle errors gracefully**
- **Provide clear, actionable metrics**
- **Make it easy for CLI tools to integrate**

The instrumentation should be **invisible to the AI agent** - it should work exactly as it normally does, but with additional metrics collection for evaluation purposes.
