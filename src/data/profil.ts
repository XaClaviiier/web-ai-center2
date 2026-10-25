// Visi & Misi resmi AI Center — mengikuti dokumen "Visi & Misi" (Arah Kami).
// KONTEN STATIS frontend (sumber satu-satunya) — diedit langsung di file ini
// (tidak lewat backend/admin); TentangKami memakai profilDummy langsung.
// petakanProfilApi() masih dipakai test pemetaan (bila suatu saat dihidupkan lagi).
export interface Profil {
  /** Kalimat pengantar section Visi & Misi. */
  introVisiMisi: string;
  visi: {
    judul: string;
    deskripsi: string;
  };
  /** Daftar poin misi, berurutan 01..n. */
  misi: string[];
}

export const profilDummy: Profil = {
  introVisiMisi:
    'Komitmen kami membangun ekosistem kecerdasan buatan yang inovatif, aplikatif, dan berdampak.',
  visi: {
    judul: 'Menjadi AI Solution Factory terdepan',
    deskripsi:
      'Menghasilkan produk, menjadi pusat riset, serta meningkatkan kapasitas sumber daya manusia ' +
      'di bidang AI yang memberikan dampak nyata bagi akademik, industri, dan masyarakat.',
  },
  misi: [
    'Mendorong riset AI yang inovatif dan aplikatif.',
    'Menghasilkan produk dan solusi AI yang siap digunakan.',
    'Mengembangkan talenta AI yang kompeten.',
    'Membangun kolaborasi strategis dengan pemerintah dan industri.',
    'Menciptakan ekosistem inovasi dan startup berbasis AI.',
  ],
};

/** Bentuk legacy untuk test mapper; halaman publik memakai `profilDummy` statis. */
export interface ProfilApi {
  nama: string;
  tagline: string;
  ringkasan: string;
  alamat: string;
  email: string;
  telepon: string;
  /** Satu baris: 'judul — deskripsi' (lihat catatan di atas profilDummy). */
  visi?: string;
  /** Beberapa baris dipisah '\n', satu poin misi per baris. */
  misi?: string;
  statistik?: Array<{ label: string; value: string }>;
}

/** Pecah string visi API ('judul — deskripsi') menjadi struktur UI. */
function pecahVisi(visi: string): Profil['visi'] {
  const pemisah = ' — ';
  const indeks = visi.indexOf(pemisah);
  if (indeks === -1) {
    // Tanpa pemisah → seluruh teks dianggap judul (deskripsi kosong).
    return { judul: visi, deskripsi: '' };
  }
  return {
    judul: visi.slice(0, indeks),
    deskripsi: visi.slice(indeks + pemisah.length),
  };
}

/**
 * Petakan respons API → struktur Profil untuk UI.
 * Aturan eksplisit:
 * - `introVisiMisi` tidak ada di kolom backend → selalu dari profilDummy;
 * - `visi` kosong/null → kartu visi memakai profilDummy (mencegah kartu
 *   kosong; backend seed selalu mengisinya);
 * - `misi` kosong/null → `[]`, supaya empty-state "Misi belum tersedia."
 *   di halaman tetap tampil bila CMS sengaja dikosongkan.
 */
export function petakanProfilApi(api: ProfilApi): Profil {
  const misi = (api.misi ?? '')
    .split('\n')
    .map((poin) => poin.trim())
    .filter((poin) => poin !== '');
  return {
    introVisiMisi: profilDummy.introVisiMisi,
    visi: api.visi !== undefined && api.visi.trim() !== '' ? pecahVisi(api.visi) : profilDummy.visi,
    misi,
  };
}
