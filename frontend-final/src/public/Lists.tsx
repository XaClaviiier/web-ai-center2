/* Island data publik. Kunci query memuat versi simulasi agar panel selalu memicu muat ulang. */
import { Fragment, useEffect, useRef, useState, type ReactNode } from 'react';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { Input } from '../components/ui/input';
import { api } from '../simulation/api';
import { useSimulationVersion } from '../simulation/state';
import { image } from '../shared/content';
import { Format } from '../shared/format';
import QueryProvider from '../shared/QueryProvider';
import { ApiError, type Album, type News, type Page } from '../shared/types';
import { Icon, Loading, Pagination, State } from '../shared/ui';
import Carousel from './Carousel';
import { AlbumCard, Breadcrumb, DocCard, ErrorState, MemberCard, NewsCard, NotFound } from './Common';

const JUDUL_SITUS = 'UBAYA AI Center';

function setMeta(title: string, description?: string): void {
  document.title = `${title} | ${JUDUL_SITUS}`;
  if (description) document.querySelector('meta[name="description"]')?.setAttribute('content', description);
}

/** Loading, galat (dengan Coba lagi), kosong, lalu isi. Coba lagi menampilkan kerangka kembali. */
function Muat<T extends { items: unknown[] }>({ query, kerangka, kosong, children }: {
  query: UseQueryResult<T>;
  kerangka: ReactNode;
  kosong: ReactNode;
  children: (data: T) => ReactNode;
}) {
  if (query.isPending || (query.isError && query.isFetching)) return kerangka;
  if (query.isError) return <ErrorState error={query.error} onRetry={() => void query.refetch()} />;
  return query.data.items.length ? children(query.data) : kosong;
}

/* === Beranda === */

function HomeAlbumsIsi() {
  const versi = useSimulationVersion();
  const query = useQuery({ queryKey: ['dokumentasi', 'beranda', versi], queryFn: () => api.daftarDokumentasi({ limit: 3 }) });
  return (
    <Muat query={query} kerangka={<Loading count={3} />} kosong={<State title="Belum ada dokumentasi" text="Dokumentasi kegiatan akan tampil di sini." />}>
      {(data) => <div className="grid-kartu grid-kartu--tiga">{data.items.map((d) => <DocCard key={d.id} album={d} />)}</div>}
    </Muat>
  );
}

function HomeTestimonialsIsi() {
  const versi = useSimulationVersion();
  const query = useQuery({ queryKey: ['testimoni', versi], queryFn: () => api.daftarTestimoni() });
  return (
    <Muat
      query={query}
      kerangka={<div className="kerangka" style={{ height: '16rem' }} role="status" aria-label="Memuat data" />}
      kosong={<State title="Belum ada testimoni" text="Testimoni akan tampil di sini." />}
    >
      {(data) => (
        <Carousel
          label="Testimoni klien"
          intervalMs={5000}
          slides={data.items.map((t) => (
            <figure className="testi__kutip">
              <blockquote>{t.teks}</blockquote>
              <figcaption>
                {t.foto
                  ? <img className="testi__foto" src={image(t.foto)} alt="" />
                  : <span className="testi__inisial" aria-hidden="true">{Format.inisial(t.nama)}</span>}
                <span><strong>{t.nama}</strong><br />{t.posisi}</span>
              </figcaption>
            </figure>
          ))}
        />
      )}
    </Muat>
  );
}

function HomeNewsIsi() {
  const versi = useSimulationVersion();
  const query = useQuery({ queryKey: ['berita', 'beranda', versi], queryFn: () => api.daftarBerita({ limit: 3 }) });
  return (
    <Muat query={query} kerangka={<Loading count={3} />} kosong={<State title="Belum ada berita" text="Berita terbaru akan tampil di sini." />}>
      {(data) => <div className="grid-kartu grid-kartu--tiga">{data.items.map((b) => <NewsCard key={b.id} news={b} heading="h3" />)}</div>}
    </Muat>
  );
}

export function HomeAlbums() {
  return <QueryProvider><HomeAlbumsIsi /></QueryProvider>;
}

export function HomeTestimonials() {
  return <QueryProvider><HomeTestimonialsIsi /></QueryProvider>;
}

export function HomeNews() {
  return <QueryProvider><HomeNewsIsi /></QueryProvider>;
}

/* === Tim === */

function TeamListIsi() {
  const versi = useSimulationVersion();
  const query = useQuery({ queryKey: ['tim', versi], queryFn: () => api.daftarTim() });
  return (
    <Muat query={query} kerangka={<Loading count={6} />} kosong={<State title="Belum ada data tim" text="Susunan tim akan tampil di sini." />}>
      {(data) => <div className="grid-tim">{data.items.map((t) => <MemberCard key={t.id} member={t} />)}</div>}
    </Muat>
  );
}

export function TeamList() {
  return <QueryProvider><TeamListIsi /></QueryProvider>;
}

/* === Daftar dengan pencarian dan pagination === */

type Kind = 'berita' | 'dokumentasi';

