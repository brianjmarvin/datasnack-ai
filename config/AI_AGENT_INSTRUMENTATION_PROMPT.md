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
- **Schema-Driven**: Complete JSON schemas for request/response validation

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
    "source_count": int,
    "visited_url_count": int,
    "research_costs": object,
    "has_citations": boolean,
    "has_structure": boolean,
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
    "total_batch_time": float,
    "average_query_time": float
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
@router.get("/providers", response_model=List[ProviderInfo])
async def list_providers()

# Get metrics schema
@router.get("/metrics/schema", response_model=MetricsSchema)
async def get_metrics_schema()

# Get service configuration
@router.get("/config")
async def get_evaluation_config()
```

### 4. Request Models with Complete Schemas
```python
class EvaluationRequest(BaseModel):
    query: str = Field(..., description="The research query to evaluate")
    provider: str = Field(..., description="AI provider (e.g., 'openai', 'anthropic')")
    model: str = Field(..., description="Specific model to use (e.g., 'gpt-4-turbo')")
    temperature: float = Field(default=0.0, ge=0.0, le=2.0, description="Temperature for response generation")
    max_tokens: Optional[int] = Field(default=None, gt=0, description="Maximum tokens in response")
    report_type: str = Field(default="research_report", description="Type of report to generate")
    report_source: str = Field(default="web", description="Source for research (web, local, etc.)")
    tone: str = Field(default="objective", description="Tone of the response")
    headers: Optional[Dict[str, str]] = Field(default=None, description="Additional headers")
    config_path: Optional[str] = Field(default=None, description="Path to custom config file")
    reasoning_effort: Optional[str] = Field(default="medium", description="Reasoning effort level")
    timeout: Optional[int] = Field(default=300, gt=0, description="Request timeout in seconds")

class BatchEvaluationRequest(BaseModel):
    queries: List[str] = Field(..., min_items=1, description="List of queries to evaluate")
    provider: str = Field(..., description="AI provider to use")
    model: str = Field(..., description="Model to use")
    temperature: float = Field(default=0.0, ge=0.0, le=2.0, description="Temperature for responses")
    max_tokens: Optional[int] = Field(default=None, gt=0, description="Maximum tokens per response")
    report_type: str = Field(default="research_report", description="Type of report to generate")
    report_source: str = Field(default="web", description="Source for research")
    tone: str = Field(default="objective", description="Tone of responses")
    headers: Optional[Dict[str, str]] = Field(default=None, description="Additional headers")
    config_path: Optional[str] = Field(default=None, description="Path to custom config file")
    reasoning_effort: Optional[str] = Field(default="medium", description="Reasoning effort level")
    timeout: Optional[int] = Field(default=300, gt=0, description="Request timeout in seconds")
    parallel: bool = Field(default=False, description="Whether to run evaluations in parallel")
```

## Implementation Steps

### Step 1: Create Directory Structure
```
backend/
├── evaluation/
│   ├── __init__.py
│   ├── models.py          # Pydantic models with complete schemas
│   ├── service.py         # Core instrumentation logic
│   ├── endpoints.py       # FastAPI endpoints
│   ├── config/
│   │   └── evaluation_config.yaml  # Complete configuration with schemas
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

### Step 5: Configuration File with Complete Schemas
Create `backend/evaluation/config/evaluation_config.yaml` with complete JSON schemas:

