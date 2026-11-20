/* Demo DocNext. Dokumen sintetis di konten.document_intelligence.
   Ekstraksi dan keterkaitan deskripsi disiapkan (tidak ada model AI); hanya perbandingan yang dihitung di sini. */
import { useRef, useState, type ReactNode } from 'react';
import { konten } from '../shared/content';
import { Icon } from '../shared/ui';

type Jenis = 'po' | 'penerimaan' | 'invoice';
type Field = 'refPo' | 'pemasok' | 'kode' | 'deskripsi' | 'qty' | 'harga' | 'subtotal';
type Status = 'sesuai' | 'ditinjau' | 'verifikasi';

interface Dokumen {
  nomor: string;
  tanggal: string;
  refPo?: string;
  pemasok: string;
  kode: string;
  deskripsi: string;
  qty: number;
  satuan: string;
  harga?: number;
  subtotal?: number;
}

type DokumenHarga = Dokumen & { harga: number; subtotal: number };

interface Temuan {
  judul: string;
  field: Field;
  dokumen: Jenis[];
  mengapa: string;
  periksa: string[];
}

interface Skenario {
  label: string;
  harapan: Status;
  penjelasan: string;
  catatan: string;
  keterkaitan?: string | undefined;
  po: DokumenHarga;
  penerimaan: Dokumen & { refPo: string };
  invoice: DokumenHarga & { refPo: string };
  temuan: Temuan[];
}

interface DocData {
  judul: string;
  subjudul: string;
  deskripsi: string;
  status: string;
  disclaimer: string;
  pembeli: string;
  lokasi: string;
  catatan_ekstraksi: string;
  label_dokumen: Record<Jenis, Partial<Record<Field, string>>>;
  skenario: Skenario[];
  masalah: { judul: string; teks: string }[];
  lapisan: { pengantar: string[]; kelompok: { label: string; judul: string; butir: string[] }[] };
  penerapan: { teks: string; contoh: string[]; catatan: string };
  deployment: {
    teks: string;
    alur: string[];
    alur_label: string;
    opsi: { judul: string; teks: string }[];
    prasyarat: string[];
    catatan: string;
  };
  batasan: string[];
}

/* JSON statis: literal union (jenis, field, status) dipersempit di batas impor; dicek penjaga di bawah. */
const D = konten.document_intelligence as DocData;
const NAMA_DOK: Record<Jenis, string> = { po: 'Purchase Order', penerimaan: 'Dokumen Penerimaan', invoice: 'Invoice' };
const JENIS: Jenis[] = ['po', 'penerimaan', 'invoice'];
const NAMA_FIELD: Record<Field, string> = {
  refPo: 'Referensi PO',
  pemasok: 'Pemasok',
  kode: 'Kode barang',
  deskripsi: 'Deskripsi barang',
  qty: 'Kuantitas',
  harga: 'Harga satuan',
  subtotal: 'Subtotal',
};
const FIELD = Object.keys(NAMA_FIELD) as Field[];
const STATUS: Record<Status, [string, string]> = {
  sesuai: ['✓', 'Sesuai'],
  ditinjau: ['!', 'Perlu Ditinjau'],
  verifikasi: ['?', 'Perlu Verifikasi'],
};
const rp = (n: number) => `Rp${n.toLocaleString('id-ID')}`;

/** Nilai field seperti tertulis di dokumen; pada PO, referensinya adalah nomor PO itu sendiri. */
function nilai(dok: Dokumen, jenis: Jenis, kunci: Field): string | null {
  switch (kunci) {
    case 'refPo':
      return jenis === 'po' ? dok.nomor : dok.refPo ?? null;
    case 'qty':
      return `${dok.qty} ${dok.satuan}`;
    case 'harga':
    case 'subtotal': {
      const angka = dok[kunci];
      return angka === undefined ? null : rp(angka);
    }
    default:
      return dok[kunci];
  }
}

interface BarisBanding {
  kunci: Field;
  oleh: 'aturan' | 'ai';
  status: Status;
  catatan?: string | null;
  nilai: (string | null)[];
}