function bacaUrl(): { q: string; page: number } {
  const params = new URLSearchParams(location.search);
  return { q: (params.get('q') ?? '').trim(), page: Math.max(1, Number.parseInt(params.get('page') ?? '', 10) || 1) };
}

function simpanUrl(q: string, page: number, ganti = false): void {
  const url = new URL(location.href);
  if (q) url.searchParams.set('q', q);
  else url.searchParams.delete('q');
  if (page > 1) url.searchParams.set('page', String(page));
  else url.searchParams.delete('page');
  if (url.href === location.href) return;
  if (ganti) history.replaceState(null, '', url);
  else history.pushState(null, '', url);
}

interface ListConfig<T> {
  kind: Kind;
  namaJamak: string;
  grid: string;
  ambil: (opsi: { page: number; q: string }) => Promise<Page<T>>;
  card: (item: T) => ReactNode;
}

function ListPageIsi<T extends { id: number }>({ kind, namaJamak, grid, ambil, card }: ListConfig<T>) {
  const versi = useSimulationVersion();
  const [awal] = useState(bacaUrl);
  const [q, setQ] = useState(awal.q);
  const [page, setPage] = useState(awal.page);
  const [nilai, setNilai] = useState(awal.q);
  const waktu = useRef<number>(undefined);
  const pindahFokus = useRef(false);
  const hasil = useRef<HTMLDivElement>(null);
  const qAktif = q.length >= 2 ? q : '';

  const query = useQuery({ queryKey: [kind, 'daftar', page, qAktif, versi], queryFn: () => ambil({ page, q: qAktif }) });
  const data = query.data;
  const terakhir = data ? Math.max(1, Math.ceil(data.total / data.limit)) : page;

  useEffect(() => {
    const kembali = () => {
      clearTimeout(waktu.current);
      const url = bacaUrl();
      setQ(url.q);
      setNilai(url.q);
      setPage(url.page);
    };
    window.addEventListener('popstate', kembali);
    return () => {
      window.removeEventListener('popstate', kembali);
      clearTimeout(waktu.current);
    };
  }, []);

  // Halaman di luar jangkauan dipindah ke halaman terakhir tanpa menambah riwayat.
  useEffect(() => {
    if (!query.isSuccess || page <= terakhir) return;
    setPage(terakhir);
    simpanUrl(q, terakhir, true);
  }, [query.isSuccess, page, terakhir, q]);

  useEffect(() => {
    if (query.isPending || query.isFetching || !pindahFokus.current) return;
    pindahFokus.current = false;
    if (document.activeElement === document.body && hasil.current) {
      hasil.current.focus({ preventScroll: true });
      hasil.current.scrollIntoView({ block: 'start' });
    }
  }, [query.isPending, query.isFetching]);

  const keHalaman = (n: number) => {
    setPage(n);
    simpanUrl(q, n);
    pindahFokus.current = true;
  };

  let status = 'Memuat hasil…';
  if (query.isError && !query.isFetching) status = 'Hasil belum dapat dimuat.';
  else if (data && !query.isFetching) status = data.total ? `Halaman ${page} dari ${terakhir}. ${data.total} ${namaJamak} ditemukan.` : 'Tidak ada hasil.';

  const kosong = qAktif
    ? <State title="Tidak ada hasil" text={`Tidak ada ${namaJamak} dengan judul "${qAktif}". Coba kata lain.`} icon="cari" />
    : <State title={`Belum ada ${namaJamak}`} text={`${namaJamak.charAt(0).toUpperCase()}${namaJamak.slice(1)} akan tampil di sini setelah dipublikasikan.`} />;

  return (
    <div className="wadah daftar">
      <div className="field cari-blok">
        <label htmlFor="cari">{`Cari ${namaJamak}`}</label>
        <div className="cari">
          <Icon name="cari" size={20} />
          <Input
            type="search"
            id="cari"
            placeholder="Cari judul"
            autoComplete="off"
            aria-describedby="petunjuk-cari"
            value={nilai}
            onChange={(e) => {
              const teks = e.target.value;
              setNilai(teks);
              clearTimeout(waktu.current);
              waktu.current = window.setTimeout(() => {
                setQ(teks.trim());
                setPage(1);
                simpanUrl(teks.trim(), 1);
              }, 300);
            }}
          />
        </div>
        <p className="field__bantuan" id="petunjuk-cari">{q.length === 1 ? 'Ketik minimal 2 karakter.' : ''}</p>
      </div>
      <p className="field__bantuan" role="status" aria-atomic="true">{status}</p>
      <div id="hasil" ref={hasil} tabIndex={-1} aria-label="Hasil pencarian">
        <Muat query={query} kerangka={<Loading count={9} />} kosong={kosong}>
          {(d) => (
            <>
              <div className={grid}>
                {d.items.map((item) => <Fragment key={item.id}>{card(item)}</Fragment>)}
              </div>
              <Pagination page={d.page} limit={d.limit} total={d.total} onPage={keHalaman} />
            </>
          )}
        </Muat>
      </div>
    </div>
  );
}

