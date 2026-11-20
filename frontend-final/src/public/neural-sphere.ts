/* Bola neural hero Beranda: awan titik berbentuk bola yang berputar pelan dan "bernapas",
   sesekali menyalakan satu node beserta garis tipis ke tetangganya, dan menarik titik
   di dekat kursor (desktop saja). Mandiri: hanya butuh <canvas> dan opsi. Semua angka ada di KONFIG. */

export const KONFIG = {
  benih: 20260,
  partikel: 300,
  partikelRingkas: 200,
  ringkas: null as boolean | null, // null = otomatis: perangkat tanpa hover/pointer halus (layar sentuh)
  fpsRingkas: 30,
  dprMaks: 2,
  dprRingkas: 1.5,
  pikselMaks: 1.5e6, // batas piksel backing store; DPR diturunkan jika kanvas sangat besar
  isi: 0.86, // cadangan --bola-skala: diameter bola terhadap sisi terpendek kanvas
  perspektif: 4, // jarak kamera, dalam satuan jari-jari bola
  putarMs: 75000,
  miring: 0.38,
  napas: { amplitudo: 0.035, aMs: 7000, bMs: 11000 },
  titik: { jari: 1.7, jariAktif: 2.8, tambahJari: 1, alfa: [0.28, 0.5, 0.9] }, // tambahan setelah kedalaman: diameter +2 px
  aktivasi: {
    jedaMs: [1200, 2800],
    jedaDiamMs: [4000, 8000],
    naikMs: 500,
    tahanMs: 900,
    turunMs: 1100,
    tetangga: 6,
    dotMin: 3,
    dotMaks: 5,
    serentak: 2,
    garisAlfa: 0.55,
    garisTebal: 0.75,
  },
  magnet: { radius: 180, kuat: 0.75, tauMs: 180 },
  kerjaMaksMs: 6, // p90 waktu kerja per frame di atas ini => partikel dikurangi, lalu diam
  jendela: 120,
  warna: {} as { titik?: string; aksen?: string }, // kosong = token --warna-putih dan --warna-mint
};

const DESKTOP_REM = 64;
const AKTIVASI_DESKTOP = { naikMs: 450, tahanMs: 2000, turunMs: 1800, jedaMs: [250, 600], serentak: 6, garisTebal: 2.75 };
const TITIK_DESKTOP_TAMBAH = 1;

export type Konfig = typeof KONFIG;
export interface Opsi extends Partial<Konfig> {
  area?: HTMLElement;
}

interface Peristiwa { i: number; mulai: number; tet: number[] }

