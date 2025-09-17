#!/bin/bash

# AI Agent Evaluator CLI - Demo Script
# This script demonstrates how to use the AI Agent Evaluator CLI
# for comprehensive testing and evaluation of Python AI agents.

set -e  # Exit on any error

echo "============================================================"
echo "🤖 AI AGENT EVALUATOR CLI - DEMO"
echo "============================================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Function to print colored output
print_step() {
    echo -e "${BLUE}📋 $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

# Check if the CLI is built
if [ ! -f "./ai-evaluator" ]; then
    print_step "Building the AI Evaluator CLI..."
    go build -o ai-evaluator
    print_success "CLI built successfully!"
    echo ""
fi

# Display configuration
print_step "Current Configuration:"
echo "  Agent Script: $(jq -r '.agentScript' config/agentConfig.json)"
echo "  Python Path: $(jq -r '.pythonPath' config/agentConfig.json)"
echo "  Tracking Enabled: $(jq -r '.trackingEnabled' config/agentConfig.json)"
echo ""

# Check AI provider configuration
print_step "AI Provider Configuration:"
if [ -f "config/aiClientConfig.json" ]; then
    echo "  Preferred Providers:"
    jq -r '.preferredOrder[] | "    - \(.description)"' config/aiClientConfig.json
    echo "  Fallback to Bedrock: $(jq -r '.fallbackToBedrock' config/aiClientConfig.json)"
else
    print_warning "No AI client configuration found, using defaults"
fi
echo ""

# Check environment variables
print_step "Environment Variables:"
if [ -n "$OPENAI_API_KEY" ]; then
    print_success "OPENAI_API_KEY is set"
else
    print_warning "OPENAI_API_KEY not set"
fi

if [ -n "$ANTHROPIC_API_KEY" ]; then
    print_success "ANTHROPIC_API_KEY is set"
else
    print_warning "ANTHROPIC_API_KEY not set"
fi

if [ -n "$GROQ_API_KEY" ]; then
    print_success "GROQ_API_KEY is set"
else
    print_warning "GROQ_API_KEY not set"
fi

# Check if Ollama is running
if command -v ollama &> /dev/null; then
    if pgrep -f "ollama serve" > /dev/null; then
        print_success "Ollama is running (local AI provider available)"
    else
        print_warning "Ollama is installed but not running"
    fi
else
    print_warning "Ollama not installed (no local AI provider)"
fi
echo ""

# Display test scenarios
print_step "Test Scenarios:"
if [ -f "config/tests.json" ]; then
    jq -r '.allTests[] | "  - \(.)"' config/tests.json
else
    print_warning "No test scenarios configured"
fi
echo ""

# Run the evaluation
print_step "Starting AI Agent Evaluation..."
echo "This will test the agent with multiple scenarios and detect vulnerabilities."
echo ""

# Run the CLI with verbose output
./ai-evaluator evaluate

# Check if results were generated
if [ -d "results" ] && [ "$(ls -A results)" ]; then
    echo ""
    print_success "Evaluation completed successfully!"
    echo ""
    
    # Show the latest results file
    latest_result=$(ls -t results/evaluation_results_*.json | head -n1)
    if [ -n "$latest_result" ]; then
        print_step "Latest Results Summary:"
        echo "  Results file: $latest_result"
        
        # Extract key metrics
        total_calls=$(jq -r '.totalCalls' "$latest_result")
        successful_calls=$(jq -r '.successfulCalls' "$latest_result")
        failed_calls=$(jq -r '.failedCalls' "$latest_result")
        avg_response_time=$(jq -r '.averageResponseTime' "$latest_result")
        vulnerability_count=$(jq -r '.vulnerabilities | length' "$latest_result")
        
        echo "  Total API calls: $total_calls"
        echo "  Successful calls: $successful_calls"
        echo "  Failed calls: $failed_calls"
        echo "  Average response time: ${avg_response_time}ms"
        echo "  Vulnerabilities detected: $vulnerability_count"
        
        if [ "$vulnerability_count" -gt 0 ]; then
            echo ""
            print_warning "Vulnerabilities Found:"
            jq -r '.vulnerabilities[] | "  - \(.Type) (\(.Severity)): \(.Description)"' "$latest_result"
        fi
        
        echo ""
        print_info "For detailed results, check: $latest_result"
    fi
else
    print_error "No results generated. Check the logs above for errors."
fi

echo ""
echo "============================================================"
echo "🎯 Demo completed! The AI Agent Evaluator CLI provides:"
echo "  • Comprehensive vulnerability testing"
echo "  • Performance analysis"
echo "  • Prompt optimization suggestions"
echo "  • Multi-provider AI support"
echo "  • Detailed reporting and analytics"
echo "============================================================"
