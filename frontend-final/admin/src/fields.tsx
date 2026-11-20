import { useRef, useState, type ReactNode } from 'react';
import { Icon } from '../../src/shared/ui';
import { image } from '../../src/shared/content';
import { Button } from '../../src/components/ui/button';
import type { FieldDef } from './entities';
import Calendar from './Calendar';

export type FieldValue = string | string[] | null;

const ACCEPT = '.jpg,.jpeg,.png,.webp';
const MAKS_GAMBAR = 6;

interface Umum {
  galat: string | null;
  bacaSaja: boolean;
  /** Elemen yang difokus bila field ini galat saat Simpan. */
  daftarkan: (el: HTMLElement | null) => void;
  tampilGalat: (pesan: string | null) => void;
}

/* === Validasi gambar === */
async function validasiBerkas(berkas: File): Promise<string | null> {
  if (berkas.size > 2 * 1024 * 1024) return 'Ukuran file melebihi 2 MB.';
  let b: Uint8Array;
  try {
    b = new Uint8Array(await berkas.slice(0, 12).arrayBuffer());
  } catch {
    return 'File tidak bisa dibaca. Coba pilih ulang.';
  }
  const teks = String.fromCharCode(...b);
  const jpg = b[0] === 0xff && b[1] === 0xd8 && b[2] === 0xff;
  const png = b[0] === 0x89 && teks.slice(1, 4) === 'PNG';
  const webp = teks.slice(0, 4) === 'RIFF' && teks.slice(8, 12) === 'WEBP';
  return jpg || png || webp ? null : 'Tipe file harus JPG, PNG, atau WEBP.';
}

function Judul({ f }: { f: FieldDef }) {
  return (
    <>
      {f.label}
      {f.wajib && <span className="field__wajib" aria-hidden="true"> *</span>}
    </>
  );
}

function Galat({ id, galat }: { id: string; galat: string | null }) {
  return <p className="field__galat" id={`${id}-galat`} hidden={!galat}>{galat}</p>;
}

// Input file bawaan browser memakai bahasa browser, jadi tombolnya dibuat sendiri.
function LabelBerkas({ id, hidden = false, children }: { id: string; hidden?: boolean; children: ReactNode }) {
  return (
    <label className="tombol tombol--kedua tombol--kecil pilih-berkas__tombol" htmlFor={id} hidden={hidden}>
      <Icon name="tambah" size={16} />
      {children}
    </label>
  );
}

type TextDef = Exclude<FieldDef, { tipe: 'gambar' | 'galeri' }>;

