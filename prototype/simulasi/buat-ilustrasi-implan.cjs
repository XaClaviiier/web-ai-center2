/* Alat tinjauan, tidak dimuat halaman. Menulis public/perencanaan-implan-gigi.html beserta empat SVG statisnya.
   Jalankan ulang setelah tinjauan anatomi: node prototype/simulasi/buat-ilustrasi-implan.cjs
   Semua geometri adalah ilustrasi konseptual, tidak berasal dari data pasien dan tidak berskala. */
const fs = require('fs');
const path = require('path');

const f = (n) => Math.round(n * 10) / 10;
const pil = (x, y, teks, kelas = '') => {
  const lebar = teks.length * 16 + 34;
  return `<g class="ig-pil ${kelas}"><rect x="${f(x - lebar / 2)}" y="${y - 23}" width="${lebar}" height="46" rx="23"/><text x="${x}" y="${y + 9}" text-anchor="middle">${teks}</text></g>`;
};
const garis = (x1, y1, x2, y2, kelas = 'ig-penghubung') => `<line x1="${f(x1)}" y1="${f(y1)}" x2="${f(x2)}" y2="${f(y2)}" class="${kelas}"/>`;
const pola = (id) => `<defs>
    <pattern id="${id}" width="16" height="16" patternUnits="userSpaceOnUse" patternTransform="rotate(45)"><rect width="16" height="16" class="ig-arsir-latar"/><line x1="0" y1="0" x2="0" y2="16" class="ig-arsir-garis"/></pattern>
    <marker id="${id}-u" viewBox="0 0 10 10" refX="5" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse"><path d="M1 1 L9 5 L1 9 z" class="ig-ujung"/></marker>
  </defs>`;

/* === Lengkung rahang tampak atas === */
function lengkung(cx, cy, a, b, skala) {
  const titik = (t) => [cx + a * Math.cos(t), cy + b * Math.sin(t)];
  const sampel = [];
  let s = 0;
  let lalu = titik(0.06);
  for (let i = 0; i <= 2000; i++) {
    const t = 0.06 + (Math.PI - 0.12) * i / 2000;
    const p = titik(t);
    s += Math.hypot(p[0] - lalu[0], p[1] - lalu[1]);
    sampel.push({ t, s, p });
    lalu = p;
  }
  const lebar = [62, 62, 62, 46, 44, 38, 34, 34, 38, 44, 46, 62, 62, 62].map((w) => w * skala);
  const jarak = 5 * skala;
  const total = lebar.reduce((x, y) => x + y, 0) + jarak * 13;
  let posisi = (s - total) / 2;
  const gigi = lebar.map((w) => {
    const o = sampel.find((x) => x.s >= posisi + w / 2) ?? sampel[2000];
    posisi += w + jarak;
    const sudut = Math.atan2(b * Math.cos(o.t), -a * Math.sin(o.t)) * 180 / Math.PI;
    const n = [Math.cos(o.t) / a, Math.sin(o.t) / b];
    const panjang = Math.hypot(n[0], n[1]);
    return { x: o.p[0], y: o.p[1], w, tinggi: (w > 50 * skala ? 68 : w > 40 * skala ? 60 : 52) * skala, sudut, n: [n[0] / panjang, n[1] / panjang], t: o.t };
  });
  const [x0, y0] = titik(0.06);
  const [x1, y1] = titik(Math.PI - 0.06);
  return { gigi, busur: `M${f(x0)} ${f(y0)} A${a} ${b} 0 0 1 ${f(x1)} ${f(y1)}` };
}
const kotakGigi = (g, kelas) => `<rect x="${f(-g.w / 2)}" y="${f(-g.tinggi / 2)}" width="${f(g.w)}" height="${f(g.tinggi)}" rx="${f(g.w / 4)}" transform="translate(${f(g.x)} ${f(g.y)}) rotate(${f(g.sudut)})" class="${kelas}"/>`;

/* Posisi slice 1 sampai 5 berurutan dari premolar ke molar belakang; posisi 3 adalah area edentulous. */
const IDX_SLICE = [4, 3, 2, 1, 0];
const POSISI_TARGET = 2;

