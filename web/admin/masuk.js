/* Halaman masuk development (/admin/masuk). Memanggil POST /api/auth/dev-login; sesi dibuat server lewat cookie
   HttpOnly. Tidak ada password di berkas statis. Di production halaman ini tidak dilayani (404). */
(() => {
  const form = document.getElementById('form-masuk');
  const galat = document.getElementById('galat-masuk');
  const tombol = form.querySelector('button[type="submit"]');

  const tampilkan = (pesan) => { galat.textContent = pesan; galat.hidden = false; };

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    galat.hidden = true;
    const password = form.elements.password.value;
    if (!password) return tampilkan('Password wajib diisi.');
    tombol.disabled = true;
    try {
      const res = await fetch('/api/auth/dev-login', {
        method: 'POST',
        credentials: 'same-origin',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ password }),
      });
      if (res.ok) { location.href = '/admin/'; return; }
      if (res.status === 401) tampilkan('Password salah.');
      else if (res.status === 429) {
        const tunggu = Number.parseInt(res.headers.get('Retry-After'), 10);
        tampilkan(`Terlalu banyak percobaan. Coba lagi${tunggu > 0 ? ` dalam ${tunggu} detik` : ' beberapa saat lagi'}.`);
      } else tampilkan('Terjadi kendala di server. Coba lagi beberapa saat lagi.');
    } catch {
      tampilkan('Tidak dapat menghubungi server.');
    } finally {
      tombol.disabled = false;
    }
  });
})();
