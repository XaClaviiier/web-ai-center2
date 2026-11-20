-- Log Aktivitas admin (docs/database.md bagian "Log Aktivitas"): riwayat tambah/ubah/hapus pada berita, dokumentasi, testimoni, tim.
-- Ditulis oleh trigger di transaksi yang SAMA dengan perubahan data: gagal menulis log menggagalkan perubahan.
-- Aktor (sub + username dari sesi server terverifikasi) dikirim aplikasi lewat set_config lokal-transaksi
-- (app.sub, app.username); tanpa aktor (job purge trash, impor, tes) tidak ada baris log. Tidak menyimpan payload,
-- isi, password, atau token; hanya label ringkas (judul/nama, maks 140 karakter). Tanpa FK ke item agar log bertahan
-- setelah item dihapus.

-- +goose Up
-- +goose StatementBegin

CREATE TABLE log_aktivitas (
    id           bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    terjadi_pada timestamptz NOT NULL DEFAULT now(),
    pengguna_sub text        NOT NULL,
    username     text        NOT NULL,
    jenis        text        NOT NULL CHECK (jenis IN ('tambah', 'ubah', 'hapus')),
    aksi         text        NOT NULL CHECK (aksi IN ('tambah', 'ubah', 'post', 'retract', 'restore', 'hapus', 'hapus_permanen')),
    modul        text        NOT NULL CHECK (modul IN ('berita', 'dokumentasi', 'testimoni', 'tim')),
    item_id      bigint      NOT NULL,
    label_item   text        NOT NULL
);

-- Urutan tampil: terbaru dulu, id sebagai pemutus seri; filter username memakai indeks kedua.
CREATE INDEX log_aktivitas_urut_idx ON log_aktivitas (terjadi_pada DESC, id DESC);
CREATE INDEX log_aktivitas_username_idx ON log_aktivitas (username, terjadi_pada DESC, id DESC);

CREATE FUNCTION catat_aktivitas() RETURNS trigger LANGUAGE plpgsql AS $fn$
DECLARE
    v_sub   text := coalesce(current_setting('app.sub', true), '');
    v_nama  text := coalesce(current_setting('app.username', true), '');
    v_baru  jsonb;
    v_lama  jsonb;
    v_ini   jsonb;
    v_aksi  text;
    v_jenis text;
BEGIN
    IF v_sub = '' THEN
        RETURN coalesce(NEW, OLD); -- aksi sistem tanpa aktor: tidak dicatat
    END IF;

    IF TG_OP = 'INSERT' THEN
        v_ini := to_jsonb(NEW);
        v_aksi := 'tambah';
    ELSIF TG_OP = 'DELETE' THEN
        v_ini := to_jsonb(OLD);
        v_aksi := 'hapus_permanen';
    ELSE
        v_baru := to_jsonb(NEW);
        v_lama := to_jsonb(OLD);
        v_ini := v_baru;
        IF (v_lama->>'deleted_at') IS NULL AND (v_baru->>'deleted_at') IS NOT NULL THEN
            v_aksi := 'hapus';
        ELSIF (v_lama->>'deleted_at') IS NOT NULL AND (v_baru->>'deleted_at') IS NULL THEN
            v_aksi := 'restore';
        ELSIF v_lama->>'status' = 'draft' AND v_baru->>'status' = 'posted' THEN
            v_aksi := 'post';
        ELSIF v_lama->>'status' = 'posted' AND v_baru->>'status' = 'draft' THEN
            v_aksi := 'retract';
        ELSE
            v_aksi := 'ubah';
        END IF;
    END IF;

    -- jenis: Post/Retract/Restore termasuk "ubah"; hapus ke trash dan hapus permanen termasuk "hapus".
    v_jenis := CASE v_aksi WHEN 'tambah' THEN 'tambah' WHEN 'hapus' THEN 'hapus' WHEN 'hapus_permanen' THEN 'hapus' ELSE 'ubah' END;

    INSERT INTO log_aktivitas (pengguna_sub, username, jenis, aksi, modul, item_id, label_item)
    VALUES (v_sub, v_nama, v_jenis, v_aksi, TG_TABLE_NAME, (v_ini->>'id')::bigint,
            left(coalesce(v_ini->>'judul', v_ini->>'nama', ''), 140));

    RETURN coalesce(NEW, OLD);
END;
$fn$;

CREATE TRIGGER berita_log AFTER INSERT OR UPDATE OR DELETE ON berita FOR EACH ROW EXECUTE FUNCTION catat_aktivitas();
CREATE TRIGGER dokumentasi_log AFTER INSERT OR UPDATE OR DELETE ON dokumentasi FOR EACH ROW EXECUTE FUNCTION catat_aktivitas();
CREATE TRIGGER testimoni_log AFTER INSERT OR UPDATE OR DELETE ON testimoni FOR EACH ROW EXECUTE FUNCTION catat_aktivitas();
CREATE TRIGGER tim_log AFTER INSERT OR UPDATE OR DELETE ON tim FOR EACH ROW EXECUTE FUNCTION catat_aktivitas();

-- +goose StatementEnd

-- +goose Down
DROP TRIGGER tim_log ON tim;
DROP TRIGGER testimoni_log ON testimoni;
DROP TRIGGER dokumentasi_log ON dokumentasi;
DROP TRIGGER berita_log ON berita;
DROP FUNCTION catat_aktivitas();
DROP TABLE log_aktivitas;
