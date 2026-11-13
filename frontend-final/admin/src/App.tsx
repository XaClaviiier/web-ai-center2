/* Dashboard admin dengan rute hash. Halaman hanya memanggil api.*; login lewat redirect
   (Keycloak, di sini simulasi pemilihan akun), aplikasi tidak punya form password. */
import { useEffect, useMemo, useRef, useState, type ReactNode } from 'react';
import { ConfirmDialog } from '../../src/components/ui/alert-dialog';
import { Button } from '../../src/components/ui/button';
import { errorMessage, Icon, State } from '../../src/shared/ui';
import { image } from '../../src/shared/content';
import { api } from '../../src/simulation/api';
import { AdminContext, is401, PESAN_TANPA_PERUBAHAN, URL_LOGIN, useTitle, type Admin, type ConfirmOptions, type FormGuard } from './context';
import { ENTITAS, isEntity } from './entities';
import ListPage from './ListPage';
import FormPage from './FormPage';
import LogPage from './LogPage';
import Shell from './Shell';

interface Toast {
  id: number;
  pesan: string;
  galat: boolean;
}

/* === Layar tanpa kerangka === */
function Layar({ judul, teks, aksi, galat = false }: { judul: string; teks: string; aksi: ReactNode; galat?: boolean }) {
  useTitle(`${judul} | Admin UBAYA AI Center`);
  return (
    <main className="layar">
      <div className="layar__kartu">
        <img className="layar__logo" src={image('AI-Center_Logo.png')} alt="UBAYA AI Center" height={44} />
        <h1 className={galat ? 'layar__judul layar__judul--galat' : 'layar__judul'}>{judul}</h1>
        <p className="teks-sekunder">{teks}</p>
        {aksi}
      </div>
    </main>
  );
}

function KeBerita() {
  useEffect(() => {
    location.hash = '#/berita';
  }, []);
  return null;
}

