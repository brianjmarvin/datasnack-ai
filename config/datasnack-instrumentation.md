# DataSnack AI Agent Instrumentation

## Overview
Lightweight, provider-agnostic instrumentation endpoints for Python AI agents. Creates endpoints for evaluation CLIs to test AI workflows without including evaluation logic in the agent. Maps and catalogs all AI prompts to generate configuration files.

## Core Requirements

### 1. Objectives
- **Prompt Discovery**: Find and catalog all AI prompts in codebase
- **Configuration Generation**: Generate two YAML config files
- **Pure Instrumentation**: Only collect metrics, no evaluation logic
- **Schema-Driven**: Complete JSON schemas for CLI integration

### 2. Generated Files

**`backend/evaluation/config/evaluation_config.yaml`**
- API endpoint schemas with request/response definitions
- CLI integration examples and validation info

**`backend/evaluation/config/prompt_config.yaml`**
- Discovered prompts catalog with metadata
- Prompt categorization by agent type

### 3. Response Schema
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
    "has_content": boolean,
    "source_count": int,
    "has_citations": boolean
  },
  "agent_info": {
    "agent_type": string,
    "report_type": string,
    "report_source": string
  },
  "error": string | null
}
```

## Implementation

### 1. Directory Structure
```
backend/evaluation/
├── __init__.py
├── models.py          # Pydantic models
├── service.py         # Core instrumentation
├── endpoints.py       # FastAPI endpoints
├── prompt_discovery.py # Prompt discovery
├── config/
│   ├── evaluation_config.yaml
│   └── prompt_config.yaml
└── README.md
```

### 2. Prompt Discovery
Key patterns to detect:
```python
prompt_patterns = [
    {'pattern': r'TEMPLATE\s*=\s*["\'](.*?)["\']', 'agent_type': 'template'},
    {'pattern': r'system_prompt\s*=\s*["\'](.*?)["\']', 'agent_type': 'system'},
    {'pattern': r'prompt\s*=\s*["\'](.*?)["\']', 'agent_type': 'general'},
    {'pattern': r'You are a.*?\.(.*?)(?=\n\n|\n[A-Z]|$)', 'agent_type': 'system', 'multiline': True}
]
```

Core methods:
- `discover_all_prompts()` - Scan codebase for prompts
- `generate_prompt_config()` - Create prompt catalog YAML
- `generate_evaluation_config()` - Create API schema YAML

### 3. Core Service
```python
class AIEvaluationService:
    async def evaluate_single(self, request: EvaluationRequest) -> EvaluationResponse:
        # CRITICAL: Only call existing AI agent workflow, collect metrics
        # NO evaluation logic should be included
        
    def _calculate_basic_metrics(self, query: str, response: str, total_time: float, **kwargs):
        # Calculate: response_length, word_count, has_content, source_count, etc.
```

### 4. FastAPI Endpoints
```python
@router.get("/health", response_model=HealthCheckResponse)
@router.post("/evaluate", response_model=EvaluationResponse)  
@router.get("/metrics/schema", response_model=MetricsSchema)
@router.get("/config")
@router.get("/prompts")  # Get discovered prompts
@router.get("/capabilities")  # Get agent capabilities
```

### 5. Pydantic Models
```python
class EvaluationRequest(BaseModel):
    query: str = Field(..., description="The research query to evaluate")
    report_type: str = Field(default="research_report", pattern="^(research_report|detailed_report|deep_research|basic_report)$")
    report_source: str = Field(default="web", pattern="^(web|local|hybrid)$")
    tone: str = Field(default="objective", pattern="^(objective|analytical|casual|formal)$")
    timeout: Optional[int] = Field(default=300, gt=0)
```

### 6. Configuration Files

**evaluation_config.yaml** - API schemas and CLI examples:
```yaml
service:
  name: "GPT-Researcher AI Evaluation Service"
  version: "1.0.0"
  base_url: "http://localhost:8000"

