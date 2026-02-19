@echo off
REM Check Node version (Node 20+ required)
node -e "const v=process.versions.node.split('.')[0]; if(parseInt(v)<20) { console.error('Node 20+ required'); process.exit(1); }"
if errorlevel 1 exit /b 1
REM Install deps
call npm install
if errorlevel 1 exit /b 1
REM Build
call npm run build
if errorlevel 1 exit /b 1
