import { useEffect, useState, type ReactNode } from 'react';
import { useQuery } from '@tanstack/react-query';
import type { Entity, Item, Status } from '../../src/shared/types';
import { Icon, Loading, State } from '../../src/shared/ui';
import { Button } from '../../src/components/ui/button';
import { api } from '../../src/api/api';
import { ENTITAS } from './entities';
import { is401, useAdmin, useTitle, type ConfirmOptions } from './context';
import Shell from './Shell';

const LABEL_STATUS: Record<Status, string> = { draft: 'Draft', posted: 'Posted' };

export default function ListPage({ sumber, trash }: { sumber: Entity; trash: boolean }) {
  const e = ENTITAS[sumber];
  const admin = useAdmin();
  useTitle(`${e.label}${trash ? ' (Trash)' : ''} | Admin UBAYA AI Center`);
  const daftar = useQuery({
    queryKey: ['admin-daftar', sumber, trash],
    queryFn: () => api.adminDaftar(sumber, { trash }),
    gcTime: 0,
  });

  useEffect(() => {
    if (is401(daftar.error)) admin.sesiBerakhir();
  }, [daftar.error, admin]);

  const tambah = (
    <Button asChild variant="utama">
      <a href={`#/${sumber}/baru`}>
        <Icon name="tambah" size={18} />
        {`Tambah ${e.nama}`}
      </a>
    </Button>
  );
  const muat = () => void daftar.refetch();

  let isi: ReactNode;
  if (daftar.isFetching) isi = <Loading rows />;
  else if (daftar.error) {
    isi = (
      <State
        error={daftar.error}
        action={
          <Button variant="kedua" onClick={muat}>
            <Icon name="ulang" size={18} />
            Coba lagi
          </Button>
        }
      />
    );
  } else if (!daftar.data?.items.length) {
    isi = trash
      ? <State title="Trash kosong" text="Tidak ada item di trash." />
      : <State title={`Belum ada ${e.nama}`} text="Data yang ditambahkan akan tampil di sini." action={tambah} />;
  } else {
    isi = <Tabel sumber={sumber} items={daftar.data.items} trash={trash} muat={muat} />;
  }

  return (
    <Shell aktif={sumber}>
      <div className="kepala">
        <h1 tabIndex={-1}>{trash ? `Trash ${e.label}` : e.label}</h1>
        {!trash && tambah}
      </div>
      {e.status && (
        <nav className="tab" aria-label="Tampilan daftar">
          <a href={`#/${sumber}`} aria-current={trash ? undefined : 'page'}>Semua</a>
          <a href={`#/${sumber}/trash`} aria-current={trash ? 'page' : undefined}>Trash</a>
        </nav>
      )}
      {trash && <p className="teks-sekunder">Item di sini terhapus permanen setelah 30 hari sejak dipindahkan ke trash.</p>}
      <div id="daftar">{isi}</div>
    </Shell>
  );
}

