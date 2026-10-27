// Native keyed SVG reconciliation. Renderers own snapshots; motion never owns state.
(() => {
  const reducedMotion = matchMedia('(prefers-reduced-motion: reduce)');
  function cancel(stage) {
    for (const animation of stage.getAnimations({ subtree: true })) animation.cancel();
  }

  function patch(parent, source) {
    const old = [...parent.childNodes];
    const keyed = new Map(old.filter(n => n.nodeType === 1 && n.hasAttribute('data-key'))
      .map(n => [n.getAttribute('data-key'), n]));
    const used = new Set();
    for (const [index, next] of [...source.childNodes].entries()) {
      const key = next.nodeType === 1 ? next.getAttribute('data-key') : null;
      let node = key === null ? old[index] : keyed.get(key);
      if (!node || used.has(node) || node.nodeType !== next.nodeType || node.nodeName !== next.nodeName ||
          (key === null && node.nodeType === 1 && node.hasAttribute('data-key'))) {
        node = next.cloneNode(true);
      } else if (node.nodeType === 1) {
        for (const attr of [...node.attributes]) {
          if (!next.hasAttribute(attr.name)) node.removeAttribute(attr.name);
        }
        for (const attr of next.attributes) node.setAttribute(attr.name, attr.value);
        patch(node, next);
      } else {
        node.nodeValue = next.nodeValue;
      }
      used.add(node);
      parent.appendChild(node);
    }
    for (const node of old) if (!used.has(node)) node.remove();
  }

  function render(stage, markup, duration) {
    cancel(stage);
    const source = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
    source.innerHTML = markup;
    if (duration <= 0 || reducedMotion.matches) {
      // Still reconcile under reduced motion: identity is useful independently of motion.
      patch(stage, source);
      return;
    }
    const before = new Map([...stage.querySelectorAll('[data-key]')]
      .map(node => [node, node.getBoundingClientRect()]));
    patch(stage, source);
    for (const [node, box] of before) {
      if (!stage.contains(node)) continue;
      const after = node.getBoundingClientRect();
      // ponytail: bounds identify translation only for unchanged size. Snap
      // resizing objects; add explicit renderer anchors if resize + travel is needed.
      if (Math.abs(box.width - after.width) > 0.01 || Math.abs(box.height - after.height) > 0.01) continue;
      const matrix = node.parentElement.getScreenCTM();
      if (!matrix || matrix.a * matrix.d === matrix.b * matrix.c) continue;
      const inverse = matrix.inverse();
      const dx = box.x - after.x, dy = box.y - after.y;
      const x = inverse.a * dx + inverse.c * dy;
      const y = inverse.b * dx + inverse.d * dy;
      if (Math.abs(x) + Math.abs(y) < 0.1) continue;
      const transform = getComputedStyle(node).transform;
      const base = transform === 'none' ? '' : transform;
      node.animate([{ transform: `translate(${x}px, ${y}px) ${base}` },
        { transform: base || 'none' }], { duration, easing: 'ease-in-out' });
    }
  }
  reducedMotion.addEventListener('change', () => {
    const stage = document.getElementById('svg-stage');
    if (stage) cancel(stage);
  });
  window.gorithmScene = { render, cancel };
})();
