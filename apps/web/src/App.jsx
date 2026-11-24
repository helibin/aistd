import React, {useEffect, useRef, useState} from 'react';
import {startScene} from './scene.js';
import {defaultStacks, projectConfiguration, stackFamilies, targetKeys} from './configuration.js';

const installUrl = window.AISTD_INSTALL_URL || '';
const installCommand = installUrl ? `curl -fsSL ${installUrl} | sh` : '';
const cliVersion = __AISTD_CLI_VERSION__;
const targetGroups = [
  {title: '服务与网页', stacks: ['apiStack', 'webStack'], options: [
    ['api', 'API'],
    ['web', 'Web']
  ]},
  {title: '后台管理', stacks: ['adminStack'], options: [
    ['admin', 'Admin']
  ]},
  {title: '移动 App', stacks: ['mobileStack'], options: [
    ['ios', 'iOS App'],
    ['android', 'Android App']
  ]},
  {title: '小程序', stacks: ['miniappStack'], options: [
    ['wechat', '微信小程序'],
    ['alipay', '支付宝小程序']
  ]}
];
const docs = {
  start: {
    title: '安装 / 升级 CLI',
    text: '通过本站安装脚本下载对应平台的二进制；使用 CLI 不需要 Go 环境。',
    code: installCommand || '安装地址尚未配置',
    note: installCommand ? '支持 macOS / Linux 的 amd64 和 arm64。重复执行同一命令即可升级；新版下载、SHA256 与运行验证成功后替换旧 CLI，失败保留旧版。' : '发行完成并配置安装地址后，这里会显示可复制的命令。',
    copyable: Boolean(installCommand)
  },
  create: {
    title: '创建新项目',
    text: '给项目起名即可使用默认组合；运行 aistd 或 aistd create 可在终端用方向键和空格选择应用目标及技术栈，查看摘要后确认创建。管道或简易终端保留编号问答。',
    code: 'aistd create --name hello-aistd\ncd hello-aistd\naistd doctor --dir .',
    note: '默认 Go + React。后端可选 NestJS，Web 可选 Vue，移动端可选 React Native / Expo 或 Flutter，小程序可选 Taro 或 uni-app。Admin 独立可选 Vben Naive UI 或 Soybean-admin；需 Node.js 22.18+，make setup 安装独立 pnpm 工作区的依赖，make run-admin 启动；模板演示的登录与权限需另行对接业务后端。可加 --openspec 初始化已安装的官方 CodeBuddy 工作流；默认关闭。'
  },
  init: {
    title: '接入已有项目',
    text: '在现有项目中安装两份规范原文和接入清单。已有代码、配置和 AGENTS.md 不会被覆盖。',
    code: 'cd /path/to/project\naistd init --dir .\naistd doctor --dir .',
    note: '可加 --openspec 接入 OpenSpec（需预先安装官方 CLI）；若已有 openspec/ 或 .codebuddy/ 则拒绝自动覆盖。可先用 --dry-run 预览。'
  },
  doctor: {
    title: '校验规范快照',
    text: 'doctor 将项目锁文件和受管规范文件逐项对照 CLI 内置的 v1.0 快照。',
    code: 'aistd doctor --dir ./hello-aistd',
    note: '文件完整性通过，不代表工程或视觉规范已通过项目验收。'
  },
  uninstall: {
    title: '卸载 CLI',
    text: '运行卸载命令会展示当前 CLI 可执行文件路径并要求确认；只卸载由安装脚本放置的 CLI。',
    code: 'aistd uninstall',
    note: '如果安装时设置了 AISTD_INSTALL_DIR，卸载时也需设置为相同目录；自动化场景可使用 --yes。不会删除已创建或接入的项目、规范文件和 task.md。'
  }
};
const tabs = [
  ['start', '安装与升级'], ['create', '创建新项目'],
  ['init', '已有项目'], ['doctor', '校验规范'], ['uninstall', '卸载 CLI']
];

