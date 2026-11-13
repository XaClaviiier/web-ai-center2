import { useEffect, useRef, useState, type FormEvent } from 'react';
import { useQuery } from '@tanstack/react-query';
import { ApiError, type Entity, type Item, type Payload } from '../../src/shared/types';
import { Icon, Loading, State } from '../../src/shared/ui';
import { Button } from '../../src/components/ui/button';
import { api } from '../../src/simulation/api';
import { useSimulationVersion } from '../../src/simulation/state';
import { daftarGambar, ENTITAS, type FieldDef } from './entities';
import { is401, PESAN_TANPA_PERUBAHAN, useAdmin, useTitle } from './context';
import { dariTeks, diMasaDepan, hariIniTeks } from './Calendar';
import { GalleryField, ImageField, TextField, type FieldValue } from './fields';
import Shell from './Shell';

type Values = Record<string, FieldValue>;
type Kunci = { milik: 'saya' | 'lain'; nama?: string };

const teks = (v: FieldValue | undefined): string => (typeof v === 'string' ? v : '');
const daftar = (v: FieldValue | undefined): string[] => (Array.isArray(v) ? v : []);

/** Form tambah: tanggal otomatis hari ini. Form edit: nilai tersimpan. */
function nilaiAwal(f: FieldDef, data: Item | null): FieldValue {
  switch (f.tipe) {
    case 'galeri':
      return daftarGambar(data).map((g) => g.key);
    case 'gambar': {
      const v = data?.[f.nama];
      return typeof v === 'string' ? v : null;
    }
    case 'angka':
      return data?.urutan === undefined ? '' : String(data.urutan);
    case 'tanggal':
      return data?.[f.nama] ?? hariIniTeks();
    default:
      return data?.[f.nama] ?? '';
  }
}

/** Nilai form sebagai payload API; juga dipakai untuk mendeteksi perubahan. */
function bacaPayload(field: FieldDef[], nilai: Values): Payload {
  const payload: Payload = {};
  for (const f of field) {
    const v = nilai[f.nama];
    switch (f.tipe) {
      case 'galeri':
        payload.gambarDaftar = daftar(v).map((key, i) => ({ key, urutan: i + 1 }));
        break;
      case 'gambar':
        if (f.nama === 'foto') payload.foto = typeof v === 'string' ? v : null;
        else payload.gambar = teks(v);
        break;
      case 'angka':
        payload.urutan = Number(v);
        break;
      default:
        payload[f.nama] = teks(v).trim();
    }
  }
  return payload;
}

function validasi(f: FieldDef, v: FieldValue | undefined, awal: FieldValue | undefined): string | null {
  if (f.tipe === 'gambar') return f.wajib && !v ? `${f.label} wajib diisi.` : null;
  if (f.tipe === 'galeri') return daftar(v).length < 1 ? 'Minimal 1 gambar.' : null;
  const s = teks(v).trim();
  if (f.wajib && !s) return `${f.label} wajib diisi.`;
  // Judul lama yang melebihi batas tetap bisa disimpan selama tidak diubah.
  if (f.tipe === 'teks' && f.maks !== undefined && s.length > f.maks && s !== teks(awal).trim()) {
    return `${f.label} maksimal ${f.maks} karakter (sekarang ${s.length}).`;
  }
  if (f.tipe === 'tanggal') {
    const d = dariTeks(s);
    if (!d) return 'Tanggal harus berformat DD-MM-YYYY dan benar-benar ada.';
    if (diMasaDepan(d)) return `${f.label} tidak boleh setelah hari ini.`;
  }
  if (f.tipe === 'angka' && !(Number.isInteger(Number(s)) && Number(s) >= 1)) return 'Urutan harus berupa angka bulat mulai dari 1.';
  return null;
}

export default function FormPage({ sumber, id }: { sumber: Entity; id: number | null }) {
  const e = ENTITAS[sumber];
  useTitle(`${id === null ? 'Tambah' : 'Edit'} ${e.nama} | Admin UBAYA AI Center`);
  return (
    <Shell aktif={sumber}>
      {id === null ? <Form sumber={sumber} data={null} kunci={{ milik: 'saya' }} /> : <MuatForm sumber={sumber} id={id} />}
    </Shell>
  );
}

