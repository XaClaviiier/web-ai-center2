import { createRoot } from 'react-dom/client';
import '../../src/styles/tokens.css';
import '../../src/styles/base.css';
import '../../src/styles/components.css';
import '../../src/styles/kalender.css';
import '../../src/styles/admin.css';
import '../../src/styles/tailwind.css';
import QueryProvider from '../../src/shared/QueryProvider';
import { api } from '../../src/api/api';
import App from './App';

const akar = document.getElementById('root');
if (!akar) throw new Error('Elemen #root tidak ditemukan.');

// Sesi dimuat dari server dulu (mengisi token CSRF). Tanpa sesi: ke halaman masuk (sementara dev-login, nanti Keycloak).
void (async () => {
  const akun = await api.siapkan().catch(() => null);
  if (!akun) {
    location.replace('/admin/masuk');
    return;
  }
  createRoot(akar).render(
    <QueryProvider>
      <App />
    </QueryProvider>,
  );
})();
