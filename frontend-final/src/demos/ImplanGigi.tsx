/* Demo Perencanaan Implan Gigi. Ilustrasi konseptual statis (ImplantDiagram1-4);
   kontrol hanya mengganti lapisan yang sudah disiapkan. Tidak ada inferensi atau input data medis. */
import { useRef, useState, type ComponentType } from 'react';
import { flushSync } from 'react-dom';
import { konten } from '../shared/content';
import { Icon } from '../shared/ui';
import ImplantDiagram1 from './ImplantDiagram1';
import ImplantDiagram2 from './ImplantDiagram2';
import ImplantDiagram3 from './ImplantDiagram3';
import ImplantDiagram4 from './ImplantDiagram4';

export interface DiagramProps {
  label: string;
  /** Lapisan toggle (data-lapisan) yang menyala. */
  nyala: Record<string, boolean>;
  /** Nilai kontrol pilihan, mis. { slice: '3' } atau { sudut: 'perspektif' }. */
  pilihan: Record<string, string>;
}

interface Opsi {
  nilai: string;
  label: string;
  keterangan: string;
}

interface Kontrol {
  tipe: string;
  kunci: string;
  label: string;
  awal?: string | undefined;
  opsi?: Opsi[] | undefined;
}

interface Pasangan {
  label: string;
  nilai: string;
}

interface Tahap {
  singkat: string;
  judul: string;
  inti: string;
  ditunjukkan: string;
  peran_ai: string;
  batasan: string;
  ilustrasi: string;
  keterangan?: string | null | undefined;
  kontrol: Kontrol[];
  legenda: { kelas: string; teks: string }[];
  nilai?: Pasangan[] | null | undefined;
  hasil?: Pasangan[] | null | undefined;
  hasil_catatan?: string | undefined;
  teknis?: { judul: string; pengantar: string; butir: string[] } | null | undefined;
}

interface ImplanData {
  judul: string;
  deskripsi: string;
  disclaimer: string;
  label: string[];
  konteks: { judul: string; teks: string }[];
  catatan_demo: string;
  tahap: Tahap[];
  cara_kerja: { judul: string; teks: string }[];
  hasil_pengantar: string;
  hasil: { judul: string; nilai: string; metrik: string; rincian: string }[];
  hasil_batasan: string[];
  statistik: string[];
  batasan: string[];
}

const D: ImplanData = konten.implan_gigi;
const DIAGRAM: ComponentType<DiagramProps>[] = [ImplantDiagram1, ImplantDiagram2, ImplantDiagram3, ImplantDiagram4];

