/* Demo JejakAlumni. Semua angka dan teks "AI" adalah data siap pakai di konten.jejak_alumni;
   tidak ada inferensi. Persentase dan total selalu dihitung dari jumlah orang agar tidak saling bertentangan. */
import { useState } from 'react';
import { konten } from '../shared/content';
import { Icon } from '../shared/ui';

const JUMLAH = ['lulusan', 'responden', 'terserap', 'bekerja', 'kuliah', 'wirausaha', 'belum', 'sesuai'] as const;
type Jumlah = (typeof JUMLAH)[number];
type Baris = Record<Jumlah, number> & { median: number };
type Total = Record<Jumlah, number> & { median: number };

interface Angkatan {
  median_semua: number;
  per_jurusan: Record<string, Baris>;
  tema: { jenis: string; teks: string; sebutan: number }[];
  ringkasan: string;
}

interface JejakData {
  status: string;
  disclaimer: string;
  jurusan: string[];
  angkatan: Record<string, Angkatan>;
  definisi: { istilah: string; arti: string }[];
  penjelasan: {
    masalah: string;
    langkah: string[];
    langkah_catatan: string;
    data: string[];
    data_catatan: string;
    hasil: string;
  };
}

const D: JejakData = konten.jejak_alumni;
const SEMUA = 'semua';
const TAHUN = Object.keys(D.angkatan).sort().reverse();
const STATUS = [['bekerja', 'Bekerja'], ['kuliah', 'Kuliah'], ['wirausaha', 'Wirausaha'], ['belum', 'Belum bekerja']] as const;
const persen = (n: number, d: number) => Math.round((100 * n) / d);
const bulan = new Intl.NumberFormat('id-ID', { minimumFractionDigits: 1, maximumFractionDigits: 1 }).format;

// Penjaga data simulasi: pelanggaran terlihat di konsol saat halaman dibuka.
for (const [th, a] of Object.entries(D.angkatan)) {
  for (const [j, b] of Object.entries(a.per_jurusan)) {
    const status = STATUS.map(([k]) => persen(b[k], b.responden)).reduce((x, y) => x + y);
    console.assert(
      b.bekerja + b.kuliah + b.wirausaha + b.belum === b.responden && status === 100
        && b.responden <= b.lulusan && b.bekerja <= b.terserap && b.sesuai <= b.bekerja,
      `Data JejakAlumni tidak konsisten: ${th} ${j}`,
    );
  }
}

function hitung(tahun: string, jurusan: string): Total | null {
  const a = D.angkatan[tahun];
  if (!a) return null;
  const satu = a.per_jurusan[jurusan];
  const baris = jurusan === SEMUA ? Object.values(a.per_jurusan) : satu ? [satu] : [];
  const median = jurusan === SEMUA ? a.median_semua : baris[0]?.median;
  if (median === undefined) return null;
  const jumlah = Object.fromEntries(JUMLAH.map((k) => [k, baris.reduce((s, b) => s + b[k], 0)])) as Record<Jumlah, number>;
  return { ...jumlah, median };
}

interface Kpi {
  label: string;
  nilai: (t: Total) => number;
  dasar: (t: Total) => string;
  bulan?: true;
}

const KPI: Kpi[] = [
  { label: 'Tingkat respons kuesioner', nilai: (t) => persen(t.responden, t.lulusan), dasar: (t) => `${t.responden} dari ${t.lulusan} lulusan` },
  { label: 'Masa tunggu kerja (median)', nilai: (t) => t.median, dasar: (t) => `${t.terserap} responden terserap`, bulan: true },
  { label: 'Terserap dunia kerja', nilai: (t) => persen(t.terserap, t.responden), dasar: (t) => `${t.terserap} dari ${t.responden} responden` },
  { label: 'Kerja sesuai bidang', nilai: (t) => persen(t.sesuai, t.bekerja), dasar: (t) => `${t.sesuai} dari ${t.bekerja} yang bekerja` },
];

/** Selisih dihitung dari nilai yang tampil. Masa tunggu lebih singkat berarti membaik. */
function Pembanding({ k, ini, lalu, tahunLalu }: { k: Kpi; ini: Total; lalu: Total | null; tahunLalu: string }) {
  if (!lalu) return <p className="ja-kpi__banding">Tidak ada data angkatan sebelumnya</p>;
  const selisih = Math.round((k.nilai(ini) - k.nilai(lalu)) * 10) / 10;
  if (!selisih) return <p className="ja-kpi__banding">{`Sama dengan ${tahunLalu}`}</p>;
  const besar = Math.abs(selisih);
  const teks = k.bulan
    ? `${bulan(besar)} bln ${selisih < 0 ? 'lebih singkat' : 'lebih lama'} dari ${tahunLalu}`
    : `${selisih > 0 ? '+' : '−'}${besar} poin dari ${tahunLalu}`;
  const membaik = k.bulan ? selisih < 0 : selisih > 0;
  return (
    <p className={`ja-kpi__banding ja-kpi__banding--${membaik ? 'baik' : 'buruk'}`}>
      <span aria-hidden="true">{selisih > 0 ? '▲ ' : '▼ '}</span>
      {teks}
    </p>
  );
}

