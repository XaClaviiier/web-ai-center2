const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, 'navbar.test.cjs'), 'utf8');
const setup = source.slice(0, source.indexOf('(async()=>{'));
vm.runInNewContext(setup + `
(async()=>{
 const until=async(fn)=>{for(let i=0;i<150;i++){if(fn())return;await new Promise(r=>setTimeout(r,10))}throw Error('Timed out')};
 const contactPages=['index.html','pelatihan.html','inference.html','jejak-alumni.html','perencanaan-implan-gigi.html','document-intelligence.html','tim.html','tentang-kami.html'];
 for(const file of contactPages){
  const t=render(file);t.flush();const sections=t.d.querySelectorAll('section.kontak');assert.equal(sections.length,1,file);
  const links=[...sections[0].querySelectorAll('a')];assert.equal(links.length,2,file);
  const [wa,email]=links;assert.equal(wa.textContent,'Hubungi lewat WhatsApp');assert.equal(email.textContent,'Kirim email');
  const url=new t.w.URL(wa.href);assert.equal(url.origin,'https://wa.me');assert.equal(url.pathname,'/62895634222240');assert(url.searchParams.get('text'));
  const mail=new t.w.URL(email.href);assert.equal(mail.protocol,'mailto:');assert.equal(mail.pathname,'aicenter@unit.ubaya.ac.id');assert(mail.searchParams.get('subject'));
  assert.equal(wa.target,'_blank');assert(wa.rel.includes('noopener'));assert(!email.hasAttribute('target'));
  for(const a of links){assert.equal(a.tabIndex,0);assert(a.querySelector('svg[aria-hidden=true]'));a.focus();assert.equal(t.d.activeElement,a)}
  if(file==='tentang-kami.html')assert(!sections[0].querySelector('.tentang-layanan'));
  assert.equal(t.errors.length,0,file);t.w.close();
 }
 console.log('PASS: delapan CTA dengan pasangan kontak, label seragam, nomor/email resmi proyek, ikon dan fokus native');
 for(const file of ['berita.html','dokumentasi.html']){
  const t=render(file);t.flush();await until(()=>t.d.querySelector('.halaman-nav'));
  const activity=t.d.querySelector('[data-nav=kabar]');assert(activity.hasAttribute('data-aktif'));
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

 const team=render('tim.html');team.flush();await until(()=>team.d.querySelectorAll('.kartu-tim').length===6);
 assert(!team.d.querySelector('input[type=search]'));assert.equal(team.d.querySelectorAll('.kartu-tim__riset').length,6);
 for(const card of team.d.querySelectorAll('.kartu-tim')){assert.equal(card.querySelector('img').getAttribute('width'),'96');assert(card.querySelector('.kartu-tim__riset').textContent.includes('Bidang Penelitian'))}
 assert(team.d.querySelector('.footer').textContent.includes('Gedung Fakultas Teknik TA'));assert(!team.d.querySelector('.footer').textContent.includes('TA 1.2'));
 for(const state of ['kosong','error','normal']){team.d.querySelector('[data-nilai="'+state+'"]').click();await until(()=>state==='normal'?team.d.querySelectorAll('.kartu-tim').length===6:!!team.d.querySelector('.daftar--tim .keadaan'));}
 assert.equal(team.errors.length,0);team.w.close();
 const mock=require('node:vm').runInNewContext(fs.readFileSync('prototype/assets/mock-data.js','utf8')+'; MOCK');
 const news=render('berita-detail.html',{hash:'?slug='+encodeURIComponent(mock.berita.find(b=>b.status==='posted'&&!b.deleted_at).slug)});news.flush();await until(()=>news.d.querySelector('.artikel__meta'));
 assert(news.d.querySelector('.artikel__meta').textContent.includes('oleh Tim Media Ubaya AI Center'));assert.equal(news.errors.length,0);news.w.close();
 const album=mock.dokumentasi.find(d=>d.status==='posted'&&!d.deleted_at);
 const doc=render('dokumentasi-detail.html',{hash:'?slug='+encodeURIComponent(album.slug)});doc.flush();await until(()=>doc.d.querySelector('.galeri img'));
 assert.equal(doc.d.querySelectorAll('.galeri img').length,mock.dokumentasi_gambar.filter(g=>g.dokumentasi_id===album.id).length);assert.equal(doc.errors.length,0);doc.w.close();
 console.log('PASS: Tim tanpa pencarian, enam profil/riset, simulasi kosong/error/pulih, alamat footer, penulis berita dan foto album');
 console.log('PASS: activity navigation, pagination focus/status/URL, search, popstate, email, dynamic reduced motion');
})().catch(e=>{console.error(e);process.exitCode=1});
`, { require, console, __dirname, setTimeout, process, URLSearchParams });

// Canvas diuji tanpa rendering browser: rekam jumlah garis setiap kelompok dan frame.
{
 const assert = require('node:assert/strict');
 const frames = [], kelompok = new Set();
 let garis = 0, garisFrame = 0, garisMaks = 0;
 const ctx = {
  setTransform(){}, clearRect(){garisMaks=Math.max(garisMaks,garisFrame);garisFrame=0},
  beginPath(){garis=0}, moveTo(){}, arc(){}, fill(){},
  lineTo(){garis++;garisFrame++}, stroke(){kelompok.add(garis)},
 };
 const canvas = {getContext:()=>ctx,clientWidth:1000,clientHeight:500};
 const context = {
  performance:{now:()=>0}, getComputedStyle:()=>({getPropertyValue:()=>''}),
  matchMedia:q=>({matches:q.includes('hover'),addEventListener(){},removeEventListener(){}}),
  window:{devicePixelRatio:1}, document:{hidden:false,addEventListener(){},removeEventListener(){}},
  requestAnimationFrame:fn=>{frames.push(fn);return frames.length}, cancelAnimationFrame(){},setTimeout,clearTimeout,
  IntersectionObserver:class{observe(){}disconnect(){}},
  ResizeObserver:class{constructor(fn){this.fn=fn}observe(){this.fn()}disconnect(){}},
 };
 vm.createContext(context);
 const neural = vm.runInContext(fs.readFileSync(path.join(__dirname,'../assets/js/bola-neural.js'),'utf8')+';BolaNeural',context);
 const kontrol = neural.pasang(canvas,{area:{addEventListener(){},removeEventListener(){}}});
 for(let frame=1;frame<=1600;frame++)frames.shift()?.(frame*50);
 assert.deepEqual([...kelompok].sort(),[2,3,4]);assert(garisMaks<=8);assert.equal(kontrol.info.partikel,300);
 kontrol.henti();
 console.log('PASS: canvas 1600 frame, kelompok 3/4/5 dot, maksimal delapan garis per frame');
}