function tahap1() {
  const { gigi, busur } = lengkung(340, 115, 228, 402, 1.2);
  const target = gigi[IDX_SLICE[POSISI_TARGET]];
  const penanda = IDX_SLICE.map((idx, i) => {
    const g = gigi[idx];
    const dalam = [g.x - g.n[0] * 78, g.y - g.n[1] * 78];
    const luar = [g.x + g.n[0] * 100, g.y + g.n[1] * 100];
    const pusat = [g.x + g.n[0] * 128, g.y + g.n[1] * 128];
    const edentulous = i === POSISI_TARGET;
    const bentuk = edentulous
      ? `<rect x="${f(pusat[0] - 22)}" y="${f(pusat[1] - 22)}" width="44" height="44" rx="5" transform="rotate(45 ${f(pusat[0])} ${f(pusat[1])})" class="ig-tanda ig-tanda--edentulous"/>`
      : `<circle cx="${f(pusat[0])}" cy="${f(pusat[1])}" r="25" class="ig-tanda"/>`;
    return `${garis(dalam[0], dalam[1], luar[0], luar[1], edentulous ? 'ig-slice ig-slice--edentulous' : 'ig-slice')}${bentuk}<text x="${f(pusat[0])}" y="${f(pusat[1] + 10)}" text-anchor="middle" class="ig-nomor">${i + 1}</text>`;
  }).join('\n      ');
  return `<svg class="ig-svg" viewBox="0 0 800 600">
  ${pola('ig1')}
  <path d="${busur}" class="ig-rahang-tepi" stroke-width="134"/>
  <path d="${busur}" class="ig-rahang-isi" stroke-width="128"/>
  <g>${gigi.filter((_, i) => i !== IDX_SLICE[POSISI_TARGET]).map((g) => kotakGigi(g, 'ig-gigi')).join('')}</g>
  <g data-lapisan="area">
    ${kotakGigi(target, 'ig-target')}
    ${garis(target.x - target.w / 2, target.y, 340, 210, 'ig-penghubung ig-penghubung--target')}
    ${pil(340, 210, 'Area edentulous', 'ig-pil--target')}
  </g>
  <g data-lapisan="penanda">
      ${penanda}
  </g>
</svg>`;
}

/* === Penampang tulang === */
const pathTulang = ({ cx, atas, bawah, w1, w2 }) => {
  const t = bawah - atas;
  const y1 = atas + t * 0.13;
  const y2 = atas + t * 0.6;
  return `M${cx} ${atas} C${cx + w1 * 0.55} ${atas} ${cx + w1 * 0.85} ${atas + t * 0.05} ${cx + w1} ${y1} C${cx + w1 * 1.15} ${atas + t * 0.3} ${cx + w2} ${atas + t * 0.42} ${cx + w2} ${y2} C${cx + w2} ${atas + t * 0.82} ${cx + w2 * 0.55} ${bawah} ${cx} ${bawah} C${cx - w2 * 0.55} ${bawah} ${cx - w2} ${atas + t * 0.82} ${cx - w2} ${y2} C${cx - w2} ${atas + t * 0.42} ${cx - w1 * 1.15} ${atas + t * 0.3} ${cx - w1} ${y1} C${cx - w1 * 0.85} ${atas + t * 0.05} ${cx - w1 * 0.55} ${atas} ${cx} ${atas} Z`;
};
function pathGigi(cx, c, l, r, duaAkar) {
  const mahkota = `M${cx - l} ${c - 66} Q${cx} ${c - 92} ${cx + l} ${c - 66} L${cx + l * 0.9} ${c + 6}`;
  if (!duaAkar) return `${mahkota} L${cx + l * 0.42} ${r} Q${cx} ${r + 16} ${cx - l * 0.42} ${r} L${cx - l * 0.9} ${c + 6} Z`;
  return `${mahkota} L${cx + l * 0.72} ${r} Q${cx + l * 0.42} ${r + 10} ${cx + l * 0.2} ${c + 44} Q${cx} ${c + 30} ${cx - l * 0.2} ${c + 44} Q${cx - l * 0.42} ${r + 10} ${cx - l * 0.72} ${r} L${cx - l * 0.9} ${c + 6} Z`;
}
const lingkarKanal = (x, y, r) => `<circle cx="${x}" cy="${y}" r="${r}" class="ig-kanal"/><line x1="${x - r + 7}" y1="${y}" x2="${x + r - 7}" y2="${y}" class="ig-kanal-garis"/>`;

