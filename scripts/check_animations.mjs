// NODE_PATH points at external Playwright. Synthetic SVG tests shared playback,
// not an algorithm implementation. Actual Bubble acceptance runs after fixtures.
import { createRequire } from 'node:module';
import assert from 'node:assert/strict';
const { chromium } = createRequire(import.meta.url)('playwright');
const browser = await chromium.launch({executablePath: process.env.CHROMIUM || '/usr/bin/chromium', headless: true, args: ['--no-sandbox']});
const page = await browser.newPage();
const errors = [];
page.on('pageerror', e => errors.push(e.message));
const base = process.env.APP_URL || 'http://localhost:8080';
const state = () => page.evaluate(() => ({
  index: document.querySelector('#scrubber').value,
  phase: document.querySelector('#svg-stage [data-phase]')?.getAttribute('data-phase'),
  source: document.querySelector('.code-line.active')?.id,
  metrics: [...document.querySelectorAll('#metric-comps,#metric-swaps,#metric-ops')].map(n => n.textContent),
}));
const stable = async () => { const before = await state(); await page.waitForTimeout(2150); assert.deepEqual(await state(), before); };
const running = () => page.evaluate(() => document.querySelector('#svg-stage').getAnimations({subtree:true}).filter(a => a.playState === 'running').length);
try {
  await page.goto(base + '#algorithm=bubble');
  await page.waitForSelector('#workspace:visible');
  await page.waitForFunction(() => !document.querySelector('#btn-next').disabled);
  // Install test snapshots through the actual RenderStep replacement path.
  await page.evaluate(() => {
    const stage = document.querySelector('#svg-stage');
    const descriptor = Object.getOwnPropertyDescriptor(Element.prototype, 'innerHTML');
    let phase = 0;
    const fixture = () => {
      const x = phase % 2 ? 300 : 100;
      return `<g data-lesson="fixture" data-phase="${phase}" data-scene-width="800" data-scene-height="450"><title>Shared motion fixture</title><g data-entity="main:0" data-x="${x}" data-y="180" data-motion="swap-up" transform="translate(${x} 180)"><rect width="50" height="50" fill="#c45d75"/><text y="70">7 · ID 0</text></g><g data-entity="main:1" data-x="${400-x}" data-y="180" data-motion="swap-down" transform="translate(${400-x} 180)"><rect width="50" height="50" fill="#387faa"/><text y="70">7 · ID 1</text></g></g>`;
    };
    descriptor.set.call(stage, fixture());
    Object.defineProperty(stage, 'innerHTML', {get() { return descriptor.get.call(this); }, set(_value) { phase++; descriptor.set.call(this, window.animationFragment ?? fixture()); }});
  });
  const position = () => page.locator('[data-entity="main:0"]').evaluate(n => { const r=n.getBoundingClientRect(); return [r.x + window.scrollX, r.y + window.scrollY]; });
  const identity = () => page.locator('#svg-stage [data-entity]').evaluateAll(nodes => nodes.map(n => [n.dataset.entity,n.querySelector('rect').getAttribute('fill')]));
  const ids = await identity();
  assert.notEqual(ids[0][0], ids[1][0]); assert.notEqual(ids[0][1], ids[1][1]);
  for (const scale of [.5, 1, 1.5]) {
    await page.locator('#svg-stage').evaluate((n,s) => { n.style.width=`${800*s}px`; n.style.minWidth='0'; }, scale);
    const start = await position();
    await page.locator('#btn-next').click();
    assert.ok(await running() > 0, 'Next animates matching identities');
    assert.ok(Math.abs((await page.locator('#svg-stage').boundingBox()).width-800*scale)<1,'requested rendered SVG scale');
    const samples = await page.locator('[data-entity="main:0"]').evaluate(async n => {
      const animation = n.getAnimations()[0];
      animation.pause(); await animation.ready;
      const duration = animation.effect.getTiming().duration;
      const sample = fraction => {
        animation.currentTime = duration * fraction;
        const r = n.getBoundingClientRect(); return [r.x + window.scrollX, r.y + window.scrollY];
      };
      const points = [sample(0),sample(.5)];
      animation.cancel();
      const r = n.getBoundingClientRect(); points.push([r.x + window.scrollX, r.y + window.scrollY]);
      return points;
    });
    const near = (got,want,label) => assert.ok(Math.abs(got-want)<1,`${label}: ${got} vs ${want}`);
    near(samples[0][0],start[0],'scaled motion begins at source x');
    near(samples[0][1],start[1],'scaled motion begins at source y');
    near(Math.abs(samples[2][0]-start[0]),200*scale,'scaled final displacement');
    near(samples[1][0],(start[0]+samples[2][0])/2,'scaled arc midpoint x');
    near(samples[1][1],start[1]-64*scale,'scaled arc midpoint y');
    assert.deepEqual(await identity(),ids);
  }
  await page.locator('#btn-next').click(); await page.waitForTimeout(150);
  // Play then Pause also settles motion started by manual Next.
  await page.locator('#btn-play').click(); await page.locator('#btn-play').click();
  assert.equal(await running(),0); await stable();
  for (const action of ['seek','reset','dialog','speed','rapid']) {
    await page.locator('#btn-next').click(); await page.waitForTimeout(150);
    if (action === 'seek') await page.locator('#scrubber').evaluate(n => { n.value='2'; n.dispatchEvent(new Event('input',{bubbles:true})); });
    if (action === 'reset') await page.locator('#btn-reset').click();
    if (action === 'dialog') {
      await page.locator('#btn-custom-input').click();
      await page.locator('#custom-input-val').fill('[7,7,2]');
      await page.locator('#btn-modal-apply').click();
    }
    if (action === 'speed') await page.locator('#speed-select').selectOption('2');
    if (action === 'rapid') {
      await page.locator('#btn-play').click(); await page.waitForTimeout(50);
      await page.locator('#btn-play').click(); await page.locator('#btn-play').click();
      await page.waitForTimeout(50); await page.locator('#btn-play').click();
    }
    assert.equal(await running(),0, action); await stable();
  }
  await page.locator('#speed-select').selectOption('1');
  await page.locator('#btn-reset').click();
  await page.locator('#btn-play').click();
  await page.waitForTimeout(200);
  const beforeSpeed = await state();
  await page.locator('#speed-select').selectOption('2');
  assert.deepEqual(await state(),beforeSpeed,'speed change does not advance');
  await page.waitForTimeout(550);
  assert.deepEqual(await state(),beforeSpeed,'speed change restarts current hold');
  await page.waitForTimeout(450);
  assert.notDeepEqual(await state(),beforeSpeed,'new speed advances after new hold');
  await page.locator('#btn-play').click(); await stable();
  await page.locator('#speed-select').selectOption('1');
  await page.locator('#btn-next').click();
  await page.emulateMedia({reducedMotion:'reduce'}); await page.waitForTimeout(50);
  assert.equal(await running(),0,'preference change cancels active motion');
  await page.locator('#btn-next').click(); assert.equal(await running(),0);
  await page.emulateMedia({reducedMotion:'no-preference'});
  await page.locator('#scrubber').evaluate(n => { n.value=String(Number(n.max)-1); n.dispatchEvent(new Event('input',{bubbles:true})); });
  await page.locator('#btn-play').click();
  await page.waitForTimeout(1950);
  assert.equal(await page.locator('#scrubber').inputValue(),await page.locator('#scrubber').getAttribute('max'));
  assert.ok(await running()>0,'final arrival remains animated during its hold');
  await page.waitForTimeout(1900);
  assert.match(await page.locator('#btn-play').getAttribute('aria-label'),/^Play/);
  await page.locator('#btn-play').click(); assert.equal(await page.locator('#scrubber').inputValue(),'0');
  await page.locator('#btn-play').click();
  const entity = (id,x,y,extra='') => `<g data-entity="${id}" data-x="${x}" data-y="${y}" transform="translate(${x} ${y})" ${extra}><rect width="30" height="30" fill="#c45d75"/><rect data-compare-outline="" width="32" height="32" fill="none" stroke="#ffe29a"/></g>`;
  const scene = (content,extra='') => `<g data-lesson="fixture" data-phase="metadata" ${extra}>${content}</g>`;
  const transition = async (before,after) => {
    await page.locator('#btn-reset').click();
    await page.evaluate(({before,after}) => {
      Object.getOwnPropertyDescriptor(Element.prototype,'innerHTML').set.call(document.querySelector('#svg-stage'),before);
      window.animationFragment=after;
    },{before,after});
    await page.locator('#btn-next').click();
  };
  for (const metadata of ['data-motion="linear"','data-motion="lift" data-clear-y="40"','data-motion="path" data-path="[[100,180],[200,40],[300,180]]"']) {
    await transition(scene(entity('source',100,180)),scene(entity('copy',300,180,`data-from-entity="source" ${metadata}`)));
    assert.ok(await running()>0,metadata);
    const duration=await page.locator('#svg-stage').evaluate(n=>n.getAnimations({subtree:true})[0].effect.getTiming().duration);
    assert.equal(duration,750);
  }
  for (const metadata of ['data-motion="path" data-path="[[0,0],[300,180]]"','data-motion="path" data-path="[[100,180],[1],[300,180]]"','data-motion="lift" data-clear-y="NaN"','data-motion="unknown"']) {
    await transition(scene(entity('same',100,180)),scene(entity('same',300,180,metadata)));
    assert.equal(await running(),0,metadata);
  }
  const fix1Failures = [];
  for (const [startX,endX,path,expected] of [
    [100,300,'[[100,180],[null,40],[300,180]]',false],
    [100,300,'[[100,180],[200,null],[300,180]]',false],
    [100,300,'[[100,180],["0",40],[300,180]]',false],
    [100,300,'[[100,180],[false,40],[300,180]]',false],
    [0,300,'[[null,180],[100,40],[300,180]]',false],
    [100,0,'[[100,180],[100,40],[null,180]]',false],
    [0,300,'[[0,180],[0,40],[300,180]]',true],
    [100,0,'[[100,180],[0,40],[0,180]]',true],
  ]) {
    await transition(scene(entity('same',startX,180)),scene(entity('same',endX,180,`data-motion="path" data-path="${path.replaceAll('"','&quot;')}"`)));
    if ((await running()>0) !== expected) fix1Failures.push(`path ${path}: expected animation ${expected}`);
  }
  await transition(scene(entity('same',100,180)),scene(entity('same',100,180,'data-emphasis="compare"')));
  assert.ok(await running()>0,'compare animates decorative outline');
  assert.equal(await page.locator('#svg-stage [data-compare-outline]:not([data-motion-outline])').count(),1,'final outline stays visible');
  await page.locator('#btn-prev').click(); assert.equal(await running(),0);
  await page.setViewportSize({width:320,height:800});
  await transition(scene(''),scene(entity('new',100,180),'data-scene-width="1000" data-scene-height="600" data-min-width="1000"'));
  assert.equal(await running(),0,'new identity snaps');
  assert.equal(await page.locator('#svg-stage').getAttribute('viewBox'),'0 0 1000 600');
  assert.equal(await page.locator('.canvas-viewport').getAttribute('tabindex'),'0');
  assert.ok(await page.locator('.canvas-viewport').getAttribute('aria-label'));
  assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));
  const viewport = page.locator('.canvas-viewport');
  await viewport.evaluate(n=>{ n.scrollLeft=100; n.focus(); });
  const focusedState = await state();
  const scrollBefore = await viewport.evaluate(n=>n.scrollLeft);
  await page.keyboard.press('ArrowRight'); await page.waitForTimeout(250);
  const scrollRight = await viewport.evaluate(n=>n.scrollLeft);
  if (!(scrollRight>scrollBefore)) fix1Failures.push('focused ArrowRight must scroll right');
  if (JSON.stringify(await state())!==JSON.stringify(focusedState)) fix1Failures.push('focused ArrowRight changed playback state');
  await page.keyboard.press('ArrowLeft'); await page.waitForTimeout(250);
  if (!((await viewport.evaluate(n=>n.scrollLeft))<scrollRight)) fix1Failures.push('focused ArrowLeft must scroll left');
  if (JSON.stringify(await state())!==JSON.stringify(focusedState)) fix1Failures.push('focused ArrowLeft changed playback state');
  await page.keyboard.press('Space'); await page.waitForTimeout(100);
  if (!(await page.locator('#btn-play').getAttribute('aria-label')).startsWith('Play')) fix1Failures.push('focused Space toggled playback');
  await page.locator('#btn-reset').click();
  await page.locator('#viz-title').focus();
  const shortcutIndex = Number((await state()).index);
  await page.keyboard.press('ArrowRight');
  assert.equal(Number((await state()).index),shortcutIndex+1,'workspace ArrowRight outside canvas still advances');
  assert.deepEqual(fix1Failures,[],'F1/F2 focused regressions');
  await transition(scene(''),scene('','data-scene-width="NaN" data-scene-height="0" data-min-width="99999"'));
  assert.equal(await page.locator('#svg-stage').getAttribute('viewBox'),'0 0 800 450');
  assert.equal(await page.locator('#svg-stage').evaluate(n=>n.style.minWidth),'0px');
  await page.evaluate(()=>{ window.animationFragment='<rect width="20" height="20"/>'; document.querySelector('#svg-stage').style.width='100%'; });
  await page.locator('#btn-next').click();
  assert.equal(await page.locator('#svg-stage').getAttribute('class'),'');
  assert.equal(await page.locator('.canvas-viewport').getAttribute('tabindex'),null);
  await page.evaluate(()=>{ window.animationFragment=null; });
  await page.locator('#btn-next').click();
  await page.locator('#back-library').click();
  await page.locator('#card-selection').click();
  await page.waitForFunction(() => document.querySelector('#viz-title').textContent.includes('Selection'));
  await stable();
  assert.equal(await running(),0);

  // Reload removes synthetic innerHTML interception; all following states come
  // from Bubble.Build through real Go/WASM dispatch, never JS algorithm logic.
  await page.goto(base + '#algorithm=bubble');
  await page.reload();
  await page.waitForSelector('#workspace:visible');
  await page.waitForFunction(() => !document.querySelector('#btn-next').disabled);
  const input = async value => {
    await page.locator('#btn-custom-input').click();
    await page.locator('#custom-input-val').fill(value);
    await page.locator('#btn-modal-apply').click();
    await page.waitForSelector('#custom-modal:not([open])',{state:'attached'});
  };
  const seek = async index => page.locator('#scrubber').evaluate((n,index) => {
    n.value=String(index); n.dispatchEvent(new Event('input',{bubbles:true}));
  },index);
  const entities = () => page.locator('#svg-stage [data-entity]').evaluateAll(nodes => nodes.map(n => ({
    id:n.dataset.entity, color:n.querySelector('rect').getAttribute('fill'),
    value:n.querySelector('text').textContent,
    checked:[...n.querySelectorAll('text')].some(t=>t.textContent==='✓'),
  })));
  const assertStep = async (index,phase,line,metrics) => {
    const actual=await state();
    assert.equal(Number(actual.index),index);
    assert.equal(actual.phase,phase);
    assert.ok(actual.source?.endsWith(line),`${actual.source} must select ${line}`);
    if (metrics) assert.deepEqual(actual.metrics,metrics.map(String));
    assert.equal(await page.locator('#svg-stage [data-lesson="bubble"]').count(),1);
  };
  await page.setViewportSize({width:1440,height:1000});
  await input('[3,1,2]');
  await seek(3);
  await assertStep(3,'compare','L5',[1,0,1]);
  assert.match(await page.locator('#step-explanation').innerText(),/Comparing arr\[0\] \(3\) and arr\[1\] \(1\)/);
  const sourcePositions=await page.locator('#svg-stage [data-entity]').evaluateAll(nodes=>Object.fromEntries(nodes.map(n=>{
    const r=n.getBoundingClientRect(); return [n.dataset.entity,[r.x + window.scrollX, r.y + window.scrollY]];
  })));
  await page.locator('#btn-next').click();
  await assertStep(4,'swap','L6',[1,1,2]);
  assert.match(await page.locator('#step-explanation').innerText(),/Swapped 3 and 1 because 3 > 1/);
  assert.deepEqual((await entities()).map(n=>n.id),['main:1','main:0','main:2']);
  const actualMotion=await page.locator('#svg-stage [data-motion]').evaluateAll(async nodes=>{
    const result=[];
    for (const n of nodes) {
      const a=n.getAnimations()[0];
      if (!a) throw new Error(`missing animation ${n.dataset.entity}`);
      a.pause(); await a.ready;
      const duration=a.effect.getTiming().duration;
      const sample=fraction=>{a.currentTime=duration*fraction;const r=n.getBoundingClientRect();return [r.x + window.scrollX, r.y + window.scrollY];};
      const start=sample(0), middle=sample(.5);
      a.cancel(); const r=n.getBoundingClientRect();
      result.push({id:n.dataset.entity,motion:n.dataset.motion,duration,scale:n.getScreenCTM().a,start,middle,end:[r.x + window.scrollX, r.y + window.scrollY]});
    }
    return result;
  });
  for (const motion of actualMotion) {
    assert.equal(motion.duration,720,'1200ms commit caps movement at 60%');
    motion.start.forEach((value,i)=>assert.ok(Math.abs(value-sourcePositions[motion.id][i])<1,'actual swap starts at previous identity'));
    assert.ok(Math.abs(motion.middle[0]-(motion.start[0]+motion.end[0])/2)<1);
    assert.ok(Math.abs(motion.middle[1]-motion.start[1]-(motion.motion==='swap-up'?-64:24)*motion.scale)<1);
    assert.notDeepEqual(motion.middle,motion.end);
  }
  await seek(7); await assertStep(7,'confirm','L8',[2,2,4]);
  assert.equal((await entities()).filter(n=>n.checked).length,1);
  assert.equal((await entities()).find(n=>n.checked).id,'main:0');
  await seek(11); await assertStep(11,'done','L10',[3,2,5]);
  assert.deepEqual((await entities()).map(n=>n.value),['1','2','3']);
  assert.ok((await entities()).every(n=>n.checked));

  await input('[2,2,1]');
  const duplicateColors=Object.fromEntries((await entities()).map(n=>[n.id,n.color]));
  assert.notEqual(duplicateColors['main:0'],duplicateColors['main:1']);
  await seek(Number(await page.locator('#scrubber').getAttribute('max')));
  const duplicates=await entities();
  assert.deepEqual(duplicates.map(n=>n.id),['main:2','main:0','main:1']);
  assert.deepEqual(duplicates.map(n=>n.value),['1','2','2']);
  duplicates.forEach(n=>assert.equal(n.color,duplicateColors[n.id]));

  for (const action of ['pause','seek','prev','reset','dialog','apply','speed','rapid','navigate','reduced']) {
    await input('[3,1,2]'); await seek(3);
    await page.locator('#btn-next').click();
    assert.ok(await running()>0,`actual swap starts before ${action}`);
    if (action==='pause') { await page.locator('#btn-play').click(); await page.locator('#btn-play').click(); }
    if (action==='seek') await seek(2);
    if (action==='prev') await page.locator('#btn-prev').click();
    if (action==='reset') await page.locator('#btn-reset').click();
    if (action==='dialog' || action==='apply') {
      await page.locator('#btn-custom-input').click();
      assert.equal(await running(),0,'opening actual data dialog cancels swap');
      if (action==='apply') {
        await page.locator('#custom-input-val').fill('[null]');
        await page.locator('#btn-modal-apply').click();
        assert.ok(await page.locator('dialog[open]').count(),'invalid input keeps dialog open');
        await assertStep(4,'swap','L6',[1,1,2]);
        await page.locator('#custom-input-val').fill('[2,2,1]');
        await page.locator('#btn-modal-apply').click();
        await assertStep(0,'ready','L1',[0,0,0]);
      } else await page.keyboard.press('Escape');
    }
    if (action==='speed') await page.locator('#speed-select').selectOption('2');
    if (action==='rapid') {
      await page.locator('#btn-play').click(); await page.waitForTimeout(50);
      await page.locator('#btn-play').click(); await page.locator('#btn-play').click();
      await page.waitForTimeout(50); await page.locator('#btn-play').click();
    }
    if (action==='navigate') {
      await page.locator('#back-library').click();
      await page.locator('#card-selection').click();
    }
    if (action==='reduced') {
      await page.emulateMedia({reducedMotion:'reduce'});
      await page.waitForTimeout(100);
    }
    assert.equal(await running(),0,`actual swap cancellation: ${action}`);
    await stable();
    if (action==='navigate') {
      await page.locator('#back-library').click(); await page.locator('#card-bubble').click();
    }
    await page.emulateMedia({reducedMotion:'no-preference'});
    await page.locator('#speed-select').selectOption('1');
  }
  await input('[3,1,2]'); await seek(3);
  await page.emulateMedia({reducedMotion:'reduce'});
  await page.locator('#btn-next').click(); assert.equal(await running(),0);
  await assertStep(4,'swap','L6',[1,1,2]);
  await page.emulateMedia({reducedMotion:'no-preference'});

  const themes=['lab','midnight','paper','plum','nord','amber','rose','monochrome'];
  for (const width of [320,390,768,1440]) {
    await page.setViewportSize({width,height:1000});
    await seek(3);
    assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),`Bubble document overflow at ${width}`);
    const geometry=await page.locator('.canvas-viewport').evaluate(n=>({width:n.clientWidth,scroll:n.scrollWidth}));
    if (width===320) {
      assert.ok(geometry.scroll>geometry.width,'Bubble stays readable through local scrolling');
      assert.equal(await page.locator('.canvas-viewport').getAttribute('tabindex'),'0');
      const before=await state();
      await page.locator('.canvas-viewport').evaluate(n=>{n.scrollLeft=0;n.focus();});
      await page.keyboard.press('ArrowRight'); await page.waitForTimeout(250);
      assert.ok(await page.locator('.canvas-viewport').evaluate(n=>n.scrollLeft>0));
      assert.deepEqual(await state(),before,'real Bubble scroll does not seek');
      await page.locator('.canvas-viewport').evaluate(n=>{n.scrollLeft=0;});
    }
    for (const theme of themes) {
      await page.evaluate(theme=>document.documentElement.dataset.theme=theme,theme);
      assert.ok(await page.locator('#svg-stage text').evaluateAll(nodes=>nodes.every(n=>{
        const scale=n.getScreenCTM().a;return parseFloat(getComputedStyle(n).fontSize)*scale>=12;
      })),'actual labels retain 12px minimum');
    }
    await page.evaluate(()=>document.documentElement.dataset.theme='lab');
    await page.screenshot({path:`/tmp/opencode/bubble-pilot-${width}.png`,fullPage:true});
  }
  assert.deepEqual(errors,[]);
  console.log('PASS shared animation fixtures plus actual Bubble source/phase/results, duplicates, deterministic arcs, cancellation, eight themes, responsive scrolling');
} finally { await browser.close(); }
