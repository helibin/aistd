import {execFileSync} from 'node:child_process';

const git = (...args) => execFileSync('git', args, {encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe']}).trim();
const read = (ref, path) => {
  try { return git('show', `${ref}:${path}`); } catch { return ''; }
};
const semver = value => /^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/.test(value);
const newer = (next, old) => {
  const a = next.split('.').map(Number), b = old.split('.').map(Number);
  const index = a.findIndex((value, i) => value !== b[i]);
  return index !== -1 && a[index] > b[index];
};

function check(ref, previous, paths) {
  if (!paths.length) return;
  const affected = new Set();
  for (const path of paths) {
    const match = /^(?:apps|changes)\/(cli|web)\//.exec(path);
    if (match) affected.add(match[1]);
    else if (path === 'package-lock.json' || path.startsWith('design/') || path.startsWith('docs/design/')) affected.add('web');
    else { affected.add('cli'); affected.add('web'); }
  }
  for (const app of affected) {
    const base = `apps/${app}`;
    const next = read(ref, `${base}/VERSION`), old = read(previous, `${base}/VERSION`);
    if (!semver(next) || (old && (!semver(old) || !newer(next, old)))) {
      throw Error(`${ref || 'staged'}: ${base}/VERSION must increase MAJOR.MINOR.PATCH (${old || 'new baseline'} -> ${next || 'missing'})`);
    }
    const changelog = read(ref, `${base}/CHANGELOG.md`);
    if (!paths.includes(`${base}/CHANGELOG.md`) || !new RegExp(`^## \\[${next.replaceAll('.', '\\.')}\\] - \\d{4}-\\d{2}-\\d{2}$`, 'm').test(changelog)) {
      throw Error(`${base}/CHANGELOG.md must include a dated entry for ${next}`);
    }
    const oldChangelog = read(previous, `${base}/CHANGELOG.md`);
    const history = oldChangelog.indexOf('\n## ');
    if (history !== -1 && !changelog.endsWith(oldChangelog.slice(history))) {
      throw Error(`${base}/CHANGELOG.md must preserve previous release entries`);
    }
    if (app === 'web') {
      const pkg = JSON.parse(read(ref, 'apps/web/package.json'));
      const lock = JSON.parse(read(ref, 'package-lock.json'));
      if (pkg.version !== next || lock.packages['apps/web'].version !== next) {
        throw Error('Website VERSION, package.json and package-lock.json versions must match');
      }
    }
  }
}

try {
  if (process.argv[2] === '--range' && process.argv[3]) {
    for (const commit of git('rev-list', '--reverse', process.argv[3]).split('\n').filter(Boolean)) {
      check(commit, `${commit}^`, git('diff', '--name-only', `${commit}^`, commit).split('\n').filter(Boolean));
    }
  } else if (process.argv.length === 2) {
    check('', 'HEAD', git('diff', '--cached', '--name-only', '--diff-filter=ACDMRT').split('\n').filter(Boolean));
  } else throw Error('Usage: node scripts/check-version.mjs [--range BASE..HEAD]');
  console.log('PASS: version and changelog checks');
} catch (error) {
  console.error(error.message);
  process.exitCode = 1;
}
