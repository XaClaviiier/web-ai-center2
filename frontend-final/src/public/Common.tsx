/* Kerangka dan kartu halaman publik. Semua teks dirender React sebagai teks, tanpa innerHTML. */
import type { ReactNode } from 'react';
import { Button } from '../components/ui/button';
import { konten, image } from '../shared/content';
import { Format } from '../shared/format';
import type { Album, Member, News } from '../shared/types';
import { Icon, State, type IconName } from '../shared/ui';

/** Nama file prototype di `konten` dipetakan ke rute akhir. */
const RUTE: Record<string, string> = {
  'index.html': '/',
  'pelatihan.html': '/layanan/pelatihan',
  'inference.html': '/layanan/inference-solutions',
  'jejak-alumni.html': '/layanan/inference-solutions/jejak-alumni',
  'perencanaan-implan-gigi.html': '/layanan/inference-solutions/perencanaan-implan-gigi',
  'document-intelligence.html': '/layanan/inference-solutions/document-intelligence',
  'berita.html': '/berita',
  'dokumentasi.html': '/dokumentasi',
  'tim.html': '/tim',
  'tentang-kami.html': '/tentang-kami',
};

export function route(href: string): string {
  const [path = '', hash] = href.split('#');
  return (RUTE[path] ?? path) + (hash === undefined ? '' : `#${hash}`);
}

type Variant = 'utama' | 'kedua' | 'terang' | 'garis-terang';

export function LinkButton({ children, href, variant = 'utama', icon, external = false }: {
  children: ReactNode;
  href: string;
  variant?: Variant;
  icon?: IconName | undefined;
  external?: boolean;
}) {
  return (
    <Button asChild variant={variant}>
      <a href={route(href)} target={external ? '_blank' : undefined} rel={external ? 'noopener' : undefined}>
        {icon && <Icon name={icon} size={18} />}
        {children}
      </a>
    </Button>
  );
}

export function Breadcrumb({ items }: { items: { label: string; href?: string }[] }) {
  const semua = [{ label: 'Beranda', href: '/' }, ...items];
  return (
    <nav className="remah" aria-label="Breadcrumb">
      <ol className="wadah">
        {semua.map((item, i) => (
          <li key={i}>
            {i === semua.length - 1 ? <span aria-current="page">{item.label}</span> : <a href={item.href}>{item.label}</a>}
          </li>
        ))}
      </ol>
    </nav>
  );
}

export function PageHeader({ title, text }: { title: string; text?: string }) {
  return (
    <div className="kepala-halaman wadah">
      <h1>{title}</h1>
      {text && <p>{text}</p>}
    </div>
  );
}

export function ContactActions() {
  const k = konten.kontak;
  return (
    <div className="kontak__aksi">
      <LinkButton href={k.whatsapp_url} variant="terang" icon="pesan" external>Hubungi lewat WhatsApp</LinkButton>
      <LinkButton href={k.email_url} variant="garis-terang" icon="surat">Kirim email</LinkButton>
    </div>
  );
}

export function Contact({ title, text, headingId = 'judul-kontak', className, textClassName }: {
  title: string;
  text: string;
  headingId?: string;
  className?: string;
  textClassName?: string;
}) {
  return (
    <section className={className ? `kontak ${className}` : 'kontak'} aria-labelledby={headingId}>
      <div className="wadah kontak__dalam">
        <div className={textClassName}>
          <h2 id={headingId}>{title}</h2>
          <p>{text}</p>
        </div>
        <ContactActions />
      </div>
    </section>
  );
}

export function Footer() {
  const k = konten.kontak;
  const jelajahi = [
    ['Pelatihan', 'pelatihan.html'],
    ['Inference and Solutions', 'inference.html'],
    ['Berita', 'berita.html'],
    ['Galeri Dokumentasi', 'dokumentasi.html'],
    ['Tim', 'tim.html'],
    ['Tentang Kami', 'tentang-kami.html'],
  ] as const;
  return (
    <footer className="footer">
      <div className="wadah footer__grid">
        <div className="footer__profil">
          <img src={image('AI-Center_Logo.png')} alt="UBAYA AI Center" width={440} height={116} className="footer__logo" />
          <p>{k.deskripsi_footer}</p>
        </div>
        <div>
          <h2>Jelajahi</h2>
          <ul>
            {jelajahi.map(([label, href]) => <li key={href}><a href={route(href)}>{label}</a></li>)}
          </ul>
        </div>
        <address>
          <h2>Kontak</h2>
          <ul>
            <li>
              <a className="footer__kontak" href={k.whatsapp_url} target="_blank" rel="noopener">
                <Icon name="pesan" size={16} />
                <span className="sr-only">WhatsApp </span>
                {k.whatsapp}
              </a>
            </li>
            <li>
              <a className="footer__kontak" href={k.email_url} target="_blank" rel="noopener">
                <Icon name="surat" size={16} />
                {k.email}
              </a>
            </li>
            <li>
              <a className="footer__kontak" href={k.instagram_url} target="_blank" rel="noopener">
                <Icon name="kamera" size={16} />
                <span className="sr-only">Instagram </span>
                {k.instagram}
              </a>
            </li>
            {k.alamat.map((baris) => <li key={baris}>{baris}</li>)}
          </ul>
        </address>
        <div className="footer__peta">
          <h2>Peta lokasi</h2>
          <iframe src={k.peta_embed} title="Peta lokasi UBAYA AI Center" loading="lazy" referrerPolicy="no-referrer-when-downgrade" />
          <a href={k.peta_url} target="_blank" rel="noopener">Buka di Google Maps</a>
        </div>
      </div>
      <div className="wadah footer__bawah"><small>© 2026 UBAYA AI Center</small></div>
    </footer>
  );
}

