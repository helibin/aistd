import 'reflect-metadata';
import { NestFactory } from '@nestjs/core';
import { randomUUID } from 'node:crypto';
import type { IncomingMessage, ServerResponse } from 'node:http';
import { AppModule } from './app.module';
import { loadConfig } from './config';

async function main() {
  const config = loadConfig();
  const app = await NestFactory.create(AppModule);
  app.use((request: IncomingMessage, response: ServerResponse, next: () => void) => {
    const requestID = randomUUID();
    const started = performance.now();
    response.setHeader('X-Request-ID', requestID);
    response.once('finish', () => console.log(JSON.stringify({ event: 'http_request', request_id: requestID,
      method: request.method, path: request.url?.split('?')[0], status: response.statusCode, duration_ms: performance.now() - started })));
    next();
  });
  const server = app.getHttpServer();
  server.headersTimeout = 5000;
  server.requestTimeout = 10000;
  server.timeout = 15000;
  server.keepAliveTimeout = 60000;
  let closing = false;
  const shutdown = async () => {
    if (closing) return;
    closing = true;
    const deadline = setTimeout(() => {
      console.error('HTTP shutdown exceeded five seconds');
      server.closeAllConnections();
      process.exit(1);
    }, 5000);
    try { await app.close(); } catch (error) { console.error(error); process.exitCode = 1; }
    finally { clearTimeout(deadline); }
  };
  process.on('SIGINT', shutdown);
  process.on('SIGTERM', shutdown);
  await app.listen(config.port, config.host);
}
main().catch(error => { console.error(error); process.exitCode = 1; });
