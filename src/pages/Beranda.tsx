import { Link } from 'react-router-dom';
import SectionHeading from '../components/SectionHeading.tsx';
import HeroCarousel from '../components/HeroCarousel.tsx';
// Our Client dikomentari sesuai briefing rapat ("di-comment dulu, keep code").
// import ClientCarousel from '../components/ClientCarousel.tsx';
import NewsCard from '../components/NewsCard.tsx';
import DokumentasiCard from '../components/DokumentasiCard.tsx';
import TestimoniSlider from '../components/TestimoniSlider.tsx';
import { heroSlidesDummy } from '../data/hero.ts';
import { layananDummy } from '../data/layanan.ts';
// PROGRESS 2: section Tentang/Portofolio/Kontak dihapus dari home (andalan
// navbar & footer) dan Fasilitas dikomentari — datanya tetap dipakai halaman lain.
// import { kontakDummy, waLinkKontak } from '../data/kontak.ts';
// import { portofolioDummy } from '../data/portofolio.ts';
// import { fasilitasDummy } from '../data/fasilitas.ts';
import { useT } from '../lib/i18n.tsx';
import { useApiDaftar } from '../lib/useApiData.ts';
import { beritaDummy } from '../data/berita.ts';
import { dokumentasiDummy } from '../data/dokumentasi.ts';
// Bagian dari Our Client yang dikomentari di atas; data tetap dipertahankan.
// import { klienDummy } from '../data/klien.ts';
import { testimoniDummy } from '../data/testimoni.ts';

function SectionBar({
  kicker,
  title,
  sub,
  actionTo,
  actionLabel,
  dark = false,
}: {
  kicker: string;
  title: string;
  sub?: string;
  actionTo: string;
  actionLabel: string;
  dark?: boolean;
}) {
  const t = useT();
  return (
    <div className="flex flex-wrap items-end justify-between gap-4">
      <SectionHeading
        kicker={t(kicker)}
        title={t(title)}
        sub={sub ? t(sub) : undefined}
        align="left"
        tone={dark ? 'dark' : 'light'}
      />
      <Link
        to={actionTo}
        className={`shrink-0 rounded-lg px-5 py-2.5 font-display text-sm font-bold ${
          dark
            ? 'border border-white text-white hover:bg-white hover:text-brand'
            : 'btn-primary text-white'
        }`}
      >
        {t(actionLabel)}
      </Link>
    </div>
  );
}

