const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, 'navbar.test.cjs'), 'utf8');
const setup = source.slice(0, source.indexOf('(async()=>{'));
vm.runInNewContext(setup + `
(async()=>{
 const until=async(fn)=>{for(let i=0;i<150;i++){if(fn())return;await new Promise(r=>setTimeout(r,10))}throw Error('Timed out')};
 for(const file of ['berita.html','dokumentasi.html']){
  const t=render(file);t.flush();await until(()=>t.d.querySelector('.halaman-nav'));
  const activity=t.d.querySelector('[data-nav=aktivitas]');assert(activity.hasAttribute('data-aktif'));
  const next=t.d.querySelector('.halaman-nav button[aria-label="Halaman 2"]');next.focus();next.click();
  await until(()=>t.d.querySelector('.halaman-nav button[aria-current=page]')?.textContent==='2');
  assert.equal(t.d.activeElement.id,'hasil');assert.equal(new URLSearchParams(t.w.location.search).get('page'),'2');
  assert(t.d.querySelector('[role=status]').textContent.includes('Halaman 2'));
  const input=t.d.querySelector('#cari');input.value='zzzzzz';input.dispatchEvent(new t.w.Event('input',{bubbles:true}));
  await until(()=>t.d.querySelector('#hasil .keadaan'));
  assert.equal(new URLSearchParams(t.w.location.search).get('q'),'zzzzzz');assert.equal(new URLSearchParams(t.w.location.search).get('page'),null);
  assert(t.d.querySelector('[role=status]').textContent.includes('Tidak ada hasil'));
  t.w.history.replaceState(null,'',file+'?page=2');t.w.dispatchEvent(new t.w.PopStateEvent('popstate'));
  await until(()=>t.d.querySelector('.halaman-nav button[aria-current=page]')?.textContent==='2');assert.equal(input.value,'');
  assert(t.d.querySelector('.footer a[href^="mailto:"]'));t.w.close();
 }
 const t=render('index.html');t.flush();const media={},listeners={};let timers=new Map(),id=0;
 t.w.matchMedia=q=>media[q]??={matches:q.includes('hover'),addEventListener(_,f){listeners[q]=f},removeEventListener(){}};
 t.w.setInterval=f=>{timers.set(++id,f);return id};t.w.clearInterval=n=>timers.delete(n);
 const c=t.w.carousel(['one','two'],{label:'Test',intervalMs:10});t.d.body.append(c);await Promise.resolve();assert.equal(timers.size,1);
 media['(prefers-reduced-motion: reduce)'].matches=true;listeners['(prefers-reduced-motion: reduce)']();assert.equal(timers.size,0);
 media['(prefers-reduced-motion: reduce)'].matches=false;listeners['(prefers-reduced-motion: reduce)']();assert.equal(timers.size,1);t.w.close();
 console.log('PASS: activity navigation, pagination focus/status/URL, search, popstate, email, dynamic reduced motion');
})().catch(e=>{console.error(e);process.exitCode=1});
`, { require, console, __dirname, setTimeout, process, URLSearchParams });
