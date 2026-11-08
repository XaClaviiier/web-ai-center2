/* Bola neural hero Beranda: awan titik berbentuk bola yang berputar pelan dan "bernapas",
   sesekali menyalakan satu node beserta garis tipis ke tetangganya, dan menarik titik
   di dekat kursor (desktop saja). Mandiri: hanya butuh <canvas> dan opsi, supaya bisa
   dibungkus island Astro di produksi. Semua angka ada di KONFIG. */
const BolaNeural = (() => {
  const KONFIG = {
    benih: 20260,
    partikel: 300, partikelRingkas: 200,
    ringkas: null,               // null = otomatis: perangkat tanpa hover/pointer halus (layar sentuh)
    fpsRingkas: 30,
    dprMaks: 2, dprRingkas: 1.5,
    pikselMaks: 1.5e6,          // batas piksel backing store; DPR diturunkan jika kanvas sangat besar
    isi: 0.86,                   // cadangan --bola-skala: diameter bola terhadap sisi terpendek kanvas
    perspektif: 4,               // jarak kamera, dalam satuan jari-jari bola
    putarMs: 75000, miring: 0.38,
    napas: { amplitudo: 0.035, aMs: 7000, bMs: 11000 },
    titik: { jari: 1.7, jariAktif: 2.8, alfa: [0.28, 0.5, 0.9] }, // alfa: belakang, tengah, depan
    aktivasi: { jedaMs: [1800, 4200], jedaDiamMs: [4000, 8000], naikMs: 500, tahanMs: 900, turunMs: 1100, tetangga: 3, serentak: 2, garisAlfa: 0.55, garisTebal: 0.75 },
    magnet: { radius: 110, kuat: 0.22, tauMs: 180 },
    kerjaMaksMs: 6, jendela: 120, // p90 waktu kerja per frame di atas ini => partikel dikurangi, lalu diam
    warna: {},                   // { titik, aksen }; kosong = token --warna-putih dan --warna-mint
  };

  function acak(benih) { // mulberry32
    return () => {
      benih |= 0; benih = (benih + 0x6D2B79F5) | 0;
      let t = Math.imul(benih ^ (benih >>> 15), 1 | benih);
      t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
      return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
    };
  }
  const halus = (x) => x * x * (3 - 2 * x);

  /* Titik tersebar rata di permukaan bola (spiral Fibonacci) dengan sedikit acak tetap. */
  function buatTitik(n, rnd, k) {
    const p = new Float32Array(n * 3), fase = new Float32Array(n), emas = Math.PI * (3 - Math.sqrt(5));
    for (let i = 0; i < n; i++) {
      const y = 1 - (2 * (i + 0.5)) / n, r = Math.sqrt(1 - y * y), a = i * emas + (rnd() - 0.5) * 0.35;
      p[i * 3] = Math.cos(a) * r; p[i * 3 + 1] = y; p[i * 3 + 2] = Math.sin(a) * r;
      fase[i] = rnd() * 6.2832;
    }
    const tet = []; // k tetangga terdekat, dihitung sekali
    for (let i = 0; i < n; i++) {
      const top = [];
      for (let j = 0; j < n; j++) {
        if (j === i) continue;
        const d = p[i * 3] * p[j * 3] + p[i * 3 + 1] * p[j * 3 + 1] + p[i * 3 + 2] * p[j * 3 + 2];
        if (top.length < k || d > top[top.length - 1][0]) { top.push([d, j]); top.sort((a, b) => b[0] - a[0]); if (top.length > k) top.pop(); }
      }
      tet.push(top.map((x) => x[1]));
    }
    return { p, fase, tet };
  }

  function pasang(kanvas, opsi = {}) {
    const k = { ...KONFIG, ...opsi };
    for (const kunci of ['napas', 'titik', 'aktivasi', 'magnet', 'warna']) k[kunci] = { ...KONFIG[kunci], ...opsi[kunci] };
    const ctx = kanvas.getContext('2d');
    const gaya = getComputedStyle(kanvas);
    const warnaTitik = k.warna.titik || gaya.getPropertyValue('--warna-putih').trim() || '#FFFFFF';
    const aksen = k.warna.aksen || gaya.getPropertyValue('--warna-mint').trim() || '#3CFBC5';
    const gerakKurang = matchMedia('(prefers-reduced-motion: reduce)');
    const ringkas = k.ringkas ?? !matchMedia('(hover: hover) and (pointer: fine)').matches;
    const area = opsi.area || kanvas;
    const rndAktif = acak(k.benih + 1);
    const a = k.aktivasi, umurPeristiwa = a.naikMs + a.tahanMs + a.turunMs;

    let n = ringkas ? k.partikelRingkas : k.partikel;
    let data, sx, sy, dz, ox, oy;
    function bangun() {
      data = buatTitik(n, acak(k.benih), a.tetangga);
      sx = new Float32Array(n); sy = new Float32Array(n); dz = new Float32Array(n); ox = new Float32Array(n); oy = new Float32Array(n);
    }
    bangun();

    let lebar = 0, tinggi = 0, R = 0, cx = 0, cy = 0;
    let magnetCss = true, dpr = 1, regangY = 1;
    let gerak = !gerakKurang.matches, tingkat = 0; // tingkat 1 = partikel separuh, 2 = diam
    let t = 0, terakhir = 0, raf = 0, tidur = 0, tampak = true, kursor = null;
    let peristiwa = [], berikut = 0, garisMaks = 0, geserMaks = 0, kerja = [];
    const jeda = () => { const [lo, hi] = gerak ? a.jedaMs : a.jedaDiamMs; return lo + rndAktif() * (hi - lo); };
    berikut = jeda();

    function hitung(dt) {
      const sudut = gerak ? (t / k.putarMs) * 6.2832 : 0.6;
      const cY = Math.cos(sudut), sY = Math.sin(sudut), cM = Math.cos(k.miring), sM = Math.sin(k.miring);
      const A = gerak ? k.napas.amplitudo : 0, ta = (t / k.napas.aMs) * 6.2832, tb = (t / k.napas.bMs) * 6.2832;
      const f = k.perspektif * R, p = data.p, fase = data.fase;
      const tarik = gerak && !ringkas && magnetCss, m = k.magnet, rad2 = m.radius * m.radius, lerp = 1 - Math.exp(-dt / m.tauMs);
      geserMaks = 0;
      for (let i = 0; i < n; i++) {
        const x = p[i * 3], y = p[i * 3 + 1], z = p[i * 3 + 2];
        const def = 1 + (A * (Math.sin(ta + 2.1 * x + 1.3 * y + fase[i] * 0.3) + 0.6 * Math.sin(tb + 1.7 * z - 2.4 * x))) / 1.6;
        const x1 = x * cY + z * sY, z1 = -x * sY + z * cY;   // putar sumbu Y
        const y2 = y * cM - z1 * sM, z2 = y * sM + z1 * cM;   // miring sumbu X
        const r = R * def, s = f / (f - z2 * r);
        const bx = cx + x1 * r * s, by = cy + y2 * r * s * regangY;
        let tx = 0, ty = 0;
        if (tarik && kursor) {
          const dx = kursor.x - bx, dy = kursor.y - by, d2 = dx * dx + dy * dy;
          if (d2 < rad2) { const w = (1 - Math.sqrt(d2) / m.radius) ** 2 * m.kuat; tx = dx * w; ty = dy * w; }
        }
        ox[i] += (tx - ox[i]) * lerp; oy[i] += (ty - oy[i]) * lerp;
        geserMaks = Math.max(geserMaks, Math.abs(ox[i]), Math.abs(oy[i]));
        sx[i] = bx + ox[i]; sy[i] = by + oy[i]; dz[i] = (z2 + 1) / 2;
      }
    }

    function perbaruiPeristiwa() {
      peristiwa = peristiwa.filter((e) => t - e.mulai < umurPeristiwa);
      if (t >= berikut && peristiwa.length < a.serentak) {
        let i = 0, coba = 0;
        do { i = Math.floor(rndAktif() * n); } while (dz[i] < 0.6 && ++coba < 12); // pilih yang menghadap depan
        peristiwa.push({ i, mulai: t });
        berikut = t + jeda();
      }
    }

    function selubung(u) {
      if (u < a.naikMs) return halus(u / a.naikMs);
      if (u < a.naikMs + a.tahanMs) return 1;
      return 1 - halus(Math.min(1, (u - a.naikMs - a.tahanMs) / a.turunMs));
    }

    const lingkar = (i, j) => { ctx.moveTo(sx[i] + j, sy[i]); ctx.arc(sx[i], sy[i], j, 0, 6.2832); };
    function gambar() {
      ctx.clearRect(0, 0, lebar, tinggi);
      ctx.fillStyle = warnaTitik;
      for (let b = 0; b < 3; b++) { // tiga tingkat kedalaman = tiga fill
        ctx.globalAlpha = k.titik.alfa[b];
        ctx.beginPath();
        for (let i = 0; i < n; i++) {
          if ((dz[i] < 0.4 ? 0 : dz[i] < 0.7 ? 1 : 2) === b) lingkar(i, k.titik.jari * (0.75 + 0.5 * dz[i]));
        }
        ctx.fill();
      }
      let garis = 0;
      ctx.strokeStyle = aksen; ctx.fillStyle = aksen; ctx.lineWidth = a.garisTebal;
      for (const e of peristiwa) {
        const nyala = selubung(t - e.mulai), tet = data.tet[e.i];
        ctx.globalAlpha = nyala * a.garisAlfa;
        ctx.beginPath();
        for (const j of tet) { ctx.moveTo(sx[e.i], sy[e.i]); ctx.lineTo(sx[j], sy[j]); garis++; }
        ctx.stroke();
        ctx.globalAlpha = nyala * 0.6;
        ctx.beginPath(); for (const j of tet) lingkar(j, k.titik.jari * 1.2); ctx.fill();
        ctx.globalAlpha = nyala;
        ctx.beginPath(); lingkar(e.i, k.titik.jariAktif); ctx.fill();
      }
      ctx.globalAlpha = 1;
      garisMaks = Math.max(garisMaks, garis);
    }

    function langkah(dt) { hitung(dt); perbaruiPeristiwa(); gambar(); }

    function p90() { const s = [...kerja].sort((x, y) => x - y); return s.length ? s[Math.floor(s.length * 0.9)] : 0; }
    function catatKerja(ms) {
      kerja.push(ms); if (kerja.length > k.jendela) kerja.shift();
      if (kerja.length < k.jendela || p90() <= k.kerjaMaksMs || !gerak) return;
      kerja = []; tingkat++;
      if (tingkat === 1) { n = Math.max(40, n >> 1); bangun(); peristiwa = []; }
      else { gerak = false; peristiwa = []; berikut = t + jeda(); }
    }

    function bingkai(now) {
      raf = 0;
      if (ringkas && terakhir && now - terakhir < 1000 / k.fpsRingkas - 4) { raf = requestAnimationFrame(bingkai); return; }
      const dt = terakhir ? Math.min(now - terakhir, 100) : 16;
      terakhir = now; t += dt;
      const mulai = performance.now();
      langkah(dt);
      catatKerja(performance.now() - mulai);
      if (gerak || peristiwa.length) { raf = requestAnimationFrame(bingkai); return; }
      // Mode diam: tidak menggambar sampai aktivasi berikutnya.
      terakhir = 0;
      tidur = setTimeout(() => { tidur = 0; t = Math.max(t, berikut); minta(); }, Math.max(0, berikut - t));
    }
    function minta() { if (!raf && !tidur && tampak && !document.hidden && lebar) raf = requestAnimationFrame(bingkai); }
    function henti() { cancelAnimationFrame(raf); clearTimeout(tidur); raf = 0; tidur = 0; terakhir = 0; }

    /* Posisi dan skala bola datang dari CSS (--bola-x, --bola-y, --bola-skala, --bola-magnet),
       dibaca hanya saat ukuran berubah, supaya komposisi per breakpoint cukup diatur di CSS. */
    const angka = (nama, cadangan) => { const v = parseFloat(gaya.getPropertyValue(nama)); return Number.isFinite(v) ? v : cadangan; };
    function ukur() {
      lebar = kanvas.clientWidth; tinggi = kanvas.clientHeight;
      if (!lebar || !tinggi) return;
      dpr = Math.min(window.devicePixelRatio || 1, ringkas ? k.dprRingkas : k.dprMaks, Math.sqrt(k.pikselMaks / (lebar * tinggi)));
      kanvas.width = Math.round(lebar * dpr); kanvas.height = Math.round(tinggi * dpr);
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
      R = (angka('--bola-skala', k.isi) * Math.min(lebar, tinggi)) / 2;
      cx = angka('--bola-x', 0.5) * lebar; cy = angka('--bola-y', 0.5) * tinggi;
      magnetCss = angka('--bola-magnet', 1) > 0;
      regangY = angka('--bola-regang-y', 1); // >1 = bola diregangkan vertikal saja (lebar tetap)
      langkah(0);
      minta();
    }

    const saatGeser = (e) => {
      if (e.pointerType === 'touch') return;
      if (e.target.closest?.('a, button')) { kursor = null; return; } // magnet tidak aktif di atas tautan/CTA
      const r = kanvas.getBoundingClientRect(), x = e.clientX - r.left, y = e.clientY - r.top;
      kursor = x >= 0 && y >= 0 && x <= r.width && y <= r.height ? { x, y } : null;
    };
    const saatKeluar = () => { kursor = null; };
    const saatTampil = () => (document.hidden ? henti() : minta());
    const saatGerak = () => { gerak = !gerakKurang.matches && tingkat < 2; peristiwa = []; berikut = t + jeda(); henti(); langkah(0); minta(); };
    const pengamat = new IntersectionObserver((e) => { tampak = e[e.length - 1].isIntersecting; tampak ? minta() : henti(); });
    const ukuran = new ResizeObserver(ukur);
    if (!ringkas) { area.addEventListener('pointermove', saatGeser, { passive: true }); area.addEventListener('pointerleave', saatKeluar); }
    document.addEventListener('visibilitychange', saatTampil);
    gerakKurang.addEventListener('change', saatGerak);
    pengamat.observe(kanvas); ukuran.observe(kanvas);

    return {
      henti, mulai: minta,
      hancurkan() {
        henti(); pengamat.disconnect(); ukuran.disconnect();
        area.removeEventListener('pointermove', saatGeser); area.removeEventListener('pointerleave', saatKeluar);
        document.removeEventListener('visibilitychange', saatTampil); gerakKurang.removeEventListener('change', saatGerak);
      },
      get info() { return { partikel: n, ringkas, gerak, tingkat, interaktif: !ringkas && gerak && magnetCss, dpr, garisMaks, geserMaks, kerjaP90: p90(), R, lebar, tinggi }; },
    };
  }

  return { pasang, KONFIG };
})();