/** Satu tahap: gambar, kontrol, dan penjelasan. Status kontrol tetap tersimpan saat tahap disembunyikan. */
function TahapDemo({ t, i, hidden, judulRef }: {
  t: Tahap;
  i: number;
  hidden: boolean;
  judulRef: (node: HTMLHeadingElement | null) => void;
}) {
  const [nyala, setNyala] = useState<Record<string, boolean>>(() =>
    Object.fromEntries(t.kontrol.filter((c) => c.tipe === 'toggle').map((c) => [c.kunci, true])));
  const [pilihan, setPilihan] = useState<Record<string, string>>(() =>
    Object.fromEntries(t.kontrol.filter((c) => c.tipe !== 'toggle').map((c) => [c.kunci, c.awal ?? ''])));

  let keterangan = t.keterangan ?? '';
  const aktif: string[] = [];
  for (const c of t.kontrol) {
    if (c.tipe === 'toggle') {
      if (nyala[c.kunci]) aktif.push(c.label.replace('Tampilkan ', ''));
    } else {
      keterangan = c.opsi?.find((o) => o.nilai === pilihan[c.kunci])?.keterangan ?? keterangan;
    }
  }
  const label = `${t.ilustrasi} ${keterangan} Lapisan yang ditampilkan: ${aktif.join(', ') || 'tidak ada'}.`;
  const Diagram = DIAGRAM[i];

  return (
    <div className="ig-tahap" hidden={hidden}>
      <figure className="ig-gambar">
        <div className="ig-gambar__kepala">
          <span className="badge badge--simulasi">Ilustrasi Konseptual</span>
          <span className="ig-gambar__nonklinis">Nonklinis, bukan citra pasien</span>
        </div>
        {Diagram && <Diagram label={label} nyala={nyala} pilihan={pilihan} />}
        <figcaption
          className="ig-gambar__keterangan"
          aria-live={t.kontrol.some((c) => c.tipe === 'pilihan') ? 'polite' : undefined}
        >
          {keterangan}
        </figcaption>
        <ul className="ig-legenda">
          {t.legenda.map((l) => (
            <li key={l.kelas}>
              <span className={`ig-sw ig-sw--${l.kelas}`} aria-hidden="true" />
              {l.teks}
            </li>
          ))}
        </ul>
      </figure>
      <div className="ig-kontrol">
        {t.kontrol.map((c) => (c.tipe === 'toggle'
          ? (
            <label key={c.kunci} className="ig-pilih">
              <input
                type="checkbox"
                checked={nyala[c.kunci] ?? false}
                data-kunci={c.kunci}
                onChange={(e) => {
                  const nilai = e.currentTarget.checked;
                  setNyala((n) => ({ ...n, [c.kunci]: nilai }));
                }}
              />
              <span>{c.label}</span>
            </label>
          )
          : (
            <fieldset key={c.kunci} className="ig-grup">
              <legend>{c.label}</legend>
              {c.opsi?.map((o) => (
                <label key={o.nilai} className="ig-pilih">
                  <input
                    type="radio"
                    name={`ig-${i}-${c.kunci}`}
                    value={o.nilai}
                    checked={pilihan[c.kunci] === o.nilai}
                    data-kunci={c.kunci}
                    onChange={() => setPilihan((p) => ({ ...p, [c.kunci]: o.nilai }))}
                  />
                  <span>{o.label}</span>
                </label>
              ))}
            </fieldset>
          )))}
        {t.nilai && (
          <dl className="ig-nilai">
            {t.nilai.map((n) => (
              <div key={n.label}>
                <dt>{n.label}</dt>
                <dd>{n.nilai}</dd>
              </div>
            ))}
          </dl>
        )}
      </div>
      <div className="ig-panel">
        <p className="tentang-label">{`Tahap ${i + 1} dari ${D.tahap.length}`}</p>
        <h3 ref={judulRef} tabIndex={-1}>{t.judul}</h3>
        <p>{t.inti}</p>
        <div>
          <h4>Apa yang ditunjukkan</h4>
          <p className="teks-sekunder">{t.ditunjukkan}</p>
        </div>
        <div>
          <h4>Peran AI dalam penelitian</h4>
          <p className="teks-sekunder">{t.peran_ai}</p>
        </div>
        {t.hasil && (
          <div>
            <h4>Hasil penelitian</h4>
            <dl className="ig-nilai ig-nilai--hasil">
              {t.hasil.map((h) => (
                <div key={h.label}>
                  <dt>{h.label}</dt>
                  <dd>{h.nilai}</dd>
                </div>
              ))}
            </dl>
            <p className="field__bantuan">{t.hasil_catatan}</p>
          </div>
        )}
        <div className="ig-panel__batasan">
          <h4>Batasan demo</h4>
          <p>{t.batasan}</p>
        </div>
        {t.teknis && (
          <details className="ig-lipat">
            <summary>{`Detail teknis: ${t.teknis.judul}`}</summary>
            <p>{t.teknis.pengantar}</p>
            <ul className="ig-daftar">
              {t.teknis.butir.map((b) => <li key={b}>{b}</li>)}
            </ul>
          </details>
        )}
      </div>
    </div>
  );
}

