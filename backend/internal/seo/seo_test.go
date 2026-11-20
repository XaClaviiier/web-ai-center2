package seo

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"
)

const kerangkaUji = `<!doctype html><html lang="id"><head><meta charset="utf-8"><!--seo:head--></head><body><div id="app"></div><!--seo:konten--><!--seo:data--></body></html>`

func TestBuatKerangkaMemerlukanTigaPenandaTepatSekali(t *testing.T) {
	if _, err := BuatKerangka(kerangkaUji); err != nil {
		t.Fatalf("kerangka sah ditolak: %v", err)
	}
	kasus := map[string]string{
		"tanpa head":   strings.Replace(kerangkaUji, PenandaHead, "", 1),
		"tanpa konten": strings.Replace(kerangkaUji, PenandaKonten, "", 1),
		"tanpa data":   strings.Replace(kerangkaUji, PenandaData, "", 1),
		"head ganda":   kerangkaUji + PenandaHead,
		"kosong":       "",
	}
	for nama, isi := range kasus {
		if _, err := BuatKerangka(isi); err == nil {
			t.Errorf("%s: harus ditolak", nama)
		}
	}
}

func TestRenderSatuLintasanTidakMenafsirkanUlangPenanda(t *testing.T) {
	k, _ := BuatKerangka(kerangkaUji)
	// Isi yang menyerupai penanda tidak boleh memicu penggantian kedua.
	out := string(k.Render("H"+PenandaKonten, "K"+PenandaData, "D"))
	if strings.Count(out, "K") != 1 || !strings.Contains(out, "H"+PenandaKonten+"</head>") {
		t.Errorf("render menafsirkan ulang penanda: %s", out)
	}
}

func TestHalaman404MenambahNoindexSekali(t *testing.T) {
	h := string(Halaman404([]byte("<html><head><title>x</title></head><body></body></html>")))
	if strings.Count(h, `name="robots" content="noindex"`) != 1 || !strings.Contains(h, `noindex"></head>`) {
		t.Errorf("noindex tidak disisipkan: %s", h)
	}
	if again := string(Halaman404([]byte(h))); again != h {
		t.Error("tidak boleh menyisipkan dua kali")
	}
	if !strings.Contains(string(Halaman404Bawaan("A<B")), `content="noindex"`) || strings.Contains(string(Halaman404Bawaan("A<B")), "A<B") {
		t.Error("halaman bawaan harus noindex dan meng-escape nama situs")
	}
}

func TestHeadMemuatSemuaMetaDanMeng_escape(t *testing.T) {
	judul := `Judul "berbahaya" </title><script>alert(1)</script> & lain`
	head, err := Meta{
		NamaSitus: "AI Center", Judul: judul + " | AI Center", Deskripsi: `desk "x" <b>`,
		Canonical: "https://situs.id/berita/a?x=1&y=2", Gambar: "https://situs.id/media/berita/a.png",
		JSONLD: JSONLDBerita("AI Center", "https://situs.id/berita/a", judul, "2026-11-09", "https://situs.id/media/berita/a.png"),
	}.Head()
	if err != nil {
		t.Fatal(err)
	}
	for _, mau := range []string{
		"<title>", `<meta name="description"`, `<link rel="canonical" href="https://situs.id/berita/a?x=1&amp;y=2">`,
		`property="og:type" content="article"`, `property="og:title"`, `property="og:description"`, `property="og:url"`,
		`property="og:site_name" content="AI Center"`, `property="og:image" content="https://situs.id/media/berita/a.png"`,
		`name="twitter:card" content="summary_large_image"`, `name="twitter:image"`, `<script type="application/ld+json">`,
	} {
		if !strings.Contains(head, mau) {
			t.Errorf("head tidak memuat %q:\n%s", mau, head)
		}
	}
	if strings.Contains(head, "<script>alert(1)") || strings.Contains(head, "</title><script>") {
		t.Errorf("judul tidak ter-escape:\n%s", head)
	}
	// JSON-LD harus JSON sah dan tidak bisa menutup tag script.
	awal := strings.Index(head, `<script type="application/ld+json">`) + len(`<script type="application/ld+json">`)
	akhir := strings.Index(head[awal:], "</script>")
	if akhir < 0 {
		t.Fatal("script JSON-LD tidak tertutup")
	}
	var v map[string]any
	if err := json.Unmarshal([]byte(head[awal:awal+akhir]), &v); err != nil {
		t.Fatalf("JSON-LD tidak sah: %v", err)
	}
	if v["@type"] != "NewsArticle" || v["headline"] != judul {
		t.Errorf("isi JSON-LD salah: %v", v)
	}
}

func TestHeadTanpaGambarTidakMemuatOGImage(t *testing.T) {
	head, _ := Meta{NamaSitus: "S", Judul: "J", Deskripsi: "D", Canonical: "https://x/y"}.Head()
	if strings.Contains(head, "og:image") || strings.Contains(head, "twitter:image") || strings.Contains(head, "ld+json") {
		t.Errorf("tidak boleh ada og:image/JSON-LD tanpa data:\n%s", head)
	}
}

func TestDataAwalMeng_escapeKarakterBerbahaya(t *testing.T) {
	s := DataAwal([]byte("{\"isi\":\"</script><!-- & \\u2028 \"}\n"))
	if !strings.HasPrefix(s, `<script id="data-awal" type="application/json">`) || !strings.HasSuffix(s, "</script>") {
		t.Fatalf("bungkus salah: %s", s)
	}
	dalam := strings.TrimSuffix(strings.TrimPrefix(s, `<script id="data-awal" type="application/json">`), "</script>")
	if strings.ContainsAny(dalam, "<>&") {
		t.Errorf("isi data masih memuat <, >, atau &: %s", dalam)
	}
	var v map[string]string
	if err := json.Unmarshal([]byte(dalam), &v); err != nil || v["isi"] != "</script><!-- & \u2028 " {
		t.Errorf("nilai harus pulih utuh setelah di-parse: %v %v", v, err)
	}
}

