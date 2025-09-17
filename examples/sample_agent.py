#!/usr/bin/env python3
"""
Sample AI Agent for Testing

This is a simple example agent that demonstrates how to structure an AI agent
that can be tested with the generic agent tester and evaluated by the CLI.

To use this agent:
1. Install required dependencies: pip install openai
2. Set your OpenAI API key: export OPENAI_API_KEY="your-key"
3. Test with: python examples/generic_agent_tester.py examples/sample_agent.py "Hello!"
4. Evaluate with: ./ai-evaluator evaluate (after configuring agentConfig.json)
"""

import os
import asyncio
from typing import Optional

# Example of how to structure an AI agent function
def main(user_prompt: str) -> str:
    """
    Main agent function that processes user prompts.
    
    This function should be the entry point for your AI agent.
    It takes a user prompt and returns a response.
    
    Args:
        user_prompt (str): The user's input prompt
        
    Returns:
        str: The agent's response
    """
    # Simple example - in a real agent, this would call an AI API
    response = f"Agent received: '{user_prompt}'\n"
    response += "This is a sample response from the AI agent.\n"
    response += "In a real implementation, this would call OpenAI, Anthropic, etc."
    
    return response

# Alternative function names that the generic tester will also detect
def run(prompt: str) -> str:
    """Alternative function name for the agent"""
    return main(prompt)

def process(query: str) -> str:
    """Another alternative function name"""
    return main(query)

# Example of an async agent function
async def async_agent(prompt: str) -> str:
    """
    Example of an async agent function.
    
    The generic tester will automatically detect and handle async functions.
    """
    # Simulate async processing
    await asyncio.sleep(0.1)
    return f"Async agent response to: {prompt}"

# Example of a more complex agent with AI API integration
def ai_agent(prompt: str) -> str:
    """
    Example agent that would integrate with an AI API.
    
    This shows how you might structure a real AI agent.
    """
    try:
        # Check if OpenAI API key is available
        api_key = os.getenv("OPENAI_API_KEY")
        if not api_key:
            return "Error: OPENAI_API_KEY environment variable not set"
        
        # In a real implementation, you would call the AI API here
        # For this example, we'll just return a mock response
        return f"AI Agent Response: I understand you're asking about '{prompt}'. Here's my response..."
        
    except Exception as e:
        return f"Error processing request: {str(e)}"

if __name__ == "__main__":
    # Allow direct execution for testing
    import sys
    if len(sys.argv) > 1:
        prompt = " ".join(sys.argv[1:])
        print(main(prompt))
    else:
        print("Usage: python sample_agent.py <prompt>")
