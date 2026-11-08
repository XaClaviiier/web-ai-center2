# Database — AI Center UBAYA CMS

Dokumen aktif. Riwayat perencanaan awal (`00-detail-rencana-web.md`, `01-kontrak-api.md`) tersedia di riwayat Git, commit `5928f05` (tag `arsip-docs-v1`). Skema lama (`D:\aicenter\notes\ext-database.md`, di luar repo) hanya dipakai untuk detail yang sudah diverifikasi terhadap plan.

## 1. Ruang lingkup

* Engine: PostgreSQL \[Lihat stack.md §3\]. Versi PostgreSQL: \[diisi: tanya infra\].  
* Entitas: Berita, Dokumentasi (dengan tabel gambar), Testimoni, Tim.  
* Tidak ada tabel pengguna; identitas dan role dari Keycloak \[Lihat §3.6\].  
* Tidak dibawa ke skema baru: `kursus`, `progress_pembelajaran`, dan tabel legacy `layanan`, `hero_slides`, `klien`, `profil`, `inference` (kontennya hardcode di frontend; \[Lihat requirement.md §2 dan §3\]).

## 2. Konvensi

* Nama kolom: snake_case, memakai istilah plan (mis. `tanggal_publikasi`, `posisi`, `teks`, `bidang_riset`).  
* Nama field JSON di API: camelCase bahasa Indonesia (mis. `tanggalPublikasi`, `dibuatPada`). Field opsional yang kosong tidak dikirim.  
* `id`: `bigint generated always as identity`.  
* `status`: `text` dengan `CHECK (status IN ('draft','posted'))`.  
* Timestamp sistem: `timestamptz`; di API berformat ISO 8601 UTC.  
* `tanggal_publikasi`: tipe `date`; di API dikirim dan diterima sebagai `DD-MM-YYYY`; format lain ditolak `400`. Tampilan di UI juga `DD-MM-YYYY`. Tanggal harus ada di kalender dan tidak boleh setelah hari ini; backend menolaknya dengan `400`. Zona waktu acuan yang menentukan "hari ini" \[terbuka\]: tanggal lokal perangkat bisa berbeda dari tanggal server. Rancangan ini tidak mendukung posting terjadwal. Item lama yang sudah bertanggal masa depan tetap tampil apa adanya, tetapi tidak bisa disimpan sampai tanggalnya diperbaiki; aturan untuk data serupa saat migrasi \[terbuka\].  
* `judul` Berita dan `judul` kegiatan Dokumentasi: maksimal 70 karakter. Panjang dihitung sebagai `String.length` pada nilai yang sudah di-trim (spasi dihitung, emoji dihitung 2), sama dengan atribut `maxlength` HTML. Judul lama yang melebihi 70 tetap boleh disimpan selama judulnya tidak diubah. \[Belum diterapkan: kode `src/` dan `server/` masih memakai batas 200. Sebelum diterapkan, periksa judul live yang lebih dari 70 karakter.\] Batas ini tidak berlaku untuk judul modul kursus, hero, inference, nama layanan, atau nama tim.

## 3. Entitas

Kolom bertanda *(diturunkan)* tidak disebut eksplisit di rencana awal, tetapi dibutuhkan oleh aturannya.

### 3.1 Berita

| Kolom | Keterangan |
|---|---|
| `id` | PK |
| `slug` | unik; aturan pembuatan \[Lihat §3.3\] |
| `judul` | wajib |
| `isi` | wajib; teks biasa |
| `tanggal_publikasi` | wajib; `date` |
| `gambar` | wajib; nama objek (key) RustFS \[Lihat §5\] |
| `status` | `draft` / `posted` |
| `version` | optimistic locking \[Lihat §4\] |
| `deleted_at` | null; waktu masuk trash *(diturunkan: aturan 30 hari, requirement.md §5.1)* |
| `created_at` | *(diturunkan: pemutus seri urutan, requirement.md §4.5)* |
| `created_by`, `updated_by` | `sub` dari session \[Lihat §3.6\] |

**Data Fetching dan Pagination**

\[Lihat stack.md §9\]

**Data dan Validasi**

