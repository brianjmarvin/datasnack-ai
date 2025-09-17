#!/usr/bin/env python3
"""
Universal AI Agent Tester Script

This script can test any Python AI agent, including:
- Function-based agents (main, run, execute, etc.)
- Class-based agents with methods
- Server-based agents
- Agents with virtual environments

Usage:
    python universal_agent_tester.py <agent_script_path> <test_prompt> [python_path]

Examples:
    python universal_agent_tester.py /path/to/agent.py "Test prompt"
    python universal_agent_tester.py /path/to/agent.py "Test prompt" /path/to/venv/bin/python
"""

import sys
import os
import asyncio
import json
import subprocess
import tempfile
from pathlib import Path

def find_python_interpreter(script_path):
    """Find the best Python interpreter to use for the agent"""
    agent_dir = os.path.dirname(script_path)
    
    # Check for virtual environments in common locations
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
                return venv_python
    
    # Fall back to system Python
    return sys.executable

def create_universal_test_script(script_path, test_prompt):
    """Create a universal test script that can handle any type of agent"""
    test_script = f'''#!/usr/bin/env python3
import sys
import os
import asyncio
import importlib.util
from pathlib import Path

# Add the agent directory to Python path
agent_dir = os.path.dirname("{script_path}")
if agent_dir not in sys.path:
    sys.path.insert(0, agent_dir)

def load_agent_module(script_path):
    """Dynamically load the agent module from the script path"""
    spec = importlib.util.spec_from_file_location("agent_module", script_path)
    if spec is None:
        raise ImportError(f"Could not load module from {{script_path}}")
    
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
                print(f"Found agent function: {{name}}")
                return func, name
    
    # If no common name found, look for any callable that might be the agent
    print("No common function names found, searching for suitable functions...")
    for attr_name in dir(module):
        if not attr_name.startswith('_') and not attr_name[0].isupper():  # Skip classes
            attr = getattr(module, attr_name)
            if callable(attr):
                # Skip server applications and utility functions
                if attr_name in ['app', 'application', 'server', 'load_dotenv', 'logger']:
                    continue
                    
                # Check if it looks like an agent function (takes string input)
                import inspect
                try:
                    sig = inspect.signature(attr)
                    params = list(sig.parameters.keys())
                    if len(params) >= 1:  # At least one parameter
                        # Additional check: make sure it's not a utility function
                        if not any(util in attr_name.lower() for util in ['load', 'config', 'setup', 'init', 'create']):
                            print(f"Found potential agent function: {{attr_name}}")
                            return attr, attr_name
                except:
                    continue
    
    # If still no function found, try to create a wrapper for server-based agents
    print("No direct function found, creating server wrapper...")
    return create_server_wrapper(module), "server_wrapper"

def create_server_wrapper(module):
    """Create a wrapper function for server-based agents like GPT Researcher"""
    async def server_wrapper(prompt):
        try:
            # Try to import GPTResearcher if available
            from gpt_researcher import GPTResearcher
            from gpt_researcher.utils.enum import ReportType, Tone
            
            researcher = GPTResearcher(
                query=prompt,
                report_type=ReportType.ResearchReport.value,
                tone=Tone.Objective
            )
            
            # Conduct research
            result = await researcher.conduct_research()
            
            # Generate report
            report = await researcher.write_report()
            
            return report
            
        except ImportError:
            # If GPTResearcher is not available, try other common patterns
            try:
                # Try to find any class that might be an agent
                for attr_name in dir(module):
                    if not attr_name.startswith('_') and attr_name[0].isupper():
                        attr = getattr(module, attr_name)
                        if hasattr(attr, '__call__'):
                            # Try to instantiate and call
                            try:
                                instance = attr()
                                if hasattr(instance, 'run') or hasattr(instance, 'execute') or hasattr(instance, 'process'):
                                    method = getattr(instance, 'run', None) or getattr(instance, 'execute', None) or getattr(instance, 'process', None)
                                    if callable(method):
                                        if asyncio.iscoroutinefunction(method):
                                            return await method(prompt)
                                        else:
                                            return method(prompt)
                            except:
                                continue
            except:
                pass
            
            # Fallback: return a message indicating the agent type
            return f"Server-based agent detected. Original prompt: {{prompt}}. This agent requires specific integration patterns."
    
    return server_wrapper

async def test_agent():
    """Test the agent with the given prompt"""
    try:
        # Load the agent module
        print(f"Loading agent from: {script_path}")
        agent_module = load_agent_module("{script_path}")
        
        # Find the main agent function
        agent_func, func_name = find_agent_function(agent_module)
        
        print(f"Testing agent with prompt: {test_prompt}")
        print("-" * 50)
        
        # Call the agent function
        if asyncio.iscoroutinefunction(agent_func):
            # Handle async functions
            print("Calling async agent function...")
            result = await agent_func("{test_prompt}")
        else:
            # Handle sync functions
            print("Calling sync agent function...")
            result = agent_func("{test_prompt}")
        
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
        print(f"Error: {{str(e)}}")
        import traceback
        traceback.print_exc()
        return None

if __name__ == "__main__":
    result = asyncio.run(test_agent())
    if result is None:
        sys.exit(1)
'''
    return test_script

def test_agent_with_subprocess(script_path, test_prompt, python_path=None):
    """Test the agent using a subprocess with the correct Python interpreter"""
    if python_path is None:
        python_path = find_python_interpreter(script_path)
    
    print(f"Using Python interpreter: {python_path}")
    
    # Create universal test script
    test_script = create_universal_test_script(script_path, test_prompt)
    
    # Write to temporary file
    with tempfile.NamedTemporaryFile(mode='w', suffix='.py', delete=False) as f:
        f.write(test_script)
        temp_script_path = f.name
    
    try:
        # Execute the test script
        result = subprocess.run(
            [python_path, temp_script_path],
            capture_output=True,
            text=True,
            cwd=os.path.dirname(script_path)
        )
        
        # Print output
        if result.stdout:
            print(result.stdout)
        
        if result.stderr:
            print("STDERR:", result.stderr)
        
        return result.returncode == 0
        
    finally:
        # Clean up temporary file
        os.unlink(temp_script_path)

def main():
    """Main function to run the universal agent tester"""
    if len(sys.argv) < 3:
        print("Usage: python universal_agent_tester.py <agent_script_path> <test_prompt> [python_path]")
        print("\nExamples:")
        print('python universal_agent_tester.py /path/to/agent.py "Test prompt"')
        print('python universal_agent_tester.py /path/to/agent.py "Test prompt" /path/to/venv/bin/python')
        print("\nSupported agent types:")
        print("- Function-based agents (main, run, execute, etc.)")
        print("- Class-based agents with methods")
        print("- Server-based agents (like GPT Researcher)")
        print("- Agents with virtual environments")
        sys.exit(1)
    
    script_path = sys.argv[1]
    test_prompt = sys.argv[2]
    python_path = sys.argv[3] if len(sys.argv) > 3 else None
    
    if not os.path.exists(script_path):
        print(f"Error: Agent script not found: {script_path}")
        sys.exit(1)
    
    print("=" * 60)
    print("UNIVERSAL AI AGENT TESTER")
    print("=" * 60)
    print(f"Agent Script: {script_path}")
    print(f"Test Prompt: {test_prompt}")
    print("=" * 60)
    
    # Test the agent
    success = test_agent_with_subprocess(script_path, test_prompt, python_path)
    
    if success:
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