const SLICE = [
  { atas: 130, w1: 60, w2: 110, kx: 8, ky: 446, kr: 25, gigi: { l: 40, akar: 300, dua: false } },
  { atas: 126, w1: 64, w2: 114, kx: 4, ky: 450, kr: 26, gigi: { l: 42, akar: 305, dua: false } },
  { atas: 172, w1: 70, w2: 120, kx: 0, ky: 454, kr: 27, gigi: null },
  { atas: 126, w1: 84, w2: 126, kx: -6, ky: 460, kr: 28, gigi: { l: 62, akar: 300, dua: true } },
  { atas: 120, w1: 92, w2: 132, kx: -12, ky: 466, kr: 29, gigi: { l: 66, akar: 296, dua: true } },
];

function tahap2() {
  const cx = 250;
  const bawah = 580;
  const peta = lengkung(668, 52, 74, 118, 0.36);
  const irisan = SLICE.map((s, i) => {
    const d = pathTulang({ cx, atas: s.atas, bawah, w1: s.w1, w2: s.w2 });
    const batasAb = s.atas + 215;
    const gigi = s.gigi ? `<path d="${pathGigi(cx, s.atas, s.gigi.l, s.gigi.akar, s.gigi.dua)}" class="ig-gigi"/>` : '';
    return `<g data-slice="${i + 1}">
    <path d="${d}" class="ig-tulang"/>
    <clipPath id="ig2-k${i}"><rect x="0" y="0" width="800" height="${batasAb}"/></clipPath>
    <g data-lapisan="ab"><path d="${d}" fill="url(#ig2)" clip-path="url(#ig2-k${i})" class="ig-ab-tepi"/>${garis(cx + s.w1 + 26, batasAb - 80, 438, 300)}${pil(482, 300, 'AB', 'ig-pil--ab')}</g>
    ${gigi}
    <g data-lapisan="mc">${lingkarKanal(cx + s.kx, s.ky, s.kr)}${garis(cx + s.kx + s.kr, s.ky, 438, 400)}${pil(482, 400, 'MC', 'ig-pil--mc')}</g>
  </g>`;
  }).join('\n  ');
  const tandaPeta = IDX_SLICE.map((idx, i) => {
    const g = peta.gigi[idx];
    const a = [g.x - g.n[0] * 26, g.y - g.n[1] * 26];
    const b = [g.x + g.n[0] * 30, g.y + g.n[1] * 30];
    const label = [g.x + g.n[0] * 48, g.y + g.n[1] * 48];
    return `${garis(a[0], a[1], b[0], b[1], 'ig-slice')}<g data-slice="${i + 1}">${garis(a[0], a[1], b[0], b[1], 'ig-slice ig-slice--aktif')}<circle cx="${f(label[0])}" cy="${f(label[1])}" r="16" class="ig-tanda ig-tanda--aktif"/><text x="${f(label[0])}" y="${f(label[1] + 7)}" text-anchor="middle" class="ig-nomor ig-nomor--kecil">${i + 1}</text></g>`;
  }).join('');
  return `<svg class="ig-svg" viewBox="0 0 800 600">
  ${pola('ig2')}
  ${irisan}
  <rect x="540" y="22" width="248" height="212" rx="20" class="ig-kotak"/>
  <path d="${peta.busur}" class="ig-rahang-tepi" stroke-width="44"/>
  <path d="${peta.busur}" class="ig-rahang-isi" stroke-width="40"/>
  ${tandaPeta}
</svg>`;
}

