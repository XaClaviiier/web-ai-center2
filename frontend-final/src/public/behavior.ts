/* Perilaku statis halaman publik: navbar (FLIP, scrollspy, menu), animasi masuk, hero canvas.
   mountPublic() mengembalikan fungsi pembersih untuk semua listener, observer, dan frame. */
import { pasang } from './neural-sphere';

// Path ikon `menu` dan `tutup` dari src/shared/ui (IKON tidak diekspor).
const IKON_MENU = 'M4 6h16M4 12h16M4 18h16';
const IKON_TUTUP = 'M6 6l12 12M18 6L6 18';

const tanpaGaring = (path: string) => path.replace(/\/+$/, '') || '/';

function mountNavbar(signal: AbortSignal): () => void {
  const akar = document.querySelector<HTMLElement>('header.nav');
  const utama = document.getElementById('utama');
  const menu = document.getElementById('menu-utama');
  const burger = akar?.querySelector<HTMLButtonElement>('.nav__burger');
  const cta = akar?.querySelector<HTMLElement>('.nav__cta');
  const wadahNav = akar?.querySelector<HTMLElement>('.nav__dalam');
  const permukaan = akar?.querySelector<HTMLElement>('.nav__permukaan');
  const logo = akar?.querySelector<HTMLElement>('.nav__logo');
  const logoPenuh = akar?.querySelector<HTMLElement>('.nav__logo-penuh');
  if (!akar || !utama || !menu || !burger || !cta || !wadahNav || !permukaan || !logo || !logoPenuh) return () => {};
  const aktif = akar.dataset['menuAktif'] ?? '';
  const halaman = document.body.dataset['halaman'] ?? '';
  const ikonBurger = burger.querySelector('path');
  const opsi = { signal };

  function tutupSub(sub: HTMLElement) {
    sub.hidden = true;
    sub.previousElementSibling?.setAttribute('aria-expanded', 'false');
  }

  for (const item of akar.querySelectorAll<HTMLElement>('.nav__punya-sub')) {
    const btn = item.querySelector<HTMLButtonElement>('button.nav__tautan');
    const sub = item.querySelector<HTMLElement>('.nav__sub');
    if (!btn || !sub) continue;
    btn.addEventListener('click', () => {
      const buka = sub.hidden;
      sub.hidden = !buka;
      btn.setAttribute('aria-expanded', String(buka));
    }, opsi);
    document.addEventListener('click', (e) => {
      if (!(e.target instanceof Node) || !item.contains(e.target)) tutupSub(sub);
    }, opsi);
    item.addEventListener('keydown', (e) => {
      if (e.key !== 'Escape' || sub.hidden) return;
      e.stopPropagation();
      tutupSub(sub);
      btn.focus();
    }, opsi);
    item.addEventListener('focusout', (e) => {
      if (!(e.relatedTarget instanceof Node) || !item.contains(e.relatedTarget)) tutupSub(sub);
    }, opsi);
  }

  function aturBurger(buka: boolean) {
    akar!.classList.toggle('nav--menu-terbuka', buka);
    burger!.setAttribute('aria-expanded', String(buka));
    burger!.setAttribute('aria-label', buka ? 'Tutup menu' : 'Buka menu');
    ikonBurger?.setAttribute('d', buka ? IKON_TUTUP : IKON_MENU);
  }

  function tutupMenu() {
    menu!.classList.remove('terbuka');
    aturBurger(false);
    akar!.querySelectorAll<HTMLElement>('.nav__sub').forEach(tutupSub);
  }

  burger.addEventListener('click', () => aturBurger(menu.classList.toggle('terbuka')), opsi);
  akar.addEventListener('keydown', (e) => {
    if (e.key !== 'Escape' || !menu.classList.contains('terbuka')) return;
    tutupMenu();
    burger.focus();
  }, opsi);

  const desktop = matchMedia('(min-width: 80rem)');
  const minimGerak = matchMedia('(prefers-reduced-motion: reduce)');
  const bagianBergerak = [logo, ...menu.children, cta];
  let animasi: Animation[] = [];

  function hentikanGerak() {
    animasi.forEach((a) => a.cancel());
    animasi = [];
  }

  function ubahKeadaan(jadiRingkas: boolean, langsung = false) {
    if (akar!.classList.contains('nav--ringkas') === jadiRingkas) return;
    const bergerak = !langsung && !minimGerak.matches && !akar!.classList.contains('nav--mobile') && typeof permukaan!.animate === 'function';
    const elemen = [...bagianBergerak, permukaan!];
    // FLIP membaca posisi yang sedang terlihat agar pembalikan scroll tidak meloncat ke awal.
    const sebelum = bergerak ? elemen.map((n) => n.getBoundingClientRect()) : [];
    const opasitas = bergerak ? getComputedStyle(permukaan!).opacity : '0';
    hentikanGerak();
    akar!.classList.toggle('nav--ringkas', jadiRingkas);
    if (!bergerak) return;
    elemen.forEach((n, i) => {
      const awal = sebelum[i]!;
      const akhir = n.getBoundingClientRect();
      if (!akhir.width || !akhir.height) return;
      const permukaanSaja = n === permukaan;
      const pindah = `translate(${awal.left - akhir.left}px, ${awal.top - akhir.top}px)`;
      const dari: Keyframe = { transform: permukaanSaja ? `${pindah} scale(${awal.width / akhir.width}, ${awal.height / akhir.height})` : pindah };
      const ke: Keyframe = { transform: 'none' };
      if (permukaanSaja) {
        dari['opacity'] = opasitas;
        ke['opacity'] = jadiRingkas ? 1 : 0;
      }
      animasi.push(n.animate([dari, ke], { duration: 360, easing: 'cubic-bezier(0.77, 0, 0.175, 1)' }));
    });
  }

  akar.classList.toggle('nav--mobile', !desktop.matches);
  let ringkas = scrollY > 96;
  let frameScroll: number | null = null;
  let frameUkuran: number | null = null;

  function aturAktif(dariScroll = false) {
    let induk = ['berita', 'dokumentasi'].includes(aktif) ? 'kabar' : aktif;
    if (halaman === 'inference' && location.hash === '#portofolio') induk = 'contoh-aplikasi';
    if (dariScroll && ['index', 'inference'].includes(halaman)) {
      const batas = parseFloat(getComputedStyle(document.documentElement).getPropertyValue('--nav-offset')) || 104;
      const sections = [...utama!.querySelectorAll<HTMLElement>('[data-nav-section]')].map((section) => ({ section, posisi: section.getBoundingClientRect() }));
      // Section tanpa pasangan menu mempertahankan pasangan terakhir yang telah dilewati.
      if (sections.some(({ posisi }) => posisi.height)) {
        induk = aktif;
        for (const { section, posisi } of sections) {
          if (posisi.height && posisi.top <= batas + 1) induk = section.dataset['navSection'] ?? induk;
        }
        if (scrollY <= 1) induk = aktif;
      }
    }
    akar!.querySelectorAll<HTMLElement>('[data-nav]').forEach((n) => {
      const dipilih = n.dataset['nav'] === induk;
      n.removeAttribute('aria-current');
      n.removeAttribute('data-aktif');
      if (dipilih && n.tagName === 'A') n.setAttribute('aria-current', induk === 'contoh-aplikasi' || induk !== aktif ? 'location' : 'page');
      else if (dipilih) n.setAttribute('data-aktif', 'ya');
    });
  }

  function aturScroll(paksa = false) {
    aturAktif(true);
    if (!paksa && scrollY >= 48 && (akar!.contains(document.activeElement) || akar!.querySelector('.nav__sub:not([hidden])'))) return;
    if (scrollY > 96) ringkas = true;
    else if (scrollY < 48) ringkas = false;
    ubahKeadaan(ringkas && !akar!.classList.contains('nav--mobile'), paksa);
  }

  function aturUkuran() {
    hentikanGerak();
    akar!.classList.add('nav--ukur');
    const sebelumnyaRingkas = akar!.classList.contains('nav--ringkas');
    akar!.classList.remove('nav--ringkas');
    akar!.classList.toggle('nav--mobile', !desktop.matches);
    if (desktop.matches) {
      const viewport = document.documentElement.clientWidth || innerWidth;
      const ruang = wadahNav!.getBoundingClientRect().width || viewport - 2 * Math.min(64, Math.max(20, viewport * 0.03));
      const lebarMenu = menu!.scrollWidth;
      const lebarCta = cta!.getBoundingClientRect().width;
      // Dua sisi sama lebar menjaga menu tepat di tengah, bukan di tengah ruang sisa.
      if (lebarMenu && lebarMenu + 2 * Math.max(152, lebarCta) + 40 > ruang) akar!.classList.add('nav--mobile');
      else {
        akar!.classList.add('nav--ringkas');
        const lebarRingkas = menu!.scrollWidth + 70 + lebarCta + 60;
        akar!.style.setProperty('--nav-lebar-ringkas', `${Math.ceil(lebarRingkas)}px`);
      }
    }
    akar!.classList.toggle('nav--ringkas', sebelumnyaRingkas && !akar!.classList.contains('nav--mobile'));
    if (!akar!.classList.contains('nav--mobile')) tutupMenu();
    document.documentElement.style.setProperty('--nav-offset', akar!.classList.contains('nav--mobile') ? '92px' : '104px');
    aturScroll(true);
    // Selesaikan pengukuran tanpa memutar ulang crossfade logo saat resize/font siap.
    void getComputedStyle(logoPenuh!).opacity;
    akar!.classList.remove('nav--ukur');
  }

  akar.addEventListener('focusin', hentikanGerak, opsi);
  akar.addEventListener('focusout', () => requestAnimationFrame(() => aturScroll()), opsi);
  minimGerak.addEventListener('change', hentikanGerak, opsi);
  akar.addEventListener('click', (e) => {
    const a = e.target instanceof Element ? e.target.closest<HTMLAnchorElement>('a[href]') : null;
    if (!a || e.defaultPrevented || e.button !== 0 || e.ctrlKey || e.metaKey || e.shiftKey || e.altKey) return;
    tutupMenu();
    const tujuan = new URL(a.href, location.href);
    if (tanpaGaring(tujuan.pathname) !== tanpaGaring(location.pathname) || tujuan.search !== location.search || tujuan.hash !== '#portofolio') return;
    requestAnimationFrame(() => {
      const target = document.getElementById('portofolio');
      if (target) {
        target.setAttribute('tabindex', '-1');
        if (e.detail > 0) target.dataset['fokusPointer'] = 'ya';
        else delete target.dataset['fokusPointer'];
        target.addEventListener('blur', () => delete target.dataset['fokusPointer'], { once: true });
        target.addEventListener('keydown', () => delete target.dataset['fokusPointer'], { once: true });
        target.focus({ preventScroll: true });
      }
      aturScroll(true);
    });
  }, opsi);
  // Jalur event tetap mengenali ikon hamburger yang diubah ketika tombol diklik.
  document.addEventListener('click', (e) => {
    if (e.composedPath().includes(akar)) return;
    tutupMenu();
    aturScroll();
  }, opsi);
  window.addEventListener('hashchange', () => {
    aturAktif();
    aturUkuran();
  }, opsi);
  window.addEventListener('scroll', () => {
    if (frameScroll !== null) return;
    frameScroll = requestAnimationFrame(() => {
      frameScroll = null;
      aturScroll();
    });
  }, { passive: true, signal });
  window.addEventListener('resize', () => {
    if (frameUkuran !== null) return;
    frameUkuran = requestAnimationFrame(() => {
      frameUkuran = null;
      aturUkuran();
    });
  }, { passive: true, signal });
  aturAktif();
  const frameAwal = requestAnimationFrame(aturUkuran);
  void document.fonts.ready.then(() => {
    if (!signal.aborted) aturUkuran();
  });
  const pengamatUkuran = new ResizeObserver(() => {
    if (frameScroll !== null) return;
    frameScroll = requestAnimationFrame(() => {
      frameScroll = null;
      aturAktif(true);
    });
  });
  pengamatUkuran.observe(utama);

  return () => {
    hentikanGerak();
    pengamatUkuran.disconnect();
    cancelAnimationFrame(frameAwal);
    if (frameScroll !== null) cancelAnimationFrame(frameScroll);
    if (frameUkuran !== null) cancelAnimationFrame(frameUkuran);
  };
}

