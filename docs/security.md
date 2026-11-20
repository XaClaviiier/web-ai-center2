# Security — UBAYA AI Center CMS

## Log Aktivitas — batas simulasi dan rancangan keamanan produksi

Disetujui 11 November 2026. Prototype hanya menawarkan pembacaan log lewat `api.adminLog`; penulisan dilakukan helper privat dari callback mutasi setelah berhasil. Admin API memeriksa akun aktif dan role, selain skenario 401/error/429. Hapus Tim dan hapus permanen memeriksa keberadaan item terlebih dahulu agar tidak mencatat sukses palsu. Respons akun/log berupa salinan sehingga perubahan objek respons tidak mengubah penyimpanan internal.

Ini **bukan batas keamanan server**: akun dipilih melalui URL, data berada di browser, dan jam perangkat bisa dimanipulasi. Label Simulasi tetap ada. Tidak ada klaim persistensi atau audit terpercaya. localStorage/sessionStorage tidak digunakan untuk log. Judul/nama ditampilkan sebagai teks melalui helper DOM, bukan HTML; payload lengkap tidak disalin. Judul/nama sendiri tetap harus bebas rahasia karena dapat menjadi label log.

Rencana produksi (belum diterapkan): ambil `sub` dan username dari sesi server yang diverifikasi; tambahkan snapshot username pada sesi dan respons identitas (`/api/auth/me`). Sumber claim username Keycloak perlu dikonfirmasi saat integrasi, tidak menerima username dari body mutasi. Timestamp berasal dari server/database. Penulisan log menyatu dengan transaksi perubahan data; gagal menulis log menggagalkan perubahan. Endpoint baca wajib autentikasi dan role CMS yang diizinkan, tanpa endpoint tulis/edit/hapus log untuk admin. Role database runtime hanya diberi SELECT/INSERT pada log, tanpa UPDATE/DELETE; pengelolaan retensi memakai akses operasional terpisah. Jangan mencatat password, cookie, token Keycloak/CSRF, atau payload form. Log admin tidak boleh terpapar endpoint publik atau cache publik.

Dokumen aktif. Riwayat perencanaan awal (`00-detail-rencana-web.md`, `01-kontrak-api.md`) tersedia di riwayat Git, commit `5928f05` (tag `arsip-docs-v1`). Peran pengguna: \[Lihat requirement.md §6\]. Audit `created_by` / `updated_by`: \[Lihat database.md §3.6\].

## 1. Akses dan login

Keycloak mengelola pengguna dan role; aplikasi tidak punya manajemen role sendiri. Semua pengguna yang berhasil login dapat menambah, mengubah, dan menghapus konten. Hanya pengguna dengan client role `direktur` atau `pegawai` pada client `ai-center-cms` (realm `ai-center`) yang diterima; lainnya tidak mendapat session dan diarahkan ke `/admin/?galat=akses_ditolak`. Role ditampilkan `direktur` bila pengguna memiliki keduanya.

**Login:** browser diarahkan ke Keycloak (OIDC Authorization Code \+ PKCE S256, scope `openid`). `state` disimpan di Redis (`oidc:login:{hash}`, 5 menit, dipakai sekali) dan diikat ke cookie binding browser. Backend menukar kode sebagai client confidential lalu memverifikasi ID token (issuer, audience, nonce, tanda tangan RS256, masa berlaku) dan access token (issuer, audience, `azp`, `sub` sama dengan ID token). Role hanya dibaca dari access token terverifikasi. `sub` Keycloak menjadi identitas audit; username Log Aktivitas dari `preferred_username`. Konfigurasi: `keycloak/README.md`.

Endpoint auth: \[Lihat stack.md §9.5\].

**Login development:** dihapus (keputusan 2026-11-20). Development juga memakai Keycloak (`keycloak/compose.dev.yml`, port 8180); tidak ada halaman login password, `ADMIN_PASSWORD`, atau `/api/auth/dev-login`.

## 2. Session

**Session login:** Key `session:{id}` berisi `sub`, nama, username, role, token CSRF, waktu dibuat, masa berlaku access token, dan refresh token Keycloak yang dienkripsi AES-256-GCM (kunci 32 byte dari `SESSION_ENCRYPTION_KEY_FILE`). Cookie hanya berisi ID acak (256 bit). Dibuat setelah login Keycloak berhasil, dihapus saat logout. Saat access token (300 detik) habis, backend me-refresh token dan memeriksa ulang role; akun yang dinonaktifkan atau kehilangan role kehilangan session paling lambat pada refresh berikutnya (`401`). Refresh dikoordinasikan per session lewat kunci Redis; logout yang terjadi bersamaan tidak dapat dihidupkan kembali. Idle 1 jam (diperpanjang tiap request), batas mutlak 8 jam tidak bertambah oleh refresh; cookie dibuat berumur sama dengan batas mutlak.

**Session:** backend membuat session di Redis (detail pada paragraf Session login di atas). Browser hanya menerima ID session acak lewat cookie `HttpOnly`, `Secure`, `SameSite=Lax`, `Path=/` (nama `__Host-session`). Nama cookie dan flag `Secure` diatur lewat environment variable: di production wajib `__Host-session` \+ `Secure` (server menolak start bila tidak); di development boleh nama tanpa prefix `__Host-` dan tanpa `Secure` agar bisa berjalan di http lokal. Mode ditentukan `APP_ENV`; override nama cookie dan `Secure` hanya berlaku bila `APP_ENV=development`; token Keycloak tidak pernah dikirim ke browser. ID session dibuat baru setiap login. Satu origin dengan frontend, sehingga tidak perlu konfigurasi CORS. \[RP-09\] Cookie berumur sama dengan batas mutlak session \[Lihat security.md §2\]. Session yang tidak ditemukan atau sudah kedaluwarsa menghasilkan `401`, lalu frontend mengarahkan pengguna ke login Keycloak.

## 3. CSRF

**CSRF:** setiap session memiliki token CSRF sendiri (disimpan di session Redis). Frontend mengirimkannya di header `X-CSRF-Token` pada setiap request POST/PUT/DELETE; backend membandingkannya dengan token di session dan memeriksa header `Origin`. Tidak cocok menghasilkan `403`.

## 4. Logout

**Logout:** session dihapus dari Redis dan cookie dibersihkan, lalu backend mengakhiri session SSO Keycloak secara back-channel memakai refresh token (best effort; kegagalan atau Keycloak mati tidak menghalangi logout lokal). \[Lihat security.md §2\]

## 5. Proteksi login

**Proteksi login:** password dimasukkan di Keycloak, sehingga pembatasan percobaan login per akun memakai Brute Force Detection Keycloak (realm: 5 kegagalan, jeda bertambah 60 detik, maksimal 15 menit). Backend membatasi endpoint mulai login dan callback per IP (`RATE_LIMIT_LOGIN_PER_MINUTE`).

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
