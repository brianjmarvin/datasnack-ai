#!/usr/bin/env python3
"""
Generic AI Agent Tester Script

This script can be used to test any Python AI agent that has been instrumented
using the AI Call Tracking and Evaluation Framework from cursor_eval_prompt.txt.

Usage:
    python generic_agent_tester.py <agent_script_path> <test_prompt>

Example:
    python generic_agent_tester.py /path/to/your/agent.py "What is artificial intelligence?"
"""

import sys
import os
import asyncio
import json
import importlib.util
from pathlib import Path

def load_agent_module(script_path):
    """Dynamically load the agent module from the script path"""
    if not os.path.exists(script_path):
        raise FileNotFoundError(f"Agent script not found: {script_path}")
    
    # Try to detect and use virtual environment
    agent_dir = os.path.dirname(script_path)
    venv_paths = [
        os.path.join(agent_dir, "venv"),
        os.path.join(agent_dir, ".venv"),
        os.path.join(agent_dir, "env"),
        os.path.join(os.path.dirname(agent_dir), "venv"),
        os.path.join(os.path.dirname(agent_dir), ".venv"),
    ]
    
    for venv_path in venv_paths:
        if os.path.exists(venv_path):
            venv_python = os.path.join(venv_path, "bin", "python")
            if os.path.exists(venv_python):
                print(f"Found virtual environment: {venv_path}")
                # Note: We can't change the Python interpreter mid-execution,
                # but we can add the venv site-packages to the path
                site_packages = os.path.join(venv_path, "lib", "python*", "site-packages")
                import glob
                for sp in glob.glob(site_packages):
                    if sp not in sys.path:
                        sys.path.insert(0, sp)
                break
    
    spec = importlib.util.spec_from_file_location("agent_module", script_path)
    if spec is None:
        raise ImportError(f"Could not load module from {script_path}")
    
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module

def find_agent_function(module):
    """Find the main agent function in the module"""
    # Common function names for AI agents
    possible_names = [
        'main', 'run', 'execute', 'process', 'generate', 'respond',
        'chat', 'query', 'research', 'analyze', 'agent', 'ai_agent',
        'handle_request', 'process_query', 'get_response', 'answer'
    ]
    
    # First, try common function names
    for name in possible_names:
        if hasattr(module, name):
            func = getattr(module, name)
            if callable(func):
                print(f"Found agent function: {name}")
                return func, name
    
    # If no common name found, look for any callable that might be the agent
    print("No common function names found, searching for suitable functions...")
    for attr_name in dir(module):
        if not attr_name.startswith('_'):
            attr = getattr(module, attr_name)
            if callable(attr):
                # Check if it looks like an agent function (takes string input)
                import inspect
                try:
                    sig = inspect.signature(attr)
                    params = list(sig.parameters.keys())
                    if len(params) >= 1:  # At least one parameter
                        print(f"Found potential agent function: {attr_name}")
                        return attr, attr_name
                except:
                    continue
    
    raise ValueError("Could not find a suitable agent function in the module")

async def test_agent(script_path, user_prompt):
    """Test the agent with the given prompt"""
    try:
        # Add the agent directory to Python path
        agent_dir = os.path.dirname(script_path)
        if agent_dir not in sys.path:
            sys.path.insert(0, agent_dir)
        
        # Load the agent module
        print(f"Loading agent from: {script_path}")
        agent_module = load_agent_module(script_path)
        
        # Find the main agent function
        agent_func, func_name = find_agent_function(agent_module)
        
        print(f"Testing agent with prompt: {user_prompt}")
        print("-" * 50)
        
        # Call the agent function
        if asyncio.iscoroutinefunction(agent_func):
            # Handle async functions
            print("Calling async agent function...")
            result = await agent_func(user_prompt)
        else:
            # Handle sync functions
            print("Calling sync agent function...")
            result = agent_func(user_prompt)
        
        print("-" * 50)
        print("AGENT RESPONSE:")
        print("-" * 50)
        
        # Print the result
        if result is not None:
            print(str(result))
        else:
            print("Agent returned None")
        
        return result
        
    except Exception as e:
        print(f"Error: {str(e)}")
        import traceback
        traceback.print_exc()
        return None

def main():
    """Main function to run the agent tester"""
    if len(sys.argv) != 3:
        print("Usage: python generic_agent_tester.py <agent_script_path> <test_prompt>")
        print("\nExample:")
        print('python generic_agent_tester.py /path/to/your/agent.py "What is artificial intelligence?"')
        sys.exit(1)
    
    script_path = sys.argv[1]
    test_prompt = sys.argv[2]
    
    print("=" * 60)
    print("GENERIC AI AGENT TESTER")
    print("=" * 60)
    print(f"Agent Script: {script_path}")
    print(f"Test Prompt: {test_prompt}")
    print("=" * 60)
    
    # Run the test
    result = asyncio.run(test_agent(script_path, test_prompt))
    
    if result is not None:
        print("\n" + "=" * 60)
        print("TEST COMPLETED SUCCESSFULLY")
        print("=" * 60)
    else:
        print("\n" + "=" * 60)
        print("TEST FAILED")
        print("=" * 60)
        sys.exit(1)

if __name__ == "__main__":
    main()
