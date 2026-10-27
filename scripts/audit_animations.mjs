import { createRequire } from 'node:module';
import assert from 'node:assert/strict';

const require = createRequire(import.meta.url);
const { chromium } = require('playwright');

const browser = await chromium.launch({
  executablePath: process.env.CHROMIUM || '/usr/bin/chromium',
  args: ['--no-sandbox']
});

const page = await browser.newPage();
const baseUrl = process.env.APP_URL || 'http://localhost:8080/';

const errors = [];
page.on('pageerror', err => {
  errors.push(`Page error: ${err.message}`);
});
page.on('console', msg => {
  if (msg.type() === 'error') {
    errors.push(`Console error: ${msg.text()}`);
  }
});

const failures = [];

try {
  await page.goto(baseUrl);
  await page.waitForSelector('.algorithm-card', { timeout: 10000 });

  const cards = await page.locator('.algorithm-card').evaluateAll(els =>
    els.map(el => ({
      href: el.href,
      id: el.href.split('#algorithm=')[1],
      name: el.querySelector('h3')?.textContent?.trim() || ''
    }))
  );

  console.log(`Discovered ${cards.length} algorithms in catalog.`);
  assert.equal(cards.length, 40, 'Expected 40 registered algorithms in catalog');

  for (let i = 0; i < cards.length; i++) {
    const card = cards[i];
    const prefix = `[${i + 1}/40] ${card.id}`;
    process.stdout.write(`Auditing ${prefix}... `);

    await page.goto(card.href);
    await page.waitForSelector('#workspace:visible', { timeout: 5000 });

    // 1. Title verification
    const title = await page.locator('#viz-title').innerText();
    if (!title || title.trim().length === 0) {
      failures.push(`${prefix}: missing title`);
    }

    // 2. Initial state verification
    const counterText = await page.locator('#step-counter').innerText();
    const match = counterText.match(/Step\s+(\d+)\s+of\s+(\d+)/i);
    if (!match) {
      failures.push(`${prefix}: invalid step counter "${counterText}"`);
      console.log('FAILED (bad counter)');
      continue;
    }

    const currentStep = parseInt(match[1], 10);
    const totalSteps = parseInt(match[2], 10);

    if (totalSteps <= 0) {
      failures.push(`${prefix}: total steps is ${totalSteps}`);
    }

    // Scrubber sync
    const scrubberMax = await page.locator('#scrubber').getAttribute('max');
    if (parseInt(scrubberMax, 10) !== totalSteps - 1) {
      failures.push(`${prefix}: scrubber max (${scrubberMax}) does not match total steps - 1 (${totalSteps - 1})`);
    }

    // SVG presence & sanity check (no NaN attributes)
    const svgHTML = await page.locator('#svg-stage').innerHTML();
    if (!svgHTML || svgHTML.trim().length < 20) {
      failures.push(`${prefix}: empty or missing SVG stage`);
    }
    if (svgHTML.includes('NaN') || svgHTML.includes('undefined')) {
      failures.push(`${prefix}: SVG stage contains NaN or undefined values`);
    }

    // Explanation
    const explanation = await page.locator('#step-explanation').innerText();
    if (!explanation || explanation.trim().length === 0) {
      failures.push(`${prefix}: missing step explanation`);
    }

    // Action pill
    const action = await page.locator('#viz-action').innerText();
    if (!action || action.trim().length === 0) {
      failures.push(`${prefix}: missing action pill`);
    }

    // Source code highlight
    const hasHighlight = await page.evaluate(() => {
      return document.querySelectorAll('.code-line.active').length > 0;
    });
    if (!hasHighlight) {
      failures.push(`${prefix}: no active line highlighted in source code`);
    }

    // Step to midpoint
    if (totalSteps > 2) {
      const mid = Math.floor(totalSteps / 2);
      await page.evaluate(val => {
        const scrubber = document.getElementById('scrubber');
        scrubber.value = val;
        scrubber.dispatchEvent(new Event('input', { bubbles: true }));
      }, mid);

      await page.waitForTimeout(60);

      const midCounter = await page.locator('#step-counter').innerText();
      const midMatch = midCounter.match(/Step\s+(\d+)\s+of\s+(\d+)/i);
      if (!midMatch || parseInt(midMatch[1], 10) !== mid + 1) {
        failures.push(`${prefix}: seeking to midpoint ${mid} failed, counter shows "${midCounter}"`);
      }

      const midSvg = await page.locator('#svg-stage').innerHTML();
      if (midSvg.includes('NaN') || midSvg.includes('undefined')) {
        failures.push(`${prefix}: SVG stage contains NaN at midpoint step ${mid}`);
      }
    }

    // Step to final step
    const lastStep = totalSteps - 1;
    await page.evaluate(val => {
      const scrubber = document.getElementById('scrubber');
      scrubber.value = val;
      scrubber.dispatchEvent(new Event('input', { bubbles: true }));
    }, lastStep);

    await page.waitForTimeout(60);

    const lastCounter = await page.locator('#step-counter').innerText();
    const lastMatch = lastCounter.match(/Step\s+(\d+)\s+of\s+(\d+)/i);
    if (!lastMatch || parseInt(lastMatch[1], 10) !== lastStep + 1) {
      failures.push(`${prefix}: seeking to last step ${lastStep} failed, counter shows "${lastCounter}"`);
    }

    // Step backward
    if (totalSteps > 1) {
      await page.locator('#btn-prev').click();
      await page.waitForTimeout(40);
      const prevCounter = await page.locator('#step-counter').innerText();
      const prevMatch = prevCounter.match(/Step\s+(\d+)\s+of\s+(\d+)/i);
      if (!prevMatch || parseInt(prevMatch[1], 10) !== lastStep) {
        failures.push(`${prefix}: btn-prev from step ${lastStep} did not go to ${lastStep}`);
      }
    }

    // Reset button
    await page.locator('#btn-reset').click();
    await page.waitForTimeout(40);
    const resetCounter = await page.locator('#step-counter').innerText();
    const resetMatch = resetCounter.match(/Step\s+(\d+)\s+of\s+(\d+)/i);
    if (!resetMatch || parseInt(resetMatch[1], 10) !== 1) {
      failures.push(`${prefix}: btn-reset did not return to step 0 (display Step 1), counter is "${resetCounter}"`);
    }

    console.log(`OK (${totalSteps} steps)`);
  }

  if (errors.length > 0) {
    failures.push(...errors);
  }

  console.log('\n--- AUDIT SUMMARY ---');
  if (failures.length === 0) {
    console.log('ALL 40 ALGORITHM ANIMATIONS VERIFIED CORRECT AND PASSING!');
  } else {
    console.error(`FAILURES (${failures.length}):`);
    for (const f of failures) {
      console.error(`  - ${f}`);
    }
  }

  assert.equal(failures.length, 0, `Animation audit had ${failures.length} failures.`);
} finally {
  await browser.close();
}