Semua field wajib terisi sejak penyimpanan pertama (hasilnya draft), termasuk 1 gambar. Nama penulis: Tim Media Ubaya AI Center (statis, hanya ditampilkan untuk berita, tidak disimpan di database); `created_by`/`updated_by` tetap dicatat dari session.

Upload divalidasi: tipe dicek dari isi file, ukuran maksimum 2 MB untuk gambar. \[RP-10\] Teks (`isi` disimpan sebagai teks biasa) \[Lihat security.md §7\].

**Format isi:** Paragraf dipisah baris kosong (`\n\n`); potongan pada kartu \= paragraf pertama, dipotong lalu diberi keterangan 'lihat detail'.

### 3.2 Dokumentasi

| Kolom | Keterangan |
|---|---|
| `id` | PK |
| `slug` | unik; aturan pembuatan \[Lihat §3.3\] |
| `judul` | wajib (judul kegiatan) |
| `tanggal_publikasi` | wajib; `date` |
| `status` | `draft` / `posted` |
| `version` | \[Lihat §4\] |
| `deleted_at` | null; waktu masuk trash *(diturunkan: aturan 30 hari, requirement.md §5.2)* |
| `created_at` | *(diturunkan: pemutus seri urutan, requirement.md §4.4)* |
| `created_by`, `updated_by` | `sub` dari session \[Lihat §3.6\] |

Tabel gambar dokumentasi `dokumentasi_gambar`:

| Kolom | Keterangan |
|---|---|
| `id` | PK |
| `dokumentasi_id` | FK ke `dokumentasi.id` |
| `key` | nama objek RustFS \[Lihat §5\] |
| `urutan` | 1..n; urutan 1 \= headline |
| keterangan / alt gambar | \[terbuka\] |

**Data Fetching dan Pagination**

\[Lihat stack.md §9\]

**Data dan Validasi**

Semua field wajib terisi sejak penyimpanan pertama (hasilnya draft). Jumlah gambar minimal 1 dan maksimal 6\.

Upload divalidasi: tipe dicek dari isi file, ukuran maksimum 2 MB tiap file gambar. \[RP-10\] \[Lihat security.md §7\]

### 3.3 SEO dan slug (Berita, Dokumentasi)

SEO berita: gambar berita menjadi OG image; judul SEO dan deskripsi SEO selalu memakai nilai default (judul, dan 150-160 karakter pertama isi). \[RP-11\]  
SEO dokumentasi: gambar headline (urutan 1) menjadi OG image; judul SEO dan deskripsi SEO selalu memakai nilai default (judul kegiatan, dan "Dokumentasi kegiatan {judul} — {tanggal publikasi}"). \[RP-11\]  
Tidak ada field SEO di form admin dan tidak ada kolom SEO di database (sebagian keputusan M-03 dibatalkan).  
Cara pembuatan slug: \[terbuka\] (usulan: otomatis dari judul, unik, dan tidak diubah setelah Post agar URL tidak rusak).

Injeksi meta SEO oleh Go: \[Lihat stack.md §3\].

### 3.4 Testimoni

| Kolom | Keterangan |
|---|---|
| `id` | PK |
| `nama` | wajib |
| `teks` | wajib |
| `posisi` | wajib |
| `foto` | null; nama objek RustFS |
| `status` | `draft` / `posted` |
| `version` | \[Lihat §4\] |
| `deleted_at` | null; waktu masuk trash *(diturunkan: aturan 30 hari, requirement.md §5.3)* |
| `created_at` | urutan tampil: terbaru di atas \[Lihat §8\] |
| `created_by`, `updated_by` | `sub` dari session \[Lihat §3.6\] |

**Data dan Validasi**

Field wajib: nama pemberi testimoni, teks testimoni, posisi. Foto opsional (maksimal 1); tanpa foto tampil inisial nama. Upload divalidasi sama seperti berita (tipe dicek dari isi file, maksimal 2 MB). \[Lihat security.md §7\]

### 3.5 Tim

| Kolom | Keterangan |
|---|---|
| `id` | PK |
| `nama` | wajib |
| `posisi` | wajib (posisi / jabatan) |
| `bidang_riset` | wajib |
| `foto` | wajib; nama objek RustFS |
| `urutan` | urutan tampil manual \[Lihat §8\] |
| `version` | \[Lihat §4\] |
| `created_at` | *(diturunkan)* |
| `created_by`, `updated_by` | `sub` dari session \[Lihat §3.6\] |