endpoints:
  health:
    path: "/api/evaluation/health"
    method: "GET"
    response_schema: {...}  # Complete JSON schema
    
  single_evaluation:
    path: "/api/evaluation/evaluate"
    method: "POST"
    request_schema: {...}   # Complete request schema
    response_schema: {...}  # Complete response schema

cli_examples:
  health_check: "curl -X GET http://localhost:8000/api/evaluation/health"
  single_evaluation: "curl -X POST http://localhost:8000/api/evaluation/evaluate -d '{\"query\": \"test\"}'"
```

**prompt_config.yaml** - Discovered prompts catalog:
```yaml
version: "1.0.0"
last_updated: "2024-01-01T00:00:00Z"

original_prompts:
  reviewer_template:
    prompt: "You are an expert research article reviewer..."
    location: "multi_agents/agents/reviewer.py:1"
    agent_type: "reviewer"
    description: "System prompt for the reviewer agent"
    
  writer_system_prompt:
    prompt: "You are a research writer..."
    location: "multi_agents/agents/writer.py:1"
    agent_type: "writer"
    description: "System prompt for the writer agent"

prompt_categories:
  multi_agent:
    description: "Prompts used in the multi-agent system"
    prompts: ["reviewer_template", "writer_system_prompt"]
    
  core_research:
    description: "Core research and query processing prompts"
    prompts: []

discovery_metadata:
  total_prompts_discovered: 8
  discovery_timestamp: "2024-01-01T00:00:00Z"
  codebase_paths_scanned: ["multi_agents/agents/", "gpt_researcher/"]
```

### 7. Integration

**Agent Workflow Integration:**
```python
# Call existing AI agent workflow - DO NOT create new logic
from gpt_researcher import GPTResearcher

researcher = GPTResearcher(
    query=request.query,
    report_type=request.report_type,
    report_source=request.report_source
)
result = await researcher.conduct_research()
```

**Error Handling:**
```python
try:
    result = await run_agent(...)
    return EvaluationResponse(success=True, ...)
except Exception as e:
    return EvaluationResponse(success=False, error=str(e), ...)
```

**Main App Integration:**
```python
# In backend/server/server.py
from backend.evaluation.endpoints import router as evaluation_router
app.include_router(evaluation_router)
```

## Key Principles

### 1. Pure Instrumentation
- **NEVER include evaluation logic** in agent endpoints
- **ONLY collect metrics** (timing, content length, etc.)
- **ONLY call existing agent workflows** - don't create new logic

### 2. Schema-Driven
- **Complete JSON schemas** for all request/response models
- **Validate all inputs** against schemas
- **Provide schema validation** for CLI tools

### 3. Error Handling
- **Catch all exceptions** and return structured error responses
- **Never crash the service** - always return a response

## Testing

```bash
# Generate config files
python backend/evaluation/prompt_discovery.py

# Test endpoints
curl -X GET "http://localhost:8000/api/evaluation/health"
curl -X POST "http://localhost:8000/api/evaluation/evaluate" \
  -H "Content-Type: application/json" \
  -d '{"query": "What is 2+2?", "report_type": "basic_report"}'
```

## Success Criteria

✅ **Prompt discovery finds all AI prompts** in codebase  
✅ **Configuration files generated** with complete schemas  
✅ **Real AI responses** (not mock responses)  
✅ **All metrics collected accurately** (timing, content length, etc.)  
✅ **Error handling works gracefully** (structured error responses)  
✅ **CLI integration straightforward** (clear API, good documentation)  
✅ **Schema validation works** (CLI tools can validate inputs)  

## Implementation Notes

- **Focus on discovery and cataloging** - not dynamic modification
- **Use existing agent workflows** - don't create new logic
- **Handle errors gracefully** - always return structured responses
- **Make CLI integration easy** - complete schemas and examples

The instrumentation should be **invisible to the AI agent** - it works exactly as normal but with additional metrics collection for evaluation purposes. :-)