/** Aturan bisnis deterministik. Hanya baris identitas yang memakai interpretasi makna. */
function periksa(s: Skenario) {
  const { po, penerimaan: gr, invoice: inv } = s;
  const sama = (...a: (string | number)[]) => a.every((x) => x === a[0]);
  const subtotalBenar = inv.subtotal === gr.qty * inv.harga;
  const aturan: Omit<BarisBanding, 'nilai'>[] = [
    { kunci: 'refPo', oleh: 'aturan', status: sama(po.nomor, gr.refPo, inv.refPo) ? 'sesuai' : 'ditinjau' },
    { kunci: 'pemasok', oleh: 'aturan', status: sama(po.pemasok, gr.pemasok, inv.pemasok) ? 'sesuai' : 'ditinjau' },
    {
      kunci: 'deskripsi',
      oleh: 'ai',
      status: sama(po.kode, gr.kode, inv.kode) && sama(po.deskripsi, gr.deskripsi, inv.deskripsi) ? 'sesuai' : 'verifikasi',
    },
    { kunci: 'qty', oleh: 'aturan', status: sama(po.qty, gr.qty, inv.qty) ? 'sesuai' : 'ditinjau' },
    { kunci: 'harga', oleh: 'aturan', status: po.harga === inv.harga ? 'sesuai' : 'ditinjau' },
    {
      kunci: 'subtotal',
      oleh: 'aturan',
      status: subtotalBenar ? 'sesuai' : 'ditinjau',
      catatan: subtotalBenar
        ? null
        : `Subtotal invoice berbeda dari jumlah diterima dikali harga satuan (${rp(gr.qty * inv.harga)}).`,
    },
  ];
  const baris: BarisBanding[] = aturan.map((b) => ({ ...b, nilai: JENIS.map((j) => nilai(s[j], j, b.kunci)) }));
  const status: Status = baris.some((b) => b.status === 'ditinjau')
    ? 'ditinjau'
    : baris.some((b) => b.status === 'verifikasi') ? 'verifikasi' : 'sesuai';
  return { baris, status, selisihUnit: inv.qty - gr.qty, selisihNilai: (inv.qty - gr.qty) * inv.harga };
}

// Penjaga data sintetis: pelanggaran terlihat di konsol saat halaman dibuka.
for (const s of D.skenario) {
  console.assert(
    s.po.qty * s.po.harga === s.po.subtotal && s.invoice.qty * s.invoice.harga === s.invoice.subtotal,
    `Subtotal tidak konsisten: ${s.label}`,
  );
  console.assert(periksa(s).status === s.harapan, `Status tidak sesuai harapan: ${s.label}`);
}

function Lencana({ status }: { status: Status }) {
  return (
    <span className={`badge dn-status dn-status--${status}`}>
      <span aria-hidden="true">{`${STATUS[status][0]} `}</span>
      {STATUS[status][1]}
    </span>
  );
}

function TagOleh({ oleh }: { oleh: 'aturan' | 'ai' }) {
  return oleh === 'ai'
    ? <span className="dn-tag dn-tag--ai">Simulasi AI</span>
    : <span className="dn-tag">Aturan Bisnis</span>;
}

function MetaDok({ pasangan }: { pasangan: [string | undefined, ReactNode][] }) {
  return (
    <dl className="dn-dok__meta">
      {pasangan.map(([l, isi], i) => (
        <div key={i}>
          <dt>{l}</dt>
          <dd>{isi}</dd>
        </div>
      ))}
    </dl>
  );
}

function TabelBarang({ kolom }: { kolom: [string | undefined, ReactNode][] }) {
  return (
    <table className="dn-tabel">
      <thead>
        <tr>
          {kolom.map(([k], i) => <th key={i} scope="col">{k}</th>)}
        </tr>
      </thead>
      <tbody>
        <tr>
          {kolom.map(([k, isi], i) => <td key={i} data-kolom={k}>{isi}</td>)}
        </tr>
      </tbody>
    </table>
  );
}

