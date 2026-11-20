import { useRef, type ReactNode } from 'react';
import { Icon } from '../../src/shared/ui';
import { image } from '../../src/shared/content';
import { Button } from '../../src/components/ui/button';
import { ENTITAS } from './entities';
import { useAdmin } from './context';

/** Kerangka admin: header akun, menu samping, dan area isi. */
export default function Shell({ aktif, children }: { aktif: string; children: ReactNode }) {
  const { akun, keluar } = useAdmin();
  const isi = useRef<HTMLElement>(null);

  return (
    <>
      {/* Hash dipakai router, jadi skip link memindahkan fokus tanpa mengubah hash. */}
      <a
        className="lewati"
        href="#isi"
        onClick={(e) => {
          e.preventDefault();
          isi.current?.focus();
        }}
      >
        Lewati ke konten
      </a>
      <header className="atas">
        <a className="atas__logo" href="#/berita" aria-label="Admin UBAYA AI Center">
          <img src={image('AI-Center_Logo.png')} alt="" height={36} />
        </a>
        <div className="atas__akun">
          <span className="atas__nama">
            <strong>{akun?.nama}</strong>
            <span className="teks-sekunder">{akun?.role}</span>
          </span>
          <Button variant="kedua" size="kecil" onClick={() => void keluar()}>
            <Icon name="keluar" size={18} />
            Keluar
          </Button>
        </div>
      </header>
      <div className="tata">
        <nav className="samping" aria-label="Menu admin">
          {Object.entries(ENTITAS).map(([kunci, e]) => (
            <a key={kunci} href={`#/${kunci}`} aria-current={kunci === aktif ? 'page' : undefined}>
              <Icon name={e.ikon} size={20} />
              {e.label}
            </a>
          ))}
          <a href="#/log-aktivitas" aria-current={aktif === 'log-aktivitas' ? 'page' : undefined}>
            <Icon name="berita" size={20} />
            Log Aktivitas
          </a>
        </nav>
        <main className="isi" id="isi" tabIndex={-1} ref={isi}>
          {children}
        </main>
      </div>
    </>
  );
}