```yaml
service:
  name: "AI Agent Instrumentation"
  version: "1.0.0"
  description: "Lightweight instrumentation for AI agent evaluation"

endpoints:
  health:
    path: "/api/evaluation/health"
    method: "GET"
    description: "Health check endpoint for evaluation service"
    response_schema:
      type: "object"
      properties:
        status:
          type: "string"
          description: "Service status"
          example: "healthy"
        version:
          type: "string"
          description: "Service version"
          example: "1.0.0"
        supported_providers:
          type: "array"
          items:
            type: "string"
          description: "List of supported AI providers"
        supported_models:
          type: "object"
          description: "Supported models per provider"
        timestamp:
          type: "string"
          format: "date-time"
          description: "Health check timestamp"
      required: ["status", "version", "supported_providers", "supported_models", "timestamp"]
    
  single_evaluation:
    path: "/api/evaluation/evaluate"
    method: "POST"
    description: "Evaluate a single AI query and return essential metrics"
    request_schema:
      type: "object"
      properties:
        query:
          type: "string"
          description: "The research query to evaluate"
        provider:
          type: "string"
          description: "AI provider (e.g., 'openai', 'anthropic')"
        model:
          type: "string"
          description: "Specific model to use (e.g., 'gpt-4-turbo')"
        temperature:
          type: "number"
          minimum: 0.0
          maximum: 2.0
          default: 0.0
          description: "Temperature for response generation"
        max_tokens:
          type: "integer"
          minimum: 1
          description: "Maximum tokens in response"
        report_type:
          type: "string"
          default: "research_report"
          description: "Type of report to generate"
          enum: ["research_report", "detailed_report", "deep_research", "basic_report"]
        report_source:
          type: "string"
          default: "web"
          description: "Source for research (web, local, etc.)"
          enum: ["web", "local", "hybrid"]
        tone:
          type: "string"
          default: "objective"
          description: "Tone of the response"
          enum: ["objective", "analytical", "casual", "formal"]
        headers:
          type: "object"
          additionalProperties:
            type: "string"
          description: "Additional headers"
        config_path:
          type: "string"
          description: "Path to custom config file"
        reasoning_effort:
          type: "string"
          default: "medium"
          description: "Reasoning effort level"
          enum: ["low", "medium", "high"]
        timeout:
          type: "integer"
          minimum: 1
          default: 300
          description: "Request timeout in seconds"
      required: ["query", "provider", "model"]
    response_schema:
      type: "object"
      properties:
        success:
          type: "boolean"
          description: "Whether the evaluation was successful"
        query:
          type: "string"
          description: "The original query"
        response:
          type: "string"
          description: "The AI agent's response"
        metrics:
          type: "object"
          properties:
            response_time:
              type: "number"
              description: "Time to generate AI response in seconds"
            total_time:
              type: "number"
              description: "Total evaluation time including overhead in seconds"
            response_length:
              type: "integer"
              description: "Character count of response"
            word_count:
              type: "integer"
              description: "Word count of response"
            character_count:
              type: "integer"
              description: "Character count of response"
            has_content:
              type: "boolean"
              description: "Boolean indicating if response has non-empty content"
            source_count:
              type: "integer"
              description: "Number of sources used in research"
            visited_url_count:
              type: "integer"
              description: "Number of URLs visited during research"
            research_costs:
              type: "object"
              description: "Cost breakdown for research operations"
            has_citations:
              type: "boolean"
              description: "Boolean indicating if response contains citations"
            has_structure:
              type: "boolean"
              description: "Boolean indicating if response has structural elements"
        provider_info:
          type: "object"
          properties:
            provider:
              type: "string"
              description: "AI provider used"
            model:
              type: "string"
              description: "Model used"
            temperature:
              type: "string"
              description: "Temperature setting"
            reasoning_effort:
              type: "string"
              description: "Reasoning effort level"
        timing:
          type: "object"
          properties:
            response_time:
              type: "number"
              description: "Response generation time"
            total_time:
              type: "number"
              description: "Total evaluation time"
        error:
          type: "string"
          nullable: true
          description: "Error message if evaluation failed"
        timestamp:
          type: "string"
          format: "date-time"
          description: "Evaluation timestamp"
      required: ["success", "query", "response", "metrics", "provider_info", "timing", "timestamp"]

cli_examples:
  health_check: "curl -X GET http://localhost:8000/api/evaluation/health"
  single_evaluation: "curl -X POST http://localhost:8000/api/evaluation/evaluate -H 'Content-Type: application/json' -d '{\"query\": \"test query\", \"provider\": \"openai\", \"model\": \"gpt-4o-mini\"}'"
  batch_evaluation: "curl -X POST http://localhost:8000/api/evaluation/evaluate/batch -H 'Content-Type: application/json' -d '{\"queries\": [\"query1\", \"query2\"], \"provider\": \"openai\", \"model\": \"gpt-4o-mini\"}'"

schema_validation:
  description: "Schema validation and CLI integration"
  cli_integration:
    description: "How CLI tools should use these schemas"
    steps:
      1: "Load the evaluation_config.yaml file"
      2: "Extract request schemas for each endpoint"
      3: "Validate CLI input against request schemas"
      4: "Generate proper HTTP requests with validated payloads"
      5: "Parse responses using response schemas"
      6: "Handle errors according to error schemas"
    
  schema_benefits:
    - "Type safety for CLI tools"
    - "Automatic validation of request parameters"
    - "Clear documentation of expected responses"
    - "Consistent error handling"
    - "Easy integration with code generation tools"
    
  example_cli_usage:
    description: "Example of how a CLI tool would use these schemas"
    pseudocode: |
      # Load configuration
      config = load_yaml("evaluation_config.yaml")
      
      # Get request schema for single evaluation
      request_schema = config["endpoints"]["single_evaluation"]["request_schema"]
      
      # Validate user input
      validated_request = validate_against_schema(user_input, request_schema)
      
      # Make HTTP request
      response = http_post("/api/evaluation/evaluate", validated_request)
      
      # Parse response using response schema
      parsed_response = parse_with_schema(response, response_schema)
      
      # Extract metrics
      metrics = parsed_response["metrics"]
      timing = parsed_response["timing"]
```

