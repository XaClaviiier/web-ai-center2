# Security — UBAYA AI Center CMS

Dokumen aktif. Riwayat perencanaan awal (`00-detail-rencana-web.md`, `01-kontrak-api.md`) tersedia di riwayat Git, commit `5928f05` (tag `arsip-docs-v1`). Peran pengguna: \[Lihat requirement.md §6\]. Audit `created_by` / `updated_by`: \[Lihat database.md §3.6\].

## 1. Akses dan login

Keycloak mengelola pengguna dan role; aplikasi tidak punya manajemen role sendiri. Semua pengguna yang berhasil login dapat menambah, mengubah, dan menghapus konten. Hanya pengguna dengan role `direktur` atau `pegawai` pada client CMS yang boleh login; lainnya `403`. \[menunggu konfirmasi Keycloak\]

**Login:** browser diarahkan ke Keycloak (OIDC Authorization Code \+ PKCE). Setelah berhasil, backend menukar kode dengan token di sisi server dan memvalidasi `state`, penerbit (issuer), tanda tangan, dan masa berlaku. \[menunggu konfirmasi Keycloak\]

Endpoint auth: \[Lihat stack.md §9.5\]. \[menunggu konfirmasi Keycloak\]

**Login development (sebelum Keycloak siap):** backend Go menyediakan login password yang mengikuti logika mekanisme login yang ada saat ini, tetapi hasilnya adalah cookie session yang sama dengan alur Keycloak (bukan Bearer token), sehingga frontend admin hanya punya satu cara autentikasi. Login ini hanya aktif bila `APP_ENV=development`; server menolak start bila login ini aktif di luar development. Saat Keycloak siap, hanya endpoint login yang diganti.

\[KONFLIK, belum diputuskan\] Catatan prototype menyebut keputusan terbaru bahwa aplikasi tidak memiliki halaman login sendiri, sedangkan paragraf Login development di atas mendefinisikan login password. Tentukan mana yang berlaku sebelum fase backend. Sampai itu jelas, paragraf tersebut tidak dihapus \[Lihat juga stack.md §9.5\].

## 2. Session

**Session login:** Key `session:{id}` berisi `sub`, nama, token CSRF, dan waktu dibuat. Cookie hanya berisi ID acak (256 bit). Dibuat setelah login Keycloak berhasil, dihapus saat logout. Pengguna yang dinonaktifkan di Keycloak tetap punya session sampai kedaluwarsa \[Diputuskan pada fase backend: validasi ulang ke Keycloak berkala\]. Idle 1 jam (diperpanjang tiap request), batas mutlak 8 jam; cookie dibuat berumur sama dengan batas mutlak.

**Session:** backend membuat session di Redis (detail pada paragraf Session login di atas). Browser hanya menerima ID session acak lewat cookie `HttpOnly`, `Secure`, `SameSite=Lax`, `Path=/` (nama `__Host-session`). Nama cookie dan flag `Secure` diatur lewat environment variable: di production wajib `__Host-session` \+ `Secure` (server menolak start bila tidak); di development boleh nama tanpa prefix `__Host-` dan tanpa `Secure` agar bisa berjalan di http lokal. Mode ditentukan `APP_ENV`; override nama cookie dan `Secure` hanya berlaku bila `APP_ENV=development`; token Keycloak tidak pernah dikirim ke browser. ID session dibuat baru setiap login. Satu origin dengan frontend, sehingga tidak perlu konfigurasi CORS. \[RP-09\] Cookie berumur sama dengan batas mutlak session \[Lihat security.md §2\]. Session yang tidak ditemukan atau sudah kedaluwarsa menghasilkan `401`, lalu frontend mengarahkan pengguna ke login Keycloak.

## 3. CSRF

**CSRF:** setiap session memiliki token CSRF sendiri (disimpan di session Redis). Frontend mengirimkannya di header `X-CSRF-Token` pada setiap request POST/PUT/DELETE; backend membandingkannya dengan token di session dan memeriksa header `Origin`. Tidak cocok menghasilkan `403`.

