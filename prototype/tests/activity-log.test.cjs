const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const { JSDOM, VirtualConsole } = require('jsdom');
const root = path.resolve(__dirname, '..');
const read = (file) => fs.readFileSync(path.join(root, file), 'utf8');
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
async function until(fn) {
  for (let n = 0; n < 150; n++) { if (fn()) return; await sleep(10); }
  throw Error('Batas waktu menunggu render terlampaui');
}

async function apiTests() {
  let now = '2026-11-10T16:59:00Z';
  const context = vm.createContext({ setTimeout, Intl, Date: class extends Date {
    constructor(...args) { super(...(args.length ? args : [now])); }
  } });
  vm.runInContext(read('assets/mock-data.js') + read('assets/js/format.js') + read('simulasi/api-simulasi.js') + ';this.test={api,SIMULASI,Format,MOCK};', context);
  const { api, SIMULASI, Format } = context.test;
  SIMULASI.jedaMs = 0;
  const rejected = (promise, status) => assert.rejects(promise, (e) => e.status === status);
  await rejected(api.adminLog(), 401);
  await rejected(api.adminTambah('tim', { nama: 'X' }), 401);
  const account = api.masuk('tanpa-role'); account.role = 'Direktur';
  await rejected(api.adminLog(), 403);
  await rejected(api.adminTambah('tim', { nama: 'X' }), 403);
  api.masuk('direktur');
  assert.equal((await api.adminLog()).total, 0);
  const secret = 'NEVER-COPY-THIS-TOKEN';
  for (const source of ['berita', 'dokumentasi', 'testimoni', 'tim']) {
    const item = await api.adminTambah(source, { judul: 'Workshop AI', nama: 'Anggota Uji', gambar_daftar: [{ key: 'foto.jpg' }], username: 'palsu', isi: secret, teks: secret });
    let log = (await api.adminLog()).items[0];
    assert.equal(log.jenis, 'tambah'); assert.equal(log.modul, source); assert.equal(log.username, 'dian.direktur');
    assert(!JSON.stringify(log).includes(secret));
    await api.adminUbah(source, item.id, { judul: 'Revisi Workshop', gambar_daftar: [{ key: 'baru.jpg' }] }, item.version);
    assert.equal((await api.adminLog()).items[0].aksi, 'ubah');
    if (source !== 'tim') {
      await api.adminPost(source, item.id); await api.adminRetract(source, item.id);
      await api.adminHapus(source, item.id); await api.adminPulihkan(source, item.id);
      await api.adminHapusPermanen(source, item.id);
      assert.equal((await api.adminLog()).items[0].label_item, 'Revisi Workshop');
    } else await api.adminHapus(source, item.id);
    await rejected(api.adminDetail(source, item.id), 404);
  }
  const all = await api.adminLog(); assert.equal(all.total, 24); assert.equal(all.items.length, 10);
  assert(all.items.every((b, i, a) => !i || a[i - 1].id > b.id));
  const page2 = await api.adminLog({ page: 2 }); assert.equal(page2.items.length, 10);
  assert(all.items[9].id > page2.items[0].id);
  const last = await api.adminLog({ page: 999 }); assert.equal(last.page, 3); assert.equal(last.items.length, 4);
  all.items[0].username = 'modifié'; assert.notEqual((await api.adminLog()).items[0].username, 'modifié');
  assert.equal((await api.adminLog({ jenis: 'tambah' })).total, 4);
  assert.equal((await api.adminLog({ jenis: 'hapus' })).total, 7);
  assert.equal((await api.adminLog({ username: 'palsu' })).total, 0);
  const before = (await api.adminLog()).total;
  await rejected(api.adminHapus('tim', 99999), 404);
  await rejected(api.adminHapusPermanen('berita', 99999), 404);
  await rejected(api.adminTambah('berita', { judul: 'x'.repeat(71) }), 400);
  await rejected(api.adminTambah('dokumentasi', { judul: 'Tanpa gambar' }), 400);
  await rejected(api.adminTambah('akun', { nama: 'X' }), 404);
  const draft = (await api.adminDaftar('berita')).items.find((b) => b.status === 'draft' && !b.dikunci_oleh);
  SIMULASI.paksa409 = true;
  await rejected(api.adminUbah('berita', draft.id, { judul: 'Konflik' }, draft.version), 409);
  const locked = (await api.adminDaftar('berita')).items.find((b) => b.dikunci_oleh);
  await rejected(api.adminHapus('berita', locked.id), 423);
  for (const [state, status] of [['error', 500], ['429', 429]]) {
    SIMULASI.keadaan = state; await rejected(api.adminTambah('tim', { nama: 'X' }), status);
  }
  SIMULASI.keadaan = 'normal'; SIMULASI.paksa401 = true;
  await rejected(api.adminTambah('tim', { nama: 'X' }), 401);
  SIMULASI.paksa401 = false;
  assert.equal((await api.adminLog()).total, before);
  now = '2026-11-10T17:00:00Z'; api.masuk('pegawai');
  await api.adminTambah('tim', { nama: '<img src=x onerror=alert(1)>' });
  assert.equal((await api.adminLog({ dari: '2026-11-11', sampai: '2026-11-11' })).total, 1);
  assert.equal((await api.adminLog({ sampai: '2026-11-10' })).total, 24);
  assert.equal((await api.adminLog({ username: 'pandu.pegawai', jenis: 'tambah', dari: '2026-11-11' })).total, 1);
  await rejected(api.adminLog({ dari: '2026-11-12', sampai: '2026-11-11' }), 400);
  assert.equal(Format.waktuJakarta(now), '11 Nov 2026, 00:00 WIB');
  assert.equal(Format.hariJakarta(now), '2026-11-11');
  const log = (await api.adminLog()).items[0];
  assert.equal(log.username, 'pandu.pegawai'); assert.equal(log.pengguna_sub, 'sub-pegawai');
  assert.equal(typeof api.adminLogUbah, 'undefined'); assert.equal(typeof api.adminLogHapus, 'undefined');
  SIMULASI.keadaan = 'kosong'; assert.equal((await api.adminLog()).total, 0);
  SIMULASI.keadaan = 'normal'; assert.equal((await api.adminLog()).total, before + 1);
  console.log('PASS: seluruh mutasi, identitas, salinan baca-saja, galat 400/401/403/404/409/423/429/500, pagination dan batas hari Jakarta');
}

