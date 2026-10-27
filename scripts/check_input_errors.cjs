// Run against freshly built WASM and server; PLAYWRIGHT_MODULE may point at an existing install.
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
    await page.locator('#algo-search').fill('selection');
    await page.locator('#btn-algo-selection').click();
    const title = await page.locator('#viz-title').innerText();
    const before = await page.locator('#svg-stage').innerHTML();
    const count = await page.locator('#step-counter').innerText();
    await page.locator('#btn-play').click();
    await page.locator('#btn-custom-input').click();

    async function reject(input) {
      await page.locator('#custom-input-val').fill(input);
      await page.locator('#btn-modal-apply').click();
      assert.equal(await page.locator('#custom-modal').isVisible(), true, 'failed input must keep modal open');
      assert.equal(await page.locator('#modal-error').isVisible(), true, 'failed input must show error');
      assert.ok((await page.locator('#modal-error').innerText()).length > 0);
      assert.equal(await page.locator('#viz-title').innerText(), title);
      assert.equal(await page.locator('#step-counter').innerText(), count);
      assert.equal(await page.locator('#svg-stage').innerHTML(), before, 'failed input must not replace trace');
    }

    await reject('[3,');
    await reject('null');
    await reject('[3,"bad"]');
    await page.waitForTimeout(650);
    assert.equal(await page.locator('#step-counter').innerText(), count, 'failed input must cancel stale playback');
    await page.locator('#custom-input-val').fill('[3,1,2]');
    await page.locator('#btn-modal-apply').click();
    assert.equal(await page.locator('#custom-modal').isVisible(), false);
    assert.equal(await page.locator('#step-counter').innerText(), 'Step 1 of 17', 'valid input must replace prior trace');
    assert.deepEqual(errors, []);
    console.log('PASS malformed JSON, null, invalid selection input, preserved trace, valid recovery');
  } finally { await browser.close(); }
})().catch(error => { console.error(error); process.exitCode = 1; });
