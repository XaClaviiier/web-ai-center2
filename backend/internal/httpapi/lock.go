package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/XaClaviiier/web-ai-center2/backend/internal/admin"
	"github.com/XaClaviiier/web-ai-center2/backend/internal/lock"
)

const (
	kodeTerkunci   = "locked"    // dipegang orang lain
	kodeLockHilang = "lock_lost" // kedaluwarsa atau bukan milik pemanggil (heartbeat); dulu lock_hilang

	statusTerkunci   = 423             // http.StatusLocked
	statusLockHilang = http.StatusGone // 410, sesuai frontend-final

	batasWaktuLepasLock = 3 * time.Second
)

type kunciLockCtx struct{}

// cekKunci memastikan item ada dan boleh dikunci (berita/dokumentasi/testimoni: draft dan tidak di-trash).
type cekKunci func(ctx context.Context, id int64) error

// pengelolaLock memasang route /{id}/lock dan menegakkan soft lock pada route ubah-data satu sumber.
type pengelolaLock struct {
	log       *slog.Logger
	l         *lock.Lock
	jenis     string
	cek       cekKunci
	heartbeat int // detik, disarankan ke klien
}

type pemegangJSON struct {
	Nama string `json:"nama"`
}

type lockJSON struct {
	Pemegang        pemegangJSON `json:"pemegang"`
	MilikSaya       bool         `json:"milikSaya"`
	KedaluwarsaPada time.Time    `json:"kedaluwarsaPada"`
	HeartbeatDetik  int          `json:"heartbeatDetik,omitempty"`
}

func bentukLock(i lock.Info, sub string, heartbeat int) *lockJSON {
	return &lockJSON{
		Pemegang:        pemegangJSON{Nama: i.Nama},
		MilikSaya:       i.Sub == sub,
		KedaluwarsaPada: i.KedaluwarsaPada.UTC(),
		HeartbeatDetik:  heartbeat,
	}
}

// denganLock membungkus router admin satu sumber: route /{id}/lock (pasang, heartbeat, lepas) dan penegakan lock.
// p nil = soft lock dimatikan (tes tanpa Redis).
func denganLock(p *pengelolaLock, router func(chi.Router)) func(chi.Router) {
	if p == nil {
		return router
	}
	return func(r chi.Router) {
		r.Use(p.penegak)
		r.Post("/{id}/lock", p.pasang)
		r.Put("/{id}/lock", p.perpanjang)
		r.Delete("/{id}/lock", p.lepas)
		router(r)
	}
}

func (p *pengelolaLock) pasang(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(r)
	if !ok {
		tulisTidakDitemukan(w)
		return
	}
	if err := p.cek(r.Context(), id); err != nil {
		petakanGalatAdmin(p.log, w, r, err)
		return
	}
	sess, _ := SessionDariContext(r.Context())
	info, err := p.l.Pasang(r.Context(), p.jenis, id, lock.Pemegang{Sub: sess.Sub, Nama: sess.Nama})
	if err != nil {
		p.galat(w, r, err)
		return
	}
	tulisJSON(w, http.StatusOK, bentukLock(info, sess.Sub, p.heartbeat))
}

func (p *pengelolaLock) perpanjang(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(r)
	if !ok {
		tulisTidakDitemukan(w)
		return
	}
	sess, _ := SessionDariContext(r.Context())
	info, err := p.l.Perpanjang(r.Context(), p.jenis, id, sess.Sub)
	if err != nil {
		p.galat(w, r, err)
		return
	}
	tulisJSON(w, http.StatusOK, bentukLock(info, sess.Sub, p.heartbeat))
}

func (p *pengelolaLock) lepas(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(r)
	if !ok {
		tulisTidakDitemukan(w)
		return
	}
	sess, _ := SessionDariContext(r.Context())
	if err := p.l.Lepas(r.Context(), p.jenis, id, sess.Sub); err != nil {
		p.galat(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// penegak menolak (423) perubahan pada item yang dikunci orang lain, mengisi info lock pada GET detail, dan
// melepas lock milik pemanggil setelah perubahan sukses. Redis bermasalah: gagal tertutup (503).
func (p *pengelolaLock) penegak(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ekor, ok := p.idDariURL(r)
		if !ok && r.Method == http.MethodGet && strings.TrimRight(r.URL.Path, "/") == "/api/admin/"+p.jenis {
			p.layaniDaftar(w, r, next) // daftar: tambahkan dikunciOleh per baris
			return
		}
		if !ok || ekor == "lock" {
			next.ServeHTTP(w, r)
			return
		}
		sess, _ := SessionDariContext(r.Context())

		if metodeAman(r.Method) {
			if ekor != "" { // hanya detail memuat info lock
				next.ServeHTTP(w, r)
				return
			}
			info, err := p.l.Baca(r.Context(), p.jenis, id)
			if err != nil {
				p.galatRedis(w, r, err)
				return
			}
			var lj *lockJSON
			if info != nil {
				lj = bentukLock(*info, sess.Sub, 0)
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), kunciLockCtx{}, lj)))
			return
		}

		info, err := p.l.Baca(r.Context(), p.jenis, id)
		if err != nil {
			p.galatRedis(w, r, err)
			return
		}
		if info != nil && info.Sub != sess.Sub {
			p.tulisTerkunci(w, *info)
			return
		}
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		if s := ww.Status(); s >= 200 && s <= 299 {
			ctx, batal := context.WithTimeout(context.WithoutCancel(r.Context()), batasWaktuLepasLock)
			defer batal()
			if err := p.l.Lepas(ctx, p.jenis, id, sess.Sub); err != nil {
				// Perubahan sudah tersimpan; kunci paling lama bertahan sampai TTL.
				p.log.Error("gagal melepas lock setelah perubahan", "err", err, "jenis", p.jenis, "id", id, "request_id", middleware.GetReqID(r.Context()))
			}
		}
	})
}

