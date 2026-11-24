const http = require('node:http');
const { loadConfig } = require('../dist/config');
try {
  const request = http.get(loadConfig().healthURL, { timeout: 2000 }, response => {
    let body = '';
    response.setEncoding('utf8');
    response.on('data', chunk => { body += chunk; if (body.length > 1024) request.destroy(new Error('health response too large')); });
    response.on('end', () => {
      try { if (response.statusCode !== 200 || JSON.parse(body).status !== 'ok') throw new Error('unhealthy HTTP response'); }
      catch (error) { console.error(error.message); process.exitCode = 1; }
    });
    response.on('error', error => { console.error(error.message); process.exitCode = 1; });
  });
  const deadline = setTimeout(() => request.destroy(new Error('health deadline exceeded')), 2000);
  request.on('close', () => clearTimeout(deadline));
  request.on('timeout', () => request.destroy(new Error('health timeout')));
  request.on('error', error => { console.error(error.message); process.exitCode = 1; });
} catch (error) { console.error(error.message); process.exitCode = 1; }
