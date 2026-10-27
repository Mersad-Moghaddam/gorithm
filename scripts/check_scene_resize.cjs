// Isolated real-browser regression; no server or WASM build required.
const assert = require('node:assert/strict');
const path = require('node:path');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
(async () => {
  const browser = await chromium.launch({ executablePath: process.env.CHROMIUM || '/usr/bin/chromium', headless: true, args: ['--no-sandbox'] });
  try {
    const page = await browser.newPage({ reducedMotion: 'no-preference' });
    await page.setContent('<svg id="svg-stage" width="800" height="450" viewBox="0 0 800 450"></svg>');
    await page.addScriptTag({ path: path.resolve(__dirname, '../web/static/scene.js') });
    const result = await page.evaluate(() => {
      const stage = document.querySelector('svg');
      const draw = (x, y, height, caption) => `<g data-key="array-slot-0"><rect x="${x}" y="${y}" width="40" height="${height}"/></g><text data-key="caption" x="400" y="24" text-anchor="middle">${caption}</text>`;
      gorithmScene.render(stage, draw(60, 70, 310, 'Short'), 0);
      const bar = stage.querySelector('g');
      gorithmScene.render(stage, draw(60, 225, 155, 'Much longer centered caption'), 300);
      const resizing = stage.getAnimations({subtree:true});
      for (const animation of resizing) { animation.pause(); animation.currentTime = 0; }
      const bottom = bar.getBoundingClientRect().bottom - stage.getBoundingClientRect().top;
      const resizeAnimations = resizing.length;
      gorithmScene.cancel(stage);
      gorithmScene.render(stage, draw(160, 225, 155, 'Much longer centered caption'), 300);
      const movement = stage.getAnimations({subtree:true});
      for (const animation of movement) { animation.pause(); animation.currentTime = 0; }
      return { resizeAnimations, bottom, moveAnimations: movement.length,
        same: bar === stage.querySelector('g'),
        displayedX: bar.getBoundingClientRect().left - stage.getBoundingClientRect().left,
        snapshotX: bar.querySelector('rect').getAttribute('x') };
    });
    assert.equal(result.resizeAnimations, 0, 'bar resizing and caption width changes must not invent translation');
    assert.equal(result.bottom, 380, 'resized bar must stay on fixed baseline');
    assert.equal(result.moveAnimations, 1, 'true equal-size translation must animate');
    assert.equal(result.same, true);
    assert.equal(result.displayedX, 60, 'movement starts at old position');
    assert.equal(result.snapshotX, '160', 'snapshot already owns final position');
    console.log('PASS fixed-baseline resizing, stationary caption width, retained identity, true translation');
  } finally { await browser.close(); }
})().catch(error => { console.error(error); process.exitCode = 1; });
