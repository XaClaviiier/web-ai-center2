/* Komponen kalender: kolom teks DD-MM-YYYY plus popup bulan. Tanggal setelah hari ini
   tidak bisa dipilih. "Hari ini" memakai tanggal lokal perangkat, bukan UTC. */
import { useEffect, useRef, useState, type ComponentProps, type KeyboardEvent, type Ref } from 'react';
import { Icon } from '../../src/shared/ui';
import { Format } from '../../src/shared/format';

const BULAN = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'];
const HARI = ['Senin', 'Selasa', 'Rabu', 'Kamis', 'Jumat', 'Sabtu', 'Minggu'];
const LANGKAH: Record<string, number> = { ArrowLeft: -1, ArrowRight: 1, ArrowUp: -7, ArrowDown: 7 };

function hariIni(): Date {
  const n = new Date();
  return new Date(n.getFullYear(), n.getMonth(), n.getDate());
}

const keTeks = (d: Date) => `${String(d.getDate()).padStart(2, '0')}-${String(d.getMonth() + 1).padStart(2, '0')}-${d.getFullYear()}`;
export const hariIniTeks = (): string => keTeks(hariIni());

/** Tanggal lokal dari "DD-MM-YYYY"; null bila format salah atau tanggal tidak ada. */
export function dariTeks(teks: string): Date | null {
  const u = Format.parseTanggal(teks);
  return u ? new Date(u.getUTCFullYear(), u.getUTCMonth(), u.getUTCDate()) : null;
}

export const diMasaDepan = (d: Date): boolean => d > hariIni();
const sama = (a: Date | null, b: Date) => a?.getTime() === b.getTime();
const indeksHari = (d: Date) => (d.getDay() + 6) % 7; // Senin = 0
const awalBulan = (d: Date) => new Date(d.getFullYear(), d.getMonth(), 1);

// Ikon kalender hanya dipakai komponen ini (prototype menambahkannya ke IKON dari kalender.js).
function IkonKalender() {
  return (
    <svg viewBox="0 0 24 24" width={20} height={20} fill="none" stroke="currentColor" strokeWidth={1.8} strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M4 6h16v14H4zM4 10h16M8 3.5V7M16 3.5V7" />
    </svg>
  );
}

type Props = Omit<ComponentProps<'input'>, 'value' | 'onChange' | 'type' | 'ref'> & {
  label: string;
  value: string;
  onChange: (value: string) => void;
  /** Dipanggil saat tanggal dipilih dari popup (setara event `change`). */
  onPick: (value: string) => void;
  inputRef?: Ref<HTMLInputElement>;
};

