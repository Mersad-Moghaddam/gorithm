// Optional audit: external playwright and @axe-core/playwright installations.
import {createRequire} from 'node:module';
import assert from 'node:assert/strict';
const require=createRequire(import.meta.url);
const {chromium}=require('playwright');
const {default:AxeBuilder}=require('@axe-core/playwright');
const browser=await chromium.launch({executablePath:'/usr/bin/chromium',args:['--no-sandbox']});
const context=await browser.newContext();
const page=await context.newPage();
const base=process.env.APP_URL||'http://localhost:8091/static/index.html';
const failures=[];
try {
  await page.goto(base);
  await page.waitForSelector('.algorithm-card');
  for(const theme of ['lab','midnight','paper','plum','nord','amber','rose','monochrome']) {
    await page.evaluate(t=>document.documentElement.dataset.theme=t,theme);
    for(const width of [320,768,1440]) {
      await page.setViewportSize({width,height:900});
      if(!await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)) failures.push(`${theme}/${width}: library overflow`);
    }
    const results=await new AxeBuilder({page}).withTags(['wcag2a','wcag2aa','wcag21aa']).analyze();
    for(const v of results.violations) failures.push(`${theme}: ${v.id}: ${v.nodes.slice(0,3).map(n=>n.target.join(' ')).join(', ')}`);
  }
  const links=await page.locator('.algorithm-card').evaluateAll(els=>els.map(el=>el.href));
  for(const link of links){await page.goto(link);await page.waitForSelector('#workspace:visible');assert.ok(await page.locator('#svg-stage').innerHTML());}
  await page.goto(base+'#algorithm=merge');
  await page.waitForSelector('#workspace:visible');
  for(const theme of ['lab','midnight','paper','plum','nord','amber','rose','monochrome']) {
    await page.evaluate(t=>document.documentElement.dataset.theme=t,theme);
    await page.setViewportSize({width:320,height:900});
    if(!await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)) failures.push(`${theme}: workspace overflow`);
    const results=await new AxeBuilder({page}).withTags(['wcag2a','wcag2aa','wcag21aa']).analyze();
    for(const v of results.violations) failures.push(`${theme} workspace: ${v.id}: ${v.nodes.slice(0,3).map(n=>n.target.join(' ')).join(', ')}`);
  }
  console.log(failures.join('\n')||'PASS: eight-theme accessibility, responsive overflow, all 40 lessons');
  assert.deepEqual(failures,[]);
} finally {await browser.close();}