## 4. Logout

**Logout:** session dihapus dari Redis dan cookie dibersihkan. \[Lihat security.md §2\]

## 5. Proteksi login

**Proteksi login:** password dimasukkan di Keycloak, sehingga pembatasan percobaan login (brute force detection) diatur di Keycloak. Pembatasan percobaan login per akun memakai Brute Force Detection di Keycloak \[nilai: diisi\]. Backend hanya membatasi endpoint mulai login dan callback per IP. \[menunggu keputusan: A-07\] \[menunggu konfirmasi Keycloak\]

## 6. Rate limiter dan aturan no-cache

**Rate limiter:** Penghitung per jendela 1 menit (`INCR` \+ `EXPIRE`) pada key `ratelimit:{bucket}:{id}`. Tiga bucket: baca publik (daftar, detail, pencarian), baca admin (halaman admin), dan tulis admin (tambah, edit, Post, Retract, hapus, restore, hapus permanen, dan heartbeat soft lock). Jika melebihi batas, backend membalas `429 Too Many Requests` dengan header `Retry-After`. Baca admin: 180 / menit per user ID. Tulis admin: 30 / menit per user ID. Baca publik: dilayani dari cache Redis dengan Cache-Control \+ ETag; limit per IP hanya pada cache miss, \[N\] / menit (env, nilai awal 300; 1.200 bila semua request dihitung). Endpoint mulai login dan callback: batas ketat per IP; per akun ditangani Keycloak. \[RP-05\] /admin dan login tidak di-cache. \[RP-04\] 429 menyertakan Retry-After; frontend menampilkan pesan jelas dan tidak retry agresif. \[RP-06\] IP klien dibaca dari X-Forwarded-For hanya dari proxy tepercaya. \[RP-03\]

Cache konten (TTL dan invalidasi): \[Lihat database.md §7\].

## 7. Sanitasi Input dan Output

* Query database: selalu berparameter (tidak menyambung string SQL dari input user) untuk mencegah SQL injection.  
* Search box: (sarankan review security audit).  
* Parameter request dipastikan bertipe benar, `page` minimal 1, `limit` dibatasi maksimum, dan nilai yang tidak valid ditolak dengan `400`.  
* `tanggal_publikasi` harus berformat `DD-MM-YYYY`, ada di kalender, dan tidak boleh setelah hari ini. Judul Berita dan Dokumentasi maksimal 70 karakter. Pelanggaran ditolak dengan `400` \[Lihat database.md §2\].  
* Konten teks (judul, `isi`): disimpan sebagai teks biasa. Format paragraf dan potongan kartu: \[Lihat database.md §3.1\]. Backend melakukan trim, membatasi panjang, dan membuang karakter kontrol. Pencegahan XSS dilakukan saat ditampilkan: frontend merender konten sebagai teks (bukan HTML mentah), dan tidak boleh memakai `innerHTML` atau `set:html` untuk konten ini.  
* Upload file ke RustFS: tipe file dicek dari isi file (bukan hanya ekstensi), ukuran dibatasi, dan nama file diganti dengan nama acak agar tidak terjadi path traversal atau tabrakan nama.  
* Aset medis publik, termasuk tangkapan layar aplikasi riset: wajib diperiksa manual agar tidak memuat nama, identitas pasien, atau citra medis nyata tanpa izin. Pemeriksaan otomatis tidak dapat membaca teks di dalam gambar.  
* Dokumen pada demo portofolio: wajib sintetis, tanpa rekening, data pribadi, atau catatan bisnis nyata. Tidak ada unggah atau pengiriman dokumen ke layanan eksternal. Pertimbangan keamanan untuk penerapan nyata (akses, retensi, audit, jaringan, hosting model) disampaikan sebagai pertimbangan, bukan fitur demo, dan tanpa klaim sertifikasi atau keamanan absolut.  
* Respons error: tidak menampilkan detail internal (stack trace, query, struktur database) ke client; detailnya hanya dicatat di log. \[RP-06\]
