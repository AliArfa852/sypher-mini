#!/usr/bin/env node
const { spawn } = require('child_process');
const path = require('path');
const extDir = path.join(__dirname, '..');
const MIN_NODE_MAJOR = 18;
function checkNodeVersion() {
  const v = process.version.slice(1).split('.');
  const major = parseInt(v[0], 10);
  if (major < MIN_NODE_MAJOR) {
    console.error('Node.js ' + MIN_NODE_MAJOR + '+ required. Current: ' + process.version);
    process.exit(1);
  }
}
function run(cmd, args) {
  return new Promise((resolve, reject) => {
    const proc = spawn(cmd, args, { cwd: extDir, stdio: 'inherit', shell: true });
    proc.on('close', (code) => (code === 0 ? resolve() : reject(new Error('exit ' + code))));
  });
}
async function main() {
  checkNodeVersion();
  console.log('Setting up whatsapp-baileys extension...');
  await run('npm', ['install']);
  await run('npm', ['run', 'build']);
  console.log('Setup complete.');
}
main().catch((e) => { console.error('Setup failed:', e.message); process.exit(1); });