function tahap3() {
  const cx = 330;
  const atas = 152;
  const bawah = 584;
  const d = pathTulang({ cx, atas, bawah, w1: 84, w2: 140 });
  const ax = 34;
  const iAtas = 132;
  const iBawah = 348;
  const ulir = Array.from({ length: 6 }, (_, i) => iAtas + 36 + i * 28).map((y) => `M${cx - ax} ${y} H${cx + ax}`).join(' ');
  return `<svg class="ig-svg" viewBox="0 0 800 600">
  ${pola('ig3')}
  <path d="${d}" class="ig-tulang"/>
  <clipPath id="ig3-k"><rect x="0" y="0" width="800" height="382"/></clipPath>
  <g data-lapisan="anatomi">
    <path d="${d}" fill="url(#ig3)" clip-path="url(#ig3-k)" class="ig-ab-tepi"/>
    ${lingkarKanal(cx, 482, 32)}
    ${garis(cx + 112, 300, 580, 300)}${pil(626, 300, 'AB', 'ig-pil--ab')}
    ${garis(cx + 32, 482, 580, 482)}${pil(626, 482, 'MC', 'ig-pil--mc')}
  </g>
  <rect x="${cx - ax}" y="${iAtas}" width="${ax * 2}" height="${iBawah - iAtas}" rx="10" class="ig-implan"/>
  <path d="${ulir}" class="ig-implan-ulir"/>
  ${garis(cx + ax, 170, 580, 170)}${pil(638, 170, 'Implan', 'ig-pil--implan')}
  <g data-lapisan="ukur">
    <path d="M${cx - ax} 118 V70 M${cx + ax} 118 V70" class="ig-bantu"/>
    <path d="M${cx - ax} 82 H${cx + ax}" class="ig-ukur" marker-start="url(#ig3-u)" marker-end="url(#ig3-u)"/>
    ${pil(cx, 38, 'Diameter', 'ig-pil--ukur')}
    <path d="M${cx - ax - 6} ${iAtas} H150 M${cx - ax - 6} ${iBawah} H150" class="ig-bantu"/>
    <path d="M168 ${iAtas} V${iBawah}" class="ig-ukur" marker-start="url(#ig3-u)" marker-end="url(#ig3-u)"/>
    ${pil(92, 240, 'Tinggi', 'ig-pil--ukur')}
  </g>
</svg>`;
}

/* === Tahap 4: satu geometri, tiga sudut pandang === */
const A = 190;
const B = 270;
const T0 = 0.04;
const SEG = 36;
const SISI = 10;
const JUMLAH_TINGKAT = 4;

const halus = (x) => { const u = Math.max(0, Math.min(1, x)); return u * u * (3 - 2 * u); };
const ramus = (t) => Math.max(halus((0.42 - t) / 0.38), halus((t - (Math.PI - 0.42)) / 0.38));
const sumbu = (t) => [A * Math.cos(t), -B * Math.sin(t) + B * 0.5];
const normal = (t) => { const n = [Math.cos(t) / A, -Math.sin(t) / B]; const p = Math.hypot(n[0], n[1]); return [n[0] / p, n[1] / p]; };

const zAtasDi = (t) => 42 + ramus(t) * 125;
const Z_BAWAH = -58;
const zTengahDi = (t) => (zAtasDi(t) + Z_BAWAH) / 2;
const pangkat = (v, e) => Math.sign(v) * Math.abs(v) ** e;

function titikPenampang(t, phi, tambah = 0) {
  const zTengah = zTengahDi(t);
  const tinggi = (zAtasDi(t) - Z_BAWAH) / 2 + tambah;
  const lebarDasar = 36 + tambah;
  const naik = (Math.sin(phi) + 1) / 2;
  const u = lebarDasar * pangkat(Math.cos(phi), 0.7) * (1 - 0.3 * naik);
  const [px, py] = sumbu(t);
  const [nx, ny] = normal(t);
  return [px + nx * u, py + ny * u, zTengah + tinggi * pangkat(Math.sin(phi), 0.7)];
}

function jalurGigi() {
  const langkah = 800;
  const data = [];
  let s = 0;
  let lalu = sumbu(0.52);
  for (let i = 0; i <= langkah; i++) {
    const t = 0.52 + (Math.PI - 1.04) * i / langkah;
    const p = sumbu(t);
    s += Math.hypot(p[0] - lalu[0], p[1] - lalu[1]);
    data.push({ t, s });
    lalu = p;
  }
  const lebar = [62, 62, 62, 46, 44, 38, 34, 34, 38, 44, 46, 62, 62, 62];
  const skala = (s - 13 * 3) / lebar.reduce((x, y) => x + y, 0);
  let pos = 0;
  return lebar.map((w) => {
    const l = w * skala;
    const o = data.find((d) => d.s >= pos + l / 2) ?? data[langkah];
    pos += l + 3;
    return { t: o.t, l, d: Math.min(26, 18 + w * 0.14) };
  });
}