export default function Calendar({ label, value, onChange, onPick, inputRef, ...props }: Props) {
  const [buka, setBuka] = useState(false);
  const [fokus, setFokus] = useState(hariIni);
  const [bulanTampil, setBulanTampil] = useState(() => awalBulan(hariIni()));
  const pembungkus = useRef<HTMLDivElement>(null);
  const tombolBuka = useRef<HTMLButtonElement>(null);
  const tombolSebelum = useRef<HTMLButtonElement>(null);
  const fokuskanHari = useRef(false);

  useEffect(() => {
    if (!buka) return;
    const diLuar = (e: Event) => {
      if (e.target instanceof Node && !pembungkus.current?.contains(e.target)) setBuka(false);
    };
    document.addEventListener('mousedown', diLuar);
    document.addEventListener('focusin', diLuar);
    return () => {
      document.removeEventListener('mousedown', diLuar);
      document.removeEventListener('focusin', diLuar);
    };
  }, [buka]);

  useEffect(() => {
    if (!buka || !fokuskanHari.current) return;
    fokuskanHari.current = false;
    pembungkus.current?.querySelector<HTMLButtonElement>(`[data-tanggal="${keTeks(fokus)}"]`)?.focus();
  }, [buka, fokus, bulanTampil]);

  function tutup(kembalikanFokus: boolean) {
    setBuka(false);
    if (kembalikanFokus) tombolBuka.current?.focus();
  }

  function bukaPopup() {
    const nilai = dariTeks(value);
    const awal = nilai && !diMasaDepan(nilai) ? nilai : hariIni();
    setFokus(awal);
    setBulanTampil(awalBulan(awal));
    fokuskanHari.current = true;
    setBuka(true);
  }

  function geserBulan(arah: number) {
    const baru = new Date(bulanTampil.getFullYear(), bulanTampil.getMonth() + arah, 1);
    if (baru > hariIni()) return;
    const akhir = new Date(baru.getFullYear(), baru.getMonth() + 1, 0).getDate();
    const fokusBaru = new Date(baru.getFullYear(), baru.getMonth(), Math.min(fokus.getDate(), akhir));
    setBulanTampil(baru);
    setFokus(diMasaDepan(fokusBaru) ? hariIni() : fokusBaru);
    // Tombol "berikutnya" menjadi nonaktif di bulan berjalan, jadi fokus dipindah agar tidak hilang.
    if (new Date(baru.getFullYear(), baru.getMonth() + 1, 1) > hariIni()) tombolSebelum.current?.focus();
  }

  function tombolKeyboard(e: KeyboardEvent<HTMLDivElement>) {
    if (e.key === 'Escape') {
      e.preventDefault();
      tutup(true);
      return;
    }
    const langkah = LANGKAH[e.key];
    if (!langkah || !(e.target instanceof HTMLElement) || !e.target.matches('.kalender__hari')) return;
    e.preventDefault();
    const baru = new Date(fokus.getFullYear(), fokus.getMonth(), fokus.getDate() + langkah);
    if (diMasaDepan(baru)) return;
    fokuskanHari.current = true;
    setFokus(baru);
    setBulanTampil(awalBulan(baru));
  }

  function pilih(d: Date) {
    const teks = keTeks(d);
    onChange(teks);
    onPick(teks);
    tutup(true);
  }

  const terpilih = dariTeks(value);
  const sekarang = hariIni();
  const jumlahHari = new Date(bulanTampil.getFullYear(), bulanTampil.getMonth() + 1, 0).getDate();
  const sel: (Date | null)[] = Array.from({ length: indeksHari(bulanTampil) }, () => null);
  for (let tgl = 1; tgl <= jumlahHari; tgl++) sel.push(new Date(bulanTampil.getFullYear(), bulanTampil.getMonth(), tgl));
  while (sel.length % 7) sel.push(null);
  const baris = Array.from({ length: sel.length / 7 }, (_, i) => sel.slice(i * 7, i * 7 + 7));

  return (
    <div className="kalender-field" ref={pembungkus}>
      <div className="kalender-field__baris">
        <input
          {...props}
          ref={inputRef}
          className="input"
          type="text"
          inputMode="numeric"
          placeholder="DD-MM-YYYY"
          maxLength={10}
          value={value}
          onChange={(e) => onChange(e.target.value)}
        />
        {!props.disabled && (
          <button
            type="button"
            ref={tombolBuka}
            className="tombol tombol--kedua kalender__buka"
            aria-label={`Buka kalender ${label}`}
            aria-expanded={buka}
            onClick={() => (buka ? tutup(true) : bukaPopup())}
          >
            <IkonKalender />
          </button>
        )}
      </div>
      <div className="kalender" role="dialog" aria-label={`Pilih ${label}`} hidden={!buka} onKeyDown={tombolKeyboard}>
        <div className="kalender__kepala">
          <button type="button" ref={tombolSebelum} className="carousel__tombol" aria-label="Bulan sebelumnya" onClick={() => geserBulan(-1)}>
            <Icon name="kiri" size={18} />
          </button>
          <p className="kalender__judul" aria-live="polite">{`${BULAN[bulanTampil.getMonth()]} ${bulanTampil.getFullYear()}`}</p>
          <button
            type="button"
            className="carousel__tombol"
            aria-label="Bulan berikutnya"
            // Bulan yang seluruhnya di masa depan tidak bisa dibuka.
            disabled={new Date(bulanTampil.getFullYear(), bulanTampil.getMonth() + 1, 1) > sekarang}
            onClick={() => geserBulan(1)}
          >
            <Icon name="kanan" size={18} />
          </button>
        </div>
        <table className="kalender__tabel" aria-label="Tanggal">
          <thead>
            <tr>
              {HARI.map((h) => (
                <th key={h} scope="col" abbr={h} title={h}>{h.slice(0, 3)}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {baris.map((minggu, i) => (
              <tr key={i}>
                {minggu.map((d, j) => (
                  <td key={j}>
                    {d && (
                      <button
                        type="button"
                        className={['kalender__hari', sama(d, sekarang) && 'kalender__hari--ini', sama(terpilih, d) && 'kalender__hari--pilih'].filter(Boolean).join(' ')}
                        data-tanggal={keTeks(d)}
                        aria-label={`${HARI[indeksHari(d)]}, ${d.getDate()} ${BULAN[d.getMonth()]} ${d.getFullYear()}`}
                        aria-current={sama(d, sekarang) ? 'date' : undefined}
                        aria-pressed={sama(terpilih, d)}
                        tabIndex={sama(fokus, d) ? 0 : -1}
                        disabled={diMasaDepan(d)}
                        onClick={() => pilih(d)}
                      >
                        {d.getDate()}
                      </button>
                    )}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
