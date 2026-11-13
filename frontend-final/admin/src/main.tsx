import { createRoot } from 'react-dom/client';
import '../../src/styles/tokens.css';
import '../../src/styles/base.css';
import '../../src/styles/components.css';
import '../../src/styles/kalender.css';
import '../../src/styles/admin.css';
import '../../src/styles/simulation.css';
import '../../src/styles/tailwind.css';
import QueryProvider from '../../src/shared/QueryProvider';
import SimulationPanel from '../../src/simulation/Panel';
import App from './App';
import LoginSimulation from './LoginSimulation';

const akar = document.getElementById('root');
if (!akar) throw new Error('Elemen #root tidak ditemukan.');

// Halaman pemilihan akun menggantikan redirect Keycloak; sama seperti prototype, tanpa panel simulasi.
createRoot(akar).render(
  new URLSearchParams(location.search).get('simulation') === 'login' ? (
    <LoginSimulation />
  ) : (
    <QueryProvider>
      <App />
      <SimulationPanel admin />
    </QueryProvider>
  ),
);
