import { createRoot } from 'react-dom/client';
import '../../src/styles/tokens.css';
import '../../src/styles/base.css';
import '../../src/styles/components.css';
import '../../src/styles/kalender.css';
import '../../src/styles/admin.css';
import '../../src/styles/tailwind.css';
import QueryProvider from '../../src/shared/QueryProvider';
import { errorMessage } from '../../src/shared/ui';
import { api } from '../../src/api/api';
import App from './App';
import { URL_LOGIN } from './context';
import { LayarGagalMuat, LayarGalatMasuk } from './Layar';

const akar = document.getElementById('root');
if (!akar) throw new Error('Elemen #root tidak ditemukan.');
const root = createRoot(akar);

/* Callback Keycloak yang gagal kembali ke /admin/?galat=<kode>. Parameter dibuang dari URL, lalu layar galat ditampilkan
   tanpa pengalihan otomatis (mencegah putaran login SSO). Tanpa galat: sesi dimuat dari server dulu (mengisi token CSRF);
   401 -> login Keycloak, galat lain -> pesan dengan tombol coba lagi. */
const kueri = new URLSearchParams(location.search);
const galat = kueri.get('galat');
if (galat !== null) {
  kueri.delete('galat');
  const sisa = kueri.toString();
  history.replaceState(null, '', `${location.pathname}${sisa ? `?${sisa}` : ''}${location.hash}`);
  root.render(<LayarGalatMasuk kode={galat} />);
} else {
  void (async () => {
    let akun;
    try {
      akun = await api.siapkan();
    } catch (g) {
      root.render(<LayarGagalMuat pesan={errorMessage(g)} />);
      return;
    }
    if (!akun) {
      location.replace(URL_LOGIN);
      return;
    }
    root.render(
      <QueryProvider>
        <App />
      </QueryProvider>,
    );
  })();
}
