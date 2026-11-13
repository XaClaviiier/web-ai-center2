/* Klien API nyata ke backend Go (pengganti API tiruan tim frontend). Nama ekspor dan tanda tangan fungsi mengikuti kontrak lama,
   jadi halaman tidak berubah.
   Satu origin, tanpa CORS. Sesi lewat cookie HttpOnly; metode ubah-data membawa X-CSRF-Token.
   Backend memakai nama field Indonesia (dibuatPada, dibuatOleh, dst.), gambar sebagai {key,url}, dan daftar
   dokumentasi hanya membawa gambar utama + jumlah; semuanya dipetakan di sini ke bentuk yang dibaca halaman. */
import {
  ApiError,
  type Account, type Album, type Entity, type Image, type Item, type Log, type Member, type News, type Page, type Payload, type Testimonial,
} from '../shared/types';

type Json = Record<string, any>; // bentuk respons backend, dipetakan di bawah

const BERSTATUS: Entity[] = ['berita', 'dokumentasi', 'testimoni'];
const MODUL: Entity[] = ['berita', 'dokumentasi', 'testimoni', 'tim'];
const AWALAN_MEDIA = '/media/';

let csrf: string | null = null;
let akunAktif: Account | null = null;
let kunciSaya: { sumber: Entity; id: number } | null = null;
// Versi terakhir yang dilihat per item: Post/Retract wajib membawa version, sedangkan halaman tidak mengirimnya.
const versiTerakhir = new Map<string, number>();
const kunciUnggah = new Map<string, string>(); // URL blob: -> key hasil unggah (percobaan ulang tidak mengunggah dua kali)

/* === Fetch dasar === */
async function minta(metode: string, url: string, opsi: { json?: unknown; form?: FormData; keepalive?: boolean } = {}): Promise<any> {
  const headers: Record<string, string> = { Accept: 'application/json' };
  let body: BodyInit | null = null;
  if (opsi.json !== undefined) {
    headers['Content-Type'] = 'application/json';
    body = JSON.stringify(opsi.json);
  }
  if (opsi.form) body = opsi.form;
  if (metode !== 'GET' && metode !== 'HEAD' && csrf) headers['X-CSRF-Token'] = csrf;
  let res: Response;
  try {
    res = await fetch(url, { method: metode, headers, body, credentials: 'same-origin', keepalive: opsi.keepalive === true });
  } catch {
    throw new ApiError(0, 'jaringan', 'Tidak dapat menghubungi server. Periksa koneksi Anda lalu coba lagi.');
  }
  if (res.status === 204) return null;
  let isi: Json | null = null;
  try {
    isi = await res.json();
  } catch {
    /* badan bukan JSON */
  }
  if (res.ok) return isi;
  const e = isi?.error ?? {};
  const tunggu = res.status === 429 ? Number.parseInt(res.headers.get('Retry-After') ?? '', 10) || 30 : undefined;
  const galat = new ApiError(res.status, e.code ?? 'galat_server', e.message ?? 'Terjadi kendala di server. Coba lagi beberapa saat lagi.', tunggu);
  (galat as ApiError & { pemegang?: { nama: string } }).pemegang = e.pemegang;
  throw galat;
}

const kueri = (opsi: Record<string, unknown>): string => {
  const p = new URLSearchParams();
  for (const [k, v] of Object.entries(opsi)) if (v !== undefined && v !== null && v !== '') p.set(k, String(v));
  const s = p.toString();
  return s ? `?${s}` : '';
};

/* === Pemetaan gambar === */
/** Key penyimpanan dari URL /media/{key}; nilai lain dikembalikan apa adanya. */
const keKey = (url: unknown): string => {
  const s = String(url ?? '');
  return s.startsWith(AWALAN_MEDIA) ? s.slice(AWALAN_MEDIA.length) : s;
};
const keKeyAtauNull = (url: unknown): string | null => (url ? keKey(url) : null);
/** Daftar publik/admin hanya membawa gambar utama + jumlah; kartu memakai gambar[0] dan gambar.length. */
const gambarRingkas = (utama: string | null, jumlah: number): Image[] => Array.from({ length: jumlah }, (_, i) => ({ key: i === 0 && utama ? utama : '', urutan: i + 1 }));

