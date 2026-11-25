package main

// Konten contoh dari seed stack lama (git show 917b2d5^:server/src/db/seed.ts), disesuaikan dengan skema sekarang:
// berita.ringkasan diturunkan dari paragraf pertama isi, dokumentasi tanpa deskripsi, tim.urutan mulai dari 1.

type berita struct{ slug, judul, isi, tanggal string }
type dokumentasi struct{ slug, judul, tanggal string }
type tim struct {
	nama, posisi string
	urutan       int
}
type testimoni struct{ nama, posisi, teks string }

var dataBerita = []berita{{
	slug:  "ai-center-dukung-layanan-kesehatan-jatim",
	judul: "AI Center Ubaya Siap Dukung Layanan Kesehatan di Jawa Timur",
	isi: "Rektor Ubaya bertemu Gubernur Jawa Timur di Grahadi; AI Center disiapkan sebagai pusat riset unggulan yang menjawab persoalan nyata, terutama di bidang kesehatan.\n\n" +
		"Pengembangan AI Center dibahas dalam audiensi Rektor Ubaya, Benny Lianto, dengan Gubernur Jawa Timur, Khofifah Indar Parawansa, di Gedung Negara Grahadi, Surabaya, Jumat (22/5/2026). Gubernur Khofifah menyambut positif langkah Ubaya mengembangkan teknologi AI untuk mendukung sektor kesehatan dan peningkatan kualitas sumber daya manusia. Rektor Benny Lianto menyebut AI Center diproyeksikan menjadi pusat riset unggulan di Jawa Timur yang menghasilkan inovasi berbasis AI yang aplikatif, antara lain teknologi implan gigi dan aplikasi AI untuk membantu diagnosis penyakit.\n\n" +
		"Pengembangan AI kesehatan di Ubaya didukung ekosistem yang lengkap, mulai dari Fakultas Kedokteran, rumah sakit milik kampus, hingga kolaborasi dengan institusi kesehatan lain termasuk RSUD Dr. Soetomo. Artikel ini disadur dari siaran pers Dinas Komunikasi dan Informatika Provinsi Jawa Timur (23 Mei 2026).",
	tanggal: "2026-05-22",
}}

var dataDokumentasi = []dokumentasi{
	{"klasifikasi-xray-pneumonia", "Klasifikasi X-Ray Pneumonia", "2026-09-15"},
	{"implan-gigi-otomatis", "Rekomendasi Implan Gigi Otomatis", "2026-08-20"},
	{"algae-finder", "Algae Finder", "2026-07-28"},
	{"translator-bahasa-isyarat", "Translator Bahasa Isyarat", "2026-07-02"},
	{"deteksi-cacat-las", "Deteksi Cacat Las (Welding)", "2026-05-28"},
	{"deteksi-kesegaran-ikan", "Deteksi Kesegaran Ikan", "2026-04-15"},
}

var dataTim = []tim{
	{"Dr. Mohammad Farid Naufal", "Ketua", 1},
	{"Dr. Monica Widiasri", "Koordinator Riset & Edukasi", 2},
	{"Prof. Joko Siswantoro", "Tim Riset", 3},
	{"Marco Ariano Kristyanto, M.M., M.Kom.", "Tim Hardware", 4},
	{"Fikri Baharuddin, M.Kom.", "Tim Software", 5},
	{"Jabesh Nehemiah Wijaya, S.Kom., M.Kom.", "Tim Software", 6},
}

var dataTestimoni = []testimoni{
	{"Peserta Pelatihan ML", "Mahasiswa", "Materi pelatihan runtut dan langsung praktik. Akses GPU lab membuat eksperimen tugas akhir jauh lebih cepat."},
	{"Mitra Industri", "Pengguna Inference Solution", "Proses deployment model didampingi sampai jalan di infrastruktur kami. Komunikasi tim responsif dan dokumentasinya jelas."},
	{"Dosen Peneliti", "Peneliti", "Pendampingan AI Center sangat membantu riset. Tim responsif dan memberi dukungan teknis saat ada kendala."},
}