/* Tata letak tiap dokumen sengaja berbeda, seperti dokumen dari pihak yang berbeda. */
function Pratinjau({ s, jenis, field }: { s: Skenario; jenis: Jenis; field: Field | null }) {
  const f = (kunci: Field, teks: string) => (
    <span className={kunci === field ? 'dn-f dn-f--aktif' : 'dn-f'} data-field={kunci}>{teks}</span>
  );
  const kepala = (teks: string) => (
    <header className="dn-dok__kepala">
      <div>
        <strong>{jenis === 'invoice' ? f('pemasok', s.invoice.pemasok) : D.pembeli}</strong>
        <p className="field__bantuan">Dokumen sintetis</p>
      </div>
      <h4>{teks}</h4>
    </header>
  );
  const L = D.label_dokumen[jenis];

  if (jenis === 'po') {
    const d = s.po;
    return (
      <>
        {kepala('PURCHASE ORDER')}
        <MetaDok pasangan={[[L.refPo, f('refPo', d.nomor)], ['Tanggal PO', d.tanggal], [L.pemasok, f('pemasok', d.pemasok)], ['Pembeli', D.pembeli]]} />
        <TabelBarang
          kolom={[
            [L.kode, f('kode', d.kode)],
            [L.deskripsi, f('deskripsi', d.deskripsi)],
            [L.qty, f('qty', `${d.qty} ${d.satuan}`)],
            [L.harga, f('harga', rp(d.harga))],
            [L.subtotal, f('subtotal', rp(d.subtotal))],
          ]}
        />
      </>
    );
  }
  if (jenis === 'penerimaan') {
    const d = s.penerimaan;
    return (
      <>
        {kepala('BUKTI PENERIMAAN BARANG')}
        <MetaDok
          pasangan={[
            ['No. Penerimaan', d.nomor],
            ['Tgl. Terima', d.tanggal],
            [L.refPo, f('refPo', d.refPo)],
            [L.pemasok, f('pemasok', d.pemasok)],
            ['Lokasi', D.lokasi],
          ]}
        />
        <div className="dn-dok__blok">
          <MetaDok pasangan={[[L.kode, f('kode', d.kode)], [L.deskripsi, f('deskripsi', d.deskripsi)], [L.qty, f('qty', `${d.qty} ${d.satuan}`)]]} />
        </div>
      </>
    );
  }
  const d = s.invoice;
  return (
    <>
      {kepala('INVOICE')}
      <MetaDok pasangan={[['Invoice No.', d.nomor], ['Tanggal', d.tanggal], [L.refPo, f('refPo', d.refPo)], ['Ditagihkan kepada', D.pembeli]]} />
      <TabelBarang
        kolom={[
          [L.kode, f('kode', d.kode)],
          [L.deskripsi, f('deskripsi', d.deskripsi)],
          [L.qty, f('qty', `${d.qty} ${d.satuan}`)],
          [L.harga, f('harga', rp(d.harga))],
        ]}
      />
      <p className="dn-dok__total">
        {`${L.subtotal}: `}
        {f('subtotal', rp(d.subtotal))}
      </p>
      <p className="field__bantuan">Pajak tidak dihitung dalam demo.</p>
    </>
  );
}