export function TextField({ f, value, awal, galat, bacaSaja, daftarkan, onValue, validasi }: Umum & {
  f: TextDef;
  value: string;
  /** Nilai awal form, untuk aturan judul lama yang melebihi batas. */
  awal: string;
  onValue: (v: string) => void;
  validasi: (v: string) => void;
}) {
  const id = `f-${f.nama}`;
  const maks = f.tipe === 'teks' ? f.maks : undefined;
  const [potong, setPotong] = useState<string | null>(null);
  const nilaiSaatFokus = useRef(value);
  // Judul lama yang sudah melebihi batas tidak dipotong otomatis: maxlength dipasang hanya bila nilai awal muat.
  const maxLength = maks !== undefined && awal.length <= maks ? maks : undefined;
  const n = value.trim().length;
  const describedby = [f.bantuan && `${id}-bantuan`, maks !== undefined && `${id}-hitung`, maks !== undefined && `${id}-potong`, `${id}-galat`].filter(Boolean).join(' ');

  function ubah(v: string) {
    onValue(v);
    if (maks === undefined) return;
    setPotong(null);
    if (v.trim().length > maks || galat) validasi(v);
  }

  const atribut = {
    id,
    className: 'input',
    disabled: bacaSaja,
    required: f.wajib,
    'aria-describedby': describedby,
    'aria-invalid': galat ? true : undefined,
  };

  let kontrol: ReactNode;
  if (f.tipe === 'tanggal') {
    // Pesan tanggal langsung tampil setelah mengetik atau memilih, bukan menunggu Simpan.
    kontrol = (
      <Calendar
        {...atribut}
        label={f.label}
        value={value}
        inputRef={daftarkan}
        onChange={onValue}
        onPick={validasi}
        onFocus={() => {
          nilaiSaatFokus.current = value;
        }}
        onBlur={() => {
          if (value !== nilaiSaatFokus.current) validasi(value);
        }}
      />
    );
  } else if (f.tipe === 'area') {
    kontrol = <textarea {...atribut} ref={daftarkan} value={value} onChange={(e) => ubah(e.target.value)} />;
  } else {
    kontrol = (
      <input
        {...atribut}
        ref={daftarkan}
        type={f.tipe === 'angka' ? 'number' : 'text'}
        min={f.tipe === 'angka' ? 1 : undefined}
        step={f.tipe === 'angka' ? 1 : undefined}
        maxLength={maxLength}
        value={value}
        onChange={(e) => ubah(e.target.value)}
        onPaste={(e) => {
          // shortcut: teks yang diseret (drop) terpotong tanpa pesan, tambahkan handler drop bila ternyata dipakai admin.
          if (maxLength === undefined) return;
          const input = e.currentTarget;
          const sisa = maxLength - (input.value.length - ((input.selectionEnd ?? 0) - (input.selectionStart ?? 0)));
          const hilang = e.clipboardData.getData('text').trimEnd().length - sisa;
          if (hilang <= 0) return;
          // Ditunda agar tampil setelah event input dari paste, yang menyembunyikan pesan lama.
          setTimeout(() => setPotong(`${hilang} karakter terakhir dipotong karena ${f.label} maksimal ${maxLength} karakter.`));
        }}
      />
    );
  }

  return (
    <div className="field">
      <label htmlFor={id}><Judul f={f} /></label>
      {kontrol}
      {/* Penghitung hanya untuk field ber-batas; pengumuman pembaca layar terpisah agar tidak berbunyi tiap ketukan. */}
      {maks !== undefined && (
        <>
          <p className={n > maks ? 'field__hitung field__hitung--lebih' : 'field__hitung'} id={`${id}-hitung`}>{`${n}/${maks}`}</p>
          <p className="sr-only" role="status">
            {n > maks ? `Melebihi batas ${maks} karakter.` : n >= maks - 10 ? `${n} dari ${maks} karakter.` : ''}
          </p>
          <p className="field__bantuan" id={`${id}-potong`} role="status" hidden={!potong}>{potong}</p>
        </>
      )}
      {f.bantuan && <p className="field__bantuan" id={`${id}-bantuan`}>{f.bantuan}</p>}
      {f.tipe === 'tanggal' && <p className="field__bantuan">Pilih dari kalender atau ketik DD-MM-YYYY. Tidak boleh setelah hari ini.</p>}
      <Galat id={id} galat={galat} />
    </div>
  );
}

export function ImageField({ f, value, galat, bacaSaja, daftarkan, tampilGalat, onValue }: Umum & {
  f: Extract<FieldDef, { tipe: 'gambar' }>;
  value: string | null;
  onValue: (v: string | null) => void;
}) {
  const id = `f-${f.nama}`;
  const input = useRef<HTMLInputElement | null>(null);
  const describedby = [f.bantuan && `${id}-bantuan`, `${id}-galat`].filter(Boolean).join(' ');

  return (
    <div className="field">
      <label htmlFor={id}><Judul f={f} /></label>
      <div className="unggah">
        <div className="unggah__pratinjau">
          {value ? <img src={image(value)} alt="Pratinjau gambar" /> : <span className="teks-sekunder">Belum ada gambar</span>}
          {!f.wajib && value && !bacaSaja && (
            <Button
              variant="bahaya"
              size="kecil"
              onClick={() => {
                onValue(null);
                if (input.current) input.current.value = '';
              }}
            >
              Hapus foto
            </Button>
          )}
        </div>
        <input
          className="pilih-berkas"
          id={id}
          type="file"
          accept={ACCEPT}
          disabled={bacaSaja}
          aria-describedby={describedby}
          ref={(el) => {
            input.current = el;
            daftarkan(el);
          }}
          onChange={async (e) => {
            const pemilih = e.currentTarget;
            const berkas = pemilih.files?.[0];
            if (!berkas) return;
            const pesan = await validasiBerkas(berkas);
            tampilGalat(pesan);
            if (pesan) {
              pemilih.value = '';
              return;
            }
            onValue(URL.createObjectURL(berkas));
          }}
        />
        {!bacaSaja && <LabelBerkas id={id}>Pilih gambar</LabelBerkas>}
      </div>
      {f.bantuan && <p className="field__bantuan" id={`${id}-bantuan`}>{f.bantuan}</p>}
      <p className="field__bantuan">JPG, PNG, atau WEBP. Maksimal 2 MB.</p>
      <Galat id={id} galat={galat} />
    </div>
  );
}

