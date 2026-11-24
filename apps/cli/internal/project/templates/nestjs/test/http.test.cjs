const { test } = require('node:test');
const assert = require('node:assert/strict');
const { spawn, spawnSync } = require('node:child_process');
const { once } = require('node:events');
const net = require('node:net');
const http = require('node:http');
const { loadConfig } = require('../dist/config');

test('validates HTTP_ADDR', () => {
  assert.equal(loadConfig(':9090').healthURL, 'http://127.0.0.1:9090/api/health/live');
  assert.equal(loadConfig('[::]:9090').healthURL, 'http://[::1]:9090/api/health/live');
  for (const value of ['bad', ':0', ':65536', '[bad]:8080', 'bad host:80']) assert.throws(() => loadConfig(value));
});

test('healthcheck has a total deadline for a trickling response', { timeout: 6000 }, async t => {
  const server = http.createServer((request, response) => {
    response.writeHead(200);
    const interval = setInterval(() => response.write(' '), 100);
    response.on('close', () => clearInterval(interval));
  });
  server.listen(0, '127.0.0.1');
  await once(server, 'listening');
  try {
    const started = performance.now();
    const child = spawn(process.execPath, ['scripts/healthcheck.cjs'], { env: { ...process.env, HTTP_ADDR: `127.0.0.1:${server.address().port}` } });
    t.after(() => { if (child.exitCode === null) child.kill('SIGKILL'); });
    const [code] = await once(child, 'exit');
    assert.equal(code, 1);
    assert.ok(performance.now() - started < 4000, 'healthcheck exceeded its total deadline');
  } finally {
    server.closeAllConnections();
    await new Promise(resolve => server.close(resolve));
  }
});

test('serves real HTTP, logs request IDs, checks health and shuts down', { timeout: 15000 }, async t => {
  const reservation = net.createServer();
  reservation.listen(0, '127.0.0.1');
  await once(reservation, 'listening');
  const port = reservation.address().port;
  await new Promise(resolve => reservation.close(resolve));
  const env = { ...process.env, HTTP_ADDR: `127.0.0.1:${port}` };
  const child = spawn(process.execPath, ['dist/main.js'], { env });
  const exit = once(child, 'exit');
  t.after(() => { if (child.exitCode === null) child.kill('SIGKILL'); });
  let logs = '';
  child.stdout.on('data', chunk => { logs += chunk; });
  child.stderr.on('data', chunk => { logs += chunk; });
  await new Promise((resolve, reject) => {
    const timer = setTimeout(() => reject(new Error(`server did not start: ${logs}`)), 10000);
    const started = () => { if (logs.includes('Nest application successfully started')) { clearTimeout(timer); child.stdout.off('data', started); resolve(); } };
    child.stdout.on('data', started); started();
    child.once('exit', code => { clearTimeout(timer); reject(new Error(`server exited ${code}: ${logs}`)); });
  });
  for (const [path, body] of [['health/live', { status: 'ok' }], ['health/ready', { status: 'ready' }], ['hello', { message: 'hello' }]]) {
    const response = await fetch(`http://127.0.0.1:${port}/api/${path}`);
    assert.equal(response.status, 200);
    assert.match(response.headers.get('x-request-id'), /^[0-9a-f-]{36}$/);
    assert.deepEqual(await response.json(), body);
  }
  assert.equal((await fetch(`http://127.0.0.1:${port}/missing`)).status, 404);
  assert.equal(spawnSync(process.execPath, ['scripts/healthcheck.cjs'], { env }).status, 0);
  child.kill('SIGTERM');
  const [code] = await exit;
  assert.equal(code, 0, logs);
  assert.match(logs, /"status":404/);
  assert.match(logs, /"request_id":/);
  assert.equal(spawnSync(process.execPath, ['scripts/healthcheck.cjs'], { env }).status, 1);
});