// idDariURL membaca {id} dan sisa path (mis. "post", "lock") dari /api/admin/{jenis}/{id}[/ekor].
func (p *pengelolaLock) idDariURL(r *http.Request) (id int64, ekor string, ok bool) {
	penanda := "/api/admin/" + p.jenis + "/"
	i := strings.Index(r.URL.Path, penanda)
	if i < 0 {
		return 0, "", false
	}
	bagian := strings.SplitN(strings.TrimRight(r.URL.Path[i+len(penanda):], "/"), "/", 2)
	n, err := strconv.ParseInt(bagian[0], 10, 64)
	if err != nil || n < 1 {
		return 0, "", false // id tidak sah: biarkan handler membalas 404
	}
	if len(bagian) == 2 {
		ekor = bagian[1]
	}
	return n, ekor, true
}

type galatLockJSON struct {
	Error struct {
		Code     string        `json:"code"`
		Message  string        `json:"message"`
		Pemegang *pemegangJSON `json:"pemegang,omitempty"`
	} `json:"error"`
}

func (p *pengelolaLock) tulisTerkunci(w http.ResponseWriter, i lock.Info) {
	var g galatLockJSON
	g.Error.Code = kodeTerkunci
	g.Error.Message = "Konten sedang diedit oleh " + potongNama(i.Nama)
	g.Error.Pemegang = &pemegangJSON{Nama: i.Nama}
	tulisJSON(w, statusTerkunci, g)
}

func potongNama(s string) string {
	if utf8.RuneCountInString(s) > 100 {
		return string([]rune(s)[:100])
	}
	return s
}

func (p *pengelolaLock) galat(w http.ResponseWriter, r *http.Request, err error) {
	var dipegang *lock.ErrDipegang
	switch {
	case errors.As(err, &dipegang):
		p.tulisTerkunci(w, dipegang.Info)
	case errors.Is(err, lock.ErrHilang):
		tulisGalat(w, statusLockHilang, kodeLockHilang, "Kunci edit sudah berakhir; buka kembali form edit")
	default:
		p.galatRedis(w, r, err)
	}
}

func (p *pengelolaLock) galatRedis(w http.ResponseWriter, r *http.Request, err error) {
	p.log.Error("soft lock bermasalah", "err", err, "path", r.URL.Path, "request_id", middleware.GetReqID(r.Context()))
	tulisGalat(w, http.StatusServiceUnavailable, kodeSessionBermasalah, "Layanan session sedang tidak tersedia")
}

// tulisDetailAdmin menulis detail admin; bila soft lock aktif, ditambah field "lock" (objek atau null).
func tulisDetailAdmin(w http.ResponseWriter, r *http.Request, v any) {
	nilai := r.Context().Value(kunciLockCtx{})
	if nilai == nil { // soft lock tidak terpasang
		tulisJSON(w, http.StatusOK, v)
		return
	}
	lj, _ := nilai.(*lockJSON)
	mentah, err := json.Marshal(v)
	if err != nil {
		slog.Error("gagal menyusun detail admin", "err", err)
		tulisGalat(w, http.StatusInternalServerError, kodeInternal, "Terjadi kesalahan pada server")
		return
	}
	var peta map[string]json.RawMessage
	if err := json.Unmarshal(mentah, &peta); err != nil {
		tulisJSON(w, http.StatusOK, v)
		return
	}
	if lj == nil {
		peta["lock"] = json.RawMessage("null")
	} else {
		peta["lock"], _ = json.Marshal(lj)
	}
	tulisJSON(w, http.StatusOK, peta)
}

