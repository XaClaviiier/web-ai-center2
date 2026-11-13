/* TIRUAN API. Dibuang saat integrasi backend. Semua data di memori dan kembali ke awal saat refresh. */
import seed from '../shared/seed.json';
import { Format } from '../shared/format';
import { SIMULASI } from './state';
import { ApiError, type Account, type Album, type Entity, type Image, type Item, type Log, type Member, type News, type Page, type Payload, type Testimonial } from '../shared/types';

interface Db { berita: Item[]; dokumentasi: Item[]; dokumentasiGambar: Image[]; testimoni: Item[]; tim: Item[]; akun: Account[]; kunci: { jenis: string; id: number; nama: string }[] }
type Aksi = 'adminTambah' | 'adminUbah' | 'adminPost' | 'adminRetract' | 'adminHapus' | 'adminPulihkan' | 'adminHapusPermanen';

const salin = <T>(nilai: T): T => structuredClone(nilai);
const awal = (): Db => salin(seed) as unknown as Db;

let db = awal();
let akunAktif: Account | null = null;
let kunciSaya: { sumber: Entity; id: number } | null = null;
let logAktivitas: Log[] = [];
let nomorLog = 0;

const BERSTATUS: Entity[] = ['berita', 'dokumentasi', 'testimoni'];
const MODUL: Entity[] = ['berita', 'dokumentasi', 'testimoni', 'tim'];
const BATAS_HALAMAN = 9;
const AKSI: Record<Aksi, string> = { adminTambah: 'tambah', adminUbah: 'ubah', adminPost: 'post', adminRetract: 'retract', adminHapus: 'hapus', adminPulihkan: 'restore', adminHapusPermanen: 'hapus_permanen' };

const bukan = () => new ApiError(404, 'not_found', 'Data tidak ditemukan.');
const bersih = (teks: unknown) => String(teks ?? '').replace(/[\u0000-\u0008\u000B\u000C\u000E-\u001F\u007F]/g, '').trim();

function cekAkun(): Account {
  if (!akunAktif) throw new ApiError(401, 'unauthorized', 'Silakan masuk kembali.');
  if (!akunAktif.role) throw new ApiError(403, 'forbidden', 'Anda tidak punya akses.');
  return akunAktif;
}

function catat(nama: Aksi, sumber: Entity, baris: Item): void {
  const akun = cekAkun();
  const jenis = nama === 'adminTambah' ? 'tambah' : nama === 'adminHapus' || nama === 'adminHapusPermanen' ? 'hapus' : 'ubah';
  logAktivitas.push({ id: ++nomorLog, terjadiPada: new Date().toISOString(), penggunaSub: akun.sub, username: akun.username, jenis, aksi: AKSI[nama], modul: sumber, itemId: baris.id, labelItem: bersih(baris.judul ?? baris.nama ?? '').slice(0, 140) });
}

// Cermin validasi server: judul berita/dokumentasi maksimal 70 karakter.
function cekJudul(sumber: Entity, data: Payload): void {
  if ((sumber === 'berita' || sumber === 'dokumentasi') && String(data.judul ?? '').trim().length > 70) throw new ApiError(400, 'validation_error', 'Field "judul" maksimal 70 karakter.');
}

/** Bungkus jeda dan keadaan panel simulasi. */
function jawab<T>(fungsi: () => T, { admin = false } = {}): Promise<T> {
  const { promise, resolve, reject } = Promise.withResolvers<T>();
  if (SIMULASI.keadaan === 'loading') return promise; // sengaja tidak pernah selesai
  setTimeout(() => {
    try {
      if (admin && SIMULASI.paksa401) throw new ApiError(401, 'unauthorized', 'Sesi Anda sudah berakhir.');
      if (SIMULASI.keadaan === 'error') throw new ApiError(500, 'server_error', 'Terjadi kendala di server. Coba lagi beberapa saat lagi.');
      if (SIMULASI.keadaan === '429') throw new ApiError(429, 'too_many_requests', 'Terlalu banyak permintaan.', 30);
      if (admin) cekAkun();
      resolve(salin(fungsi()));
    } catch (e) {
      reject(e);
    }
  }, SIMULASI.jedaMs);
  return promise;
}

