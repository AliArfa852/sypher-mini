
I need to build Sypher-mini, a coding-centric agent pipeline combining the lightweight efficiency of PicoClaw and the flexible API support of OpenClaw. Key specifications are as follows:

1. API Flexibility: Sypher-mini must support multiple LLM providers via API keys. It should integrate Cerebras, OpenAI, Anthropic, and Gemini LLMs so that it can act as its own smart decision-making agent. Users must supply API keys for these LLMs to enable agentic reasoning.

2. Agent and CLI Integration: Sypher-mini must manage and control multiple agents (like Gemini CLI, Cursor) and other CLI tools. Every agent must be set up by the user. The agent should access and control commands on any CLI environment chosen, with each CLI command logged in a dedicated file for audit.

3. Terminal and Process Tracking: Sypher-mini must monitor all active terminals, record which commands are run, tag and categorize them, and ensure all actions are logged for each task in a separate file. It must track which processes it started and only allow killing of tasks it created.

4. Command Execution & Reporting: It must allow live CLI monitoring (e.g., tail -n 100) and live reports on any server. If a chosen server throws specific errors (e.g., 400, 500 status codes), it must send an immediate WhatsApp alert (only on selected CLIs chosen by me).

5. User Setup & Control: All settings—agents, commands, CLIs—must be accessible via both CLI and WhatsApp. Users can set up, modify, and monitor all configurations from these interfaces.

6. Secure CLI Access: Sypher-mini must never have full unrestricted CLI access. It should only operate on the terminals I authorize, log all activities per task, and only allow it to kill tasks it initiated.

7. Automation & Config Files: Sypher-mini must auto-start via CLI on system boot, using a config file per command. The user will define which files, agents, and tasks it can access. The agent will guide the code development pipeline by relaying prompts and tasks between me (via WhatsApp) and the selected coding agents.

8. Live Reporting: It should generate live reports on chosen servers, alerting me via WhatsApp on critical errors, and allowing me to track code execution status across all terminals.

This prompt should guide Cursor to produce a detailed plan.md, ensuring all tasks, API keys, and agent controls are included, and that all settings are accessible both via CLI and WhatsApp.
