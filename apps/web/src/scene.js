export function startScene(canvas) {
  const context = canvas.getContext('2d');
  if (!context) return () => {};

  const reducedMotion = matchMedia('(prefers-reduced-motion: reduce)');
  const colors = getComputedStyle(document.documentElement);
  const line = colors.getPropertyValue('--scene-line').trim();
  const text = colors.getPropertyValue('--scene-text').trim();
  const accent = colors.getPropertyValue('--color-accent').trim();
  let frame = 0;
  let visible = true;
  let pointerShift = 0;

  function draw(time = 0) {
    const width = canvas.clientWidth;
    const height = canvas.clientHeight;
    const ratio = Math.min(devicePixelRatio || 1, 2);
    if (canvas.width !== Math.round(width * ratio) || canvas.height !== Math.round(height * ratio)) {
      canvas.width = Math.round(width * ratio);
      canvas.height = Math.round(height * ratio);
    }
    context.setTransform(ratio, 0, 0, ratio, 0, 0);
    context.clearRect(0, 0, width, height);
    const mobile = width < 760;
    const scale = Math.min(1.06, width / 1350);
    const x = Math.max(width * .49, width - 690) + pointerShift;
    const y = height * .17;
    const start = mobile
      ? document.querySelector('.hero-status').getBoundingClientRect().bottom - canvas.getBoundingClientRect().top + 20
      : 0;

    context.lineWidth = 1;
    context.strokeStyle = line;
    context.globalAlpha = mobile ? .34 : .44;
    for (let gx = mobile ? 20 : x - 40 * scale; gx < width; gx += 56 * scale) {
      context.beginPath(); context.moveTo(gx, start); context.lineTo(gx, height); context.stroke();
    }
    for (let gy = start; gy < height; gy += 56 * scale) {
      context.beginPath(); context.moveTo(mobile ? 0 : x - 40 * scale, gy); context.lineTo(width, gy); context.stroke();
    }
    context.globalAlpha = 1;

    if (mobile && height - start < 175) return;
    const nodes = mobile ? [
      [32, start, 'engineering-v1.0.zip', false],
      [32, start + 42, 'visual-v1.0.zip', false],
      [146, start + 94, 'aistd create', true],
      [72, start + 148, '.aistd/standards', false],
      [158, start + 178, 'standards.lock.json', false]
    ] : [
      [x + 20 * scale, y, 'engineering-v1.0.zip', false],
      [x + 20 * scale, y + 82 * scale, 'visual-v1.0.zip', false],
      [x + 250 * scale, y + 148 * scale, 'aistd create', true],
      [x + 130 * scale, y + 260 * scale, '.aistd/standards', false],
      [x + 315 * scale, y + 325 * scale, 'standards.lock.json', false]
    ];
    const size = mobile ? 20 : 27 * scale;
    [[0, 2], [1, 2], [2, 3], [2, 4]].forEach(([from, to], index) => {
      const [sx, sy] = nodes[from];
      const [ex, ey] = nodes[to];
      context.beginPath();
      context.moveTo(sx + size, sy + size / 2);
      context.lineTo(ex + size / 2, sy + size / 2);
      context.lineTo(ex + size / 2, ey);
      context.strokeStyle = line;
      context.globalAlpha = .9;
      context.stroke();
      if (!mobile && !reducedMotion.matches) {
        const progress = (time / 2600 + index * .22) % 1;
        context.fillStyle = accent;
        context.fillRect(sx + (ex - sx) * progress, sy + size / 2 - 2, 4, 4);
      }
    });
    context.globalAlpha = 1;
    nodes.forEach(([nx, ny, label, active]) => {
      context.strokeStyle = active ? accent : text;
      context.strokeRect(nx, ny, size, size);
      context.fillStyle = active ? accent : text;
      context.fillRect(nx + size * .35, ny + size * .35, size * .3, size * .3);
      context.font = `${mobile ? 10 : Math.max(10, Math.round(12 * scale))}px ui-monospace, SFMono-Regular, Menlo, monospace`;
      context.fillText(label, nx + size + (mobile ? 8 : 12 * scale), ny + size * .69);
    });
  }

  function animate(time) {
    draw(time);
    frame = visible && canvas.clientWidth >= 760 && !reducedMotion.matches ? requestAnimationFrame(animate) : 0;
  }

  function resume() {
    cancelAnimationFrame(frame);
    if (visible && canvas.clientWidth >= 760 && !reducedMotion.matches) frame = requestAnimationFrame(animate);
    else draw();
  }

  function move(event) {
    if (canvas.clientWidth >= 760 && !reducedMotion.matches) pointerShift = (event.clientX / canvas.clientWidth - .5) * 18;
  }

  const observer = new IntersectionObserver(entries => {
    visible = entries[0].isIntersecting;
    resume();
  });
  observer.observe(canvas);
  canvas.parentElement.addEventListener('pointermove', move);
  window.addEventListener('resize', resume);
  reducedMotion.addEventListener('change', resume);
  resume();
  return () => {
    observer.disconnect();
    canvas.parentElement.removeEventListener('pointermove', move);
    window.removeEventListener('resize', resume);
    reducedMotion.removeEventListener('change', resume);
    cancelAnimationFrame(frame);
  };
}
