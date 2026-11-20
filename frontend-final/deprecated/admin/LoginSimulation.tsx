import { useTitle } from './context';
import './simulasi-login.css';

const AKUN = [
  { id: 'direktur', label: 'Dian Direktur (Direktur)', varian: 'utama' },
  { id: 'pegawai', label: 'Pandu Pegawai (Pegawai)', varian: 'utama' },
  { id: 'tanpa-role', label: 'Tamu Tanpa Role (hasilnya 403)', varian: 'kedua' },
] as const;

/** Pengganti redirect Keycloak: pilih akun contoh, lalu kembali ke admin dengan `?akun=`. */
export default function LoginSimulation() {
  useTitle('Simulasi Keycloak');
  const admin = import.meta.env.BASE_URL;
  return (
    <div className="sim-login">
      <main className="sim-login__kartu">
        <p className="sim-login__label">Simulasi, bukan bagian aplikasi</p>
        <h1>Pilih akun untuk tinjauan</h1>
        <p className="teks-sekunder">Di produksi, langkah ini adalah halaman login Keycloak. Tidak ada password di simulasi ini.</p>
        <ul>
          {AKUN.map((a) => (
            <li key={a.id}>
              <a className={`tombol tombol--${a.varian}`} href={`${admin}?akun=${a.id}`}>{a.label}</a>
            </li>
          ))}
        </ul>
        <a href={admin}>Batal</a>
      </main>
    </div>
  );
}