function Tabel({ sumber, items, trash, muat }: { sumber: Entity; items: Item[]; trash: boolean; muat: () => void }) {
  const e = ENTITAS[sumber];
  const kolomStatus = e.status && !trash;
  const judulTabel = `${trash ? 'Trash' : 'Daftar'} ${e.nama}`;
  // Wadah bisa difokus agar tabel lebar bisa digulir dengan keyboard.
  return (
    <div className="tabel-wadah" role="region" tabIndex={0} aria-label={judulTabel}>
      <p className="tabel-petunjuk">Geser tabel ke samping untuk melihat seluruh kolom dan tindakan.</p>
      <table className="tabel">
        <caption className="sr-only">{judulTabel}</caption>
        <thead>
          <tr>
            {e.kolom.map((k) => <th key={k.label} scope="col">{k.label}</th>)}
            {kolomStatus && <th scope="col">Status</th>}
            <th className="tabel__aksi" scope="col">Aksi</th>
          </tr>
        </thead>
        <tbody>
          {items.map((b) => (
            <tr key={b.id}>
              {e.kolom.map((k) => k.utama ? (
                <td key={k.label} className="tabel__utama">
                  <span className="tabel__judul">{k.isi(b)}</span>
                  {b.dikunciOleh && (
                    <span className="tabel__kunci">
                      <Icon name="kunci" size={16} />
                      {`Sedang diedit oleh ${b.dikunciOleh}`}
                    </span>
                  )}
                </td>
              ) : (
                <td key={k.label}>{k.isi(b)}</td>
              ))}
              {kolomStatus && (
                <td>{b.status && <span className={`badge badge--${b.status}`}>{LABEL_STATUS[b.status]}</span>}</td>
              )}
              <td className="tabel__aksi">
                <div className="tabel__tombol">
                  <AksiBaris sumber={sumber} b={b} trash={trash} muat={muat} />
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

/** Tombol sesuai status. Item dikunci orang lain: hanya bisa dibaca. */
function AksiBaris({ sumber, b, trash, muat }: { sumber: Entity; b: Item; trash: boolean; muat: () => void }) {
  const e = ENTITAS[sumber];
  const admin = useAdmin();
  const [sibuk, setSibuk] = useState<string | null>(null);
  const namaBaris = b.judul ?? b.nama ?? '';
  const terkunci = Boolean(b.dikunciOleh);

  async function jalankan(label: string, aksi: () => Promise<unknown>, pesan: string, konfirmasi?: ConfirmOptions) {
    if (konfirmasi && !(await admin.konfirmasi(konfirmasi))) return;
    setSibuk(label);
    try {
      await aksi();
      admin.toast(pesan);
      muat();
    } catch (galat) {
      setSibuk(null);
      admin.tanganiGalat(galat);
    }
  }

  const kecil = (label: string, varian: 'utama' | 'kedua' | 'bahaya', onClick: () => void, nonaktif = false) => (
    <Button key={label} variant={varian} size="kecil" aria-label={`${label}: ${namaBaris}`} disabled={nonaktif || sibuk === label} onClick={onClick}>
      {label}
    </Button>
  );

  if (trash) {
    return (
      <>
        {kecil('Restore', 'kedua', () => void jalankan('Restore', () => api.adminPulihkan(sumber, b.id), 'Dipulihkan sebagai draft.'), terkunci)}
        {kecil('Hapus Permanen', 'bahaya', () => void jalankan('Hapus Permanen', () => api.adminHapusPermanen(sumber, b.id), 'Dihapus permanen.', {
          judul: 'Hapus permanen?',
          teks: 'Data dan semua gambarnya hilang dan tidak bisa dikembalikan.',
          labelYa: 'Hapus Permanen',
          varianYa: 'bahaya',
        }), terkunci)}
      </>
    );
  }
  if (e.status && b.status === 'posted') {
    return kecil('Retract', 'kedua', () => void jalankan('Retract', () => api.adminRetract(sumber, b.id), 'Dikembalikan ke draft.', {
      judul: 'Tarik konten dari publik?',
      teks: `“${namaBaris}” tidak lagi terlihat oleh pengunjung dan kembali menjadi draft.`,
      labelYa: 'Retract',
      labelTidak: 'Batal',
      varianYa: 'bahaya',
    }));
  }
  const labelEdit = terkunci ? 'Lihat' : 'Edit';
  const hapus = sumber === 'tim'
    ? { judul: 'Hapus anggota tim?', teks: `${b.nama} akan dihapus permanen beserta fotonya.`, labelYa: 'Hapus Permanen', varianYa: 'bahaya' as const }
    : { judul: `Hapus ${e.nama} ini?`, teks: 'Item dipindahkan ke trash dan terhapus permanen setelah 30 hari.', labelYa: 'Pindahkan ke trash', varianYa: 'bahaya' as const };
  return (
    <>
      <Button asChild variant="kedua" size="kecil">
        <a href={`#/${sumber}/${b.id}/edit`} aria-label={`${labelEdit}: ${namaBaris}`}>{labelEdit}</a>
      </Button>
      {kecil('Hapus', 'bahaya', () => void jalankan('Hapus', () => api.adminHapus(sumber, b.id), sumber === 'tim' ? 'Anggota tim dihapus.' : 'Dipindahkan ke trash.', hapus), terkunci)}
      {e.status && kecil('Post', 'utama', () => void jalankan('Post', () => api.adminPost(sumber, b.id), 'Konten tampil di publik.'), terkunci)}
    </>
  );
}