**Data dan Validasi**

Field wajib: foto, posisi / jabatan, nama, bidang riset. Upload divalidasi sama seperti berita.

### 3.6 Identitas pengguna dan audit

`created_by` / `updated_by` dicatat dari identitas (`sub`) di session sebagai jejak audit.

Role dan login: \[Lihat security.md §1\].

## 4. Race Condition (Soft Locking)

Saat pengguna membuka form edit, backend mengunci konten tersebut (atas nama pengguna itu). Pengguna lain hanya bisa membaca dan melihat info siapa yang sedang mengedit.

Kunci punya TTL, diperpanjang otomatis (heartbeat), dan dilepas saat data disimpan atau edit dibatalkan.

**Penerapan kunci di Redis**

Key `lock:{jenis}:{id}` berisi pengguna pemegang kunci, dibuat hanya jika belum ada (`SET NX`). Hanya pemegang yang dapat memperpanjang (heartbeat) atau melepas. Kunci dilepas saat Simpan atau batal. TTL 5 menit, heartbeat tiap 1 menit. Heartbeat dihitung ke bucket tulis admin rate limiter (30 / menit per user ID) \[Lihat security.md §6\].

**Kapan kunci dipasang**

Edit hanya untuk data berstatus draft (konten yang sudah terbit di-Retract dulu). Saat form edit dibuka, data dikunci (soft lock). Aturan ini berlaku untuk Dokumentasi, Berita, dan Testimoni. Tim memakai soft lock tanpa aturan draft \[Lihat database.md §6\].

**Pengaman tambahan: optimistic locking**

Kolom `version` dicek saat Simpan. Semua perubahan diterapkan dalam satu request dan satu transaksi bersama pengecekan `version`, lalu `version` dinaikkan. Jika versi tidak cocok, backend mengembalikan `409 Conflict`. `version` naik 1 setiap Simpan, Post, dan Retract.

**Perilaku di antarmuka admin**

* Item yang dikunci pengguna lain: tombol Hapus dan Post nonaktif; Edit berganti "Lihat" untuk mode baca saja.  
* Sesi edit berakhir (heartbeat gagal): Simpan nonaktif, isian tetap ada di form, dan pengguna diminta menyalin lalu memuat ulang. Batas waktu pengguna diam \[terbuka\].  
* Konten posted harus di-Retract sebelum diedit \[terbuka, perlu didiskusikan sebelum backend karena berdampak ke API; lihat requirement.md §5\].

## 5. Penyimpanan gambar (RustFS)

### 5.1 Dokumentasi

File di RustFS memakai nama unik yang tidak berubah: `dokumentasi/{id}/{uuid}.ext`, mengikuti pola Berita, Testimoni, dan Tim. \[RP-11\]

**Tambah dan Edit**

Tambah: mengunggah 1-6 gambar; urutan upload menjadi `urutan`, gambar pertama menjadi headline.

Edit hanya untuk data berstatus draft (konten yang sudah terbit di-Retract dulu). \[Lihat database.md §4\]

Operasi gambar saat edit:

* Replace: satu gambar diganti dan menempati `urutan` yang sama.  
* Hapus: satu atau beberapa gambar dihapus; `urutan` dirapikan menjadi 1..n (gambar berikutnya menjadi headline bila headline dihapus).  
* Tambah: gambar baru di urutan paling akhir.

Jumlah akhir harus tetap 1-6. Tanpa operasi gambar, gambar lama tetap dipakai.

Penguncian saat form edit dibuka dan pengecekan `version` saat Simpan (termasuk `409 Conflict`): \[Lihat database.md §4\]

Urutan eksekusi: yang masuk, file dulu baru database; yang keluar, database dulu baru file.

1. Upload file baru ke RustFS.  
2. Transaksi database: hapus baris gambar lama, tambah baris baru, rapikan `urutan`, naikkan `version`.  
3. Setelah commit, hapus file lama di RustFS.