function GrafikPenyerapan({ a, tahun, jurusan }: { a: Angkatan; tahun: string; jurusan: string }) {
  return (
    <section className="ja-panel" aria-labelledby="ja-j-serap">
      <h3 className="ja-judul" id="ja-j-serap">Terserap dunia kerja per jurusan</h3>
      <p className="field__bantuan">{`Persentase responden yang bekerja dalam 6 bulan setelah lulus · angkatan ${tahun}`}</p>
      <ul className="ja-batang" data-pilih={jurusan === SEMUA ? undefined : 'ya'}>
        {D.jurusan.map((j) => {
          const b = a.per_jurusan[j];
          if (!b) return null;
          const nilai = persen(b.terserap, b.responden);
          return (
            <li key={j} aria-current={j === jurusan ? 'true' : undefined}>
              <span className="ja-batang__label">{j}</span>
              <span className="ja-batang__jalur" aria-hidden="true">
                <span className="ja-batang__isi" style={{ width: `${nilai}%` }} />
              </span>
              <span className="ja-batang__nilai">{`${nilai}%`}</span>
            </li>
          );
        })}
      </ul>
    </section>
  );
}

function GrafikStatus({ t, tahun, jurusan }: { t: Total; tahun: string; jurusan: string }) {
  return (
    <section className="ja-panel" aria-labelledby="ja-j-status">
      <h3 className="ja-judul" id="ja-j-status">Status alumni saat ini</h3>
      <p className="field__bantuan">
        {`Saat mengisi kuesioner · angkatan ${tahun} · ${jurusan === SEMUA ? 'semua jurusan' : jurusan}`}
      </p>
      <div className="ja-segmen" aria-hidden="true">
        {STATUS.map(([k]) => <span key={k} className={`ja-warna--${k}`} style={{ flexGrow: t[k] }} />)}
      </div>
      <ul className="ja-status">
        {STATUS.map(([k, label]) => (
          <li key={k}>
            <span className={`ja-status__warna ja-warna--${k}`} aria-hidden="true" />
            <span>{label}</span>
            <strong>{`${persen(t[k], t.responden)}%`}</strong>
            <span className="ja-status__jumlah">{`${t[k]} orang`}</span>
          </li>
        ))}
      </ul>
    </section>
  );
}

function PanelAi({ a, tahun, jurusan }: { a: Angkatan; tahun: string; jurusan: string }) {
  const cakupan = `Cakupan: seluruh jurusan, angkatan ${tahun}${jurusan === SEMUA ? '' : ' (tidak mengikuti filter jurusan)'}.`;
  return (
    <div className="ja-ai">
      <section className="ja-panel" aria-labelledby="ja-j-tema">
        <div className="ja-panel__kepala">
          <h3 className="ja-judul" id="ja-j-tema">Tema jawaban terbuka</h3>
          <span className="badge badge--simulasi">Simulasi Analisis AI</span>
        </div>
        <p className="field__bantuan">
          Contoh bagaimana AI dapat membantu mengelompokkan jawaban terbuka alumni menjadi tema yang lebih mudah ditinjau.
        </p>
        <ul className="ja-tema">
          {a.tema.map((t) => (
            <li key={t.teks}>
              <span className={`badge ja-lencana--${t.jenis}`}>{t.jenis === 'positif' ? 'Positif' : 'Keluhan'}</span>
              <span>{t.teks}</span>
              <span className="ja-tema__jumlah">{`${t.sebutan} sebutan`}</span>
            </li>
          ))}
        </ul>
        <p className="field__bantuan">{cakupan}</p>
      </section>
      <section className="ja-panel ja-panel--ringkasan" aria-labelledby="ja-j-ringkas">
        <div className="ja-panel__kepala">
          <h3 className="ja-judul" id="ja-j-ringkas">Ringkasan untuk rapat kurikulum</h3>
          <span className="badge badge--simulasi">Simulasi Ringkasan AI</span>
        </div>
        <p>{a.ringkasan}</p>
        <p className="field__bantuan">{cakupan}</p>
      </section>
    </div>
  );
}