function RuangKerja() {
  const [indeks, setIndeks] = useState(0);
  const [jenis, setJenis] = useState<Jenis>('po');
  const [field, setField] = useState<Field | null>(null);
  const pratinjau = useRef<HTMLElement>(null);
  const s = D.skenario[indeks];
  if (!s) return null;
  const h = periksa(s);
  const label = D.label_dokumen[jenis];
  const nilaiAktif = field && nilai(s[jenis], jenis, field);

  let sumber: string;
  if (!field) sumber = `Menampilkan ${NAMA_DOK[jenis]}. Pilih hasil ekstraksi untuk melihat sumbernya.`;
  else if (nilaiAktif === null) sumber = `${NAMA_FIELD[field]} tidak tercantum pada ${NAMA_DOK[jenis]}.`;
  else sumber = `Sumber: ${NAMA_DOK[jenis]}, label "${label[field]}", nilai ${nilaiAktif}.`;

  function keSumber(j: Jenis, k: Field) {
    setJenis(j);
    setField(k);
    if ((pratinjau.current?.getBoundingClientRect().top ?? 0) < 0) pratinjau.current?.scrollIntoView({ block: 'center' });
  }

  return (
    <div className="dn-kerja">
      <fieldset className="dn-pilih">
        <legend>Skenario transaksi</legend>
        {D.skenario.map((x, i) => (
          <label key={x.label} className="dn-chip">
            <input
              type="radio"
              name="dn-skenario"
              value={String(i)}
              checked={i === indeks}
              onChange={() => {
                setIndeks(i);
                setJenis('po');
                setField(null);
              }}
            />
            <span>{`${i + 1}. ${x.label}`}</span>
          </label>
        ))}
      </fieldset>
      <p className="sr-only" role="status" aria-atomic="true">
        {`${s.label}. ${STATUS[h.status][0]} ${STATUS[h.status][1]}. ${s.penjelasan}`}
      </p>
      <div className="dn-banner">
        <Lencana status={h.status} />
        <div>
          <p>{s.penjelasan}</p>
          <p className="field__bantuan">{s.catatan}</p>
        </div>
      </div>
      <div className="dn-kerja__kolom">
        <div className="dn-kerja__dokumen">
          <fieldset className="dn-pilih">
            <legend>Dokumen</legend>
            {JENIS.map((j) => (
              <label key={j} className="dn-chip">
                <input type="radio" name="dn-dokumen" value={j} checked={j === jenis} onChange={() => setJenis(j)} />
                <span>{NAMA_DOK[j]}</span>
              </label>
            ))}
          </fieldset>
          <p className="dn-sumber" role="status">{sumber}</p>
          <article ref={pratinjau} className="dn-dok" aria-label="Pratinjau dokumen sintetis">
            <Pratinjau s={s} jenis={jenis} field={field} />
          </article>
        </div>
        <div className="dn-kerja__analisis">
          <section className="dn-panel" aria-labelledby="dn-j-ekstraksi">
            <div className="dn-panel__kepala">
              <h3 id="dn-j-ekstraksi">Hasil ekstraksi</h3>
              <TagOleh oleh="ai" />
            </div>
            <ul className="dn-field">
              {FIELD.filter((k) => label[k] !== undefined).map((k) => (
                <li key={k}>
                  <button
                    type="button"
                    className="dn-field__tombol"
                    data-field={k}
                    aria-pressed={k === field}
                    onClick={() => setField(k)}
                  >
                    <span className="dn-field__nama">{NAMA_FIELD[k]}</span>
                    <span className="dn-field__nilai">{nilai(s[jenis], jenis, k)}</span>
                    <span className="field__bantuan">{`Label di dokumen: ${label[k]}`}</span>
                  </button>
                </li>
              ))}
            </ul>
            <p className="field__bantuan">{D.catatan_ekstraksi}</p>
          </section>
          <section className="dn-panel" aria-labelledby="dn-j-banding">
            <h3 id="dn-j-banding">Perbandingan</h3>
            <ul className="dn-banding">
              {h.baris.map((b) => (
                <li key={b.kunci}>
                  <div className="dn-banding__kepala">
                    <strong>{b.kunci === 'deskripsi' ? 'Identitas barang' : NAMA_FIELD[b.kunci]}</strong>
                    <Lencana status={b.status} />
                  </div>
                  <dl className="dn-banding__nilai">
                    {JENIS.map((j, i) => (
                      <div key={j}>
                        <dt>{NAMA_DOK[j]}</dt>
                        <dd>{b.nilai[i] ?? 'Tidak tercantum'}</dd>
                      </div>
                    ))}
                  </dl>
                  {b.oleh === 'ai' && s.keterkaitan && <p className="dn-keterkaitan">{s.keterkaitan}</p>}
                  {b.catatan && <p className="dn-keterkaitan">{b.catatan}</p>}
                  <TagOleh oleh={b.oleh} />
                </li>
              ))}
            </ul>
            <details className="dn-lipat">
              <summary>Apa bedanya AI dan aturan bisnis?</summary>
              <p>
                Simulasi AI menandai informasi yang dalam penerapan nyata dapat dibantu model, seperti membaca dokumen dan
                menafsirkan deskripsi. Aturan Bisnis adalah perhitungan biasa yang dijalankan browser pada demo ini.
              </p>
            </details>
          </section>
          <section className="dn-panel" aria-labelledby="dn-j-temuan">
            <h3 id="dn-j-temuan">Temuan untuk peninjau</h3>
            <div className="dn-temuan">
              {s.temuan.length
                ? s.temuan.map((t) => (
                  <article key={t.judul} className="dn-temuan__kartu">
                    <h4>{t.judul}</h4>
                    <dl className="dn-banding__nilai">
                      {t.dokumen.map((j) => (
                        <div key={j}>
                          <dt>{NAMA_DOK[j]}</dt>
                          <dd>{nilai(s[j], j, t.field)}</dd>
                        </div>
                      ))}
                    </dl>
                    {t.field === 'qty' && (
                      <p>{`Selisih ${h.selisihUnit} ${s.invoice.satuan}, senilai ${rp(h.selisihNilai)} sebelum pajak.`}</p>
                    )}
                    <p>{t.mengapa}</p>
                    <p className="dn-temuan__label">Yang perlu diperiksa peninjau</p>
                    <ul className="dn-daftar">
                      {t.periksa.map((p) => <li key={p}>{p}</li>)}
                    </ul>
                    <div className="dn-temuan__aksi">
                      {t.dokumen.map((j) => (
                        <button key={j} type="button" className="tombol tombol--kedua tombol--kecil" onClick={() => keSumber(j, t.field)}>
                          {`Lihat di ${NAMA_DOK[j]}`}
                        </button>
                      ))}
                    </div>
                  </article>
                ))
                : <p className="teks-sekunder">Tidak ada temuan berdasarkan aturan pemeriksaan yang disimulasikan.</p>}
            </div>
          </section>
        </div>
      </div>
    </div>
  );
}