/** Tampilkan elemen sekali saat masuk viewport (Beranda: hanya lembar di bawah layar awal). */
function mountReveal(selektor: string, kelasMasuk: string, kelasSiap?: string): () => void {
  if (matchMedia('(prefers-reduced-motion: reduce)').matches) return () => {};
  const pengamat = new IntersectionObserver((entri) => entri.forEach((e) => {
    if (!e.isIntersecting) return;
    e.target.classList.add(kelasMasuk);
    pengamat.unobserve(e.target);
  }), { rootMargin: '0px 0px -8% 0px' });
  document.querySelectorAll(selektor).forEach((el) => {
    if (!kelasSiap) pengamat.observe(el);
    else if (el.getBoundingClientRect().top > innerHeight) {
      el.classList.add(kelasSiap);
      pengamat.observe(el);
    }
  });
  return () => pengamat.disconnect();
}

/** Penekanan alur sekali ketika seluruh daftar terlihat, hanya pada layar lebar tanpa reduced motion. */
function mountFlow(selektor: string, query: string, kelas: string, namaAnimasi: string, signal: AbortSignal): () => void {
  const alur = document.querySelector<HTMLElement>(selektor);
  const lebar = matchMedia(query);
  const kurangGerak = matchMedia('(prefers-reduced-motion: reduce)');
  if (!alur || !lebar.matches || kurangGerak.matches) return () => {};
  const pengamat = new IntersectionObserver((entri) => {
    if (!entri.some((e) => e.isIntersecting && e.intersectionRatio >= 1)) return;
    alur.classList.add(kelas);
    pengamat.disconnect();
  }, { threshold: 1 });
  pengamat.observe(alur);
  alur.addEventListener('animationend', (e) => {
    if (e.animationName === namaAnimasi && e.target === alur.lastElementChild?.querySelector('.alur__nomor')) alur.classList.remove(kelas);
  }, { signal });
  const hentikan = () => {
    if (lebar.matches && !kurangGerak.matches) return;
    pengamat.disconnect();
    alur.classList.remove(kelas);
  };
  lebar.addEventListener('change', hentikan, { signal });
  kurangGerak.addEventListener('change', hentikan, { signal });
  return () => pengamat.disconnect();
}