function Dasbor() {
  const [tahun, setTahun] = useState(TAHUN[0] ?? '');
  const [jurusan, setJurusan] = useState(SEMUA);
  const a = D.angkatan[tahun];
  const ini = hitung(tahun, jurusan);
  const tahunLalu = String(Number(tahun) - 1);
  const lalu = hitung(tahunLalu, jurusan);

  return (
    <section className="ja-bingkai" aria-labelledby="ja-j-dasbor">
      <div className="ja-panel__kepala">
        <h2 className="ja-bingkai__judul" id="ja-j-dasbor">Dasbor tracer study</h2>
      </div>
      <details className="ja-fitur-lanjutan">
        <summary>Fitur lain dalam persiapan</summary>
        <ul className="ja-tab">
          <li>
            <span className="ja-tab__aktif" aria-current="true">Ringkasan</span>
          </li>
          {['Responden', 'Mitra kerja'].map((t) => (
            <li key={t}>
              <button type="button" className="ja-tab__mati" disabled>
                {t}
                <small>Belum tersedia dalam demo</small>
              </button>
            </li>
          ))}
        </ul>
      </details>
      <div className="ja-filter">
        <div className="field">
          <label htmlFor="ja-angkatan">Angkatan lulus</label>
          <select className="input" id="ja-angkatan" value={tahun} onChange={(e) => setTahun(e.currentTarget.value)}>
            {TAHUN.map((t) => <option key={t} value={t}>{t}</option>)}
          </select>
        </div>
        <div className="field">
          <label htmlFor="ja-jurusan">Jurusan</label>
          <select className="input" id="ja-jurusan" value={jurusan} onChange={(e) => setJurusan(e.currentTarget.value)}>
            <option value={SEMUA}>Semua jurusan</option>
            {D.jurusan.map((j) => <option key={j} value={j}>{j}</option>)}
          </select>
        </div>
      </div>
      <p className="ja-cakupan" role="status">
        {ini && `Angkatan ${tahun} · ${jurusan === SEMUA ? 'Semua jurusan' : jurusan}: ${ini.responden} responden dari ${ini.lulusan} lulusan`}
      </p>
      <div className="ja-kpi">
        {ini && KPI.map((k) => (
          <div key={k.label} className="ja-panel ja-kpi__kartu">
            <h3 className="ja-kpi__label">{k.label}</h3>
            <p className="ja-kpi__nilai">
              {k.bulan ? bulan(k.nilai(ini)) : `${k.nilai(ini)}%`}
              {k.bulan && <small> bln</small>}
            </p>
            <p className="field__bantuan">{k.dasar(ini)}</p>
            <Pembanding k={k} ini={ini} lalu={lalu} tahunLalu={tahunLalu} />
          </div>
        ))}
      </div>
      <details className="ja-definisi">
        <summary>Definisi indikator</summary>
        <dl>
          {D.definisi.map((d) => (
            <div key={d.istilah}>
              <dt>{d.istilah}</dt>
              <dd>{d.arti}</dd>
            </div>
          ))}
        </dl>
      </details>
      <div className="ja-grafik">
        {a && <GrafikPenyerapan a={a} tahun={tahun} jurusan={jurusan} />}
        {ini && <GrafikStatus t={ini} tahun={tahun} jurusan={jurusan} />}
      </div>
      {a && <PanelAi a={a} tahun={tahun} jurusan={jurusan} />}
    </section>
  );
}

function Penjelasan() {
  const p = D.penjelasan;
  return (
    <aside className="ja-penjelasan" aria-labelledby="ja-j-tentang">
      <div>
        <p className="tentang-label">Data</p>
        <h2 id="ja-j-tentang">Tentang JejakAlumni</h2>
      </div>
      <div>
        <h3>Masalah nyata di sekolah</h3>
        <p className="teks-sekunder">{p.masalah}</p>
      </div>
      <div>
        <h3>Cara kerjanya</h3>
        <ol className="ja-langkah">
          {p.langkah.map((l, i) => (
            <li key={l}>
              <span className="alur__nomor" aria-hidden="true">{String(i + 1)}</span>
              <span className="teks-sekunder">{l}</span>
            </li>
          ))}
        </ol>
        <p className="field__bantuan">{p.langkah_catatan}</p>
      </div>
      <div>
        <h3>Data sekolah yang dipakai</h3>
        <ul className="ja-chip">
          {p.data.map((d) => <li key={d}>{d}</li>)}
        </ul>
        <p className="field__bantuan">{p.data_catatan}</p>
      </div>
      <div>
        <h3>Yang berubah</h3>
        <p className="teks-sekunder">{p.hasil}</p>
      </div>
    </aside>
  );
}

export default function JejakAlumni() {
  return (
    <div className="ja">
      <header className="wadah ja-kepala">
        <a className="tautan-panah ja-kembali" href="/layanan/inference-solutions#portofolio">
          <Icon name="kiri" size={18} />
          Kembali ke Inference and Solutions
        </a>
        <div className="ja-kepala__judul">
          <h1>JejakAlumni</h1>
          <span className="badge badge--simulasi">{D.status}</span>
        </div>
        <p className="ja-kepala__sub">Tracer Study Otomatis</p>
        <div className="banner">
          <div className="banner__teks">
            <Icon name="peringatan" size={18} />
            <span>{D.disclaimer}</span>
          </div>
        </div>
      </header>
      <div className="wadah ja-tata">
        <Dasbor />
        <Penjelasan />
      </div>
    </div>
  );
}