function Rute({ hash }: { hash: string }) {
  const [sumber, a, b] = hash.replace(/^#\/?/, '').split('/');
  if (sumber === 'log-aktivitas' && !a) return <LogPage />;
  if (!isEntity(sumber)) return <KeBerita />;
  if (a === 'baru') return <FormPage sumber={sumber} id={null} />;
  if (a === 'trash' && ENTITAS[sumber].status) return <ListPage sumber={sumber} trash />;
  if (/^\d+$/.test(a ?? '') && (b === 'edit' || b === undefined)) return <FormPage sumber={sumber} id={Number(a)} />;
  if (a) {
    return (
      <Shell aktif={sumber}>
        <State
          title="Halaman tidak ditemukan"
          text="Tautan admin ini tidak dikenal."
          action={<Button asChild variant="kedua"><a href={`#/${sumber}`}>Kembali ke daftar</a></Button>}
        />
      </Shell>
    );
  }
  return <ListPage sumber={sumber} trash={false} />;
}

export default function App() {
  const [akun] = useState(() => api.masuk(new URLSearchParams(location.search).get('akun') ?? ''));
  const [rute, setRute] = useState({ hash: location.hash, render: 0 });
  const [sesiHabis, setSesiHabis] = useState(false);
  const [toasts, setToasts] = useState<Toast[]>([]);
  const [dialog, setDialog] = useState<{ opsi: ConfirmOptions; jawab: (ya: boolean) => void } | null>(null);
  const formAktif = useRef<FormGuard | null>(null);
  const nomorToast = useRef(0);

  const admin = useMemo<Admin>(() => {
    const toast = (pesan: string, galat = false) => {
      const id = ++nomorToast.current;
      setToasts((t) => [...t, { id, pesan, galat }]);
      setTimeout(() => setToasts((t) => t.filter((x) => x.id !== id)), 5000);
    };
    const konfirmasi = (opsi: ConfirmOptions) => {
      const { promise, resolve } = Promise.withResolvers<boolean>();
      setDialog({
        opsi,
        jawab: (ya) => {
          setDialog(null);
          resolve(ya);
        },
      });
      return promise;
    };
    const sesiBerakhir = () => {
      formAktif.current = null;
      setSesiHabis(true);
    };
    return {
      akun,
      toast,
      konfirmasi,
      sesiBerakhir,
      tanganiGalat: (galat) => (is401(galat) ? sesiBerakhir() : toast(errorMessage(galat), true)),
      render: () => {
        setSesiHabis(false);
        setRute((r) => ({ hash: location.hash, render: r.render + 1 }));
      },
      keluar: async () => {
        const teks = `Anda perlu masuk kembali untuk mengelola konten.${formAktif.current?.dirty() ? ' Perubahan yang belum disimpan akan hilang.' : ''}`;
        if (!(await konfirmasi({ judul: 'Keluar dari admin?', teks, labelYa: 'Keluar' }))) return;
        formAktif.current = null;
        location.href = import.meta.env.BASE_URL;
      },
      formAktif,
    };
  }, [akun]);

  /* Hash sudah berubah saat hashchange terpanggil, jadi membatalkan pindah rute berarti mengembalikan hash.
     Form tidak dirender ulang, sehingga isian dan kunci tetap utuh.
     shortcut: pengembalian hash menambah satu entri riwayat browser, cukup untuk simulasi; di produksi pakai blocker router. */
  useEffect(() => {
    // Hash yang terakhir dirender, bukan location.hash: efek anak (KeBerita) bisa sudah mengubah hash sebelum efek ini jalan.
    let hashAktif = rute.hash;
    let menungguKonfirmasi = false;
    async function pindahRute() {
      if (location.hash === hashAktif) return;
      if (menungguKonfirmasi) {
        location.hash = hashAktif;
        return;
      }
      if (formAktif.current?.dirty()) {
        const tujuan = location.hash;
        location.hash = hashAktif;
        menungguKonfirmasi = true;
        const ya = await admin.konfirmasi({
          judul: 'Tinggalkan halaman ini?',
          teks: 'Perubahan yang belum disimpan akan hilang.',
          labelYa: 'Tinggalkan',
          labelTidak: 'Tetap di sini',
          varianYa: 'bahaya',
        });
        menungguKonfirmasi = false;
        if (!ya) return;
        formAktif.current = null;
        location.hash = tujuan;
        return;
      }
      // Form terbuka tanpa perubahan; setelah simpan formAktif sudah null.
      if (formAktif.current) admin.toast(PESAN_TANPA_PERUBAHAN);
      hashAktif = location.hash;
      admin.render();
    }
    const sebelumTutup = (e: BeforeUnloadEvent) => {
      if (formAktif.current?.dirty()) e.preventDefault();
    };
    const catat = (e: Event) => {
      if (e instanceof CustomEvent && typeof e.detail === 'string') admin.toast(e.detail);
    };
    const onHash = () => void pindahRute();
    window.addEventListener('hashchange', onHash);
    window.addEventListener('beforeunload', sebelumTutup);
    document.addEventListener('simulasi:ubah', admin.render);
    document.addEventListener('simulasi:catat', catat);
    return () => {
      window.removeEventListener('hashchange', onHash);
      window.removeEventListener('beforeunload', sebelumTutup);
      document.removeEventListener('simulasi:ubah', admin.render);
      document.removeEventListener('simulasi:catat', catat);
    };
  }, [admin]);

  let isi: ReactNode;
  if (!akun) {
    isi = (
      <Layar
        judul="Masuk ke admin"
        teks="Masuk memakai akun UBAYA AI Center Anda untuk mengelola konten website."
        aksi={<Button asChild variant="utama"><a href={URL_LOGIN}>Masuk</a></Button>}
      />
    );
  } else if (!akun.role) {
    isi = (
      <Layar
        judul="Anda tidak punya akses"
        teks="Akun ini belum memiliki role untuk mengelola konten. Hubungi pengelola akun Anda."
        galat
        aksi={
          <Button variant="kedua" onClick={() => void admin.keluar()}>
            <Icon name="keluar" size={18} />
            Keluar
          </Button>
        }
      />
    );
  } else if (sesiHabis) {
    isi = (
      <Layar
        judul="Sesi Anda sudah berakhir"
        teks="Silakan masuk kembali untuk melanjutkan."
        aksi={<Button asChild variant="utama"><a href={URL_LOGIN}>Masuk kembali</a></Button>}
      />
    );
  } else {
    isi = <Rute key={`${rute.hash}|${rute.render}`} hash={rute.hash} />;
  }

  return (
    <AdminContext value={admin}>
      {isi}
      <div className="toast-wadah" role="status" aria-live="polite">
        {toasts.map((t) => (
          <div key={t.id} className={t.galat ? 'toast toast--galat' : 'toast'}>
            <Icon name={t.galat ? 'peringatan' : 'cek'} size={18} />
            <span>{t.pesan}</span>
          </div>
        ))}
      </div>
      {dialog && (
        <ConfirmDialog
          open
          title={dialog.opsi.judul}
          text={dialog.opsi.teks}
          confirmLabel={dialog.opsi.labelYa}
          variant={dialog.opsi.varianYa}
          cancelLabel={dialog.opsi.labelTidak}
          onAnswer={dialog.jawab}
        />
      )}
    </AdminContext>
  );
}