function hull(titik) {
  const p = [...titik].sort((a, b) => a[0] - b[0] || a[1] - b[1]);
  const silang = (o, a, b) => (a[0] - o[0]) * (b[1] - o[1]) - (a[1] - o[1]) * (b[0] - o[0]);
  const bawah = [];
  for (const q of p) { while (bawah.length >= 2 && silang(bawah[bawah.length - 2], bawah[bawah.length - 1], q) <= 0) bawah.pop(); bawah.push(q); }
  const atas = [];
  for (const q of p.reverse()) { while (atas.length >= 2 && silang(atas[atas.length - 2], atas[atas.length - 1], q) <= 0) atas.pop(); atas.push(q); }
  return [...bawah.slice(0, -1), ...atas.slice(0, -1)];
}

function adegan() {
  const dt = (Math.PI - 2 * T0) / SEG;
  const seg = [];
  for (let i = 0; i < SEG; i++) {
    const t1 = T0 + dt * i;
    const t2 = t1 + dt;
    const sisi = [];
    for (let j = 0; j < SISI; j++) {
      const p1 = (2 * Math.PI * j) / SISI - Math.PI / 2;
      const p2 = (2 * Math.PI * (j + 1)) / SISI - Math.PI / 2;
      sisi.push([titikPenampang(t1, p1), titikPenampang(t2, p1), titikPenampang(t2, p2), titikPenampang(t1, p2)]);
    }
    seg.push({ t: (t1 + t2) / 2, sisi });
  }
  const cincin = (t) => Array.from({ length: SISI }, (_, j) => titikPenampang(t, (2 * Math.PI * j) / SISI - Math.PI / 2));
  const luar = (t, arah) => { const [x0, y0] = sumbu(t); const [x1, y1] = sumbu(t + arah * 0.01); return [x1 - x0, y1 - y0, 0]; };
  seg[0].tutup = { titik: cincin(T0), luar: luar(T0, -1) };
  seg[SEG - 1].tutup = { titik: cincin(Math.PI - T0), luar: luar(Math.PI - T0, 1) };
  const gigiTahap = jalurGigi();
  return { seg, gigi: gigiTahap, target: gigiTahap[2] };
}

function kotakGigi3d(g) {
  const [px, py] = sumbu(g.t);
  const [nx, ny] = normal(g.t);
  const [tx, ty] = [-ny, nx];
  const z0 = 36;
  const z1 = 60;
  const sudut = (u, v, z) => [px + tx * u + nx * v, py + ty * u + ny * v, z];
  const l = g.l / 2;
  const d = g.d / 2 + 3;
  return [[-l, -d], [l, -d], [l, d], [-l, d]].flatMap(([u, v]) => [sudut(u, v, z0), sudut(u, v, z1)]);
}

function tampilanSudut(nama, az, el, a) {
  const ca = Math.cos(az); const sa = Math.sin(az); const ce = Math.cos(el); const se = Math.sin(el);
  const pro = ([x, y, z]) => {
    const X = x * ca - y * sa;
    const Y = x * sa + y * ca;
    return { x: X, y: -(z * ce + Y * se), d: Y * ce - z * se };
  };
  const kamera = [sa * ce, ca * ce, -se].map((v) => -v);
  const cahaya = (() => { const c = [-0.45, -0.55, 0.7]; const p = Math.hypot(...c); return c.map((v) => v / p); })();
  const noktah = (u, v, w) => {
    const e1 = [v[0] - u[0], v[1] - u[1], v[2] - u[2]];
    const e2 = [w[0] - u[0], w[1] - u[1], w[2] - u[2]];
    const n = [e1[1] * e2[2] - e1[2] * e2[1], e1[2] * e2[0] - e1[0] * e2[2], e1[0] * e2[1] - e1[1] * e2[0]];
    const p = Math.hypot(...n) || 1;
    return n.map((c) => c / p);
  };
  const dot = (p, q) => p[0] * q[0] + p[1] * q[1] + p[2] * q[2];
  const [ix, iy] = sumbu(a.target.t);
  const irR = 15;
  const iZ1 = 62;
  const iZ0 = 6;
  const lingkar = (z) => Array.from({ length: 18 }, (_, i) => { const u = (2 * Math.PI * i) / 18; return [ix + irR * Math.cos(u), iy + irR * Math.sin(u), z]; });
  const titikKanal = [];
  for (let i = 0; i <= 28; i++) titikKanal.push(0.12 + (Math.PI / 2 - 0.44) * i / 28);
  const kanalKiri = titikKanal.map((t) => { const [x, y] = sumbu(t); return [x, y, -36 + ramus(t) * 40]; });
  const kanalKanan = titikKanal.map((t) => { const [x, y] = sumbu(Math.PI - t); return [x, y, -36 + ramus(t) * 40]; });

  const semua = [...a.seg.flatMap((s) => s.sisi.flat()), ...a.gigi.flatMap(kotakGigi3d), ...lingkar(iZ1), ...lingkar(iZ0)].map(pro);
  return { nama, pro, kamera, cahaya, noktah, dot, ix, iy, irR, iZ1, iZ0, lingkar, kanalKiri, kanalKanan, batas: { x0: Math.min(...semua.map((p) => p.x)), x1: Math.max(...semua.map((p) => p.x)), y0: Math.min(...semua.map((p) => p.y)), y1: Math.max(...semua.map((p) => p.y)) } };
}

