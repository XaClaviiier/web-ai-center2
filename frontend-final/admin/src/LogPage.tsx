import { useEffect, useRef, useState, type ReactNode } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Format } from '../../src/shared/format';
import { Icon, Loading, Pagination, State } from '../../src/shared/ui';
import { Button } from '../../src/components/ui/button';
import { api } from '../../src/simulation/api';
import { useSimulationVersion } from '../../src/simulation/state';
import { is401, useAdmin, useTitle } from './context';
import Shell from './Shell';

type Filter = { username: string; jenis: string; dari: string; sampai: string };
const KOSONG: Filter = { username: '', jenis: '', dari: '', sampai: '' };
const JENIS = [['', 'Semua aktivitas'], ['tambah', 'Tambah Data'], ['ubah', 'Ubah Data'], ['hapus', 'Hapus Data']] as const;

function Field({ label, id, children }: { label: string; id: string; children: ReactNode }) {
  return (
    <div className="field">
      <label htmlFor={id}>{label}</label>
      {children}
    </div>
  );
}

/* === Log aktivitas simulasi, baca saja === */
export default function LogPage() {
  useTitle('Log Aktivitas | Admin UBAYA AI Center');
  const admin = useAdmin();
  const versi = useSimulationVersion();
  const [isian, setIsian] = useState<Filter>(KOSONG);
  // `muat` selalu membuat permintaan baru, walau filter sama (nomor permintaan bagian dari key).
  const [opsi, setOpsi] = useState({ ...KOSONG, page: 1, permintaan: 0 });
  const [usernames, setUsernames] = useState<string[]>([]);
  const fokusHasil = useRef(false);
  const hasil = useRef<HTMLDivElement>(null);
  const sampai = useRef<HTMLInputElement>(null);

  const log = useQuery({
    queryKey: ['admin-log', opsi, versi],
    queryFn: () => api.adminLog({ page: opsi.page, username: opsi.username, jenis: opsi.jenis, dari: opsi.dari, sampai: opsi.sampai }),
    gcTime: 0,
  });

  useEffect(() => {
    if (log.data) setUsernames(log.data.usernames);
  }, [log.data]);

  useEffect(() => {
    if (is401(log.error)) admin.sesiBerakhir();
  }, [log.error, admin]);

  useEffect(() => {
    if (log.isFetching || !fokusHasil.current) return;
    fokusHasil.current = false;
    hasil.current?.focus();
  }, [log.isFetching]);

  function muat(filter: Filter, page: number) {
    fokusHasil.current = true;
    setOpsi((o) => ({ ...filter, page, permintaan: o.permintaan + 1 }));
  }

  const ubah = (kunci: keyof Filter, nilai: string) => setIsian((f) => ({ ...f, [kunci]: nilai }));

  let isi: ReactNode;
  let status: string;
  if (log.isFetching) {
    isi = <Loading rows />;
    status = 'Memuat log aktivitas…';
  } else if (log.error || !log.data) {
    isi = (
      <State
        error={log.error}
        action={
          <Button variant="kedua" onClick={() => muat(isian, opsi.page)}>
            <Icon name="ulang" size={18} />
            Coba lagi
          </Button>
        }
      />
    );
    status = 'Log aktivitas belum dapat dimuat.';
  } else {
    const data = log.data;
    const terfilter = opsi.username || opsi.jenis || opsi.dari || opsi.sampai;
    isi = (
      <>
        {data.items.length ? (
          <div className="tabel-wadah log-tabel-wadah" role="region" tabIndex={0} aria-label="Log aktivitas">
            <p className="tabel-petunjuk">Geser tabel ke samping untuk melihat seluruh kolom.</p>
            <table className="tabel log-tabel">
              <caption className="sr-only">Riwayat aktivitas simulasi, terbaru dahulu</caption>
              <thead>
                <tr>
                  {['Tanggal & Waktu', 'Aktivitas', 'Username'].map((teks) => <th key={teks} scope="col">{teks}</th>)}
                </tr>
              </thead>
              <tbody>
                {data.items.map((b) => (
                  <tr key={b.id}>
                    <td><time dateTime={b.terjadiPada}>{Format.waktuJakarta(b.terjadiPada)}</time></td>
                    <td className="log-tabel__aktivitas">{Format.aktivitas(b)}</td>
                    <td className="log-tabel__username">{b.username}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <State
            title={terfilter ? 'Tidak ada aktivitas yang cocok' : 'Belum ada aktivitas'}
            text={terfilter ? 'Coba ubah atau reset filter.' : 'Perubahan data yang berhasil pada simulasi ini akan tercatat di sini.'}
          />
        )}
        <Pagination page={data.page} limit={data.limit} total={data.total} onPage={(n) => muat(isian, n)} />
      </>
    );
    status = data.total
      ? `${data.total} aktivitas. Halaman ${data.page} dari ${Math.ceil(data.total / data.limit)}. Waktu Asia/Jakarta (WIB).`
      : 'Tidak ada aktivitas untuk ditampilkan.';
  }

  return (
    <Shell aktif="log-aktivitas">
      <div className="kepala">
        <h1 tabIndex={-1}>Log Aktivitas</h1>
      </div>
      <p className="log-simulasi">
        <strong>Simulasi. </strong>
        Aktivitas simulasi hanya tersimpan selama halaman ini terbuka dan kembali kosong setelah refresh. Identitas akun dan waktu perangkat belum diverifikasi server.
      </p>
      <form
        className="log-filter"
        aria-label="Filter log aktivitas"
        onSubmit={(ev) => {
          ev.preventDefault();
          sampai.current?.setCustomValidity(isian.dari && isian.sampai && isian.dari > isian.sampai ? 'Tanggal akhir harus sama dengan atau setelah tanggal awal.' : '');
          if (!ev.currentTarget.reportValidity()) return;
          muat(isian, 1);
        }}
      >
        <Field label="Username" id="log-username">
          <select className="input" id="log-username" value={isian.username} onChange={(e) => ubah('username', e.target.value)}>
            <option value="">Semua username</option>
            {usernames.map((u) => <option key={u} value={u}>{u}</option>)}
          </select>
        </Field>
        <Field label="Jenis aktivitas" id="log-jenis">
          <select className="input" id="log-jenis" value={isian.jenis} onChange={(e) => ubah('jenis', e.target.value)}>
            {JENIS.map(([value, label]) => <option key={value} value={value}>{label}</option>)}
          </select>
        </Field>
        <Field label="Dari tanggal (WIB)" id="log-dari">
          <input
            className="input"
            id="log-dari"
            type="date"
            value={isian.dari}
            onChange={(e) => {
              sampai.current?.setCustomValidity('');
              ubah('dari', e.target.value);
            }}
          />
        </Field>
        <Field label="Sampai tanggal (WIB)" id="log-sampai">
          <input
            className="input"
            id="log-sampai"
            type="date"
            ref={sampai}
            value={isian.sampai}
            onChange={(e) => {
              e.target.setCustomValidity('');
              ubah('sampai', e.target.value);
            }}
          />
        </Field>
        <div className="log-filter__aksi">
          <Button type="submit" variant="utama">Terapkan</Button>
          <Button
            variant="kedua"
            onClick={() => {
              setIsian(KOSONG);
              sampai.current?.setCustomValidity('');
              muat(KOSONG, 1);
            }}
          >
            Reset filter
          </Button>
        </div>
      </form>
      <p className="teks-sekunder" role="status" aria-live="polite">{status}</p>
      <div className="log-hasil" id="hasil-log" tabIndex={-1} ref={hasil}>{isi}</div>
    </Shell>
  );
}
