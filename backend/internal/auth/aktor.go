package auth

import "context"

// Aktor adalah pengguna yang melakukan perubahan data. Dibawa lewat context ke lapisan database agar setiap
// perubahan konten tercatat di log aktivitas dalam transaksi yang sama (docs/database.md, Log Aktivitas).
// Identitas HANYA berasal dari sesi server yang terverifikasi, tidak pernah dari body request.
type Aktor struct {
	Sub      string
	Username string // snapshot username saat kejadian; kosong jatuh ke Sub
}

type kunciAktor struct{}

// DenganAktor mengembalikan context yang membawa aktor.
func DenganAktor(ctx context.Context, a Aktor) context.Context {
	if a.Username == "" {
		a.Username = a.Sub
	}
	return context.WithValue(ctx, kunciAktor{}, a)
}

// AktorDariContext mengambil aktor; ok=false untuk aksi sistem (job, tes) sehingga tidak dicatat di log.
func AktorDariContext(ctx context.Context) (Aktor, bool) {
	a, ok := ctx.Value(kunciAktor{}).(Aktor)
	return a, ok && a.Sub != ""
}