// cekDraft adalah aturan "boleh dikunci" untuk berita, dokumentasi, dan testimoni.
func cekDraft(status string, dihapus *time.Time) error {
	if status != "draft" || dihapus != nil {
		return admin.ErrStatusTidakValid
	}
	return nil
}

// Pembuat pengelola lock per sumber; nil bila soft lock dimatikan atau layanan sumbernya tidak ada.

func (d Deps) pengelola(jenis string, cek cekKunci) *pengelolaLock {
	if d.Lock == nil || cek == nil {
		return nil
	}
	return &pengelolaLock{log: d.Log, l: d.Lock, jenis: jenis, cek: cek, heartbeat: d.LockHeartbeatDetik}
}

func (d Deps) lockBerita() *pengelolaLock {
	svc := d.Admin.Berita
	if svc == nil {
		return nil
	}
	return d.pengelola(lock.Berita, func(ctx context.Context, id int64) error {
		b, err := svc.Detail(ctx, id)
		if err != nil {
			return err
		}
		return cekDraft(b.Status, b.DihapusPada)
	})
}

func (d Deps) lockDokumentasi() *pengelolaLock {
	svc := d.Admin.Dokumentasi
	if svc == nil {
		return nil
	}
	return d.pengelola(lock.Dokumentasi, func(ctx context.Context, id int64) error {
		b, err := svc.Detail(ctx, id)
		if err != nil {
			return err
		}
		return cekDraft(b.Status, b.DihapusPada)
	})
}

func (d Deps) lockTestimoni() *pengelolaLock {
	svc := d.Admin.Testimoni
	if svc == nil {
		return nil
	}
	return d.pengelola(lock.Testimoni, func(ctx context.Context, id int64) error {
		b, err := svc.Detail(ctx, id)
		if err != nil {
			return err
		}
		return cekDraft(b.Status, b.DihapusPada)
	})
}

// Tim tidak punya status: cukup ada.
func (d Deps) lockTim() *pengelolaLock {
	svc := d.Admin.Tim
	if svc == nil {
		return nil
	}
	return d.pengelola(lock.Tim, func(ctx context.Context, id int64) error {
		_, err := svc.Detail(ctx, id)
		return err
	})
}

// penyanggaRespons menampung respons handler agar badan JSON daftar dapat diperkaya sebelum dikirim.
type penyanggaRespons struct {
	header http.Header
	status int
	badan  bytes.Buffer
}

func (b *penyanggaRespons) Header() http.Header         { return b.header }
func (b *penyanggaRespons) WriteHeader(s int)           { b.status = s }
func (b *penyanggaRespons) Write(d []byte) (int, error) { return b.badan.Write(d) }

// layaniDaftar menjalankan handler daftar lalu menambahkan "dikunciOleh" (nama pemegang) pada setiap item yang
// sedang dikunci ORANG LAIN, sehingga UI dapat menonaktifkan aksi dan menampilkan "Sedang diedit oleh ...".
// Kunci milik pemanggil sendiri tidak ditandai. Redis bermasalah: gagal tertutup (503), sama seperti detail.
func (p *pengelolaLock) layaniDaftar(w http.ResponseWriter, r *http.Request, next http.Handler) {
	buf := &penyanggaRespons{header: http.Header{}, status: http.StatusOK}
	next.ServeHTTP(buf, r)

	salin := func(badan []byte) {
		for k, v := range buf.header {
			w.Header()[k] = v
		}
		w.WriteHeader(buf.status)
		_, _ = w.Write(badan)
	}
	var peta map[string]json.RawMessage
	if buf.status != http.StatusOK || json.Unmarshal(buf.badan.Bytes(), &peta) != nil || peta["items"] == nil {
		salin(buf.badan.Bytes())
		return
	}
	var items []map[string]json.RawMessage
	if json.Unmarshal(peta["items"], &items) != nil {
		salin(buf.badan.Bytes())
		return
	}
	ids := make([]int64, 0, len(items))
	for _, it := range items {
		var id int64
		if json.Unmarshal(it["id"], &id) == nil && id > 0 {
			ids = append(ids, id)
		}
	}
	kunci, err := p.l.BacaBanyak(r.Context(), p.jenis, ids)
	if err != nil {
		p.galatRedis(w, r, err)
		return
	}
	sess, _ := SessionDariContext(r.Context())
	for _, it := range items {
		var id int64
		if json.Unmarshal(it["id"], &id) != nil {
			continue
		}
		if pm, ada := kunci[id]; ada && pm.Sub != sess.Sub {
			it["dikunciOleh"], _ = json.Marshal(potongNama(pm.Nama))
		}
	}
	peta["items"], _ = json.Marshal(items)
	badan, err := json.Marshal(peta)
	if err != nil {
		salin(buf.badan.Bytes())
		return
	}
	buf.header.Del("Content-Length")
	salin(append(badan, '\n'))
}