/* === Pemetaan publik === */
function berita(b: Json): News {
  return { id: b.id, slug: b.slug, judul: b.judul, isi: b.isi ?? b.ringkasan ?? '', tanggalPublikasi: b.tanggalPublikasi, gambar: keKey(b.gambar), version: 1, createdAt: '' };
}
function dokumentasi(d: Json): Album {
  const gambar = Array.isArray(d.gambar)
    ? d.gambar.map((g: Json, i: number) => ({ key: keKey(g.url ?? g.key), urutan: i + 1 }))
    : gambarRingkas(keKeyAtauNull(d.gambarUtama), d.jumlahGambar ?? (d.gambarUtama ? 1 : 0));
  return { id: d.id, slug: d.slug, judul: d.judul, tanggalPublikasi: d.tanggalPublikasi, gambar, version: 1, createdAt: '' };
}
function halaman<T>(d: Json, items: T[]): Page<T> {
  return { items, page: d.page, limit: d.limit, total: d.total };
}

/* === Pemetaan admin === */
function dasar(sumber: Entity, b: Json): Item {
  versiTerakhir.set(`${sumber}:${b.id}`, b.version);
  return {
    id: b.id, version: b.version, createdAt: b.dibuatPada, createdBy: b.dibuatOleh, updatedBy: b.diubahOleh, dikunciOleh: b.dikunciOleh ?? null,
  };
}
function dariAdmin(sumber: Entity, b: Json): Item {
  const out = dasar(sumber, b);
  if (BERSTATUS.includes(sumber)) {
    out.status = b.status;
    out.deletedAt = b.dihapusPada ?? null;
  }
  if (sumber === 'berita') Object.assign(out, { slug: b.slug, judul: b.judul, isi: b.isi, tanggalPublikasi: b.tanggalPublikasi, gambar: b.gambar?.key ?? '' });
  else if (sumber === 'dokumentasi') {
    Object.assign(out, { slug: b.slug, judul: b.judul, tanggalPublikasi: b.tanggalPublikasi });
    out.gambar = Array.isArray(b.gambar)
      ? b.gambar.map((g: Json, i: number) => ({ key: g.key, urutan: i + 1 }))
      : gambarRingkas(b.gambarUtama?.key ?? null, b.jumlahGambar ?? 0);
  } else if (sumber === 'testimoni') Object.assign(out, { nama: b.nama, teks: b.teks, posisi: b.posisi, foto: b.foto?.key ?? null });
  else Object.assign(out, { nama: b.nama, posisi: b.posisi, bidangRiset: b.bidangRiset, foto: b.foto?.key ?? '', urutan: b.urutan });
  return out;
}

/** Unggah gambar baru (URL blob: dari pemilih berkas) dan kembalikan key; key yang sudah ada dikembalikan apa adanya. */
async function pastikanKey(sumber: Entity, nilai: unknown): Promise<string | null> {
  if (nilai === null || nilai === undefined || nilai === '') return null;
  const s = String(nilai);
  if (!s.startsWith('blob:')) return s;
  const ada = kunciUnggah.get(s);
  if (ada) return ada;
  let berkas: Blob;
  try {
    berkas = await (await fetch(s)).blob();
  } catch {
    throw new ApiError(400, 'invalid_request', 'Gambar tidak ditemukan. Pilih ulang gambar tersebut.');
  }
  const form = new FormData();
  form.append('sumber', sumber);
  form.append('gambar', berkas, 'gambar');
  const hasil = await minta('POST', '/api/admin/upload', { form });
  kunciUnggah.set(s, hasil.key);
  return hasil.key;
}