function render() {
  const errors = [], vc = new VirtualConsole(); vc.on('jsdomError', (e) => errors.push(e.message));
  const dom = new JSDOM(read('admin/index.html'), { url: 'http://localhost/prototype/admin/index.html?akun=direktur#/log-aktivitas', runScripts: 'dangerously', pretendToBeVisual: true, virtualConsole: vc });
  const w = dom.window;
  // JSDOM tidak mengimplementasikan dialog modal native; hanya shim siklus buka/tutup.
  w.HTMLDialogElement.prototype.showModal = function () { this.setAttribute('open', ''); };
  w.HTMLDialogElement.prototype.close = function () { this.removeAttribute('open'); };
  const scripts = [...w.document.querySelectorAll('script[src]')].map((s) => read(s.getAttribute('src').replace('../', '')));
  w.eval(scripts.join('\n') + ';SIMULASI.jedaMs=0;window.test={api,SIMULASI};');
  return { w, d: w.document, errors };
}

async function uiTests() {
  const { w, d, errors } = render();
  try {
    await until(() => d.querySelector('.keadaan h3'));
    assert.equal(d.querySelector('.keadaan h3').textContent, 'Belum ada aktivitas');
    assert.equal(d.querySelector('.samping [aria-current=page]').textContent, 'Log Aktivitas');
    assert(d.querySelector('.log-simulasi').textContent.includes('refresh'));
    assert(d.querySelector('.sim-buka').textContent.includes('Simulasi'));
    const { api } = w.test;
    for (let i = 0; i < 12; i++) await api.adminTambah('tim', { nama: i ? `Tim ${i}` : '<img src=x onerror=alert(1)>' });
    const form = d.querySelector('.log-filter');
    const submit = () => form.dispatchEvent(new w.Event('submit', { bubbles: true, cancelable: true }));
    submit(); await until(() => d.querySelectorAll('.log-tabel tbody tr').length === 10);
    assert.equal(d.querySelectorAll('.log-tabel th').length, 3);
    assert.equal(d.querySelectorAll('.log-tabel button, .log-tabel a').length, 0);
    assert.equal(d.querySelector('.log-tabel-wadah').tabIndex, 0);
    d.querySelector('[aria-label="Halaman 2"]').click();
    await until(() => d.querySelectorAll('.log-tabel tbody tr').length === 2);
    assert(d.querySelector('.log-tabel').textContent.includes('<img src=x onerror=alert(1)>'));
    assert.equal(d.querySelectorAll('.log-tabel img').length, 0);
    assert.equal(d.activeElement.id, 'hasil-log');
    d.querySelector('#log-jenis').value = 'hapus'; submit();
    await until(() => d.querySelector('.keadaan h3')?.textContent === 'Tidak ada aktivitas yang cocok');
    [...form.querySelectorAll('button')].find((b) => b.textContent === 'Reset filter').click();
    await until(() => d.querySelectorAll('.log-tabel tbody tr').length === 10);
    d.querySelector('#log-dari').value = '2026-11-12'; d.querySelector('#log-sampai').value = '2026-11-11'; submit();
    assert.equal(d.querySelector('#log-sampai').validity.customError, true);
    d.querySelector('#log-dari').dispatchEvent(new w.Event('input'));
    assert.equal(d.querySelector('#log-sampai').validity.customError, false);
    w.test.SIMULASI.keadaan = 'error'; w.document.dispatchEvent(new w.CustomEvent('simulasi:ubah'));
    await until(() => d.querySelector('.keadaan--galat'));
    w.test.SIMULASI.keadaan = 'normal'; d.querySelector('.keadaan button').click();
    await until(() => d.querySelector('.log-tabel'));
    w.location.hash = '#/berita';
    await until(() => d.querySelector('.kepala h1')?.textContent === 'Berita' && d.querySelector('tbody'));
    // Membatalkan hapus tidak boleh menambah log.
    const before = (await api.adminLog()).total;
    d.querySelector('button[aria-label^="Hapus:"]:not(:disabled)').click();
    await until(() => d.querySelector('dialog[open]'));
    [...d.querySelectorAll('dialog button')].find((b) => b.textContent === 'Batal').click();
    assert.equal((await api.adminLog()).total, before);
    assert.equal(errors.length, 0, errors.join('\n'));
  } finally { w.close(); }
  const fresh = render();
  try { await until(() => fresh.d.querySelector('.keadaan h3')); assert.equal((await fresh.w.test.api.adminLog()).total, 0); }
  finally { fresh.w.close(); }
  console.log('PASS: UI baca-saja, empty state, filter/reset, pagination/fokus, XSS sebagai teks, galat/retry, batal hapus, refresh kosong');
}

