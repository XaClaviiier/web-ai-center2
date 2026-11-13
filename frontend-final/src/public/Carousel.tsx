/* Carousel: otomatis hanya di perangkat ber-hover (berhenti saat disorot atau difokus);
   layar sentuh dan reduced motion manual lewat panah, karena tanpa tombol jeda (WCAG 2.2.2). */
import { useEffect, useState, useSyncExternalStore, type ReactNode } from 'react';
import useEmblaCarousel from 'embla-carousel-react';
import { Icon } from '../shared/ui';

function useMedia(query: string): boolean {
  return useSyncExternalStore(
    (listener) => {
      const media = matchMedia(query);
      media.addEventListener('change', listener);
      return () => media.removeEventListener('change', listener);
    },
    () => matchMedia(query).matches,
    () => false,
  );
}

export default function Carousel({ slides, label, intervalMs }: { slides: ReactNode[]; label: string; intervalMs: number }) {
  const [viewportRef, embla] = useEmblaCarousel({ align: 'start' });
  const [posisi, setPosisi] = useState(0);
  const [disorot, setDisorot] = useState(false);
  const [difokus, setDifokus] = useState(false);
  const hover = useMedia('(hover: hover)');
  const kurangGerak = useMedia('(prefers-reduced-motion: reduce)');
  const jumlah = slides.length;

  useEffect(() => {
    if (!embla) return;
    const pilih = () => setPosisi(embla.selectedScrollSnap());
    pilih();
    embla.on('select', pilih).on('reInit', pilih);
    return () => {
      embla.off('select', pilih).off('reInit', pilih);
    };
  }, [embla]);

  const tampil = (i: number) => embla?.scrollTo((i + jumlah) % jumlah, kurangGerak);

  useEffect(() => {
    if (!embla || !hover || kurangGerak || disorot || difokus || jumlah <= 1) return;
    const timer = setInterval(() => embla.scrollTo((embla.selectedScrollSnap() + 1) % jumlah), intervalMs);
    return () => clearInterval(timer);
  }, [embla, hover, kurangGerak, disorot, difokus, jumlah, intervalMs]);

  return (
    <div
      className="carousel"
      role="region"
      aria-roledescription="carousel"
      aria-label={label}
      onMouseEnter={() => setDisorot(true)}
      onMouseLeave={() => setDisorot(false)}
      onFocus={() => setDifokus(true)}
      onBlur={(e) => {
        if (!e.currentTarget.contains(e.relatedTarget)) setDifokus(false);
      }}
    >
      <div className="overflow-hidden" ref={viewportRef}>
        <div className="flex items-start">
          {slides.map((isi, i) => (
            <div
              key={i}
              className="min-w-0 shrink-0 grow-0 basis-full"
              role="group"
              aria-roledescription="slide"
              aria-label={`${i + 1} dari ${jumlah}`}
              aria-hidden={i !== posisi}
              inert={i !== posisi}
            >
              {isi}
            </div>
          ))}
        </div>
      </div>
      {jumlah > 1 && (
        <div className="carousel__kontrol">
          <button className="carousel__tombol" type="button" aria-label="Sebelumnya" onClick={() => tampil(posisi - 1)}>
            <Icon name="kiri" size={18} />
          </button>
          <span className="carousel__posisi" aria-hidden="true">{`${posisi + 1} / ${jumlah}`}</span>
          <button className="carousel__tombol" type="button" aria-label="Berikutnya" onClick={() => tampil(posisi + 1)}>
            <Icon name="kanan" size={18} />
          </button>
        </div>
      )}
    </div>
  );
}

/** Island pertanyaan Inference: teks statis dari `konten`, dibungkus kutip editorial. */
export function QuestionCarousel({ questions }: { questions: string[] }) {
  return (
    <Carousel
      label="Pertanyaan untuk Anda"
      intervalMs={6000}
      slides={questions.map((teks) => <p className="pertanyaan">{`“${teks}”`}</p>)}
    />
  );
}
