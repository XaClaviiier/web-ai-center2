package httpapi

import "time"

// tglAPI mengubah tanggal internal YYYY-MM-DD menjadi DD-MM-YYYY, format tanggal di JSON API
// (keputusan pengguna 2026-11-13 agar sesuai frontend-final; menyimpang dari docs/stack.md §9).
// Nilai yang tidak dapat dibaca dikembalikan apa adanya.
func tglAPI(iso string) string {
	t, err := time.Parse("2006-01-02", iso)
	if err != nil {
		return iso
	}
	return t.Format("02-01-2006")
}