/** Sapuan gradient headline Inference sekali; dilewati pada reduced motion, forced colors, atau tanpa dukungan. */
function mountHeadline(signal: AbortSignal): void {
  const teks = document.querySelector<HTMLElement>('.inference-headline');
  const gerak = matchMedia('(prefers-reduced-motion: reduce)');
  const warnaPaksa = matchMedia('(forced-colors: active)');
  if (!teks || gerak.matches || warnaPaksa.matches || !CSS.supports('background-clip', 'text') || !CSS.supports('box-decoration-break', 'slice')) return;
  const selesai = new AbortController();
  const berhenti = AbortSignal.any([signal, selesai.signal]);
  const akhiri = () => {
    teks.classList.remove('inference-headline--reveal');
    selesai.abort();
  };
  teks.classList.add('inference-headline--reveal');
  teks.addEventListener('animationend', akhiri, { signal: berhenti });
  teks.addEventListener('animationcancel', akhiri, { signal: berhenti });
  gerak.addEventListener('change', akhiri, { signal: berhenti });
  warnaPaksa.addEventListener('change', akhiri, { signal: berhenti });
}

/** Tautan masuk ke #portofolio: gulir setelah font siap, kecuali pengguna sudah berinteraksi. */
function mountPortfolioHash(signal: AbortSignal): void {
  if (location.hash !== '#portofolio') return;
  const interaksi = ['wheel', 'touchstart', 'pointerdown', 'keydown'] as const;
  const batal = new AbortController();
  const berhenti = AbortSignal.any([signal, batal.signal]);
  interaksi.forEach((jenis) => window.addEventListener(jenis, () => batal.abort(), { once: true, passive: true, signal: berhenti }));
  void document.fonts.ready.then(() => requestAnimationFrame(() => {
    if (berhenti.aborted) return;
    batal.abort();
    if (location.hash === '#portofolio') document.getElementById('portofolio')?.scrollIntoView({ block: 'start', behavior: 'instant' });
  }));
}

export function mountPublic(): () => void {
  const pengendali = new AbortController();
  const { signal } = pengendali;
  const lepas = [mountNavbar(signal)];
  switch (document.body.dataset['halaman']) {
    case 'index': {
      const kanvas = document.querySelector<HTMLCanvasElement>('.hero__bola');
      const hero = document.querySelector<HTMLElement>('.hero');
      if (kanvas && hero) lepas.push(pasang(kanvas, { area: hero }).hancurkan);
      lepas.push(mountReveal('.lembar', 'lembar--masuk', 'lembar--siap'));
      break;
    }
    case 'pelatihan':
      lepas.push(mountFlow('.pelatihan-workflow .alur', '(min-width: 64rem)', 'alur--mengalir', 'pelatihan-indikator', signal));
      break;
    case 'inference':
      mountHeadline(signal);
      lepas.push(mountFlow('.alur--inference', '(min-width: 80rem)', 'inference--mengalir', 'inference-indikator', signal));
      mountPortfolioHash(signal);
      break;
    case 'tentang-kami':
      lepas.push(mountReveal('.tentang-ungkap', 'tentang--masuk'));
      break;
  }
  return () => {
    pengendali.abort();
    lepas.forEach((fn) => fn());
  };
}