/** Badan buat/ubah menurut kontrak backend (version ditambahkan pemanggil). */
async function keAdmin(sumber: Entity, d: Payload): Promise<Json> {
  if (sumber === 'berita') return { judul: d.judul, isi: d.isi, tanggalPublikasi: d.tanggalPublikasi, gambar: await pastikanKey(sumber, d.gambar) };
  if (sumber === 'dokumentasi') {
    const gambarDaftar: { key: string | null }[] = [];
    for (const g of d.gambarDaftar ?? []) gambarDaftar.push({ key: await pastikanKey(sumber, g.key) });
    return { judul: d.judul, tanggalPublikasi: d.tanggalPublikasi, gambarDaftar };
  }
  if (sumber === 'testimoni') return { nama: d.nama, teks: d.teks, posisi: d.posisi, foto: await pastikanKey(sumber, d.foto) };
  return { nama: d.nama, posisi: d.posisi, bidangRiset: d.bidangRiset, foto: await pastikanKey(sumber, d.foto), urutan: Number(d.urutan) };
}

function url(sumber: Entity): string {
  if (!MODUL.includes(sumber)) throw new ApiError(404, 'not_found', 'Data tidak ditemukan.');
  return `/api/admin/${sumber}`;
}

/** Version terakhir yang dilihat halaman; bila belum ada, ambil dari server. */
async function versiDari(sumber: Entity, id: number): Promise<number> {
  const v = versiTerakhir.get(`${sumber}:${id}`);
  if (v !== undefined) return v;
  const b = await minta('GET', `${url(sumber)}/${id}`);
  versiTerakhir.set(`${sumber}:${id}`, b.version);
  return b.version;
}

