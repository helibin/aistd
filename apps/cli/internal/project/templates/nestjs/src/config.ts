import { isIP } from 'node:net';

export function loadConfig(value = process.env.HTTP_ADDR || ':8080') {
  const match = /^(?:\[([^\]]+)\]|([^:]*)):(\d+)$/.exec(value);
  if (!match) throw new Error('HTTP_ADDR must be host:port or [IPv6]:port');
  const host = match[1] ?? match[2];
  const port = Number(match[3]);
  if (port < 1 || port > 65535) throw new Error('HTTP_ADDR port must be 1..65535');
  if ((match[1] && isIP(host) !== 6) || (!match[1] && host && !isIP(host) && !host.split('.').every(label => /^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$/.test(label)))) {
    throw new Error('HTTP_ADDR contains an invalid host');
  }
  const healthHost = !host || host === '0.0.0.0' ? '127.0.0.1' : host === '::' ? '::1' : host;
  return { host: host || '0.0.0.0', port, healthURL: `http://${isIP(healthHost) === 6 ? `[${healthHost}]` : healthHost}:${port}/api/health/live` };
}
