// Sesi tersimpan dari auth.setup membuka admin langsung, dengan role yang benar di header. Tanpa langkah agen.
import { test } from '@e2e-dev/web';
import { expect } from 'e2e';

for (const role of ['direktur', 'pegawai'] as const) {
  test(`admin terbuka langsung dengan sesi ${role}`, { session: role }, async ({ app, browser }) => {
    await app.open('/admin/');
    await expect(browser.locator('nav.samping')).toContainText('Berita');
    await expect(browser.locator('.atas__nama')).toContainText(role);
  });
}