// Struktur home (PROGRESS 2): hero → layanan → dokumentasi → testimoni → berita.
// Tentang Kami, Portofolio, Kontak dihapus dari home (dialihkan ke navbar/footer);
// Fasilitas dikomentari (keep code, jangan tampil dulu di home).
export default function Beranda() {
  const t = useT();

  // Hero & layanan: konten statis dari src/data/.
  // Dokumentasi/testimoni/berita memakai backend same-origin dengan fallback.
  const slidesHero = heroSlidesDummy;
  const daftarLayanan = layananDummy;
  // Urut tanggal terbaru dulu supaya slice(0, 3) benar-benar "3 terbaru"
  // (konsul PROGRESS 2; karya portofolio ikut sebagai entri dokumentasi ber-tanggal).
  const highlightDokumentasi = [...useApiDaftar('/dokumentasi', dokumentasiDummy)].sort((a, b) =>
    b.tanggal.localeCompare(a.tanggal),
  );
  // Section Our Client nonaktif; klien bukan konten CMS.
  // Testimoni tampil TEPAT 3 (urutan paling atas dari backend; item baru via
  // admin otomatis jadi #1, yang lama turun — tersimpan semua, tidak ada yang
  // dihapus, hanya tidak tampil di beranda bila keluar dari 3 besar).
  const daftarTestimoni = useApiDaftar('/testimoni', testimoniDummy).slice(0, 3);
  const highlightBerita = useApiDaftar('/berita', beritaDummy);

  return (
    <div>
      {/* Slide hero — konten statis src/data/hero.ts (diedit via frontend). */}
      <HeroCarousel slides={slidesHero} />

      <section id="layanan" className="relative isolate scroll-mt-20 overflow-hidden bg-sky py-16">
        {/* Transisi hero → sky dihaluskan (fade navy, -z-10 di bawah konten) + cincin ala referensi. */}
        <div
          aria-hidden="true"
          className="pointer-events-none absolute inset-x-0 top-0 -z-10 h-16"
          style={{ background: 'linear-gradient(to bottom, var(--color-navy), transparent)' }}
        />
        <div aria-hidden="true" className="pointer-events-none absolute -right-28 -top-24 h-72 w-72 rounded-full border border-brand/15" />
        <div aria-hidden="true" className="pointer-events-none absolute -bottom-32 -left-24 h-80 w-80 rounded-full border border-brand/10" />
        <div className="mx-auto max-w-6xl px-6">
          <SectionHeading kicker={t('Layanan')} title={t('Pilih Jalur Kolaborasimu')} sub={t('Pelatihan dan inference solution untuk kebutuhan nyata.')} />
          <div className="mt-8 grid gap-5 md:grid-cols-2">
            {daftarLayanan.map((c, i) => (
              <article key={c.slug} className="card-accent flex flex-col rounded-2xl border border-line bg-surface p-6 text-center">
                <span
                  aria-hidden="true"
                  className={`mx-auto inline-block rounded-full px-3 py-1 font-mono text-[11px] font-bold tracking-[0.2em] ${
                    i % 2 === 0 ? 'bg-yellow text-navy' : 'bg-mint text-navy'
                  }`}
                >
                  {String(i + 1).padStart(2, '0')}
                </span>
                <h3 className="mt-3 font-display font-bold">{t(c.nama)}</h3>
                <p className="mt-2 flex-1 text-sm text-muted">{t(c.tagline)}</p>
                <Link
                  to={`/layanan/${c.slug}`}
                  className="btn-accent mx-auto mt-4 inline-block rounded-lg px-5 py-2 text-sm font-bold"
                >
                  {t('Detail Layanan →')}
                </Link>
              </article>
            ))}
          </div>
        </div>
      </section>

      {/* Fasilitas: dikomentari sesuai konsul PROGRESS 2 ("keep codenya namun
          jangan tampil di home"). Aktifkan kembali dengan menghapus komentar. */}
      {/* <section aria-label={t('Fasilitas')} className="bg-soft py-16">
        <div className="mx-auto max-w-6xl px-6">
          <SectionHeading
            kicker={t('Lihat Ruangnya')}
            title={t('Fasilitas yang dirancang untuk berkarya')}
            sub={t('Ruang demo, ruang pelatihan, dan ruang diskusi — dilengkapi untuk presentasi, kelas, dan kerja tim.')}
          />
          <ul className="mt-8 grid gap-5 md:grid-cols-3">
            {fasilitasDummy.map((r) => (
              <li key={r.slug} className="card-accent overflow-hidden rounded-2xl border border-line bg-surface">
                <img
                  src={aset(r.gambar)}
                  alt={r.alt}
                  width={r.lebar}
                  height={r.tinggi}
                  loading="lazy"
                  className="aspect-[16/10] w-full object-cover"
                />
                <div className="p-5 text-center">
                  <p className="text-xs font-bold uppercase tracking-wider text-brand">{t(r.label)}</p>
                  <h3 className="mt-1 font-display font-bold">{t(r.judul)}</h3>
                  <p className="mt-2 text-sm text-muted">{t(r.deskripsi)}</p>
                </div>
              </li>
            ))}
          </ul>
        </div>
      </section> */}

      {/* aria-label langsung pada <section> (jadi region bernama); tanpa wrapper div. */}
      <section aria-label={t('Dokumentasi Kegiatan')} className="relative isolate overflow-hidden bg-gradient-to-br from-brand via-brand to-navy py-16 text-white">
        {/* Jembatan warna antar section (atas: sky → gradasi, bawah: gradasi → krem). */}
        <div
          aria-hidden="true"
          className="pointer-events-none absolute inset-x-0 top-0 -z-10 h-16"
          style={{ background: 'linear-gradient(to bottom, var(--color-sky), transparent)' }}
        />
        <div
          aria-hidden="true"
          className="pointer-events-none absolute inset-x-0 bottom-0 -z-10 h-16"
          style={{ background: 'linear-gradient(to top, var(--color-soft), transparent)' }}
        />
        {/* Dekorasi gradient profesional: blob terang + aksen mint/kuning di atas navy. */}
        <div aria-hidden="true" className="pointer-events-none absolute -right-24 -top-24 h-72 w-72 rounded-full bg-brand-bright opacity-50 blur-3xl" />
        <div aria-hidden="true" className="pointer-events-none absolute -bottom-28 -left-20 h-72 w-72 rounded-full bg-mint opacity-25 blur-3xl" />
        <div aria-hidden="true" className="pointer-events-none absolute right-1/4 top-8 h-20 w-20 rounded-full bg-yellow opacity-30 blur-2xl" />
        <div className="relative mx-auto max-w-6xl px-6">
          {/* Highlight 3 dokumentasi teratas — GET /api/dokumentasi (limit penampilan di frontend). */}
          <SectionBar
            kicker="Kegiatan"
            title="Dokumentasi Kegiatan"
            sub="Sorotan kegiatan terbaru AI Center."
            actionTo="/dokumentasi"
            actionLabel="Lihat Semua →"
            dark
          />
          <div className="mt-8 grid gap-5 md:grid-cols-3">
            {highlightDokumentasi.slice(0, 3).map((d) => (
              <DokumentasiCard key={d.slug} item={d} showLink={false} />
            ))}
          </div>
        </div>
      </section>

      {/* Our Client dikomentari sesuai briefing rapat ("our client di-comment dulu, keep code").
          Testimoni (slot berikutnya) menempati posisinya; aktifkan kembali section bila dibutuhkan.
      {/* <section aria-label={t('Klien Kami')} className="py-16">
        <div className="mx-auto max-w-6xl px-6">
          <SectionHeading kicker={t('Dipercaya')} title={t('Klien Kami')} sub={t('Mitra yang berkolaborasi dengan AI Center.')} />
          <ClientCarousel items={daftarKlien} perPage={3} />
        </div>
      </section> */}

      <section aria-label={t('Testimoni')} className="bg-soft py-16">
        <div className="mx-auto max-w-6xl px-6 text-center">
          <SectionHeading kicker={t('Testimoni')} title={t('Apa Kata Mereka?')} sub={t('Cerita peserta dan mitra AI Center.')} />
          {/* Testimoni — GET /api/testimoni. */}
          <TestimoniSlider items={daftarTestimoni} dark={false} />
        </div>
      </section>

      <section aria-label={t('Berita Terkini')} className="bg-gradient-to-b from-white via-sky/70 to-navy py-16">
        <div className="mx-auto max-w-6xl px-6">
          {/* Highlight 3 berita teratas — GET /api/berita (limit penampilan di frontend). */}
          <SectionBar
            kicker="Berita"
            title="Berita Terkini"
            sub="Kabar terbaru AI Center."
            actionTo="/berita"
            actionLabel="Lihat Semua →"
          />
          <div className="mt-8 grid gap-5 md:grid-cols-3">
            {highlightBerita.slice(0, 3).map((b) => (
              <NewsCard key={b.slug} item={b} showLink={false} />
            ))}
          </div>
        </div>
      </section>
    </div>
  );
}