function tahap4() {
  const a = adegan();
  const sudut = [['perspektif', -0.72, 0.68], ['samping', -Math.PI / 2, 0.14], ['depan', 0, 0.2]].map(([n, az, el]) => tampilanSudut(n, az, el, a));
  const skala = Math.min(...sudut.map((v) => Math.min(700 / (v.batas.x1 - v.batas.x0), 500 / (v.batas.y1 - v.batas.y0))));
  const isi = sudut.map((v) => {
    const ox = 400 - ((v.batas.x0 + v.batas.x1) / 2) * skala;
    const oy = 310 - ((v.batas.y0 + v.batas.y1) / 2) * skala;
    const layar = (q) => { const p = v.pro(q); return [Math.round(p.x * skala + ox), Math.round(p.y * skala + oy)]; };
    const poligon = (t) => `M${t.map((q) => q.join(' ')).join('L')}Z`;
    const urut = [...a.seg].sort((p, q) => v.pro([...sumbu(q.t), 0]).d - v.pro([...sumbu(p.t), 0]).d);
    const pita = (ambil, awalan) => urut.map((s) => {
      const per = {};
      const pusat = [...sumbu(s.t), zTengahDi(s.t)];
      ambil(s).forEach((quad) => {
        let n = v.noktah(quad[0], quad[1], quad[2]);
        const c = [0, 1, 2].map((i) => quad.reduce((j, q) => j + q[i], 0) / 4 - pusat[i]);
        if (v.dot(n, c) < 0) n = n.map((x) => -x);
        if (v.dot(n, v.kamera) <= 0) return;
        const k = Math.min(JUMLAH_TINGKAT - 1, Math.floor((1 - (v.dot(n, v.cahaya) + 1) / 2) * JUMLAH_TINGKAT));
        (per[k] ??= []).push(poligon(quad.map(layar)));
      });
      const tutup = awalan === 'r' && s.tutup && v.dot(s.tutup.luar, v.kamera) > 0 ? `<path class="r2" d="${poligon(s.tutup.titik.map(layar))}"/>` : '';
      return Object.entries(per).map(([k, d]) => `<path class="${awalan}${k}" d="${d.join('')}"/>`).join('') + tutup;
    }).join('');
    const rahang = pita((s) => s.sisi, 'r');
    const ab = pita((s) => (ramus(s.t) > 0.05 ? [] : s.sisi.slice(3, 7).map((q) => q.map(([x, y, z]) => {
      const [px, py] = sumbu(s.t);
      return [px + (x - px) * 1.05, py + (y - py) * 1.05, z + 1];
    }))), 'a');
    const urutGigi = [...a.gigi].filter((g) => g !== a.target).sort((p, q) => v.pro([...sumbu(q.t), 50]).d - v.pro([...sumbu(p.t), 50]).d);
    const gigi = urutGigi.map((g) => {
      const titik = kotakGigi3d(g);
      const atas = [titik[1], titik[3], titik[5], titik[7]].map(layar);
      return `<path class="ig-gigi" d="${poligon(hull(titik.map(layar)))}"/><path class="ig-gigi-atas" d="${poligon(atas)}"/>`;
    }).join('');
    const kanal = [v.kanalKiri, v.kanalKanan].map((j) => `M${j.map(layar).map((q) => q.join(' ')).join('L')}`).join('');
    const hIm = hull([...v.lingkar(v.iZ1), ...v.lingkar(v.iZ0)].map(layar));
    const ulir = [20, 34, 48].map((z) => `<path class="ig-implan-ulir" d="${poligon(v.lingkar(z).map(layar))}"/>`).join('');
    const tutup = poligon(v.lingkar(v.iZ1).map(layar));
    const jepit = (x, y, l) => [Math.max(l / 2 + 8, Math.min(800 - l / 2 - 8, x)), Math.max(30, Math.min(570, y))];
    const labelDi = (q, dx, dy, teks, kelas) => {
      const [x, y] = layar(q);
      const [lx, ly] = jepit(x + dx, y + dy, teks.length * 16 + 34);
      return `${garis(x, y, lx, ly)}${pil(lx, ly, teks, kelas)}`;
    };
    const titikAb = titikPenampang(Math.PI * 0.74, Math.PI / 2);
    const titikMc = v.kanalKiri[6];
    const posisiLabel = {
      perspektif: { ab: [-110, -70], mc: [120, 60], im: [40, -96] },
      samping: { ab: [-20, -140], mc: [-110, 110], im: [110, -90] },
      depan: { ab: [-120, -110], mc: [150, 110], im: [90, -110] },
    }[v.nama];
    return `<g data-sudut="${v.nama}">
    <g data-lapisan="rahang">${rahang}</g>
    <g data-lapisan="ab">${ab}${labelDi(titikAb, ...posisiLabel.ab, 'AB', 'ig-pil--ab')}</g>
    <g data-lapisan="mc"><path class="ig-kanal-3d" d="${kanal}"/><path class="ig-kanal-3d-garis" d="${kanal}"/>${labelDi(titikMc, ...posisiLabel.mc, 'MC', 'ig-pil--mc')}</g>
    <g data-lapisan="rahang">${gigi}</g>
    <g data-lapisan="implan"><path class="ig-implan" d="${poligon(hIm)}"/>${ulir}<path class="ig-implan-tutup" d="${tutup}"/>${labelDi([v.ix, v.iy, v.iZ1], ...posisiLabel.im, 'Implan', 'ig-pil--implan')}</g>
  </g>`;
  }).join('\n  ');
  return `<svg class="ig-svg ig-svg--3d" viewBox="0 0 800 600">
  ${isi}
</svg>`;
}