const hidup = (b: Item) => !b.deletedAt;
const terbit = (b: Item) => hidup(b) && b.status === 'posted';
const kosongkan = <T>(daftar: T[]): T[] => (SIMULASI.keadaan === 'kosong' ? [] : daftar);
const lengkapiGambar = (dok: Item): Item => ({ ...dok, gambar: db.dokumentasiGambar.filter((g) => g.dokumentasiId === dok.id).sort((a, b) => a.urutan - b.urutan) });
const terbaruDibuat = (a: Item, b: Item) => Date.parse(b.createdAt) - Date.parse(a.createdAt);
const urutanTim = (a: Item, b: Item) => (a.urutan ?? 0) - (b.urutan ?? 0) || a.id - b.id;
const idBaru = (daftar: { id?: number }[]) => Math.max(0, ...daftar.map((b) => b.id ?? 0)) + 1;
const pengguna = () => akunAktif?.sub ?? 'sub-tidak-dikenal';

function halaman<T>(daftar: T[], page = 1, limit = BATAS_HALAMAN): Page<T> {
  return { items: daftar.slice((page - 1) * limit, page * limit), page, limit, total: daftar.length };
}

type OpsiDaftar = { q?: string; page?: number; limit?: number };
/** Daftar publik: pencarian judul berlaku minimal 2 karakter. */
function daftarPublik(tabel: 'berita' | 'dokumentasi', { q = '', page, limit }: OpsiDaftar): Page<Item> {
  let hasil = kosongkan(db[tabel].filter(terbit)).sort(Format.bandingkanTerbaru);
  const kata = q.trim().toLowerCase();
  if (kata.length >= 2) hasil = hasil.filter((b) => (b.judul ?? '').toLowerCase().includes(kata));
  if (tabel === 'dokumentasi') hasil = hasil.map(lengkapiGambar);
  return halaman(hasil, page, limit);
}

function cari(sumber: Entity, id: number): Item {
  if (!MODUL.includes(sumber)) throw bukan();
  const baris = db[sumber].find((b) => b.id === id);
  if (!baris) throw bukan();
  return baris;
}

const kunciLain = (sumber: Entity, id: number) => (SIMULASI.kunciLain ? db.kunci.find((k) => k.jenis === sumber && k.id === id) : undefined);
function cekTerkunci(sumber: Entity, id: number): void {
  const k = kunciLain(sumber, id);
  if (k) throw new ApiError(423, 'locked', `Sedang diedit oleh ${k.nama}.`);
}

function slugUnik(judul: string, tabel: 'berita' | 'dokumentasi'): string {
  // Simulasi. Aturan slug final masih [terbuka] di database.md §3.3.
  const dasar = judul.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '') || 'konten';
  let slug = dasar;
  for (let n = 2; db[tabel].some((b) => b.slug === slug); n++) slug = `${dasar}-${n}`;
  return slug;
}

function simpanGambarDokumentasi(id: number, gambar: Image[]): void {
  db.dokumentasiGambar = db.dokumentasiGambar.filter((g) => g.dokumentasiId !== id);
  gambar.forEach((g, i) => db.dokumentasiGambar.push({ id: idBaru(db.dokumentasiGambar), dokumentasiId: id, key: g.key, urutan: i + 1 }));
}

function terapkan(baris: Item, data: Payload): void {
  const tujuan = baris as unknown as Record<string, unknown>;
  for (const [k, v] of Object.entries(data)) if (k !== 'gambarDaftar') tujuan[k] = typeof v === 'string' ? bersih(v) : v;
}

const tandaiKunci = (sumber: Entity) => (b: Item): Item => ({ ...b, dikunciOleh: kunciLain(sumber, b.id)?.nama ?? null });