func TestDeskripsiBerita(t *testing.T) {
	if got := DeskripsiBerita("  Halo\n\n dunia\tini  "); got != "Halo dunia ini" {
		t.Errorf("rapikan spasi: %q", got)
	}
	panjang := strings.Repeat("kata ", 80)
	got := DeskripsiBerita(panjang)
	if utf8.RuneCountInString(got) > 160 || !strings.HasSuffix(got, "…") || strings.Contains(got, "kat…") {
		t.Errorf("potong salah (%d): %q", utf8.RuneCountInString(got), got)
	}
	// Tepat 160 karakter tidak dipotong; 161 dipotong.
	tepat := strings.Repeat("a", 160)
	if DeskripsiBerita(tepat) != tepat {
		t.Error("160 karakter tidak boleh dipotong")
	}
	if g := DeskripsiBerita(tepat + "b"); utf8.RuneCountInString(g) > 160 || !strings.HasSuffix(g, "…") {
		t.Errorf("161 karakter harus dipotong: %q", g)
	}
	// Multibyte tidak boleh terpotong di tengah rune.
	if g := DeskripsiBerita(strings.Repeat("é", 300)); !utf8.ValidString(g) || utf8.RuneCountInString(g) > 160 {
		t.Errorf("multibyte rusak: %q", g)
	}
}

func TestDeskripsiDokumentasi(t *testing.T) {
	if got := DeskripsiDokumentasi("Workshop AI", "2026-11-09"); got != "Dokumentasi kegiatan Workshop AI — 2026-11-09" {
		t.Errorf("got %q", got)
	}
}

func TestKontenBeritaParagrafDanEscape(t *testing.T) {
	got := KontenBerita(`J<udul>`, "Satu <b>.\r\n\r\nDua & tiga.\n\n\n\nEmpat")
	mau := `<article id="konten-awal"><h1>J&lt;udul&gt;</h1><p>Satu &lt;b&gt;.</p><p>Dua &amp; tiga.</p><p>Empat</p></article>`
	if got != mau {
		t.Errorf("got  %s\nmau  %s", got, mau)
	}
}

func TestKontenDokumentasiAltSelaluJudul(t *testing.T) {
	got := KontenDokumentasi(`Acara "A"`, "Desk", []Gambar{{URL: "/media/a.png"}, {URL: "/media/b.png"}})
	if n := strings.Count(got, `alt="Acara &#34;A&#34;"`); n != 2 {
		t.Errorf("alt semua gambar harus judul (2x), dapat %d: %s", n, got)
	}
	for _, mau := range []string{`src="/media/a.png"`, `src="/media/b.png"`, "<h1>Acara &#34;A&#34;</h1>"} {
		if !strings.Contains(got, mau) {
			t.Errorf("tidak memuat %q: %s", mau, got)
		}
	}
}

func TestJSONLDDokumentasi(t *testing.T) {
	v := JSONLDDokumentasi("https://x/dokumentasi/a", "A", "Desk", "2026-01-02", []string{"https://x/media/1.png", "https://x/media/2.png"})
	b, _ := json.Marshal(v)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	if m["@type"] != "ImageGallery" || len(m["image"].([]any)) != 2 || m["datePublished"] != "2026-01-02" {
		t.Errorf("JSON-LD dokumentasi salah: %s", b)
	}
}

func TestSitemapDanRobots(t *testing.T) {
	data, dipotong, err := Sitemap("https://situs.id", []URL{{Path: "/"}, {Path: "/berita/a&b", Lastmod: "2026-11-09"}})
	if err != nil || dipotong {
		t.Fatal(err, dipotong)
	}
	var set struct {
		URL []struct {
			Loc     string `xml:"loc"`
			Lastmod string `xml:"lastmod"`
		} `xml:"url"`
	}
	if err := xml.Unmarshal(data, &set); err != nil {
		t.Fatalf("XML tidak sah: %v\n%s", err, data)
	}
	if len(set.URL) != 2 || set.URL[0].Loc != "https://situs.id/" || set.URL[0].Lastmod != "" ||
		set.URL[1].Loc != "https://situs.id/berita/a&b" || set.URL[1].Lastmod != "2026-11-09" {
		t.Errorf("isi sitemap salah: %+v", set.URL)
	}
	if !strings.HasPrefix(string(data), "<?xml") || !strings.Contains(string(data), `xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"`) {
		t.Errorf("header/xmlns salah:\n%s", data)
	}

	banyak := make([]URL, BatasSitemap+5)
	for i := range banyak {
		banyak[i] = URL{Path: fmt.Sprintf("/berita/s-%d", i)}
	}
	data, dipotong, err = Sitemap("https://situs.id", banyak)
	if err != nil || !dipotong || strings.Count(string(data), "<url>") != BatasSitemap {
		t.Errorf("harus dipotong ke %d (dipotong=%v err=%v n=%d)", BatasSitemap, dipotong, err, strings.Count(string(data), "<url>"))
	}

	r := Robots("https://situs.id")
	for _, mau := range []string{"User-agent: *", "Allow: /", "Disallow: /admin", "Disallow: /api/", "Sitemap: https://situs.id/sitemap.xml"} {
		if !strings.Contains(r, mau) {
			t.Errorf("robots tidak memuat %q:\n%s", mau, r)
		}
	}
	if strings.Contains(r, "Disallow: /media") {
		t.Error("/media harus tetap boleh diambil crawler")
	}
}