## Critical Success Factors

### 1. Pure Instrumentation
- **NEVER include evaluation logic** in the agent endpoints
- **ONLY collect basic metrics** (timing, content length, etc.)
- **ONLY call existing agent workflows** - don't create new logic
- **ONLY return response data** - let CLI handle evaluation

### 2. Schema-Driven Development
- **Include complete JSON schemas** for all request/response models
- **Validate all inputs** against schemas before processing
- **Document all fields** with descriptions and examples
- **Provide schema validation** for CLI tools

### 3. Model Configuration
- **Always use valid, available models** (e.g., gpt-4o-mini, not gpt-5)
- **Set environment variables dynamically** based on request parameters
- **Handle model fallbacks gracefully** if primary model fails

### 4. Error Handling
- **Catch all exceptions** and return structured error responses
- **Log errors appropriately** for debugging
- **Never crash the service** - always return a response

### 5. Testing Strategy
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

### 4. Schema Validation
- **Always validate inputs** against schemas
- **Handle validation errors gracefully**
- **Provide clear error messages** for invalid inputs
- **Test schema validation** with edge cases

## Documentation Requirements

### 1. API Documentation
- Complete endpoint documentation with examples
- Request/response schemas with field descriptions
- Error codes and messages
- Rate limiting information

### 2. CLI Integration Guide
- How to call endpoints from CLI tools
- Schema validation examples
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

### 5. Schema Reference
- Complete JSON schema documentation
- Field descriptions and examples
- Validation rules and constraints
- CLI integration examples

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
    "parallel": true
  }'
```

### Schema Validation Example
```python
import yaml
import jsonschema
from jsonschema import validate

# Load configuration
with open('evaluation_config.yaml', 'r') as f:
    config = yaml.safe_load(f)

# Get request schema
request_schema = config['endpoints']['single_evaluation']['request_schema']

# Validate user input
user_input = {
    "query": "What is AI?",
    "provider": "openai",
    "model": "gpt-4o-mini",
    "temperature": 0.0
}

try:
    validate(instance=user_input, schema=request_schema)
    print("Input is valid")
except jsonschema.exceptions.ValidationError as e:
    print(f"Validation error: {e.message}")
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
9. ✅ **Schema validation works** (CLI tools can validate inputs)
10. ✅ **Complete documentation** (schemas, examples, error handling)

## Final Notes

This prompt is designed to be **generic and reusable** for any Python AI agent. The key is to:
- **Focus on instrumentation, not evaluation**
- **Use existing agent workflows**
- **Handle errors gracefully**
- **Provide clear, actionable metrics**
- **Make it easy for CLI tools to integrate**
- **Include complete schemas for validation**

The instrumentation should be **invisible to the AI agent** - it should work exactly as it normally does, but with additional metrics collection for evaluation purposes. The schemas ensure that CLI tools can properly validate inputs and parse responses, making the integration robust and reliable.