function MuatForm({ sumber, id }: { sumber: Entity; id: number }) {
  const e = ENTITAS[sumber];
  const admin = useAdmin();
  const versi = useSimulationVersion();
  // Mengambil kunci adalah efek samping: tidak di-cache dan tidak diulang otomatis.
  const muat = useQuery({
    queryKey: ['admin-form', sumber, id, versi],
    gcTime: 0,
    staleTime: Infinity,
    queryFn: async (): Promise<{ data: Item; kunci: Kunci | null }> => {
      const data = await api.adminDetail(sumber, id);
      if (e.status && data.deletedAt) throw new ApiError(404, 'not_found', 'Data ada di trash. Pulihkan dulu sebelum mengedit.');
      if (e.status && data.status !== 'draft') return { data, kunci: null };
      return { data, kunci: await api.kunciAmbil(sumber, id) };
    },
  });

  useEffect(() => {
    if (is401(muat.error)) admin.sesiBerakhir();
  }, [muat.error, admin]);

  if (muat.isFetching) return <Loading rows count={6} />;
  if (muat.error || !muat.data) {
    return (
      <State
        error={muat.error}
        action={
          <Button variant="kedua" onClick={() => void muat.refetch()}>
            <Icon name="ulang" size={18} />
            Coba lagi
          </Button>
        }
      />
    );
  }
  if (!muat.data.kunci) {
    return (
      <State
        title="Konten ini sudah tayang"
        text="Retract dulu untuk mengembalikannya ke draft, lalu edit."
        action={<Button asChild variant="kedua"><a href={`#/${sumber}`}>Kembali ke daftar</a></Button>}
      />
    );
  }
  return <Form sumber={sumber} data={muat.data.data} kunci={muat.data.kunci} />;
}