function BagianDemo() {
  const [aktif, setAktif] = useState(0);
  const judul = useRef<(HTMLHeadingElement | null)[]>([]);
  const stepper = useRef<HTMLOListElement>(null);
  const demo = useRef<HTMLElement>(null);
  const akhir = D.tahap.length - 1;

  function pilih(n: number) {
    flushSync(() => setAktif(n));
    judul.current[n]?.focus({ preventScroll: true });
    if ((stepper.current?.getBoundingClientRect().top ?? 0) < 0) demo.current?.scrollIntoView({ block: 'start' });
  }

  return (
    <section ref={demo} className="bagian bagian--muda ig-demo" id="demo" aria-labelledby="ig-j-demo">
      <div className="wadah">
        <div className="ig-kepala-demo">
          <h2 id="ig-j-demo">Jelajahi empat tahap penelitian.</h2>
          <p className="bagian__sub">{D.catatan_demo}</p>
        </div>
        <ol ref={stepper} className="ig-stepper">
          {D.tahap.map((t, i) => (
            <li key={t.singkat}>
              <button
                type="button"
                className="ig-langkah"
                aria-current={i === aktif ? 'step' : undefined}
                onClick={() => pilih(i)}
              >
                <span className="ig-langkah__nomor" aria-hidden="true">{String(i + 1)}</span>
                <span>
                  <span className="sr-only">{`Tahap ${i + 1}: `}</span>
                  {t.singkat}
                </span>
              </button>
            </li>
          ))}
        </ol>
        {D.tahap.map((t, i) => (
          <TahapDemo
            key={t.singkat}
            t={t}
            i={i}
            hidden={i !== aktif}
            judulRef={(node) => {
              judul.current[i] = node;
            }}
          />
        ))}
        <div className="ig-nav">
          <button type="button" className="tombol tombol--kedua" disabled={aktif === 0} onClick={() => pilih(aktif - 1)}>
            Sebelumnya
          </button>
          <button type="button" className="tombol tombol--utama" hidden={aktif === akhir} onClick={() => pilih(aktif + 1)}>
            Berikutnya
          </button>
          <a className="tombol tombol--utama" href="#hasil" hidden={aktif !== akhir}>
            <Icon name="panah" size={18} />
            Lihat Hasil Penelitian
          </a>
        </div>
      </div>
    </section>
  );
}

export default function ImplanGigi() {
  return (
    <>
      <header className="wadah ig-kepala">
        <a className="tautan-panah ig-kembali" href="/layanan/inference-solutions#portofolio">
          <Icon name="kiri" size={18} />
          Kembali ke Inference and Solutions
        </a>
        <h1>{D.judul}</h1>
        <p className="ig-kepala__sub">{D.deskripsi}</p>
        <ul className="ig-label">
          {D.label.map((l) => (
            <li key={l}>
              <span className="badge badge--simulasi">{l}</span>
            </li>
          ))}
        </ul>
        <div className="banner">
          <div className="banner__teks">
            <Icon name="peringatan" size={18} />
            <span>{D.disclaimer}</span>
          </div>
        </div>
      </header>
      <section className="bagian wadah" aria-labelledby="ig-j-konteks">
        <h2 id="ig-j-konteks">Mengapa perlu analisis yang teliti?</h2>
        <ul className="alur">
          {D.konteks.map((k) => (
            <li key={k.judul}>
              <h3>{k.judul}</h3>
              <p className="teks-sekunder">{k.teks}</p>
            </li>
          ))}
        </ul>
      </section>
      <BagianDemo />
      <section className="bagian wadah" aria-labelledby="ig-j-cara">
        <h2 id="ig-j-cara">Cara kerja AI dalam penelitian.</h2>
        <div className="ig-daftar-lipat">
          {D.cara_kerja.map((c) => (
            <details key={c.judul} className="ig-lipat">
              <summary>{c.judul}</summary>
              <p>{c.teks}</p>
            </details>
          ))}
        </div>
      </section>
      <section className="bagian bagian--muda" id="hasil" aria-labelledby="ig-j-hasil">
        <div className="wadah">
          <h2 id="ig-j-hasil">Hasil penelitian yang dilaporkan.</h2>
          <p className="bagian__sub">{D.hasil_pengantar}</p>
          <ul className="ig-metrik">
            {D.hasil.map((h) => (
              <li key={h.judul}>
                <h3>{h.judul}</h3>
                <p className="ig-metrik__nilai">{h.nilai}</p>
                <p>{h.metrik}</p>
                <p className="field__bantuan">{h.rincian}</p>
              </li>
            ))}
          </ul>
          <ul className="ig-daftar">
            {D.hasil_batasan.map((b) => <li key={b}>{b}</li>)}
          </ul>
          <details className="ig-lipat ig-lipat--lebar">
            <summary>Catatan statistik</summary>
            <ul className="ig-daftar">
              {D.statistik.map((s) => <li key={s}>{s}</li>)}
            </ul>
          </details>
        </div>
      </section>
      <section className="bagian wadah" aria-labelledby="ig-j-batasan">
        <h2 id="ig-j-batasan">Batasan demo.</h2>
        <ul className="ig-daftar">
          {D.batasan.map((b) => <li key={b}>{b}</li>)}
        </ul>
      </section>
    </>
  );
}