export default function DocNext() {
  const p = D.deployment;
  return (
    <>
      <header className="wadah dn-kepala">
        <a className="tautan-panah dn-kembali" href="/layanan/inference-solutions#portofolio">
          <Icon name="kiri" size={18} />
          Kembali ke Inference and Solutions
        </a>
        <h1>{D.judul}</h1>
        <p className="dn-kepala__sub">{D.subjudul}</p>
        <p className="teks-sekunder dn-kepala__teks">{D.deskripsi}</p>
        <a className="tautan-panah" href="#demo">
          Langsung ke Demo
          <Icon name="panah" size={18} />
        </a>
        <span className="badge badge--simulasi">{D.status}</span>
        <div className="banner">
          <div className="banner__teks">
            <Icon name="peringatan" size={18} />
            <span>{D.disclaimer}</span>
          </div>
        </div>
      </header>
      <section className="bagian wadah" aria-labelledby="dn-j-masalah">
        <h2 id="dn-j-masalah">Satu transaksi, banyak dokumen untuk diperiksa.</h2>
        <ul className="alur">
          {D.masalah.map((m) => (
            <li key={m.judul}>
              <h3>{m.judul}</h3>
              <p className="teks-sekunder">{m.teks}</p>
            </li>
          ))}
        </ul>
      </section>
      <section className="bagian bagian--muda dn-demo" id="demo" aria-labelledby="dn-j-demo">
        <div className="wadah">
          <h2 id="dn-j-demo">Telusuri tiga contoh transaksi.</h2>
          <RuangKerja />
        </div>
      </section>
      <section className="bagian wadah" aria-labelledby="dn-j-lapisan">
        <h2 id="dn-j-lapisan">Peran AI, Aturan Bisnis, dan Peninjau.</h2>
        {D.lapisan.pengantar.map((t) => <p key={t} className="bagian__sub">{t}</p>)}
        <div className="dn-tiga">
          {D.lapisan.kelompok.map((k) => (
            <article key={k.judul} className="jenis">
              <p className="tentang-label">{k.label}</p>
              <h3>{k.judul}</h3>
              <ul className="dn-daftar">
                {k.butir.map((b) => <li key={b}>{b}</li>)}
              </ul>
            </article>
          ))}
        </div>
      </section>
      <section className="bagian bagian--muda" aria-labelledby="dn-j-penerapan">
        <div className="wadah">
          <h2 id="dn-j-penerapan">Satu Pendekatan, Berbagai Kebutuhan Dokumen.</h2>
          <p className="bagian__sub">{D.penerapan.teks}</p>
          <ul className="dn-contoh">
            {D.penerapan.contoh.map((c) => <li key={c}>{c}</li>)}
          </ul>
          <p className="field__bantuan">{D.penerapan.catatan}</p>
        </div>
      </section>
      <section className="bagian wadah" aria-labelledby="dn-j-deploy">
        <h2 id="dn-j-deploy">AI yang Dapat Dirancang untuk Infrastruktur Internal Perusahaan.</h2>
        <p className="bagian__sub">{p.teks}</p>
        <figure className="dn-alur">
          <ol>
            {p.alur.map((a) => <li key={a}>{a}</li>)}
          </ol>
          <figcaption className="field__bantuan">{p.alur_label}</figcaption>
        </figure>
        <div className="jenis-grid">
          {p.opsi.map((o) => (
            <article key={o.judul} className="jenis">
              <h3>{o.judul}</h3>
              <p className="teks-sekunder">{o.teks}</p>
            </article>
          ))}
        </div>
        <details className="dn-lipat">
          <summary>Hal yang perlu dipertimbangkan</summary>
          <ul className="dn-daftar">
            {p.prasyarat.map((x) => <li key={x}>{x}</li>)}
          </ul>
        </details>
        <p className="teks-sekunder">{p.catatan}</p>
      </section>
      <section className="bagian bagian--muda" aria-labelledby="dn-j-batasan">
        <div className="wadah">
          <h2 id="dn-j-batasan">Batasan demo.</h2>
          <ul className="dn-daftar">
            {D.batasan.map((b) => <li key={b}>{b}</li>)}
          </ul>
        </div>
      </section>
    </>
  );
}