export default function App() {
  const canvasRef = useRef(null);
  const toastTimer = useRef(null);
  const [name, setName] = useState('hello-aistd');
  const [targets, setTargets] = useState({api: true, web: true, ios: false, android: false, wechat: false, alipay: false, admin: false});
  const [stacks, setStacks] = useState(defaultStacks);
  const [visual, setVisual] = useState(true);
  const [openSpec, setOpenSpec] = useState(false);
  const [activeDoc, setActiveDoc] = useState('start');
  const [previewMode, setPreviewMode] = useState('create');
  const [toast, setToast] = useState('');
  const validName = /^[a-z][a-z0-9-]{0,62}$/.test(name);
  const selectedTargets = targetKeys.filter(key => targets[key]);
  const valid = validName && selectedTargets.length > 0;
  const selection = projectConfiguration(name, targets, stacks, visual, openSpec);
  const command = selection.command;
  const config = JSON.stringify(selection.config, null, 2);
  const doc = docs[activeDoc];

  useEffect(() => {
    const stopScene = startScene(canvasRef.current);
    document.documentElement.classList.add('js');
    const targets = document.querySelectorAll('.reveal');
    const observer = new IntersectionObserver(entries => {
      entries.forEach(entry => {
        if (!entry.isIntersecting) return;
        entry.target.classList.add('is-visible');
        observer.unobserve(entry.target);
      });
    }, {threshold: .08});
    targets.forEach(target => observer.observe(target));
    return () => {
      stopScene();
      observer.disconnect();
      clearTimeout(toastTimer.current);
    };
  }, []);

  function notify(message) {
    setToast(message);
    clearTimeout(toastTimer.current);
    toastTimer.current = setTimeout(() => setToast(''), 2600);
  }

  async function copy(value) {
    try {
      await navigator.clipboard.writeText(value);
      notify('已复制');
    } catch {
      notify('复制不可用，请手动选择文本');
    }
  }

  function download() {
    if (!valid) return;
    const url = URL.createObjectURL(new Blob([config + '\n'], {type: 'application/json'}));
    const link = document.createElement('a');
    link.href = url;
    link.download = 'project.json';
    link.click();
    setTimeout(() => URL.revokeObjectURL(url), 1000);
    notify('配置已生成');
  }

  function toggleTarget(key) {
    setTargets(current => ({...current, [key]: !current[key]}));
  }

  function selectTab(event, index) {
    const direction = ['ArrowRight', 'ArrowDown'].includes(event.key) ? 1 :
      ['ArrowLeft', 'ArrowUp'].includes(event.key) ? -1 : 0;
    if (!direction) return;
    event.preventDefault();
    const next = (index + direction + tabs.length) % tabs.length;
    setActiveDoc(tabs[next][0]);
    document.getElementById(`tab-${tabs[next][0]}`).focus();
  }

  return <>
    <a className="skip-link" href="#main">跳转到内容</a>
    <header className="site-header">
      <div className="header-inner">
        <a className="brand" href="#top" aria-label="aistd 首页">aistd<span aria-hidden="true">.</span></a>
        <nav className="main-nav" aria-label="主导航">
          <a href="#process">开始使用</a><a href="#standards">规范</a><a href="#docs">文档</a>
        </nav>
        <a className="header-action" href="#install">安装 / 升级 CLI <span aria-hidden="true">↓</span></a>
      </div>
    </header>

    <main id="main">
      <section className="hero" id="top" aria-labelledby="hero-title">
        <canvas id="system-canvas" ref={canvasRef} aria-hidden="true" />
        <div className="hero-content">
          <p className="overline hero-overline"><b>aistd</b> {cliVersion} × ENGINEERING / VISUAL V1.0</p>
          <h1 id="hero-title">让标准，<br />成为项目的起点<em aria-hidden="true">.</em></h1>
          <p className="hero-description">一个 CLI，创建多平台项目，或把工程与视觉规范接入现有代码。</p>
          <div className="hero-actions">
            {installCommand ? <button className="button button-accent" type="button" onClick={() => copy(installCommand)}>安装 / 升级 CLI <span aria-hidden="true">↓</span></button> : <a className="button button-accent" href="#install">查看安装方式</a>}
            <a className="button button-ghost" href="#configure">配置新项目</a>
          </div>
          <p className="hero-status"><code>{installCommand ? <><span>curl -fsSL </span><span>{installUrl} | sh</span></> : '安装地址尚未配置'}</code></p>
          <p className="hero-install-note">{installCommand ? '复制命令到终端执行 · 已安装？重复执行即可升级。' : '安装地址配置完成后，这里会提供安装与升级命令。'}</p>
          <div className="artifact-preview" aria-label="CLI 生成结果预览">
            <div className="artifact-toolbar">
              <span className="artifact-title">aistd <span>/</span> project output</span>
              <div className="artifact-modes" role="group" aria-label="项目路径">
                <button type="button" aria-pressed={previewMode === 'create'} onClick={() => setPreviewMode('create')}>新建项目</button>
                <button type="button" aria-pressed={previewMode === 'init'} onClick={() => setPreviewMode('init')}>已有项目</button>
              </div>
              <span className="artifact-version">v{cliVersion}</span>
            </div>
            <div className="artifact-body">
              <div className="artifact-command">
                <span className="artifact-label">01 / COMMAND</span>
                <pre>{previewMode === 'create' ? <><span>$</span> aistd create --name hello-aistd<br /><span>$</span> cd hello-aistd<br /></> : <><span>$</span> aistd init --dir .<br /></>}<span>$</span> aistd doctor --dir .</pre>
                <p>{previewMode === 'create' ? '默认生成 Go API、React Web 与视觉草稿。' : '原地安装规范来源，不改动已有业务文件。'}</p>
              </div>
              <div className="artifact-files">
                <span className="artifact-label">02 / FILES</span>
                {previewMode === 'create' ? <div className="file-tree" aria-label="默认项目目录">
                  <div className="tree-root">hello-aistd/</div>
                  <div><span>├─</span> apps/<em>api/</em> <small>Go</small></div>
                  <div><span>├─</span> apps/<em>web/</em> <small>React</small></div>
                  <div><span>├─</span> design/tokens/<em>brand.json</em></div>
                  <div><span>└─</span> .aistd/<em>standards/</em> <small>v1.0</small></div>
                </div> : <div className="file-tree" aria-label="已有项目新增文件">
                  <div className="tree-root">your-project/</div>
                  <div><span>├─</span> AISTD-ADOPTION.md</div>
                  <div><span>└─</span> .aistd/<em>standards/</em> <small>v1.0</small></div>
                  <div><span>　</span> .aistd/<em>standards.lock.json</em></div>
                </div>}
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="process" id="process" aria-labelledby="process-title"><div className="section-shell">
        <div className="section-heading reveal">
          <p className="overline">01 / THE WORKFLOW</p>
          <h2 id="process-title">安装，然后开始。</h2>
          <p>新项目直接创建；已有项目原地接入。两条路径都不需要手写配置。</p>
        </div>
        <div className="install-step reveal" id="install">
          <div><span className="process-number">01 / INSTALL & UPGRADE</span><h3>安装 / 升级 CLI</h3><p>首次安装，或重复执行升级。新版验证成功后自动替换，失败保留旧版。</p></div>
          <div className="install-step-command"><code>{installCommand || '安装地址尚未配置'}</code><button type="button" disabled={!installCommand} onClick={() => copy(installCommand)}>复制</button></div>
        </div>
        <div className="process-list">
          <article className="process-row reveal"><span className="process-number">02 / NEW</span><div><h3>创建新项目</h3><p>默认包含 Go API、React Web 和视觉草稿。项目名可替换，选项可另行配置。</p><code>aistd create --name hello-aistd</code></div><button type="button" onClick={() => copy('aistd create --name hello-aistd')}>复制命令</button></article>
          <article className="process-row reveal"><span className="process-number">03 / EXISTING</span><div><h3>接入已有项目</h3><p>在项目目录新增规范和审计清单，不覆盖现有文件。</p><code>aistd init --dir .</code></div><button type="button" onClick={() => copy('aistd init --dir .')}>复制命令</button></article>
        </div>
        <p className="process-after reveal">完成后在项目目录运行 <code>aistd doctor --dir .</code>，校验规范原文与锁文件。项目是否符合规范仍需评审。<a href="#docs">查看完整命令 ↗</a></p>
      </div></section>

      <section className="configure-band" id="configure" aria-labelledby="configure-title">
        <div className="section-shell">
          <div className="section-heading reveal"><p className="overline">02 / PROJECT TARGETS</p><h2 id="configure-title">选一个，或组合多个。</h2><p>先选应用目标，再选技术栈；命令和 JSON 配置同步生成。</p></div>
          <div className="config-tool reveal">
            <div className="config-form">
              <div className="tool-heading"><span>PROJECT SETUP</span><span>TARGETS / STACKS</span></div>
              <label className="field-label" htmlFor="name">项目名称</label>
              <input id="name" type="text" value={name} onChange={event => setName(event.target.value)} maxLength="63" spellCheck="false" autoComplete="off" aria-describedby="error" aria-invalid={!validName} />
              <div id="error" className="field-error" aria-live="polite">{!validName ? '使用小写字母开头，仅含小写字母、数字、短横线。' : selectedTargets.length === 0 ? '至少选择一种应用目标。' : ''}</div>
              {targetGroups.map(group => <fieldset key={group.title}><legend>{group.title}</legend>
                {group.options.map(([key, label]) => {
                  const family = stackFamilies.find(family => family.targets.includes(key));
                  const option = family.options.find(option => option[0] === stacks[family.key]);
                  return <label className="option-row" htmlFor={key} key={key}><span><strong>{label}</strong><small>{option[1]}</small></span><input id={key} type="checkbox" checked={targets[key]} onChange={() => toggleTarget(key)} /></label>;
                })}
                {stackFamilies.filter(family => group.stacks.includes(family.key) && family.targets.some(key => targets[key])).map(family => <div className="stack-field" key={family.key}>
                  <label className="field-label" htmlFor={family.key}>{family.label}</label>
                  <select id={family.key} value={stacks[family.key]} aria-describedby={`${family.key}-note`} onChange={event => setStacks(current => ({...current, [family.key]: event.target.value}))}>
                    {family.options.map(([value, label]) => <option key={value} value={value}>{label}</option>)}
                  </select>
                  <p id={`${family.key}-note`} className="stack-note">{family.options.find(option => option[0] === stacks[family.key])[2]}</p>
                </div>)}
              </fieldset>)}
              <fieldset><legend>设计文件</legend><label className="option-row" htmlFor="visual"><span><strong>视觉主题草稿</strong><small>产品方向文档与空白语义 Tokens</small></span><input id="visual" type="checkbox" checked={visual} onChange={event => setVisual(event.target.checked)} /></label></fieldset>
              <fieldset><legend>项目管理（可选）</legend><label className="option-row" htmlFor="openspec"><span><strong>OpenSpec 工作流</strong><small>需预先安装官方 CLI；生成 CodeBuddy 规范与变更工作流</small></span><input id="openspec" type="checkbox" checked={openSpec} onChange={event => setOpenSpec(event.target.checked)} /></label></fieldset>
            </div>
            <div className="config-preview">
              <div className="tool-heading"><span>CREATE COMMAND</span><span>{selectedTargets.length} TARGETS</span></div>
              <pre className="selected-command" aria-label="创建命令">{valid ? command : '完成项目名称和目标选择后显示命令'}</pre>
              <div className="tool-actions"><button className="button button-dark" type="button" onClick={() => copy(command)} disabled={!valid}>复制命令</button></div>
              <details className="config-json"><summary>查看 JSON 配置</summary><pre id="config" aria-label="项目配置 JSON">{config}</pre><div className="tool-actions"><button className="text-button" id="download" type="button" onClick={download} disabled={!valid}>下载 JSON</button><button className="text-button" id="copyconfig" type="button" onClick={() => copy(config)} disabled={!valid}>复制 JSON</button></div><p className="tool-command">本地执行 <code>aistd create --config project.json --yes</code></p></details>
            </div>
          </div>
        </div>
      </section>

      <section className="standards-section section-shell" id="standards" aria-labelledby="standards-title">
          <div className="section-heading reveal"><p className="overline">03 / THE FOUNDATION</p><h2 id="standards-title">一套工程语言，<br />一套视觉语言。</h2><p>两份规范独立维护，可用于新建项目和已有项目。</p></div>
        <div className="standard-row reveal"><span className="standard-index">01 / ENGINEERING</span><div><h3>工程规范</h3><p>架构边界、构建入口、契约、CI 与交付方式，随实际应用类型启用。</p></div><span className="standard-version">v1.0 ↗</span></div>
        <div className="standard-row reveal"><span className="standard-index">02 / VISUAL DESIGN</span><div><h3>视觉规范</h3><p>品牌方向、语义 Tokens、组件外观与屏幕验收，由产品自己定义视觉身份。</p></div><span className="standard-version">v1.0 ↗</span></div>
        <p className="standard-note">规范文件完整性可由 <code>aistd doctor</code> 校验；项目是否符合规范仍需按实际实现评审。</p>
      </section>

      <section className="docs-band" id="docs" aria-labelledby="docs-title">
        <div className="section-shell">
          <div className="section-heading reveal"><p className="overline">04 / REFERENCE</p><h2 id="docs-title">命令参考。</h2><p>安装、创建、接入、规范完整性校验与卸载的完整命令。OpenSpec 可选；需 Node.js 20.19+ 并先运行 npm install -g @fission-ai/openspec@latest。OpenSpec 管需求变更，task.md 管里程碑与验收。</p></div>
          <div className="docs-layout reveal">
            <div className="tablist" role="tablist" aria-label="CLI 使用文档">{tabs.map(([key, label], index) => <button key={key} className="tab" id={`tab-${key}`} role="tab" aria-controls="panel" aria-selected={activeDoc === key} tabIndex={activeDoc === key ? 0 : -1} type="button" onClick={() => setActiveDoc(key)} onKeyDown={event => selectTab(event, index)}><span>{String(index + 1).padStart(2, '0')}</span> {label}</button>)}</div>
            <div className="doc-body" id="panel" role="tabpanel" aria-labelledby={`tab-${activeDoc}`}>
              <h3>{doc.title}</h3><p>{doc.text}</p>
              <div className="code-block"><button className="copy-code" type="button" aria-label={`复制${doc.title}命令`} disabled={doc.copyable === false} onClick={() => copy(doc.code)}>复制</button><pre>{doc.code}</pre></div>
              <div className="doc-note">{doc.note}</div>
            </div>
          </div>
        </div>
      </section>

      <section className="release-band" aria-labelledby="release-title">
        <div className="section-shell release-inner">
          <div><p className="overline">CLI RELEASE / {cliVersion}</p><h2 id="release-title">一条命令，安装与升级。</h2><p>{installCommand ? '支持 macOS / Linux，自动校验 SHA256。已安装用户重复执行即可升级，无需手动卸载。' : '公开安装地址尚未配置。发布后从公网下载并校验对应平台的 CLI。'}</p></div>
          <div className="release-command"><code id="install-command">{installCommand || '安装地址尚未配置'}</code><button type="button" id="copyinstall" aria-label="复制安装命令" disabled={!installCommand} onClick={() => copy(installCommand)}>复制</button></div>
        </div>
      </section>
    </main>

    <footer className="site-footer"><div className="section-shell footer-inner"><a href="#top" className="footer-brand">aistd<span>.</span></a><span>工程规范 × 视觉规范 × 项目 CLI</span><span>v{cliVersion} · 2026</span></div></footer>
    <div role="status" aria-live="polite" className={`toast${toast ? ' is-visible' : ''}`} id="toast">{toast}</div>
  </>;
}