async function browserTests() {
  const { spawn } = require('node:child_process');
  const { pathToFileURL } = require('node:url');
  const os = require('node:os');
  const chrome = 'C:/Program Files/Google/Chrome/Application/chrome.exe';
  const profile = fs.mkdtempSync(path.join(os.tmpdir(), 'aicenter-log-check-'));
  const child = spawn(chrome, ['--headless=new', '--no-first-run', '--no-default-browser-check', '--remote-debugging-port=0', `--user-data-dir=${profile}`, 'about:blank'], { windowsHide: true, stdio: ['ignore', 'ignore', 'pipe'] });
  let socket;
  try {
    const endpoint = await new Promise((resolve, reject) => {
      let output = '';
      const timer = setTimeout(() => reject(Error('Chrome CDP tidak merespons')), 15000);
      child.on('error', (e) => { clearTimeout(timer); reject(e); });
      child.stderr.on('data', (data) => {
        output += data.toString(); const match = output.match(/DevTools listening on (ws:\/\/[^\s]+)/);
        if (match) { clearTimeout(timer); resolve(match[1]); }
      });
    });
    socket = new WebSocket(endpoint);
    await new Promise((resolve, reject) => { socket.addEventListener('open', resolve, { once: true }); socket.addEventListener('error', reject, { once: true }); });
    let next = 0; const pending = new Map(), exceptions = [];
    socket.addEventListener('message', (event) => {
      const data = JSON.parse(event.data);
      if (data.method === 'Runtime.exceptionThrown') exceptions.push(data.params.exceptionDetails.text);
      if (!data.id) return;
      const request = pending.get(data.id); if (!request) return;
      clearTimeout(request.timer); pending.delete(data.id);
      if (data.error) request.reject(Error(data.error.message)); else request.resolve(data.result);
    });
    const send = (method, params = {}, sessionId) => new Promise((resolve, reject) => {
      const id = ++next;
      const timer = setTimeout(() => { pending.delete(id); reject(Error(`CDP timeout: ${method}`)); }, 15000);
      pending.set(id, { resolve, reject, timer }); socket.send(JSON.stringify({ id, method, params, ...(sessionId ? { sessionId } : {}) }));
    });
    const { targetId } = await send('Target.createTarget', { url: 'about:blank' });
    const { sessionId } = await send('Target.attachToTarget', { targetId, flatten: true });
    const cdp = (method, params) => send(method, params, sessionId);
    const evaluate = async (expression) => {
      const result = await cdp('Runtime.evaluate', { expression, awaitPromise: true, returnByValue: true });
      if (result.exceptionDetails) throw Error(JSON.stringify(result.exceptionDetails));
      return result.result.value;
    };
    await cdp('Page.enable'); await cdp('Runtime.enable');
    const waitFor = (condition) => evaluate(`(async()=>{for(let n=0;n<300;n++){if(${condition})return true;await new Promise(r=>setTimeout(r,20));}throw Error('Browser render timeout: '+location.href+' '+document.body?.innerText.slice(0,1000))})()`);
    const screenshots = path.join(root, 'tests/.browser-check'); fs.mkdirSync(screenshots, { recursive: true });
    const report = [];
    for (const width of [1440, 820, 390]) {
      await cdp('Emulation.setDeviceMetricsOverride', { width, height: 1000, deviceScaleFactor: 1, mobile: false });
      // Query berbeda memaksa dokumen baru; perpindahan hash saja mempertahankan data simulasi.
      await cdp('Page.navigate', { url: pathToFileURL(path.join(root, 'admin/index.html')).href + `?akun=direktur&check=${width}#/log-aktivitas` });
      await waitFor("document.querySelector('.keadaan h3')");
      assert.equal(await evaluate("document.querySelector('.keadaan h3').textContent"), 'Belum ada aktivitas');
      await evaluate(`(async()=>{SIMULASI.jedaMs=0;for(let i=0;i<12;i++)await api.adminTambah('tim',{nama:i?'Anggota Tim '+i:'Nama anggota sangat panjang '+ 'x'.repeat(200)});document.querySelector('.log-filter').requestSubmit()})()`);
      await waitFor("document.querySelectorAll('.log-tabel tbody tr').length===10");
      await evaluate('document.fonts.ready');
      const layout = await evaluate(`(()=>{const table=document.querySelector('.log-tabel-wadah');return {width:innerWidth,pageWidth:document.documentElement.scrollWidth,tableScrollable:table.scrollWidth>table.clientWidth,columns:document.querySelectorAll('.log-tabel th').length,actions:document.querySelectorAll('.log-tabel button,.log-tabel a').length}})()`);
      assert(layout.pageWidth <= width, JSON.stringify(layout)); assert.equal(layout.columns, 3); assert.equal(layout.actions, 0);
      if (width === 390) assert(layout.tableScrollable);
      await evaluate('window.scrollTo(0,0)');
      const { cssContentSize } = await cdp('Page.getLayoutMetrics');
      const shot = await cdp('Page.captureScreenshot', { format: 'png', captureBeyondViewport: true, clip: { x: 0, y: 0, width: cssContentSize.width, height: cssContentSize.height, scale: 1 } });
      fs.writeFileSync(path.join(screenshots, `log-aktivitas-${width}.png`), Buffer.from(shot.data, 'base64'));
      await evaluate(`document.querySelector('[aria-label="Halaman 2"]').click()`);
      await waitFor("document.querySelectorAll('.log-tabel tbody tr').length===2");
      assert(await evaluate('document.documentElement.scrollWidth<=innerWidth'), 'Label panjang menyebabkan overflow halaman');
      await evaluate(`document.querySelector('#log-jenis').value='hapus';document.querySelector('.log-filter').requestSubmit()`);
      await waitFor("document.querySelector('.keadaan h3')?.textContent==='Tidak ada aktivitas yang cocok'");
      await evaluate(`document.querySelector('#log-jenis').value='';document.querySelector('.log-filter').requestSubmit()`);
      await waitFor("document.querySelector('.log-tabel')");
      await evaluate(`location.hash='#/berita'`);
      await waitFor("document.querySelector('button[aria-label^=\"Hapus:\"]:not(:disabled)')");
      await evaluate(`document.querySelector('button[aria-label^="Hapus:"]:not(:disabled)').click()`);
      await waitFor("document.querySelector('dialog[open]')");
      await evaluate(`[...document.querySelectorAll('dialog button')].find(b=>b.textContent==='Batal').click()`);
      assert.equal(await evaluate('(async()=> (await api.adminLog()).total)()'), 12);
      report.push(layout);
    }
    assert.deepEqual(exceptions, []);
    console.log('PASS browser Chrome: ' + JSON.stringify(report));
    await send('Browser.close');
  } finally { socket?.close(); child.kill(); }
}

(async () => { await apiTests(); await uiTests(); if (process.argv.includes('--browser')) await browserTests(); })().catch((e) => { console.error(e); process.exitCode = 1; });