/** mulberry32 */
function acak(awal: number): () => number {
  let benih = awal;
  return () => {
    benih |= 0;
    benih = (benih + 0x6D2B79F5) | 0;
    let t = Math.imul(benih ^ (benih >>> 15), 1 | benih);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

const halus = (x: number) => x * x * (3 - 2 * x);

/* Titik tersebar rata di permukaan bola (spiral Fibonacci) dengan sedikit acak tetap. */
function buatTitik(n: number, rnd: () => number, k: number) {
  const p = new Float32Array(n * 3);
  const fase = new Float32Array(n);
  const emas = Math.PI * (3 - Math.sqrt(5));
  for (let i = 0; i < n; i++) {
    const y = 1 - (2 * (i + 0.5)) / n;
    const r = Math.sqrt(1 - y * y);
    const a = i * emas + (rnd() - 0.5) * 0.35;
    p[i * 3] = Math.cos(a) * r;
    p[i * 3 + 1] = y;
    p[i * 3 + 2] = Math.sin(a) * r;
    fase[i] = rnd() * 6.2832;
  }
  const tet: number[][] = []; // k tetangga terdekat, dihitung sekali
  for (let i = 0; i < n; i++) {
    const top: [number, number][] = [];
    for (let j = 0; j < n; j++) {
      if (j === i) continue;
      const d = p[i * 3]! * p[j * 3]! + p[i * 3 + 1]! * p[j * 3 + 1]! + p[i * 3 + 2]! * p[j * 3 + 2]!;
      if (top.length < k || d > top[top.length - 1]![0]) {
        top.push([d, j]);
        top.sort((a, b) => b[0] - a[0]);
        if (top.length > k) top.pop();
      }
    }
    tet.push(top.map((x) => x[1]));
  }
  return { p, fase, tet };
}

export function pasang(kanvas: HTMLCanvasElement, opsi: Opsi = {}): { henti: () => void; mulai: () => void; hancurkan: () => void } {
  const k: Konfig = {
    ...KONFIG,
    ...opsi,
    napas: { ...KONFIG.napas, ...opsi.napas },
    titik: { ...KONFIG.titik, ...opsi.titik },
    aktivasi: { ...KONFIG.aktivasi, ...opsi.aktivasi },
    magnet: { ...KONFIG.magnet, ...opsi.magnet },
    warna: { ...KONFIG.warna, ...opsi.warna },
  };
  const ctx = kanvas.getContext('2d');
  if (!ctx) return { henti() {}, mulai() {}, hancurkan() {} };
  const gaya = getComputedStyle(kanvas);
  const warnaTitik = k.warna.titik || gaya.getPropertyValue('--warna-putih').trim() || '#FFFFFF';
  const aksen = k.warna.aksen || gaya.getPropertyValue('--warna-mint').trim() || '#3CFBC5';
  const gerakKurang = matchMedia('(prefers-reduced-motion: reduce)');
  const ringkas = k.ringkas ?? !matchMedia('(hover: hover) and (pointer: fine)').matches;
  const area = opsi.area ?? kanvas;
  const rndAktif = acak(k.benih + 1);
  // Desktop (>= 64rem): titik +2 px, garis mint lebih tebal, lebih lama, lebih cepat muncul; dibaca saat pasang.
  const desktop = matchMedia(`(min-width: ${DESKTOP_REM}rem)`).matches;
  if (desktop) {
    k.titik.jari += TITIK_DESKTOP_TAMBAH;
    k.titik.jariAktif += TITIK_DESKTOP_TAMBAH;
  }
  const a = desktop ? { ...k.aktivasi, ...AKTIVASI_DESKTOP } : k.aktivasi;
  const umurPeristiwa = a.naikMs + a.tahanMs + a.turunMs;

  let n = ringkas ? k.partikelRingkas : k.partikel;
  let data = buatTitik(n, acak(k.benih), a.tetangga);
  let sx = new Float32Array(n);
  let sy = new Float32Array(n);
  let dz = new Float32Array(n);
  let ox = new Float32Array(n);
  let oy = new Float32Array(n);
  function bangun() {
    data = buatTitik(n, acak(k.benih), a.tetangga);
    sx = new Float32Array(n);
    sy = new Float32Array(n);
    dz = new Float32Array(n);
    ox = new Float32Array(n);
    oy = new Float32Array(n);
  }

  let lebar = 0;
  let tinggi = 0;
  let R = 0;
  let cx = 0;
  let cy = 0;
  let magnetCss = true;
  let dpr = 1;
  let regangY = 1;
  let gerak = !gerakKurang.matches;
  let tingkat = 0; // tingkat 1 = partikel separuh, 2 = diam
  let t = 0;
  let terakhir = 0;
  let raf = 0;
  let tidur = 0;
  let tampak = true;
  let kursor: { x: number; y: number } | null = null;
  let peristiwa: Peristiwa[] = [];
  let kerja: number[] = [];
  let pusatTerakhir = -1;
  const jeda = () => {
    const [lo, hi] = gerak ? a.jedaMs : a.jedaDiamMs;
    return lo! + rndAktif() * (hi! - lo!);
  };
  let berikut = jeda();

  function hitung(dt: number) {
    const sudut = gerak ? (t / k.putarMs) * 6.2832 : 0.6;
    const cY = Math.cos(sudut);
    const sY = Math.sin(sudut);
    const cM = Math.cos(k.miring);
    const sM = Math.sin(k.miring);
    const A = gerak ? k.napas.amplitudo : 0;
    const ta = (t / k.napas.aMs) * 6.2832;
    const tb = (t / k.napas.bMs) * 6.2832;
    const f = k.perspektif * R;
    const { p, fase } = data;
    const tarik = gerak && !ringkas && magnetCss;
    const m = k.magnet;
    const rad2 = m.radius * m.radius;
    const lerp = 1 - Math.exp(-dt / m.tauMs);
    for (let i = 0; i < n; i++) {
      const x = p[i * 3]!;
      const y = p[i * 3 + 1]!;
      const z = p[i * 3 + 2]!;
      const def = 1 + (A * (Math.sin(ta + 2.1 * x + 1.3 * y + fase[i]! * 0.3) + 0.6 * Math.sin(tb + 1.7 * z - 2.4 * x))) / 1.6;
      const x1 = x * cY + z * sY; // putar sumbu Y
      const z1 = -x * sY + z * cY;
      const y2 = y * cM - z1 * sM; // miring sumbu X
      const z2 = y * sM + z1 * cM;
      const r = R * def;
      const s = f / (f - z2 * r);
      const bx = cx + x1 * r * s;
      const by = cy + y2 * r * s * regangY;
      let tx = 0;
      let ty = 0;
      if (tarik && kursor) {
        const dx = kursor.x - bx;
        const dy = kursor.y - by;
        const d2 = dx * dx + dy * dy;
        if (d2 < rad2) {
          const w = (1 - Math.sqrt(d2) / m.radius) ** 2 * m.kuat;
          tx = dx * w;
          ty = dy * w;
        }
      }
      ox[i]! += (tx - ox[i]!) * lerp;
      oy[i]! += (ty - oy[i]!) * lerp;
      sx[i] = bx + ox[i]!;
      sy[i] = by + oy[i]!;
      dz[i] = (z2 + 1) / 2;
    }
  }

  function perbaruiPeristiwa() {
    peristiwa = peristiwa.filter((e) => t - e.mulai < umurPeristiwa);
    if (t < berikut || peristiwa.length >= a.serentak) return;
    let i = 0;
    let coba = 0;
    do {
      i = Math.floor(rndAktif() * n);
    } while ((dz[i]! < 0.6 || i === pusatTerakhir || peristiwa.some((e) => e.i === i)) && ++coba < 12); // pilih yang menghadap depan
    const kandidat = [...data.tet[i]!];
    for (let j = kandidat.length - 1; j > 0; j--) {
      const pilih = Math.floor(rndAktif() * (j + 1));
      [kandidat[j], kandidat[pilih]] = [kandidat[pilih]!, kandidat[j]!];
    }
    const jumlah = a.dotMin + Math.floor(rndAktif() * (a.dotMaks - a.dotMin + 1));
    peristiwa.push({ i, mulai: t, tet: kandidat.slice(0, jumlah - 1) });
    pusatTerakhir = i;
    berikut = t + jeda();
  }

  function selubung(u: number) {
    if (u < a.naikMs) return halus(u / a.naikMs);
    if (u < a.naikMs + a.tahanMs) return 1;
    return 1 - halus(Math.min(1, (u - a.naikMs - a.tahanMs) / a.turunMs));
  }

  function lingkar(c: CanvasRenderingContext2D, i: number, jari: number) {
    const r = jari + k.titik.tambahJari;
    c.moveTo(sx[i]! + r, sy[i]!);
    c.arc(sx[i]!, sy[i]!, r, 0, 6.2832);
  }

  function gambar(c: CanvasRenderingContext2D) {
    c.clearRect(0, 0, lebar, tinggi);
    c.fillStyle = warnaTitik;
    for (let b = 0; b < 3; b++) { // tiga tingkat kedalaman = tiga fill
      c.globalAlpha = k.titik.alfa[b]!;
      c.beginPath();
      for (let i = 0; i < n; i++) {
        const lapis = dz[i]! < 0.4 ? 0 : dz[i]! < 0.7 ? 1 : 2;
        if (lapis === b) lingkar(c, i, k.titik.jari * (0.75 + 0.5 * dz[i]!));
      }
      c.fill();
    }
    c.strokeStyle = aksen;
    c.fillStyle = aksen;
    c.lineWidth = a.garisTebal;
    for (const e of peristiwa) {
      const nyala = selubung(t - e.mulai);
      c.globalAlpha = nyala * a.garisAlfa;
      c.beginPath();
      for (const j of e.tet) {
        c.moveTo(sx[e.i]!, sy[e.i]!);
        c.lineTo(sx[j]!, sy[j]!);
      }
      c.stroke();
      c.globalAlpha = nyala * 0.6;
      c.beginPath();
      for (const j of e.tet) lingkar(c, j, k.titik.jari * 1.2);
      c.fill();
      c.globalAlpha = nyala;
      c.beginPath();
      lingkar(c, e.i, k.titik.jariAktif);
      c.fill();
    }
    c.globalAlpha = 1;
  }

  function langkah(dt: number) {
    hitung(dt);
    perbaruiPeristiwa();
    gambar(ctx!);
  }

  function p90() {
    const s = [...kerja].sort((x, y) => x - y);
    return s.length ? s[Math.floor(s.length * 0.9)]! : 0;
  }

  function catatKerja(ms: number) {
    kerja.push(ms);
    if (kerja.length > k.jendela) kerja.shift();
    if (kerja.length < k.jendela || p90() <= k.kerjaMaksMs || !gerak) return;
    kerja = [];
    tingkat++;
    if (tingkat === 1) {
      n = Math.max(40, n >> 1);
      bangun();
      peristiwa = [];
    } else {
      gerak = false;
      peristiwa = [];
      berikut = t + jeda();
    }
  }

  function bingkai(now: number) {
    raf = 0;
    if (ringkas && terakhir && now - terakhir < 1000 / k.fpsRingkas - 4) {
      raf = requestAnimationFrame(bingkai);
      return;
    }
    const dt = terakhir ? Math.min(now - terakhir, 100) : 16;
    terakhir = now;
    t += dt;
    const mulai = performance.now();
    langkah(dt);
    catatKerja(performance.now() - mulai);
    if (gerak || peristiwa.length) {
      raf = requestAnimationFrame(bingkai);
      return;
    }
    // Mode diam: tidak menggambar sampai aktivasi berikutnya.
    terakhir = 0;
    tidur = window.setTimeout(() => {
      tidur = 0;
      t = Math.max(t, berikut);
      minta();
    }, Math.max(0, berikut - t));
  }

  function minta() {
    if (!raf && !tidur && tampak && !document.hidden && lebar) raf = requestAnimationFrame(bingkai);
  }

  function henti() {
    cancelAnimationFrame(raf);
    clearTimeout(tidur);
    raf = 0;
    tidur = 0;
    terakhir = 0;
  }

  /* Posisi dan skala bola datang dari CSS (--bola-x, --bola-y, --bola-skala, --bola-magnet, --bola-regang-y),
     dengan --bola-tinggi-acuan opsional dalam CSS px untuk membatasi radius tanpa mengecilkan kanvas;
     dibaca hanya saat ukuran berubah, supaya komposisi per breakpoint cukup diatur di CSS. */
  function angka(nama: string, cadangan: number) {
    const v = parseFloat(gaya.getPropertyValue(nama));
    return Number.isFinite(v) ? v : cadangan;
  }

  function ukur() {
    lebar = kanvas.clientWidth;
    tinggi = kanvas.clientHeight;
    if (!lebar || !tinggi) return;
    dpr = Math.min(window.devicePixelRatio || 1, ringkas ? k.dprRingkas : k.dprMaks, Math.sqrt(k.pikselMaks / (lebar * tinggi)));
    kanvas.width = Math.round(lebar * dpr);
    kanvas.height = Math.round(tinggi * dpr);
    ctx!.setTransform(dpr, 0, 0, dpr, 0, 0);
    const tinggiAcuan = Math.max(1, angka('--bola-tinggi-acuan', tinggi));
    R = (angka('--bola-skala', k.isi) * Math.min(lebar, tinggi, tinggiAcuan)) / 2;
    cx = angka('--bola-x', 0.5) * lebar;
    cy = angka('--bola-y', 0.5) * tinggi;
    magnetCss = angka('--bola-magnet', 1) > 0;
    regangY = angka('--bola-regang-y', 1); // >1 = bola diregangkan vertikal saja (lebar tetap)
    langkah(0);
    minta();
  }

  const saatGeser = (e: PointerEvent) => {
    if (e.pointerType === 'touch') return;
    if (e.target instanceof Element && e.target.closest('a, button')) { // magnet tidak aktif di atas tautan/CTA
      kursor = null;
      return;
    }
    const r = kanvas.getBoundingClientRect();
    const x = e.clientX - r.left;
    const y = e.clientY - r.top;
    kursor = x >= 0 && y >= 0 && x <= r.width && y <= r.height ? { x, y } : null;
  };
  const saatKeluar = () => {
    kursor = null;
  };
  const saatTampil = () => (document.hidden ? henti() : minta());
  const saatGerak = () => {
    gerak = !gerakKurang.matches && tingkat < 2;
    peristiwa = [];
    berikut = t + jeda();
    henti();
    langkah(0);
    minta();
  };
  const pengamat = new IntersectionObserver((e) => {
    tampak = e[e.length - 1]?.isIntersecting ?? tampak;
    if (tampak) minta();
    else henti();
  });
  const ukuran = new ResizeObserver(ukur);
  if (!ringkas) {
    area.addEventListener('pointermove', saatGeser, { passive: true });
    area.addEventListener('pointerleave', saatKeluar);
  }
  document.addEventListener('visibilitychange', saatTampil);
  gerakKurang.addEventListener('change', saatGerak);
  pengamat.observe(kanvas);
  ukuran.observe(kanvas);

  return {
    henti,
    mulai: minta,
    hancurkan() {
      henti();
      pengamat.disconnect();
      ukuran.disconnect();
      area.removeEventListener('pointermove', saatGeser);
      area.removeEventListener('pointerleave', saatKeluar);
      document.removeEventListener('visibilitychange', saatTampil);
      gerakKurang.removeEventListener('change', saatGerak);
    },
  };
}