export function GalleryField({ f, value, galat, bacaSaja, daftarkan, tampilGalat, onValue }: Umum & {
  f: Extract<FieldDef, { tipe: 'galeri' }>;
  value: string[];
  onValue: (v: string[]) => void;
}) {
  const id = `f-${f.nama}`;
  const penuh = bacaSaja || value.length >= MAKS_GAMBAR;

  function gantiDi(i: number) {
    const pemilih = document.createElement('input');
    pemilih.type = 'file';
    pemilih.accept = ACCEPT;
    pemilih.addEventListener('change', async () => {
      const berkas = pemilih.files?.[0];
      if (!berkas) return;
      const pesan = await validasiBerkas(berkas);
      tampilGalat(pesan);
      if (!pesan) onValue(value.map((key, n) => (n === i ? URL.createObjectURL(berkas) : key)));
    });
    pemilih.click();
  }

  return (
    <div className="field">
      <label htmlFor={id}><Judul f={f} /></label>
      <ol className="galeri-admin">
        {value.map((key, i) => (
          <li key={`${i}-${key}`}>
            <img src={image(key)} alt={`Gambar ${i + 1}`} />
            <span>
              {`Gambar ${i + 1}`}
              {i === 0 && <strong>{' · Headline'}</strong>}
            </span>
            {!bacaSaja && (
              <span className="galeri-admin__aksi">
                <Button variant="kedua" size="kecil" aria-label={`Ganti gambar ${i + 1}`} onClick={() => gantiDi(i)}>Ganti</Button>
                <Button
                  variant="bahaya"
                  size="kecil"
                  aria-label={`Hapus gambar ${i + 1}`}
                  onClick={() => {
                    onValue(value.filter((_, n) => n !== i));
                    tampilGalat(null);
                  }}
                >
                  Hapus
                </Button>
              </span>
            )}
          </li>
        ))}
      </ol>
      <input
        className="pilih-berkas"
        id={id}
        type="file"
        accept={ACCEPT}
        multiple
        disabled={penuh}
        aria-describedby={`${id}-galat`}
        ref={daftarkan}
        onChange={async (e) => {
          const pemilih = e.currentTarget;
          const berkas = [...(pemilih.files ?? [])];
          pemilih.value = '';
          const daftar = [...value];
          for (const b of berkas) {
            if (daftar.length >= MAKS_GAMBAR) {
              tampilGalat(`Maksimal ${MAKS_GAMBAR} gambar.`);
              break;
            }
            const pesan = await validasiBerkas(b);
            tampilGalat(pesan);
            if (pesan) break;
            daftar.push(URL.createObjectURL(b));
          }
          onValue(daftar);
        }}
      />
      <LabelBerkas id={id} hidden={penuh}>Tambah gambar</LabelBerkas>
      <p className="field__bantuan">1 sampai 6 gambar. Gambar pertama menjadi headline. JPG, PNG, atau WEBP, maksimal 2 MB per file.</p>
      <Galat id={id} galat={galat} />
    </div>
  );
}