export function ListPage({ kind }: { kind: Kind }) {
  return (
    <QueryProvider>
      {kind === 'berita'
        ? (
          <ListPageIsi
            kind="berita"
            namaJamak="berita"
            grid="grid-kartu grid-kartu--berita"
            ambil={(opsi) => api.daftarBerita(opsi)}
            card={(b: News) => <NewsCard news={b} />}
          />
        )
        : (
          <ListPageIsi
            kind="dokumentasi"
            namaJamak="album"
            grid="grid-kartu grid-kartu--album"
            ambil={(opsi) => api.daftarDokumentasi(opsi)}
            card={(d: Album) => <AlbumCard album={d} />}
          />
        )}
    </QueryProvider>
  );
}

/* === Detail === */

function DetailSkeleton() {
  return (
    <div className="wadah detail-kerangka" role="status" aria-label="Memuat data">
      <div className="kerangka" style={{ height: '3rem', width: '70%' }} />
      <div className="kerangka" style={{ height: '24rem' }} />
    </div>
  );
}

/** Slug tidak dikenal menjadi 404 sisi client; galat lain dapat dicoba lagi. */
function Detail<T>({ query, label, children }: { query: UseQueryResult<T>; label: string; children: (data: T) => ReactNode }) {
  const tidakAda = query.isError && !query.isFetching && query.error instanceof ApiError && query.error.status === 404;
  useEffect(() => {
    if (!tidakAda) return;
    setMeta('Halaman tidak ditemukan');
    document.querySelector('meta[name="robots"]')?.setAttribute('content', 'noindex');
  }, [tidakAda]);

  if (query.isPending || (query.isError && query.isFetching)) return <DetailSkeleton />;
  if (tidakAda) return <NotFound label={label} />;
  if (query.isError) return <div className="wadah daftar"><ErrorState error={query.error} onRetry={() => void query.refetch()} /></div>;
  return children(query.data);
}

function NewsDetailIsi({ slug }: { slug: string }) {
  const versi = useSimulationVersion();
  const query = useQuery({ queryKey: ['berita', 'detail', slug, versi], queryFn: () => api.detailBerita(slug) });
  const berita = query.data;
  useEffect(() => {
    if (berita) setMeta(berita.judul, Format.potongan(berita.isi, 160));
  }, [berita]);
  return (
    <Detail query={query} label="Berita">
      {(b) => (
        <div>
          <Breadcrumb items={[{ label: 'Berita', href: '/berita' }, { label: b.judul }]} />
          <article className="artikel">
            <header className="artikel__kepala">
              <h1>{b.judul}</h1>
              <p className="artikel__meta">
                <time className="tanggal">{b.tanggalPublikasi}</time>
                {' '}
                <span>oleh Tim Media Ubaya AI Center</span>
              </p>
            </header>
            <img className="artikel__gambar" src={image(b.gambar)} alt="" />
            <div className="artikel__isi">
              {Format.paragraf(b.isi).map((p, i) => <p key={i}>{p}</p>)}
            </div>
            <a className="tautan-panah artikel__kembali" href="/berita">Semua berita<Icon name="panah" size={18} /></a>
          </article>
        </div>
      )}
    </Detail>
  );
}

function AlbumDetailIsi({ slug }: { slug: string }) {
  const versi = useSimulationVersion();
  const query = useQuery({ queryKey: ['dokumentasi', 'detail', slug, versi], queryFn: () => api.detailDokumentasi(slug) });
  const album = query.data;
  useEffect(() => {
    if (album) setMeta(album.judul, `Dokumentasi kegiatan ${album.judul} — ${album.tanggalPublikasi}`);
  }, [album]);
  return (
    <Detail query={query} label="Dokumentasi">
      {(d) => (
        <div>
          <Breadcrumb items={[{ label: 'Galeri Dokumentasi', href: '/dokumentasi' }, { label: d.judul }]} />
          <article className="wadah dok">
            <header className="artikel__kepala">
              <h1>{d.judul}</h1>
              <p className="artikel__meta"><time className="tanggal">{d.tanggalPublikasi}</time></p>
            </header>
            <div className={`galeri galeri--${d.gambar.length}`}>
              {d.gambar.map((g, i) => (
                <img key={g.key} src={image(g.key)} alt={`Foto ${i + 1} dari ${d.gambar.length}: ${d.judul}`} loading={i ? 'lazy' : undefined} />
              ))}
            </div>
            <a className="tautan-panah artikel__kembali" href="/dokumentasi">Semua Galeri Dokumentasi<Icon name="panah" size={18} /></a>
          </article>
        </div>
      )}
    </Detail>
  );
}

export function NewsDetail({ slug }: { slug: string }) {
  return <QueryProvider><NewsDetailIsi slug={slug} /></QueryProvider>;
}

export function AlbumDetail({ slug }: { slug: string }) {
  return <QueryProvider><AlbumDetailIsi slug={slug} /></QueryProvider>;
}
