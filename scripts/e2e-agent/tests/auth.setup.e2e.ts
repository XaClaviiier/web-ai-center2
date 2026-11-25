// Login sekali per role lewat form Keycloak sungguhan; tes lain memulihkan sesi bernama.
import { test } from '@e2e-dev/web';
import { expect, credentials } from 'e2e';

for (const role of ['direktur', 'pegawai'] as const) {
  test.setup(`login sebagai ${role}`, { sessions: [role] }, async ({ app, browser, session }) => {
    const akun = credentials.user(role);
    await app.open('/admin/');
    await expect(browser.locator('#username')).toBeVisible();
    await browser.locator('#username').fill(akun.username);
    await browser.locator('#password').fill(akun.password);
    await browser.locator('#kc-login').tap();
    await expect(browser.locator('nav.samping')).toBeVisible();
    await expect(browser.locator('.atas__nama')).toContainText(role);
    await session.save(role);
  });
}