function Form({ sumber, data, kunci }: { sumber: Entity; data: Item | null; kunci: Kunci }) {
  const e = ENTITAS[sumber];
  const admin = useAdmin();
  const baru = data === null;
  const bacaSaja = kunci.milik === 'lain';
  const kembali = `#/${sumber}`;
  const judulHalaman = `${baru ? 'Tambah' : 'Edit'} ${e.nama}`;

  const [awal] = useState<Values>(() => Object.fromEntries(e.field.map((f) => [f.nama, nilaiAwal(f, data)] as const)));
  const [nilai, setNilai] = useState(awal);
  const nilaiTerbaru = useRef(awal);
  const [galat, setGalat] = useState<Record<string, string | null>>({});
  const [menyimpan, setMenyimpan] = useState(false);
  const [kunciHilang, setKunciHilang] = useState(false);
  const [detakTerakhir, setDetakTerakhir] = useState(() => new Date());
  const kontrol = useRef<Record<string, HTMLElement | null>>({});
  const judul = useRef<HTMLHeadingElement>(null);

  // Dibandingkan dengan nilai awal, sehingga isian yang dikembalikan seperti semula tidak dianggap berubah.
  const [nilaiAwalForm] = useState(() => JSON.stringify(bacaPayload(e.field, awal)));
  const formKotor = () => JSON.stringify(bacaPayload(e.field, nilaiTerbaru.current)) !== nilaiAwalForm;

  function ubah(nama: string, v: FieldValue) {
    nilaiTerbaru.current = { ...nilaiTerbaru.current, [nama]: v };
    setNilai(nilaiTerbaru.current);
  }
  const tampilGalat = (nama: string, pesan: string | null) => setGalat((g) => ({ ...g, [nama]: pesan }));

  useEffect(() => {
    judul.current?.focus();
    if (bacaSaja) return;
    const guard = { dirty: formKotor };
    admin.formAktif.current = guard;
    return () => {
      if (admin.formAktif.current === guard) admin.formAktif.current = null;
    };
    // Dipasang sekali per form; formKotor membaca nilai terbaru lewat ref.
  }, []);

  /* Heartbeat: kunci diperpanjang tiap 1 menit selama form terbuka. */
  useEffect(() => {
    if (baru || bacaSaja) return;
    let aktif = true;
    const timer = setInterval(detak, 60000);
    async function detak() {
      try {
        await api.kunciHeartbeat();
        if (aktif) setDetakTerakhir(new Date());
      } catch (g) {
        if (!aktif) return;
        if (is401(g)) {
          admin.sesiBerakhir();
          return;
        }
        if (!(g instanceof ApiError && g.status === 410)) return; // galat sementara: coba lagi pada detak berikutnya
        clearInterval(timer);
        setKunciHilang(true);
      }
    }
    document.addEventListener('simulasi:heartbeat', detak);
    return () => {
      aktif = false;
      clearInterval(timer);
      document.removeEventListener('simulasi:heartbeat', detak);
      api.kunciLepas();
    };
  }, [baru, bacaSaja, admin]);

  async function kirim(ev: FormEvent<HTMLFormElement>) {
    ev.preventDefault();
    if (menyimpan || kunciHilang) return;
    const hasil = e.field.map((f) => ({ f, pesan: validasi(f, nilaiTerbaru.current[f.nama], awal[f.nama]) }));
    setGalat(Object.fromEntries(hasil.map(({ f, pesan }) => [f.nama, pesan] as const)));
    const salah = hasil.find((h) => h.pesan);
    if (salah) {
      kontrol.current[salah.f.nama]?.focus();
      return;
    }
    if (!baru && !formKotor()) {
      admin.formAktif.current = null;
      admin.toast(PESAN_TANPA_PERUBAHAN);
      location.hash = kembali;
      return;
    }
    // Nonaktif setelah klik pertama, mencegah klik ganda.
    setMenyimpan(true);
    const payload = bacaPayload(e.field, nilaiTerbaru.current);
    try {
      if (data === null) await api.adminTambah(sumber, payload);
      else await api.adminUbah(sumber, data.id, payload, data.version);
      admin.formAktif.current = null;
      admin.toast(!baru ? 'Perubahan tersimpan.' : e.status ? 'Tersimpan sebagai draft.' : 'Anggota tim ditambahkan.');
      location.hash = kembali;
    } catch (g) {
      setMenyimpan(false);
      if (g instanceof ApiError && g.status === 409) {
        const muatUlang = await admin.konfirmasi({
          judul: 'Data sudah berubah',
          teks: 'Pengguna lain menyimpan perubahan lebih dulu. Muat data terbaru untuk melanjutkan; isian Anda yang belum tersimpan akan hilang.',
          labelYa: 'Muat ulang data terbaru',
          labelTidak: 'Tutup',
        });
        if (muatUlang) admin.render();
      } else {
        admin.tanganiGalat(g);
      }
    }
  }

  return (
    <>
      <div className="kepala">
        <h1 tabIndex={-1} ref={judul}>{bacaSaja ? `Lihat ${e.nama}` : judulHalaman}</h1>
      </div>
      {!baru && !bacaSaja && (
        <p className="field__bantuan" role="status">
          {kunciHilang ? '' : `Data dikunci untuk Anda. Diperpanjang otomatis, terakhir ${detakTerakhir.toLocaleTimeString('id-ID')}.`}
        </p>
      )}
      <form className="form" noValidate onSubmit={(ev) => void kirim(ev)}>
        <div>
          {bacaSaja && (
            <div className="banner banner--kunci" role="status">
              <p className="banner__teks">
                <Icon name="kunci" size={20} />
                {`Sedang diedit oleh ${kunci.nama}. Anda hanya bisa membaca sampai kunci dilepas.`}
              </p>
            </div>
          )}
          {kunciHilang && (
            <div className="banner banner--galat" role="alert">
              <p className="banner__teks">
                <Icon name="peringatan" size={20} />
                Sesi edit berakhir. Isian Anda masih ada di form, tetapi tidak bisa disimpan. Salin yang perlu, lalu muat ulang.
              </p>
              <Button variant="kedua" size="kecil" onClick={admin.render}>
                <Icon name="ulang" size={18} />
                Muat ulang
              </Button>
            </div>
          )}
        </div>
        {e.field.map((f) => {
          const v = nilai[f.nama];
          const umum = {
            galat: galat[f.nama] ?? null,
            bacaSaja,
            daftarkan: (el: HTMLElement | null) => {
              kontrol.current[f.nama] = el;
            },
            tampilGalat: (pesan: string | null) => tampilGalat(f.nama, pesan),
          };
          if (f.tipe === 'gambar') {
            return <ImageField key={f.nama} f={f} {...umum} value={typeof v === 'string' ? v : null} onValue={(baru) => ubah(f.nama, baru)} />;
          }
          if (f.tipe === 'galeri') {
            return <GalleryField key={f.nama} f={f} {...umum} value={daftar(v)} onValue={(baru) => ubah(f.nama, baru)} />;
          }
          return (
            <TextField
              key={f.nama}
              f={f}
              {...umum}
              value={teks(v)}
              awal={teks(awal[f.nama])}
              onValue={(baru) => ubah(f.nama, baru)}
              validasi={(baru) => tampilGalat(f.nama, validasi(f, baru, awal[f.nama]))}
            />
          );
        })}
        <div className="form__aksi">
          {bacaSaja ? (
            <Button asChild variant="kedua"><a href={kembali}>Kembali ke daftar</a></Button>
          ) : (
            <>
              <Button type="submit" variant="utama" disabled={menyimpan || kunciHilang}>{menyimpan ? 'Menyimpan…' : 'Simpan'}</Button>
              <Button asChild variant="kedua"><a href={kembali}>Batal</a></Button>
            </>
          )}
        </div>
      </form>
    </>
  );
}
