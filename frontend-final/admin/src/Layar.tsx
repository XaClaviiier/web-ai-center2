/* Layar tanpa kerangka admin: dipakai App (sesi berakhir, tanpa akses) dan main (galat login, gagal memuat sesi). */
import type { ReactNode } from 'react';
import { Button } from '../../src/components/ui/button';
import { image } from '../../src/shared/content';
import { URL_LOGIN, useTitle } from './context';

export function Layar({ judul, teks, aksi, galat = false }: { judul: string; teks: string; aksi: ReactNode; galat?: boolean }) {
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

/** Hasil gagal dari /api/auth/callback (`/admin/?galat=<kode>`). Tidak pernah mengalihkan otomatis agar tidak berputar di SSO. */
export function LayarGalatMasuk({ kode }: { kode: string }) {
  if (kode === 'akses_ditolak') {
    return (
      <Layar
        judul="Anda tidak punya akses"
        teks="Akun ini tidak punya akses ke CMS UBAYA AI Center. Masuk dengan akun lain atau hubungi pengelola akun Anda."
        galat
        aksi={<Button asChild variant="utama"><a href={`${URL_LOGIN}?pilih=1`}>Masuk dengan akun lain</a></Button>}
      />
    );
  }
  const kedaluwarsa = kode === 'login_kedaluwarsa';
  return (
    <Layar
      judul={kedaluwarsa ? 'Proses masuk kedaluwarsa' : 'Gagal masuk'}
      teks={kedaluwarsa ? 'Waktu proses masuk sudah habis atau tautannya sudah dipakai. Silakan masuk lagi.' : 'Terjadi kendala saat masuk. Silakan coba lagi beberapa saat lagi.'}
      galat
      aksi={<Button asChild variant="utama"><a href={URL_LOGIN}>{kedaluwarsa ? 'Masuk lagi' : 'Coba lagi'}</a></Button>}
    />
  );
}

/** Sesi tidak bisa dimuat karena galat selain 401 (jaringan, server): tampilkan pesan, bukan alihkan ke login. */
export function LayarGagalMuat({ pesan }: { pesan: string }) {
  return (
    <Layar
      judul="Admin tidak dapat dimuat"
      teks={pesan}
      galat
      aksi={<Button variant="utama" onClick={() => location.reload()}>Coba lagi</Button>}
    />
  );
}