Jika gagal sebelum commit, file baru yang sudah terunggah dihapus kembali. File yatim dibersihkan job berkala (objek yang tidak ada di tabel gambar dan lebih tua dari 24 jam \[usulan; diputuskan pada fase backend\]).

Retract dan Hapus (soft delete) tidak menyentuh gambar. Hapus permanen (manual atau otomatis setelah 30 hari) menghapus baris dan semua file gambarnya.

### 5.2 Berita

Satu berita memiliki tepat 1 gambar, disimpan sebagai nama objek di kolom `gambar` pada tabel `berita` (tanpa tabel tambahan dan tanpa `urutan`).

File di RustFS memakai nama unik yang tidak berubah: `berita/{id}/{uuid}.ext`. \[RP-11\]

**Tambah dan Edit**

Tambah: mengunggah 1 gambar.

Edit hanya untuk data berstatus draft. \[Lihat database.md §4\] Satu-satunya operasi gambar adalah replace; gambar tidak dapat dihapus atau ditambah. Tanpa gambar baru, gambar lama tetap dipakai.

Penguncian saat form edit dibuka dan pengecekan `version` saat Simpan (termasuk `409 Conflict`): \[Lihat database.md §4\]

Urutan eksekusi: yang masuk, file dulu baru database; yang keluar, database dulu baru file.

1. Upload gambar baru ke RustFS.  
2. Update database: `gambar` diganti nama objek baru, `version` dinaikkan, `updated_by` diisi.  
3. Setelah commit, hapus file lama di RustFS.

Jika gagal sebelum commit, file baru yang sudah terunggah dihapus kembali. File yatim dibersihkan oleh job berkala yang sama dengan dokumentasi.

Retract dan Hapus (soft delete) tidak menyentuh gambar. Hapus permanen (manual atau otomatis setelah 30 hari) menghapus baris dan file gambarnya.

### 5.3 Testimoni

Foto disimpan sebagai nama objek pada kolom `foto` (boleh kosong). File di RustFS: `testimoni/{id}/{uuid}.ext`. Urutan eksekusi, pembersihan file yatim, dan hapus permanen sama dengan berita.

### 5.4 Tim

Foto disimpan sebagai nama objek pada kolom `foto`. File di RustFS: `tim/{id}/{uuid}.ext`. Urutan eksekusi sama dengan berita.

## 6. Status dan transisi

Berita dan Dokumentasi: alur draft, post, retract, trash 30 hari, pulihkan, dan hapus permanen \[Lihat requirement.md §5.1 dan §5.2\].

**Testimoni:** Alur draft, post, retract, trash 30 hari sama dengan berita. Edit hanya untuk draft dengan soft lock dan pengecekan `version`. \[Lihat database.md §4\]

**Tim:** Tim tidak memiliki status draft/posted dan tidak memiliki trash. Saat form edit dibuka, data dikunci (soft lock) dan `version` dicek saat Simpan. Hapus langsung permanen (dengan konfirmasi) dan menghapus file fotonya.

## 7. Cache konten

Key Redis memakai prefix `cache:` \[Lihat stack.md §6\]. Semua TTL disimpan sebagai environment variable.

**Dokumentasi dan Berita:** Daftar dan detail di-cache di Redis (cache-aside \+ TTL 20 menit) dan di-invalidate saat Post dan Retract. Draft dan item terhapus tidak pernah masuk cache publik.

**Testimoni:** Daftar testimoni posted di-cache di Redis (cache-aside \+ TTL 20 menit) dan di-invalidate saat Post dan Retract.

**Tim:** Daftar tim di-cache di Redis (cache-aside \+ TTL 20 menit) dan di-invalidate saat tambah, edit, dan hapus.

**Cache server:** Pola cache-aside \+ TTL untuk daftar dan detail berita/dokumentasi; aturan invalidasi (Post dan Retract) \[Lihat database.md §7\]. \[RP-04\]

## 8. Urutan tampil

* Tim: urutan manual lewat kolom `urutan`.  
* Testimoni: terbaru di atas berdasarkan `created_at`.  
* Berita dan Dokumentasi: \[Lihat requirement.md §4.4 dan §4.5\].

## 9. Index

\[diisi\]