export const api = {
  // === Publik ===
  daftarBerita: (opsi: OpsiDaftar = {}) => jawab(() => daftarPublik('berita', opsi) as Page<News>),
  daftarDokumentasi: (opsi: OpsiDaftar = {}) => jawab(() => daftarPublik('dokumentasi', opsi) as Page<Album>),
  detailBerita: (slug: string) => jawab(() => {
    const b = db.berita.find((x) => x.slug === slug && terbit(x));
    if (!b || SIMULASI.keadaan === 'kosong') throw bukan();
    return b as News;
  }),
  detailDokumentasi: (slug: string) => jawab(() => {
    const d = db.dokumentasi.find((x) => x.slug === slug && terbit(x));
    if (!d || SIMULASI.keadaan === 'kosong') throw bukan();
    return lengkapiGambar(d) as Album;
  }),
  daftarTestimoni: () => jawab(() => ({ items: kosongkan(db.testimoni.filter(terbit)).sort(terbaruDibuat) as Testimonial[] })),
  daftarTim: () => jawab(() => ({ items: kosongkan([...db.tim]).sort(urutanTim) as Member[] })),

  // === Sesi (simulasi pemilihan akun, tanpa cookie) ===
  masuk(idAkun: string): Account | null {
    akunAktif = db.akun.find((a) => a.id === idAkun) ?? null;
    return akunAktif ? { ...akunAktif } : null;
  },
  akun: (): Account | null => (akunAktif ? { ...akunAktif } : null),
  /** Keluar mengembalikan data, akun, dan log ke keadaan awal (setara refresh prototype). */
  keluar(): void {
    db = awal();
    akunAktif = null;
    kunciSaya = null;
    logAktivitas = [];
    nomorLog = 0;
  },

  adminLog: (opsi: { page?: number; username?: string; jenis?: string; dari?: string; sampai?: string } = {}) => jawab(() => {
    const { username = '', jenis = '', dari = '', sampai = '' } = opsi;
    if (dari && sampai && dari > sampai) throw new ApiError(400, 'validation_error', 'Tanggal awal harus sebelum atau sama dengan tanggal akhir.');
    const hasil = kosongkan(logAktivitas).filter((b) => {
      const hari = Format.hariJakarta(b.terjadiPada);
      return (!username || b.username === username) && (!jenis || b.jenis === jenis) && (!dari || hari >= dari) && (!sampai || hari <= sampai);
    }).sort((a, b) => Date.parse(b.terjadiPada) - Date.parse(a.terjadiPada) || b.id - a.id);
    const limit = 10;
    const page = Math.min(Math.max(1, Math.floor(Number(opsi.page)) || 1), Math.max(1, Math.ceil(hasil.length / limit)));
    return { ...halaman(hasil, page, limit), usernames: [...new Set(logAktivitas.map((b) => b.username))].sort() };
  }, { admin: true }),

  // === Admin ===
  adminDaftar: (sumber: Entity, { trash = false } = {}) => jawab(() => {
    if (sumber === 'tim') return { items: kosongkan([...db.tim]).sort(urutanTim).map(tandaiKunci(sumber)) };
    let hasil = kosongkan(db[sumber].filter((b) => (trash ? !hidup(b) : hidup(b)))).sort(sumber === 'testimoni' ? terbaruDibuat : Format.bandingkanTerbaru);
    if (sumber === 'dokumentasi') hasil = hasil.map(lengkapiGambar);
    return { items: hasil.map(tandaiKunci(sumber)) };
  }, { admin: true }),

  adminDetail: (sumber: Entity, id: number) => jawab(() => {
    const b = cari(sumber, id);
    return sumber === 'dokumentasi' ? lengkapiGambar(b) : b;
  }, { admin: true }),

  adminTambah: (sumber: Entity, data: Payload) => jawab(() => {
    if (!MODUL.includes(sumber)) throw bukan();
    cekJudul(sumber, data);
    if (sumber === 'dokumentasi' && !Array.isArray(data.gambarDaftar)) throw new ApiError(400, 'validation_error', 'Daftar gambar wajib diisi.');
    const id = idBaru(db[sumber]);
    const baris: Item = { id, version: 1, createdAt: new Date().toISOString(), createdBy: pengguna(), updatedBy: pengguna() };
    terapkan(baris, data);
    if (BERSTATUS.includes(sumber)) { baris.status = 'draft'; baris.deletedAt = null; }
    if (sumber === 'berita' || sumber === 'dokumentasi') baris.slug = slugUnik(baris.judul ?? '', sumber);
    db[sumber].push(baris);
    if (sumber === 'dokumentasi') simpanGambarDokumentasi(id, data.gambarDaftar ?? []);
    catat('adminTambah', sumber, baris);
    return baris;
  }, { admin: true }),

  adminUbah: (sumber: Entity, id: number, data: Payload, version: number) => jawab(() => {
    const b = cari(sumber, id);
    if (BERSTATUS.includes(sumber) && b.status !== 'draft') throw new ApiError(400, 'bukan_draft', 'Hanya data berstatus draft yang bisa diedit.');
    if (data.judul !== b.judul) cekJudul(sumber, data); // judul lama >70 tetap bisa disimpan bila tidak diubah
    if (SIMULASI.paksa409) { b.version += 1; SIMULASI.paksa409 = false; }
    if (b.version !== version) throw new ApiError(409, 'conflict', 'Data sudah diubah oleh pengguna lain.');
    if (sumber === 'dokumentasi' && !Array.isArray(data.gambarDaftar)) throw new ApiError(400, 'validation_error', 'Daftar gambar wajib diisi.');
    terapkan(b, data);
    if (sumber === 'dokumentasi') simpanGambarDokumentasi(id, data.gambarDaftar ?? []);
    b.version += 1;
    b.updatedBy = pengguna();
    kunciSaya = null;
    catat('adminUbah', sumber, b);
    return b;
  }, { admin: true }),

  adminPost: (sumber: Entity, id: number) => jawab(() => { cekTerkunci(sumber, id); const b = cari(sumber, id); b.status = 'posted'; b.version += 1; catat('adminPost', sumber, b); return b; }, { admin: true }),
  adminRetract: (sumber: Entity, id: number) => jawab(() => { const b = cari(sumber, id); b.status = 'draft'; b.version += 1; catat('adminRetract', sumber, b); return b; }, { admin: true }),
  adminHapus: (sumber: Entity, id: number) => jawab((): Record<string, never> => {
    cekTerkunci(sumber, id);
    const b = cari(sumber, id);
    if (sumber === 'tim') db.tim = db.tim.filter((x) => x.id !== id);
    else b.deletedAt = new Date().toISOString();
    catat('adminHapus', sumber, b);
    return {};
  }, { admin: true }),
  adminPulihkan: (sumber: Entity, id: number) => jawab(() => { const b = cari(sumber, id); b.deletedAt = null; b.status = 'draft'; b.version += 1; catat('adminPulihkan', sumber, b); return b; }, { admin: true }),
  adminHapusPermanen: (sumber: Entity, id: number) => jawab((): Record<string, never> => {
    const b = cari(sumber, id);
    db[sumber] = db[sumber].filter((x) => x.id !== id);
    if (sumber === 'dokumentasi') db.dokumentasiGambar = db.dokumentasiGambar.filter((g) => g.dokumentasiId !== id);
    catat('adminHapusPermanen', sumber, b);
    return {};
  }, { admin: true }),

  // === Soft lock ===
  kunciAmbil: (sumber: Entity, id: number) => jawab((): { milik: 'saya' | 'lain'; nama?: string } => {
    const k = kunciLain(sumber, id);
    if (k) return { milik: 'lain', nama: k.nama };
    kunciSaya = { sumber, id };
    SIMULASI.heartbeatGagal = false;
    return { milik: 'saya' };
  }, { admin: true }),
  kunciHeartbeat: () => jawab(() => {
    if (SIMULASI.heartbeatGagal || !kunciSaya) { kunciSaya = null; throw new ApiError(410, 'lock_lost', 'Sesi edit berakhir.'); }
    return { ok: true };
  }, { admin: true }),
  kunciLepas: (): void => { kunciSaya = null; },
};
