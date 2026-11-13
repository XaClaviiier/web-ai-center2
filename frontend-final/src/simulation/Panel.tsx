import { useState } from 'react';
import { SIMULASI, useSimulationVersion } from './state';
import '../styles/simulation.css';
export default function SimulationPanel({ admin = false }: { admin?: boolean }) {
  const [open, setOpen] = useState(false);
  const [, refresh] = useState(0);
  useSimulationVersion();
  const change = () => document.dispatchEvent(new Event('simulasi:ubah'));
  const states = [['normal', 'Normal'], ['loading', 'Loading'], ['kosong', 'Kosong'], ['error', 'Error'], ['429', '429']] as const;
  return <aside className="sim-panel" aria-label="Panel simulasi"><div className="sim-isi" hidden={!open}>
    <strong>Simulasi, bukan bagian aplikasi</strong><p>Keadaan data</p><div className="sim-grup">{states.map(([value, label]) => <button type="button" key={value} aria-pressed={SIMULASI.keadaan === value} onClick={() => { SIMULASI.keadaan = value; change(); }}>{label}</button>)}</div>
    {admin && <><p>Admin</p><div className="sim-grup sim-grup--tegak">
      <button type="button" aria-pressed={SIMULASI.kunciLain} onClick={() => { SIMULASI.kunciLain = !SIMULASI.kunciLain; change(); }}>Item dikunci pengguna lain</button>
      <button type="button" onClick={() => { SIMULASI.paksa409 = true; document.dispatchEvent(new CustomEvent('simulasi:catat', { detail: '409 aktif untuk simpan berikutnya.' })); refresh(n => n + 1); }}>409 pada simpan berikutnya</button>
      <button type="button" onClick={() => { SIMULASI.heartbeatGagal = true; document.dispatchEvent(new Event('simulasi:heartbeat')); }}>Heartbeat gagal sekarang</button>
      <button type="button" onClick={() => { SIMULASI.paksa401 = true; change(); }}>Sesi berakhir (401)</button>
    </div></>}
  </div><button type="button" className="sim-buka" aria-expanded={open} onClick={() => setOpen(!open)}>Simulasi</button></aside>;
}
