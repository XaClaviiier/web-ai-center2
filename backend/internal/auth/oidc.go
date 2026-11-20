package auth

import (
	"cmp"
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// ErrAksesDitolak berarti pengguna terverifikasi Keycloak tetapi tidak memiliki role klien direktur atau pegawai.
var ErrAksesDitolak = errors.New("pengguna tidak memiliki role yang diizinkan")

// Role klien Keycloak yang boleh masuk admin. Keduanya berhak sama; Role di session hanya label tampilan.
const (
	RoleDirektur = "direktur"
	RolePegawai  = "pegawai"
)

// batasHTTPOIDC membatasi setiap panggilan ke Keycloak (discovery, JWKS, token, logout).
const batasHTTPOIDC = 10 * time.Second

// KonfigOIDC adalah konfigurasi klien confidential Keycloak.
type KonfigOIDC struct {
	Issuer       string // harus sama persis dengan "issuer" hasil discovery
	ClientID     string
	ClientSecret string
	RedirectURL  string       // PUBLIC_BASE_URL + /api/auth/callback
	HTTP         *http.Client // nil = klien dengan batas waktu batasHTTPOIDC
}

// OIDC adalah klien Keycloak: membangun URL login (PKCE S256 + nonce), menukar kode, menyegarkan token, dan
// mengakhiri session SSO. Semua token diverifikasi (RS256, issuer, audience, azp, masa berlaku) sebelum dipakai.
type OIDC struct {
	klien     string
	rahasia   string
	oauth     oauth2.Config
	verif     *oidc.IDTokenVerifier
	akhirSesi string // end_session_endpoint; kosong = penyedia tidak mendukung logout back-channel
	http      *http.Client
}

// BukaOIDC menjalankan discovery dari issuer (wajib sama persis) dan menyiapkan verifikator token.
func BukaOIDC(ctx context.Context, k KonfigOIDC) (*OIDC, error) {
	hc := k.HTTP
	if hc == nil {
		hc = &http.Client{Timeout: batasHTTPOIDC}
	}
	// Klien di context juga dipakai go-oidc untuk mengambil JWKS selanjutnya (tanpa ikut pembatalan ctx).
	p, err := oidc.NewProvider(oidc.ClientContext(ctx, hc), k.Issuer)
	if err != nil {
		return nil, fmt.Errorf("discovery OIDC %s: %w", k.Issuer, err)
	}
	var meta struct {
		AkhirSesi string `json:"end_session_endpoint"`
	}
	if err := p.Claims(&meta); err != nil {
		return nil, fmt.Errorf("membaca metadata OIDC: %w", err)
	}
	ep := p.Endpoint()
	ep.AuthStyle = oauth2.AuthStyleInParams
	return &OIDC{
		klien:   k.ClientID,
		rahasia: k.ClientSecret,
		oauth: oauth2.Config{
			ClientID:     k.ClientID,
			ClientSecret: k.ClientSecret,
			Endpoint:     ep,
			RedirectURL:  k.RedirectURL,
			// Klien Keycloak tidak punya client scope tambahan: meminta selain openid ditolak (invalid_scope).
			Scopes: []string{oidc.ScopeOpenID},
		},
		verif:     p.Verifier(&oidc.Config{ClientID: k.ClientID, SupportedSigningAlgs: []string{oidc.RS256}}),
		akhirSesi: meta.AkhirSesi,
		http:      hc,
	}, nil
}

// URLMasuk membangun URL authorize Keycloak. pilihAkun menambahkan prompt=login (ganti akun).
func (o *OIDC) URLMasuk(state, nonce, verifier string, pilihAkun bool) string {
	opsi := []oauth2.AuthCodeOption{oidc.Nonce(nonce), oauth2.S256ChallengeOption(verifier)}
	if pilihAkun {
		opsi = append(opsi, oauth2.SetAuthURLParam("prompt", "login"))
	}
	return o.oauth.AuthCodeURL(state, opsi...)
}

// Tukar menukar kode otorisasi (dengan PKCE verifier) lalu memverifikasi ID token (termasuk nonce) dan access token.
// ErrAksesDitolak bila pengguna tidak memiliki role yang diizinkan.
func (o *OIDC) Tukar(ctx context.Context, kode, verifier, nonce string) (Identitas, Token, error) {
	tok, err := o.oauth.Exchange(o.konteks(ctx), kode, oauth2.VerifierOption(verifier))
	if err != nil {
		return Identitas{}, Token{}, fmt.Errorf("menukar kode otorisasi: %w", err)
	}
	if tok.RefreshToken == "" {
		return Identitas{}, Token{}, errors.New("respons token tanpa refresh token")
	}
	return o.periksa(ctx, tok, nonce, "")
}

// Segarkan memakai refresh token session (grant refresh_token) lalu memverifikasi ulang token baru, role, dan
// kesamaan sub. invalid_grant atau token/role yang tidak lolos membungkus ErrSegarDitolak; galat lain berarti
// Keycloak tidak terjangkau.
func (o *OIDC) Segarkan(ctx context.Context, s Session) (Identitas, Token, error) {
	if s.Refresh == "" || s.Sub == "" {
		return Identitas{}, Token{}, fmt.Errorf("%w: session tanpa refresh token", ErrSegarDitolak)
	}
	tok, err := o.oauth.TokenSource(o.konteks(ctx), &oauth2.Token{RefreshToken: s.Refresh}).Token()
	if err != nil {
		var re *oauth2.RetrieveError
		if errors.As(err, &re) && re.ErrorCode == "invalid_grant" {
			return Identitas{}, Token{}, fmt.Errorf("%w: invalid_grant", ErrSegarDitolak)
		}
		return Identitas{}, Token{}, fmt.Errorf("menyegarkan token: %w", err)
	}
	ident, t, err := o.periksa(ctx, tok, "", s.Sub)
	if err != nil {
		return Identitas{}, Token{}, fmt.Errorf("%w: %w", ErrSegarDitolak, err)
	}
	return ident, t, nil
}

// AkhiriSSO mengakhiri session SSO Keycloak lewat end_session_endpoint (POST dengan kredensial klien dan refresh
// token). Upaya terbaik: pemanggil tidak menggagalkan logout lokal karena galat ini.
func (o *OIDC) AkhiriSSO(ctx context.Context, refresh string) error {
	if o.akhirSesi == "" || refresh == "" {
		return nil
	}
	form := url.Values{"client_id": {o.klien}, "client_secret": {o.rahasia}, "refresh_token": {refresh}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.akhirSesi, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("mengakhiri session SSO: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := o.http.Do(req)
	if err != nil {
		return fmt.Errorf("mengakhiri session SSO: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("mengakhiri session SSO: status %d", resp.StatusCode)
	}
	return nil
}

func (o *OIDC) konteks(ctx context.Context) context.Context {
	return context.WithValue(ctx, oauth2.HTTPClient, o.http)
}

// klaimToken adalah claim Keycloak yang dipakai dari ID token dan access token.
type klaimToken struct {
	Typ            string `json:"typ"`
	Azp            string `json:"azp"`
	Nama           string `json:"name"`
	Username       string `json:"preferred_username"`
	ResourceAccess map[string]struct {
		Roles []string `json:"roles"`
	} `json:"resource_access"`
}

// periksa memverifikasi respons token. nonce kosong = penyegaran (ID token opsional, tanpa nonce); subLama tidak
// kosong = sub wajib sama dengan session. Role hanya diambil dari access token yang terverifikasi.
func (o *OIDC) periksa(ctx context.Context, tok *oauth2.Token, nonce, subLama string) (Identitas, Token, error) {
	akses, err := o.verif.Verify(ctx, tok.AccessToken)
	if err != nil {
		return Identitas{}, Token{}, fmt.Errorf("access token tidak sah: %w", err)
	}
	var ka klaimToken
	if err := akses.Claims(&ka); err != nil {
		return Identitas{}, Token{}, fmt.Errorf("claim access token: %w", err)
	}
	switch {
	case ka.Typ != "Bearer":
		return Identitas{}, Token{}, errors.New("access token: typ bukan Bearer")
	case ka.Azp != o.klien:
		return Identitas{}, Token{}, errors.New("access token: azp bukan klien ini")
	case akses.Subject == "":
		return Identitas{}, Token{}, errors.New("access token tanpa sub")
	case subLama != "" && akses.Subject != subLama:
		return Identitas{}, Token{}, errors.New("access token: sub berbeda dengan session")
	}

	sumber := ka
	rawID, _ := tok.Extra("id_token").(string)
	switch {
	case rawID == "" && nonce != "":
		return Identitas{}, Token{}, errors.New("respons token tanpa ID token")
	case rawID != "":
		id, err := o.verif.Verify(ctx, rawID)
		if err != nil {
			return Identitas{}, Token{}, fmt.Errorf("ID token tidak sah: %w", err)
		}
		var ki klaimToken
		if err := id.Claims(&ki); err != nil {
			return Identitas{}, Token{}, fmt.Errorf("claim ID token: %w", err)
		}
		switch {
		case ki.Typ != "" && ki.Typ != "ID":
			return Identitas{}, Token{}, errors.New("ID token: typ bukan ID")
		case ki.Azp != "" && ki.Azp != o.klien:
			return Identitas{}, Token{}, errors.New("ID token: azp bukan klien ini")
		case id.Subject != akses.Subject:
			return Identitas{}, Token{}, errors.New("ID token: sub berbeda dengan access token")
		case nonce != "" && subtle.ConstantTimeCompare([]byte(id.Nonce), []byte(nonce)) != 1:
			return Identitas{}, Token{}, errors.New("ID token: nonce tidak cocok")
		}
		sumber = ki
	}

	role := ""
	for _, r := range ka.ResourceAccess[o.klien].Roles {
		if r == RoleDirektur {
			role = RoleDirektur
			break
		}
		if r == RolePegawai {
			role = RolePegawai
		}
	}
	if role == "" {
		return Identitas{}, Token{}, ErrAksesDitolak
	}
	username := cmp.Or(sumber.Username, ka.Username, akses.Subject)
	ident := Identitas{Sub: akses.Subject, Nama: cmp.Or(sumber.Nama, username), Role: role, Username: username}
	return ident, Token{Refresh: tok.RefreshToken, AksesHabis: akses.Expiry.UTC()}, nil
}
