/* Pemilih jalur Pelatihan: satu radio menampilkan layanan, alasan, dan tautan ke bagiannya. */
import { useState } from 'react';
import { konten } from '../shared/content';
import { Icon } from '../shared/ui';
import { route } from './Common';

export default function TrainingChooser() {
  const jalur = konten.pelatihan.jalur;
  const [pilihan, setPilihan] = useState<number | null>(null);
  const terpilih = pilihan === null ? undefined : jalur[pilihan];
  return (
    <>
      <fieldset className="pelatihan-pilihan">
        <legend className="sr-only">Pilih kebutuhan pelatihan</legend>
        {jalur.map((j, i) => (
          <label key={i} className="pelatihan-pilihan__item" htmlFor={`jalur-${i}`}>
            <input
              type="radio"
              name="jalur-pelatihan"
              id={`jalur-${i}`}
              value={String(i)}
              checked={pilihan === i}
              onChange={(e) => {
                if (e.target.checked) setPilihan(i);
              }}
            />
            <span>{j.kebutuhan}</span>
          </label>
        ))}
      </fieldset>
      <div className="pelatihan-hasil">
        <div role="status" aria-live="polite" aria-atomic="true">
          {terpilih
            ? (
              <div key={pilihan} className="pelatihan-hasil__isi">
                <h3>{terpilih.layanan}</h3>
                <p>{terpilih.alasan}</p>
              </div>
            )
            : <p className="teks-sekunder">Pilih satu kebutuhan di atas untuk melihat jalur yang dapat dipertimbangkan.</p>}
        </div>
        <a className="tautan-panah" href={terpilih ? route(terpilih.href) : undefined} hidden={!terpilih}>
          {terpilih && <>{terpilih.tautan}<Icon name="panah" size={18} /></>}
        </a>
      </div>
    </>
  );
}