export const api = {
  // === Publik ===
  daftarBerita: async ({ q, page, limit }: { q?: string; page?: number; limit?: number } = {}): Promise<Page<News>> => {
    const d = await minta('GET', `/api/berita${kueri({ q, page, limit })}`);
    return halaman(d, d.items.map(berita));
  },
  daftarDokumentasi: async ({ q, page, limit }: { q?: string; page?: number; limit?: number } = {}): Promise<Page<Album>> => {
    const d = await minta('GET', `/api/dokumentasi${kueri({ q, page, limit })}`);
    return halaman(d, d.items.map(dokumentasi));
  },
  detailBerita: async (slug: string): Promise<News> => berita(await minta('GET', `/api/berita/${encodeURIComponent(slug)}`)),
  detailDokumentasi: async (slug: string): Promise<Album> => dokumentasi(await minta('GET', `/api/dokumentasi/${encodeURIComponent(slug)}`)),
  daftarTestimoni: async (): Promise<{ items: Testimonial[] }> => {
    const d = await minta('GET', '/api/testimoni');
    return { items: d.items.map((t: Json) => ({ id: t.id, nama: t.nama, posisi: t.posisi, teks: t.teks, foto: keKeyAtauNull(t.foto), version: 1, createdAt: '' })) };
  },
  daftarTim: async (): Promise<{ items: Member[] }> => {
    const d = await minta('GET', '/api/tim');
    return { items: d.items.map((t: Json, i: number) => ({ id: t.id, nama: t.nama, posisi: t.posisi, bidangRiset: t.bidangRiset, foto: keKey(t.foto), urutan: t.urutan ?? i + 1, version: 1, createdAt: '' })) };
  },

  // === Sesi ===
  /** Muat sesi dari server (mengisi token CSRF). Dipanggil sekali sebelum admin dirender. */
  async siapkan(): Promise<Account | null> {
    try {
      const s = await minta('GET', '/api/auth/me');
      csrf = s.csrfToken;
      akunAktif = { id: s.username, sub: s.username, username: s.username, nama: s.nama, role: s.role || null };
    } catch (g) {
      if (!(g instanceof ApiError && g.status === 401)) throw g;
      akunAktif = null;
    }
    return akunAktif;
  },
  masuk(_idAkun?: string): Account | null {
    return akunAktif ? { ...akunAktif } : null;
  },
  akun: (): Account | null => (akunAktif ? { ...akunAktif } : null),
  async keluar(): Promise<void> {
    try {
      await minta('POST', '/api/auth/logout');
    } catch {
      /* sudah keluar atau sesi hilang: tujuan tercapai */
    }
    csrf = null;
    akunAktif = null;
    kunciSaya = null;
  },

  // === Admin ===
  adminLog: async (opsi: { page?: number; username?: string; jenis?: string; dari?: string; sampai?: string } = {}): Promise<Page<Log> & { usernames: string[] }> => {
    const d = await minta('GET', `/api/admin/log-aktivitas${kueri(opsi)}`);
    return { ...halaman(d, d.items as Log[]), usernames: d.usernames };
  },
  adminDaftar: async (sumber: Entity, { trash = false } = {}): Promise<{ items: Item[] }> => {
    const d = await minta('GET', `${url(sumber)}${BERSTATUS.includes(sumber) ? kueri({ status: trash ? 'trash' : 'semua' }) : ''}`);
    return { items: d.items.map((b: Json) => dariAdmin(sumber, b)) };
  },
  adminDetail: async (sumber: Entity, id: number): Promise<Item> => dariAdmin(sumber, await minta('GET', `${url(sumber)}/${id}`)),
  adminTambah: async (sumber: Entity, data: Payload): Promise<Item> => dariAdmin(sumber, await minta('POST', url(sumber), { json: await keAdmin(sumber, data) })),
  adminUbah: async (sumber: Entity, id: number, data: Payload, version: number): Promise<Item> =>
    dariAdmin(sumber, await minta('PUT', `${url(sumber)}/${id}`, { json: { ...(await keAdmin(sumber, data)), version } })),
  adminPost: async (sumber: Entity, id: number): Promise<Item> =>
    dariAdmin(sumber, await minta('POST', `${url(sumber)}/${id}/post`, { json: { version: await versiDari(sumber, id) } })),
  adminRetract: async (sumber: Entity, id: number): Promise<Item> =>
    dariAdmin(sumber, await minta('POST', `${url(sumber)}/${id}/retract`, { json: { version: await versiDari(sumber, id) } })),
  adminHapus: async (sumber: Entity, id: number): Promise<Record<string, never>> => {
    await minta('DELETE', `${url(sumber)}/${id}`);
    return {};
  },
  adminPulihkan: async (sumber: Entity, id: number): Promise<Item> => dariAdmin(sumber, await minta('POST', `${url(sumber)}/${id}/restore`)),
  adminHapusPermanen: async (sumber: Entity, id: number): Promise<Record<string, never>> => {
    await minta('DELETE', `${url(sumber)}/${id}/permanen`);
    return {};
  },

  // === Soft lock ===
  kunciAmbil: async (sumber: Entity, id: number): Promise<{ milik: 'saya' | 'lain'; nama?: string }> => {
    try {
      await minta('POST', `${url(sumber)}/${id}/lock`);
      kunciSaya = { sumber, id };
      return { milik: 'saya' };
    } catch (g) {
      if (g instanceof ApiError && g.status === 423 && g.error.code === 'locked') {
        return { milik: 'lain', nama: (g as ApiError & { pemegang?: { nama: string } }).pemegang?.nama ?? 'pengguna lain' };
      }
      throw g;
    }
  },
  kunciHeartbeat: async (): Promise<{ ok: true }> => {
    if (!kunciSaya) throw new ApiError(410, 'lock_lost', 'Sesi edit berakhir.');
    await minta('PUT', `${url(kunciSaya.sumber)}/${kunciSaya.id}/lock`);
    return { ok: true };
  },
  /** Lepas kunci yang dipegang (best effort, tanpa menunggu). */
  kunciLepas: (): void => {
    if (!kunciSaya) return;
    const { sumber, id } = kunciSaya;
    kunciSaya = null;
    minta('DELETE', `${url(sumber)}/${id}/lock`, { keepalive: true }).catch(() => {});
  },
};
