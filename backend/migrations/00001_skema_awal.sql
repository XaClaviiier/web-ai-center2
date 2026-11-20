-- Skema awal sesuai docs/database.md §3-§4 dan keputusan sesi desain (lihat docs/backend/perubahan-kode.md).
-- Nilai `status` mengikuti docs: 'draft' / 'posted'. Hanya 00001 yang dipakai sampai ada data nyata.
-- Kolom key gambar (berita.gambar, dokumentasi_gambar.key, testimoni.foto, tim.foto) UNIQUE: satu key hanya boleh dipakai satu baris.

-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- berita dan dokumentasi: tanggal_publikasi dipilih editor di form admin dan wajib sejak simpan pertama (docs/database.md §3.1-§3.2).
CREATE TABLE berita (
    id                bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    slug              text        NOT NULL UNIQUE,
    judul             text        NOT NULL,
    -- Potongan kartu: paragraf pertama isi (sebelum baris kosong), maksimal 200 karakter. Turunan, tidak diisi editor.
    ringkasan         text        GENERATED ALWAYS AS (left(split_part(isi, E'\n\n', 1), 200)) STORED,
    isi               text        NOT NULL,
    tanggal_publikasi date        NOT NULL,
    gambar            text        NOT NULL UNIQUE,
    status            text        NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'posted')),
    version           integer     NOT NULL DEFAULT 1 CHECK (version >= 1),
    created_at        timestamptz NOT NULL DEFAULT now(),
    created_by        text        NOT NULL,
    updated_by        text        NOT NULL,
    deleted_at        timestamptz
);

CREATE TABLE dokumentasi (
    id                bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    slug              text        NOT NULL UNIQUE,
    judul             text        NOT NULL,
    tanggal_publikasi date        NOT NULL,
    status            text        NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'posted')),
    version           integer     NOT NULL DEFAULT 1 CHECK (version >= 1),
    created_at        timestamptz NOT NULL DEFAULT now(),
    created_by        text        NOT NULL,
    updated_by        text        NOT NULL,
    deleted_at        timestamptz
);

CREATE TABLE dokumentasi_gambar (
    id             bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    dokumentasi_id bigint   NOT NULL REFERENCES dokumentasi (id) ON DELETE CASCADE,
    key            text     NOT NULL UNIQUE,
    urutan         smallint NOT NULL CHECK (urutan BETWEEN 1 AND 6),
    -- Teks alternatif; NULL = frontend memakai judul dokumentasi sebagai cadangan. Caption belum dibuat.
    alt            text,
    -- DEFERRABLE: urutan boleh dirapikan ulang (1..n) di dalam satu transaksi.
    CONSTRAINT dokumentasi_gambar_urutan_unik UNIQUE (dokumentasi_id, urutan) DEFERRABLE INITIALLY DEFERRED
);

-- Urutan tampil testimoni: terbaru di atas berdasarkan created_at (tanpa kolom urutan).
CREATE TABLE testimoni (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    nama       text        NOT NULL,
    teks       text        NOT NULL,
    posisi     text        NOT NULL,
    foto       text        UNIQUE,
    status     text        NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'posted')),
    version    integer     NOT NULL DEFAULT 1 CHECK (version >= 1),
    created_at timestamptz NOT NULL DEFAULT now(),
    created_by text        NOT NULL,
    updated_by text        NOT NULL,
    deleted_at timestamptz
);

-- Tim: tanpa status dan tanpa trash (hapus langsung permanen).
CREATE TABLE tim (
    id           bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    nama         text        NOT NULL,
    posisi       text        NOT NULL,
    bidang_riset text        NOT NULL,
    foto         text        NOT NULL UNIQUE,
    urutan       integer     NOT NULL,
    version      integer     NOT NULL DEFAULT 1 CHECK (version >= 1),
    created_at   timestamptz NOT NULL DEFAULT now(),
    created_by   text        NOT NULL,
    updated_by   text        NOT NULL
);

-- Index: daftar publik (item posted, belum di-trash, terbaru dulu)
-- Urutan: tanggal_publikasi terbaru, waktu buat sebagai pemutus seri (requirement.md §4.4-§4.5).
CREATE INDEX berita_terbit_idx ON berita (tanggal_publikasi DESC, created_at DESC)
    WHERE status = 'posted' AND deleted_at IS NULL;
CREATE INDEX dokumentasi_terbit_idx ON dokumentasi (tanggal_publikasi DESC, created_at DESC)
    WHERE status = 'posted' AND deleted_at IS NULL;

-- Index: pencarian judul (ILIKE '%q%'); efektif untuk kata kunci >= 3 karakter.
CREATE INDEX berita_judul_trgm_idx ON berita USING gin (judul gin_trgm_ops);
CREATE INDEX dokumentasi_judul_trgm_idx ON dokumentasi USING gin (judul gin_trgm_ops);

-- Index: job hapus otomatis trash 30 hari
CREATE INDEX berita_trash_idx ON berita (deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX dokumentasi_trash_idx ON dokumentasi (deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX testimoni_trash_idx ON testimoni (deleted_at) WHERE deleted_at IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
DROP TABLE tim;
DROP TABLE testimoni;
DROP TABLE dokumentasi_gambar;
DROP TABLE dokumentasi;
DROP TABLE berita;
-- Ekstensi pg_trgm sengaja tidak dihapus (bisa dipakai objek lain).
