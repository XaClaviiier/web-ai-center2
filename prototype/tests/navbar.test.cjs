const path = require("node:path");
const root = path.resolve(__dirname, "../..");
const nativeFs = require("node:fs");
const fs = {
  readFileSync: (p, ...args) => nativeFs.readFileSync(path.join(root, p), ...args),
  readdirSync: (p, ...args) => nativeFs.readdirSync(path.join(root, p), ...args),
};
const assert = require("node:assert/strict");
const { JSDOM, VirtualConsole } = require("jsdom");
function render(file,{wide=true,hash="",reduced=false}={}){
 const errors=[],vc=new VirtualConsole();vc.on("jsdomError",e=>errors.push(e.message));
 const dom=new JSDOM(fs.readFileSync("prototype/public/"+file,"utf8"),{url:"http://localhost/prototype/public/"+file+hash,runScripts:"dangerously",pretendToBeVisual:true,virtualConsole:vc});
 const w=dom.window,frames=[],media={};w.requestAnimationFrame=f=>{frames.push(f);return frames.length};w.matchMedia=q=>media[q]??={matches:q==="(min-width: 80rem)"?wide:q==="(prefers-reduced-motion: reduce)"?reduced:false,addEventListener(){},removeEventListener(){}};
 Object.defineProperty(w,"innerWidth",{value:wide?1440:390,writable:true});Object.defineProperty(w,"scrollY",{value:0,writable:true});w.CSS={supports:()=>false};w.HTMLElement.prototype.scrollIntoView=function(opts){this.lastScroll=opts};w.BolaNeural={pasang(){}};
 const paths=[...w.document.querySelectorAll("script[src]")].map(s=>"prototype/"+s.getAttribute("src").replace("../",""));
 w.eval(paths.map(path=>path.endsWith("bola-neural.js")?"const BolaNeural={pasang(){}};":fs.readFileSync(path,"utf8")).join("\n"));
 return {w,d:w.document,errors,media,flush(){for(let n=0;frames.length&&n<10;n++){const batch=frames.splice(0);batch.forEach(f=>f())}}};
}
(async()=>{
 const files=fs.readdirSync("prototype/public").filter(f=>f.endsWith(".html"));
 for(const f of files){const t=render(f);t.flush();const nav=t.d.querySelector("header.nav");assert(nav,f);assert.equal(nav.querySelectorAll("[data-nav]").length,6,f);assert.equal(nav.querySelector('[data-nav="contoh-aplikasi"]').getAttribute("href"),"inference.html#portofolio");assert.equal(new URL(nav.querySelector(".nav__cta").href).pathname,"/62895634222240");assert.equal(nav.querySelectorAll('#menu-layanan').length,1);assert.equal(nav.querySelectorAll('#menu-kabar').length,1);assert.equal(nav.querySelector('[data-nav="kabar"]').textContent,'Kabar');assert(!nav.querySelector('[data-nav="profil"]'));assert.equal(t.errors.length,0,f+": "+t.errors.join());t.w.close();}
 const t=render("inference.html",{hash:"#portofolio"});await Promise.resolve();t.flush();const {d,w}=t,nav=d.querySelector("header.nav"),app=nav.querySelector('[data-nav="contoh-aplikasi"]'),service=nav.querySelector('[data-nav="layanan"]');
 assert.equal(app.getAttribute("aria-current"),"location");assert(!service.hasAttribute("data-aktif"));assert.equal(d.querySelectorAll("#portofolio").length,1);assert.equal(d.querySelector("#portofolio").lastScroll.behavior,"instant");
 w.history.replaceState(null,"","inference.html");w.dispatchEvent(new w.HashChangeEvent("hashchange"));assert(service.hasAttribute("data-aktif"));assert(!app.hasAttribute("aria-current"));
 w.scrollY=110;w.dispatchEvent(new w.Event("scroll"));t.flush();assert(nav.classList.contains("nav--ringkas"));
 w.scrollY=70;w.dispatchEvent(new w.Event("scroll"));t.flush();assert(nav.classList.contains("nav--ringkas"));
 w.scrollY=20;w.dispatchEvent(new w.Event("scroll"));t.flush();assert(!nav.classList.contains("nav--ringkas"));
 service.click();assert.equal(service.getAttribute("aria-expanded"),"true");assert(!d.querySelector("#menu-layanan").hidden);
 w.scrollY=200;w.dispatchEvent(new w.Event("scroll"));t.flush();assert(!nav.classList.contains("nav--ringkas"));
 service.dispatchEvent(new w.KeyboardEvent("keydown",{key:"Escape",bubbles:true}));assert(d.querySelector("#menu-layanan").hidden);assert.equal(d.activeElement,service);
 service.blur();d.body.click();t.flush();assert(nav.classList.contains("nav--ringkas"));w.close();
 const m=render("inference.html",{wide:false});m.flush();const mn=m.d.querySelector("header.nav"),burger=mn.querySelector(".nav__burger");assert(mn.classList.contains("nav--mobile"));burger.querySelector("path").dispatchEvent(new m.w.MouseEvent("click",{bubbles:true}));assert.equal(burger.getAttribute("aria-expanded"),"true");assert(mn.classList.contains("nav--menu-terbuka"));mn.querySelector('[data-nav="layanan"]').click();assert(!m.d.querySelector("#menu-layanan").hidden);mn.querySelector('[data-nav="contoh-aplikasi"]').click();m.flush();assert(!mn.classList.contains("nav--menu-terbuka"));assert.equal(burger.getAttribute("aria-expanded"),"false");assert.equal(m.d.activeElement,m.d.querySelector("#portofolio"));m.w.close();

 for(const f of files){
  const raw=fs.readFileSync('prototype/public/'+f,'utf8'),dom=new JSDOM(raw),d=dom.window.document;
  const fallback=d.querySelector('.nav-sans-js');assert(fallback.querySelector('a[href="inference.html#portofolio"]'));
  assert([...fallback.querySelectorAll('summary')].some(n=>n.textContent==='Kabar'));assert(!raw.includes('<summary>Profil</summary>'));
  for(const page of ['tim.html','tentang-kami.html'])assert(fallback.querySelector('ul > li > a[href="'+page+'"]'));
  dom.window.close();
 }
 for(const file of ['tim.html','tentang-kami.html'])for(const wide of [true,false]){
  const p=render(file,{wide});p.flush();const head=p.d.querySelector('header.nav'),key=file.replace('.html','');
  assert.equal(head.querySelector('[data-nav="'+key+'"]').getAttribute('aria-current'),'page');
  assert(!head.querySelector('[data-nav="layanan"]').hasAttribute('data-aktif'));
  if(!wide)head.querySelector('.nav__burger').click();
  head.querySelector('[data-nav="kabar"]').click();assert(!head.querySelector('#menu-kabar').hidden);
  head.querySelector('[data-nav="layanan"]').click();assert(head.querySelector('#menu-kabar').hidden);
  p.w.close();
 }
 // Regresi: fokus atau dropdown tidak boleh menahan navbar ringkas di atas halaman.
 for(const dropdown of [false,true]){
  const p=render('inference.html');p.flush();const h=p.d.querySelector('header.nav'),button=h.querySelector('[data-nav="layanan"]');
  p.w.scrollY=200;p.w.dispatchEvent(new p.w.Event('scroll'));p.flush();assert(h.classList.contains('nav--ringkas'));
  button.focus();if(dropdown)button.click();p.w.scrollY=0;p.w.dispatchEvent(new p.w.Event('scroll'));p.flush();
  assert(!h.classList.contains('nav--ringkas'));assert.equal(p.d.activeElement,button);
  if(dropdown)assert(!h.querySelector('#menu-layanan').hidden);p.w.close();
 }
 // Posisi nyata digantikan geometry deterministik untuk pengujian scrollspy dua arah.
 for(const page of ['index.html','inference.html']){
  const p=render(page);p.flush();const sections=[...p.d.querySelectorAll('[data-nav-section]')];
  sections.forEach((section,i)=>section.getBoundingClientRect=()=>({top:i*500-p.w.scrollY,height:400}));
  const check=(y,key)=>{p.w.scrollY=y;p.w.dispatchEvent(new p.w.Event('scroll'));p.flush();const active=p.d.querySelectorAll('.nav__tautan[aria-current],.nav__tautan[data-aktif]');assert.equal(active.length,1);assert.equal(active[0].dataset.nav,key)};
  check(0,page==='index.html'?'index':'layanan');
  check(500,page==='index.html'?'layanan':'contoh-aplikasi');
  check(1100,page==='index.html'?'kabar':'contoh-aplikasi');
  check(1700,page==='index.html'?'kabar':'contoh-aplikasi');
  check(500,page==='index.html'?'layanan':'contoh-aplikasi');check(0,page==='index.html'?'index':'layanan');
  if(page==='inference.html'){
   const a=p.d.querySelector('[data-nav="contoh-aplikasi"]'),target=p.d.querySelector('#portofolio');
   a.dispatchEvent(new p.w.MouseEvent('click',{bubbles:true,cancelable:true,detail:1}));p.flush();assert.equal(target.dataset.fokusPointer,'ya');
   target.dispatchEvent(new p.w.KeyboardEvent('keydown',{key:'Tab',bubbles:true}));assert(!target.hasAttribute('data-fokus-pointer'));
   a.click();p.flush();assert(!target.hasAttribute('data-fokus-pointer'));assert.equal(p.d.activeElement,target);
  }
  p.w.close();
 }
 const fit=render("inference.html");fit.flush();const fn=fit.d.querySelector("header.nav"),fm=fn.querySelector(".nav__menu");let needed=1400;Object.defineProperty(fm,"scrollWidth",{get:()=>needed});fit.w.dispatchEvent(new fit.w.Event("resize"));fit.flush();assert(fn.classList.contains("nav--mobile"));needed=500;fit.w.dispatchEvent(new fit.w.Event("resize"));fit.flush();assert(!fn.classList.contains("nav--mobile"));fit.w.close();
 const motion=render("inference.html");motion.flush();const h=motion.d.querySelector("header.nav"),surface=h.querySelector(".nav__permukaan"),played=[];
 assert.equal(surface.parentElement,h.querySelector(".nav__dalam"));assert.equal(surface.getAttribute("aria-hidden"),"true");
 const parts=[h.querySelector(".nav__logo"),...h.querySelector(".nav__menu").children,h.querySelector(".nav__cta"),surface];
 parts.forEach((part,i)=>{part.getBoundingClientRect=()=>({left:(h.classList.contains("nav--ringkas")?100:20)+i*50,top:0,width:part===surface?900:70,height:44});part.animate=(frames,options)=>{const animation={cancelled:false,cancel(){this.cancelled=true}};played.push({frames,options,animation});return animation}});
 motion.w.scrollY=200;motion.w.dispatchEvent(new motion.w.Event("scroll"));motion.flush();assert(h.classList.contains("nav--ringkas"));assert.equal(played.length,parts.length);assert.equal(played[0].options.duration,360);assert(played[0].frames[0].transform.includes("-80px"));assert.equal(played.at(-1).frames[1].opacity,1);
 const first=played.map(a=>a.animation);motion.w.scrollY=0;motion.w.dispatchEvent(new motion.w.Event("scroll"));motion.flush();assert(!h.classList.contains("nav--ringkas"));assert(first.every(a=>a.cancelled));assert.equal(played.at(-1).frames[1].opacity,0);
 const count=played.length;motion.media["(prefers-reduced-motion: reduce)"].matches=true;motion.w.scrollY=200;motion.w.dispatchEvent(new motion.w.Event("scroll"));motion.flush();assert(h.classList.contains("nav--ringkas"));assert.equal(played.length,count);motion.w.close();
 console.log("PASS: 13 halaman/nav+CTA; hash aktif dan direct-load; scrollspy dua arah; navbar kembali di atas saat fokus/dropdown; Escape/fokus; menu mobile dan anchor; 13 fallback tanpa JS.");
 console.log("PASS: satu permukaan bersama; FLIP dua arah dan pembatalan animasi; reduced motion tanpa animasi spasial.");
})().catch(error => { console.error(error); process.exitCode = 1; });