export function NotFound({ label = 'Halaman' }: { label?: string }) {
  return (
    <>
      <Breadcrumb items={[{ label: 'Halaman tidak ditemukan' }]} />
      <section className="wadah tak-ditemukan">
        <p className="tak-ditemukan__kode" aria-hidden="true">404</p>
        <h1>{label} tidak ditemukan</h1>
        <p className="teks-sekunder">Tautan mungkin sudah berubah atau kontennya tidak lagi tersedia.</p>
        <LinkButton href="/">Kembali ke Beranda</LinkButton>
      </section>
    </>
  );
}

/** Keadaan galat dengan tombol Coba lagi; 429 menyebut waktu tunggu. */
export function ErrorState({ error, onRetry }: { error: unknown; onRetry: () => void }) {
  return (
    <State
      error={error}
      action={(
        <Button variant="kedua" onClick={onRetry}>
          <Icon name="ulang" size={18} />
          Coba lagi
        </Button>
      )}
    />
  );
}

export function NewsCard({ news, heading: Heading = 'h2' }: { news: News; heading?: 'h2' | 'h3' }) {
  return (
    <article className="kartu kartu-berita">
      <a className="kartu-berita__tautan" href={`/berita/${encodeURIComponent(news.slug)}`}>
        <span className="kartu-berita__media">
          <img className="kartu__gambar" src={image(news.gambar)} alt="" loading="lazy" />
        </span>
        <div className="kartu__isi">
          <time className="tanggal" dateTime={Format.parseTanggal(news.tanggalPublikasi)?.toISOString().slice(0, 10)}>{news.tanggalPublikasi}</time>
          <Heading>{news.judul}</Heading>
          <p className="teks-sekunder">{Format.potongan(news.isi)}</p>
          <span className="tautan-panah">Lihat detail<Icon name="panah" size={18} /></span>
        </div>
      </a>
    </article>
  );
}

export function DocCard({ album }: { album: Album }) {
  const href = `/dokumentasi/${encodeURIComponent(album.slug)}`;
  return (
    <article className="kartu">
      <img className="kartu__gambar" src={image(album.gambar[0]?.key)} alt="" loading="lazy" />
      <div className="kartu__isi">
        <time className="tanggal">{album.tanggalPublikasi}</time>
        <h3><a className="tautan-kartu" href={href}>{album.judul}</a></h3>
        <a className="tautan-panah" href={href} aria-label={`Lihat detail: ${album.judul}`}>Lihat detail<Icon name="panah" size={18} /></a>
      </div>
    </article>
  );
}

export function AlbumCard({ album }: { album: Album }) {
  return (
    <article className="kartu kartu-album">
      <a className="kartu-album__tautan" href={`/dokumentasi/${encodeURIComponent(album.slug)}`}>
        <span className="kartu-album__media">
          <img className="kartu__gambar" src={image(album.gambar[0]?.key)} alt="" loading="lazy" />
        </span>
        <div className="kartu__isi">
          <div className="kartu-album__meta">
            <time className="tanggal" dateTime={Format.parseTanggal(album.tanggalPublikasi)?.toISOString().slice(0, 10)}>{album.tanggalPublikasi}</time>
            <span className="kartu-album__jumlah"><Icon name="galeri" size={16} />{`${album.gambar.length} Foto`}</span>
          </div>
          <h2>{album.judul}</h2>
        </div>
      </a>
    </article>
  );
}

export function MemberCard({ member }: { member: Member }) {
  return (
    <article className="kartu-tim">
      <img src={image(member.foto)} alt={`Foto ${member.nama}`} loading="lazy" width={96} height={120} />
      <div className="kartu-tim__identitas">
        <h2>{member.nama}</h2>
        <p className="kartu-tim__jabatan">{member.posisi}</p>
      </div>
      <div className="kartu-tim__riset">
        <p className="kartu-tim__label">Bidang Penelitian</p>
        <p>{member.bidangRiset}</p>
      </div>
    </article>
  );
}
