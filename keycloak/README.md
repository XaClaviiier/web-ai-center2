# Keycloak — UBAYA AI Center CMS

Identity provider untuk admin CMS (backend Go + Chi, admin React di `/admin`). Image Keycloak **26.8.1** dan PostgreSQL **18.6** dipin ke digest; database, volume, realm, dan client Keycloak terpisah dari database aplikasi.

## Ringkasan desain

- Realm `ai-center`, client OIDC **confidential** `ai-center-cms`, Authorization Code + PKCE S256 wajib; implicit, password grant, dan service account nonaktif. Backend hanya meminta scope `openid`.
- Client role **`direktur`** dan **`pegawai`** (huruf kecil, case-sensitive), izin CRUD sama; tampilan memakai `direktur` bila keduanya ada. Akun tanpa salah satu role ditolak (`/admin/?galat=akses_ditolak`), termasuk operator Keycloak.
- Mapper client: `cms-subject` (`sub` di access token; Keycloak tidak mengisinya tanpa scope `basic`), `cms-username` (`preferred_username` di ID dan access token, dipakai sebagai username Log Aktivitas), `cms-display-name` (`name` di ID token), `cms-client-roles`, `cms-api-audience`.
- Backend menukar code, memverifikasi token (issuer, audience, `azp`, nonce, tanda tangan RS256, subject sama), lalu membuat session Redis yang sudah ada (`session:{id}`, cookie HttpOnly, CSRF). Browser tidak pernah menerima token Keycloak. `sub` Keycloak menjadi identitas audit (`created_by`, Log Aktivitas, soft lock, rate limit admin).
- Session idle 1 jam / maksimal 8 jam. Refresh token disimpan terenkripsi (AES-256-GCM) di Redis; saat access token (300 detik) habis, backend me-refresh dan memeriksa ulang role. Akun yang dinonaktifkan atau kehilangan role kehilangan akses paling lambat saat refresh berikutnya (±5 menit).
- Logout menghapus session CMS lalu mengakhiri session SSO Keycloak secara back-channel (best effort; IdP mati tidak menghalangi logout lokal).
- Tidak ada pengguna, password, atau client secret di realm JSON. Login password development (`dev-login`, `ADMIN_PASSWORD`, `/admin/masuk`) sudah dihapus.

## 1. Development lokal

| Layanan | URL |
| --- | --- |
| CMS (Go, publik + admin + API) | `http://localhost:8080` |
| Keycloak (operator lokal) | `http://localhost:8180` |
| Issuer | `http://localhost:8180/realms/ai-center` |
| Callback | `http://localhost:8080/api/auth/callback` |

Pakai `localhost` secara konsisten (bukan `127.0.0.1`); issuer dan cookie bergantung pada hostname.

1. Nyalakan Docker Desktop. Buat folder `keycloak/secrets/` (diabaikan Git) berisi:
   - `db-password.txt`: password acak untuk database Keycloak.
   - `operator.conf`: operator sementara untuk bootstrap pertama:
     ```properties
     bootstrap-admin-username=operator-lokal
     bootstrap-admin-password=<password-acak-buatan-sendiri>
     ```
2. Dari folder `keycloak/` (variabel menunjuk file di atas):
   ```bash
   export KEYCLOAK_DB_PASSWORD_FILE=./secrets/db-password.txt
   export KEYCLOAK_OPERATOR_CONFIG_FILE=./secrets/operator.conf
   docker compose -f compose.dev.yml up -d --build
   ```
   Hanya `127.0.0.1:8180` dipublish. Import startup membuat realm bila belum ada dan **tidak menimpa** realm yang sudah ada; perubahan realm berikutnya dilakukan granular di console.
3. Buka `http://localhost:8180/admin/`, masuk sebagai operator. Di realm **ai-center**:
   - **Clients → ai-center-cms → Credentials**: salin secret ke `keycloak/secrets/cms-client-secret.txt`.
   - **Users**: buat pengguna (isi nama), atur password, lalu **Role mapping → Assign role → Filter by clients → ai-center-cms → `direktur`/`pegawai`**.
4. Buat kunci enkripsi session (tepat 32 byte acak), misalnya `openssl rand -out keycloak/secrets/session-key.bin 32`.

   Langkah 3-4 untuk dev lokal dapat diotomatiskan: `bash keycloak/siapkan-dev.sh` menyalin client secret, membuat kunci session, dan membuat akun uji `uji-direktur`, `uji-pegawai`, dan `uji-tanpa-role` (password di `keycloak/secrets/akun-uji.txt`).