const halaman = `<!doctype html>
<html lang="id">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Perencanaan Implan Gigi Berbasis AI (Konseptual) | AI Center UBAYA</title>
  <meta name="description" content="Demo penelitian dengan visualisasi konseptual. Nonklinis.">
  <meta name="robots" content="noindex">
  <link rel="stylesheet" href="../assets/css/tokens.css">
  <link rel="stylesheet" href="../assets/css/base.css">
  <link rel="stylesheet" href="../assets/css/components.css">
  <link rel="stylesheet" href="../assets/css/public.css">
  <link rel="stylesheet" href="../assets/css/implan-gigi.css">
</head>
<body data-halaman="perencanaan-implan-gigi" data-sisi="publik">
  <main id="utama" tabindex="-1"></main>
  <!-- Dibuat oleh simulasi/buat-ilustrasi-implan.cjs. Ilustrasi konseptual, bukan citra atau data pasien. -->
  <template id="ig-gambar-1">${tahap1()}</template>
  <template id="ig-gambar-2">${tahap2()}</template>
  <template id="ig-gambar-3">${tahap3()}</template>
  <template id="ig-gambar-4">${tahap4()}</template>
  <script src="../assets/mock-data.js"></script>
  <script src="../assets/js/components.js"></script>
  <script src="../assets/js/implan-gigi.js"></script>
  <script src="../assets/js/public.js"></script>
</body>
</html>
`;

const tujuan = path.join(__dirname, '..', 'public', 'perencanaan-implan-gigi.html');
fs.writeFileSync(tujuan, halaman);
console.log(`${path.basename(tujuan)}: ${(Buffer.byteLength(halaman) / 1024).toFixed(1)} KB`);
