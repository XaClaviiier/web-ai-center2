// Setiap tautan menu utama mengarah ke halaman yang ada (bukan 404). Tanpa langkah agen.
import { test } from '@e2e-dev/web';
import { expect } from 'e2e';

test('semua tautan menu utama di beranda dapat dibuka', async ({ app, browser }) => {
  await app.open('/');
  const tautan: string[] = await browser.evaluate(() =>
    [...new Set([...document.querySelectorAll('header a[href^="/"], nav a[href^="/"]')].map((a) => a.getAttribute('href')!))],
  );
  expect(tautan.length).toBeGreaterThan(0);
  const rusak: string[] = [];
  for (const href of tautan) {
    const status: number = await browser.evaluate(async (u: string) => (await fetch(u, { method: 'GET' })).status, href);
    if (status >= 400) rusak.push(`${href} -> ${status}`);
  }
  expect(rusak).toEqual([]);
});