5. Di `backend/.env` (salin dari `backend/.env.example`), isi:
   ```dotenv
   PUBLIC_BASE_URL=http://localhost:8080
   KEYCLOAK_ISSUER=http://localhost:8180/realms/ai-center
   KEYCLOAK_CLIENT_ID=ai-center-cms
   KEYCLOAK_CLIENT_SECRET_FILE=../keycloak/secrets/cms-client-secret.txt
   SESSION_ENCRYPTION_KEY_FILE=../keycloak/secrets/session-key.bin
   ```
   Path relatif terhadap folder `backend/`. Secret hanya untuk backend.
6. Jalankan CMS seperti di `README.md` root (`cd backend && go run ./cmd/api`), buka `http://localhost:8080/admin/`: browser diarahkan ke Keycloak, lalu kembali ke admin dengan nama dan role di header.
7. Setelah operator permanen/MFA disiapkan, hapus operator bootstrap dan kosongkan opsi bootstrap di `operator.conf`.

Menghentikan: `docker compose -f compose.dev.yml down` (tanpa `-v`, kecuali ingin menghapus semua akun lokal).

### Realm yang sudah ada (dibuat dari JSON versi lama)

Import tidak memperbarui realm existing. Tambahkan granular di **Clients → ai-center-cms → Client scopes → ai-center-cms-dedicated → Add mapper**: mapper **User Property** `cms-username` (property `username`, claim `preferred_username`, ID token + access token), dan pastikan mapper **Subject (sub)** `cms-subject` ada. Ubah **Valid redirect URIs** client menjadi `http://localhost:8080/api/auth/callback` (dev) atau callback produksi.

## 2. Produksi

`compose.yml` tidak mengimpor otomatis dan tidak punya kredensial bawaan.

1. Siapkan di luar repository: origin HTTPS publik identitas, origin HTTPS admin privat, sertifikat/key PEM, password database, `operator.conf`. Isi path absolut di `keycloak/.env` (lihat `.env.example`).
2. `CMS_CALLBACK_URL` = `PUBLIC_BASE_URL` produksi + `/api/auth/callback` (HTTPS, persis).
3. Hanya untuk realm baru pada database Keycloak khusus CMS:
   ```bash
   docker compose -f compose.yml config --quiet
   docker compose -f compose.yml build
   docker compose -f compose.yml up -d keycloak-db
   docker compose -f compose.yml run --rm keycloak import --optimized --file=/opt/keycloak/cms-config/ai-center-realm.json --override=false
   docker compose -f compose.yml up -d keycloak
   ```
   Jangan memakai `override=true` atau menimpa realm organisasi bersama; untuk realm existing tambahkan client, role, dan mapper secara granular.
4. Backend: `APP_ENV=production`, `KEYCLOAK_ISSUER` HTTPS, secret client dan kunci session sebagai file privat. Cookie otomatis `__Host-session` + Secure.
5. Keycloak HTTPS dipublish hanya ke `127.0.0.1:8443` untuk reverse proxy. Ingress publik hanya `/realms/ai-center/` dan `/resources/`; blokir `/admin/`, realm lain, `/health`, `/metrics`. Hostname admin bukan kontrol akses: lindungi dengan VPN/jaringan operator.
6. `sslRequired=external` mengizinkan HTTP loopback dev; produksi tetap HTTPS. Cadangkan database Keycloak terpisah dari database CMS.

## 3. Pemeriksaan penerimaan

- `/admin/` tanpa session → Keycloak (PKCE S256); login berhasil → admin dengan nama/role; username Log Aktivitas = `preferred_username`.
- Akun tanpa role → pesan akses ditolak, tanpa session, tanpa loop redirect; "Masuk dengan akun lain" memaksa form login.
- Callback/state dipakai ulang atau cookie binding berbeda → ditolak.
- Mutasi tanpa CSRF/Origin cocok → 403; tanpa session → 401.
- Disable akun/hapus role → akses hilang paling lambat saat refresh token berikutnya.
- Logout → session CMS hilang dan login berikutnya meminta kredensial lagi; Keycloak mati tidak menghalangi logout lokal.

## Rujukan vendor

- [Container Keycloak](https://www.keycloak.org/server/containers)
- [Import/export realm](https://www.keycloak.org/server/importExport)
- [Bootstrap operator](https://www.keycloak.org/server/bootstrap-admin-recovery)
- [Hostname](https://www.keycloak.org/server/hostname)
