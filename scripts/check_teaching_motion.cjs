// Run server first. PLAYWRIGHT_MODULE may point at an existing Playwright install.
const assert = require('node:assert/strict');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
(async () => {
  const browser = await chromium.launch({ executablePath: process.env.CHROMIUM || '/usr/bin/chromium', headless: true, args: ['--no-sandbox'] });
  try {
    const page = await browser.newPage();
    const errors = [];
    page.on('pageerror', error => errors.push(error.message));
    await page.goto(process.env.BASE_URL || 'http://localhost:8080');
    await page.waitForFunction(() => document.querySelector('#scrubber').max > 0);
    // Equal labels, different identities: node references survive an actual swap.
    const motion = await page.evaluate(() => {
      const stage = document.querySelector('#svg-stage');
      const scene = (a, b) => `<g data-key="a"><rect x="${a}" y="50" width="30" height="40"/><text x="${a}" y="40">4</text></g><g data-key="b"><rect x="${b}" y="50" width="30" height="40"/><text x="${b}" y="40">4</text></g>`;
      gorithmScene.render(stage, scene(20, 120), 0);
      const a = stage.querySelector('[data-key="a"]');
      gorithmScene.render(stage, scene(120, 20), 300);
      const same = a === stage.querySelector('[data-key="a"]');
      const moving = stage.getAnimations({subtree:true}).length;
      gorithmScene.cancel(stage);
      return { same, moving, cancelled: stage.getAnimations({subtree:true}).length, x: a.querySelector('rect').getAttribute('x') };
    });
    assert.equal(motion.same, true);
    assert.equal(motion.moving, 2);
    assert.equal(motion.cancelled, 0);
    assert.equal(motion.x, '120');
    await page.locator('#btn-reset').click();
    await page.locator('#btn-play').click();
    await page.waitForTimeout(80);
    await page.locator('#btn-play').click();
    const paused = await page.locator('#scrubber').inputValue();
    await page.waitForTimeout(650);
    assert.equal(await page.locator('#scrubber').inputValue(), paused, 'pause must cancel pending advance');
    await page.locator('#btn-play').click();
    await page.locator('#scrubber').evaluate(el => { el.value = '3'; el.dispatchEvent(new Event('input', { bubbles:true })); });
    await page.waitForTimeout(650);
    assert.equal(await page.locator('#scrubber').inputValue(), '3', 'seek must cancel playback');
    for (let i = 0; i < 3; i++) {
      await page.locator('#btn-play').click();
      await page.locator('#btn-play').click();
    }
    await page.waitForTimeout(650);
    assert.equal(await page.locator('#scrubber').inputValue(), '3', 'rapid play/pause must not leave old callbacks');
    await page.locator('#btn-next').click();
    assert.equal(await page.locator('#scrubber').inputValue(), '4');
    await page.keyboard.press('ArrowLeft');
    assert.equal(await page.locator('#scrubber').inputValue(), '3');
    await page.emulateMedia({ reducedMotion: 'reduce' });
    await page.locator('#btn-next').click();
    assert.equal(await page.locator('#svg-stage').evaluate(el => el.getAnimations({subtree:true}).length), 0);
    await page.locator('#btn-play').click();
    await page.locator('#btn-reset').click();
    await page.waitForTimeout(650);
    assert.equal(await page.locator('#scrubber').inputValue(), '0', 'reset must cancel playback');
    await page.setViewportSize({width:390, height:844});
    assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), true, 'mobile page overflow');
    assert.deepEqual(errors, []);
    console.log('PASS keyed duplicate identity, real motion, cancellation, seek, keyboard, reduced motion, mobile overflow');
  } finally { await browser.close(); }
})().catch(error => { console.error(error); process.exitCode = 1; });
