package main

import "testing"

func TestBolehJalanMenolakProduction(t *testing.T) {
	if bolehJalan("production") == nil {
		t.Fatal("seed harus menolak APP_ENV=production")
	}
	for _, env := range []string{"development", "test", ""} {
		if err := bolehJalan(env); err != nil {
			t.Fatalf("APP_ENV=%q seharusnya boleh: %v", env, err)
		}
	}
}

func TestDataSeedUnik(t *testing.T) {
	slug := map[string]bool{}
	for _, b := range dataBerita {
		slug["b/"+b.slug] = true
	}
	for _, d := range dataDokumentasi {
		if slug["d/"+d.slug] {
			t.Fatalf("slug dokumentasi ganda: %s", d.slug)
		}
		slug["d/"+d.slug] = true
	}
	nama := map[string]bool{}
	for _, x := range dataTim {
		if nama[x.nama] || x.urutan < 1 {
			t.Fatalf("tim ganda atau urutan < 1: %+v", x)
		}
		nama[x.nama] = true
	}
}
