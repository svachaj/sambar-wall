import { Page } from '@playwright/test';
import { test, expect, row } from '../lib/fixtures';
import { loginWithoutEmail } from '../lib/auth';
import { deployedStateFor, Role } from '../lib/constants';

// Smoke tests for the deployed test site. The deployed app sends real emails (including copies to the admin
// inbox), so nothing here submits an application, toggles "paid" or asks for a login/agreement code.
// Exports are intercepted, so no data from the test database is downloaded.

const as = (role: Role) => test.use({ storageState: deployedStateFor(role) });

/** Stops a download request before it reaches the server and returns its URL. */
async function interceptDownload(page: Page, path: string, action: () => Promise<void>): Promise<URL> {
  const matches = (url: URL) => url.pathname === path;
  await page.route(matches, (route) => route.fulfill({ status: 204 }));
  const response = page.waitForResponse((r) => matches(new URL(r.url())));
  await action();
  const intercepted = await response;
  expect(intercepted.status(), 'the export reached the server instead of being intercepted').toBe(204);
  return new URL(intercepted.url());
}

test.describe('public pages', () => {
  test('the course list shows the seeded e2e course from the shared DB', async ({ page, seed }) => {
    const { open } = seed.courses;
    await page.goto('/');
    await expect(row(page, `${open.timeFrom} - ${open.timeTo}`).getByRole('button', { name: /volných míst/ }).last()).toBeVisible();
  });

  test('login, agreement and 404 pages render', async ({ page }) => {
    await page.goto('/prihlaseni');
    await expect(page.getByLabel('Zadej email pro přihlášení')).toBeVisible();
    await page.goto('/souhlas-s-provoznim-radem');
    await expect(page.getByLabel('Nejprve zadej svůj email, prosím')).toBeVisible();
    expect((await page.goto('/neexistuje'))?.status()).toBe(404);
  });

  test('inline field validation works (#70)', async ({ page }) => {
    await page.goto('/souhlas-s-provoznim-radem');
    const validated = page.waitForResponse((r) => r.url().endsWith('/validate-form-field'));
    await page.getByLabel('Nejprve zadej svůj email, prosím').fill('spatny-email');
    await page.getByLabel('Nejprve zadej svůj email, prosím').blur();
    expect((await validated).status()).toBe(200);
    await expect(page.getByText('Neplatný email')).toBeVisible();
  });
});

test.describe('login', () => {
  test('a code older than 10 minutes is rejected (#71)', async ({ request, seed }) => {
    const response = await loginWithoutEmail(request, seed.users.user.email, 11);
    expect(response.headers()['hx-redirect']).toBeUndefined();
    expect(await response.text()).toContain('Neplatný kód');
  });

  test('a code issued 9 minutes ago still works', async ({ request, seed }) => {
    const response = await loginWithoutEmail(request, seed.users.user.email, 9);
    expect(response.headers()['hx-redirect']).toBe('/');
  });

  test('protected pages redirect anonymous visitors to login', async ({ page }) => {
    for (const path of ['/moje-prihlasky', '/prihlasky', '/dochazka', '/navstevnici']) {
      await page.goto(path);
      await expect(page).toHaveURL('/prihlaseni');
    }
  });
});

test.describe('parent', () => {
  as('user');

  test('sees their seeded applications', async ({ page, seed }) => {
    await page.goto('/moje-prihlasky');
    const { main } = seed.applications;
    await expect(row(page, `${main.lastName} ${main.firstName}`)).toBeVisible();
  });

  test('an unknown course shows a friendly message (#72)', async ({ page }) => {
    await page.goto('/prihlaska/999999999');
    await expect(page.getByText('Požadovaný kurz nebyl nalezen', { exact: false })).toBeVisible();
  });

  test('cannot open admin pages', async ({ page }) => {
    await page.goto('/prihlasky');
    await expect(page).toHaveURL('/prihlaseni');
  });
});

test.describe('admin', () => {
  as('admin');

  test('finds the seeded application', async ({ page, seed }) => {
    await page.goto(`/prihlasky?search=${seed.applications.main.lastName}`);
    await expect(row(page, seed.applications.main.lastName)).toBeVisible();
  });

  test('the export does not execute script smuggled in through the search text (#72)', async ({ page }) => {
    const payload = `x'+(window.__xss=1)+'</script><script>window.__xss=2</script>`;
    await page.goto(`/prihlasky?search=${encodeURIComponent(payload)}`);
    await expect(page.getByPlaceholder('Hledat...')).toHaveValue(payload);

    const exportUrl = await interceptDownload(page, '/prihlasky-export', () =>
      page.getByRole('button', { name: 'Export vše' }).click(),
    );
    expect(await page.evaluate(() => (window as unknown as { __xss?: number }).__xss)).toBeUndefined();
    expect(exportUrl.searchParams.get('search')).toBe(payload);
  });

  test('the newsletter export works from a nested page (#72)', async ({ page, seed }) => {
    await page.goto(`/prihlaska-edit/${seed.applications.main.id}`);
    const init = page.waitForResponse((r) => new URL(r.url()).pathname === '/commercial-emails-export-init');
    const exportUrl = await interceptDownload(page, '/commercial-emails-export', () =>
      page.getByTitle('Stáhnout emaily pro komerční sdělení').click(),
    );
    expect((await init).status()).toBe(200);
    expect(exportUrl.pathname).toBe('/commercial-emails-export');
  });
});

test.describe('instructor', () => {
  as('instructor');

  test('sees the seeded child on the attendance sheet', async ({ page, seed }) => {
    await page.goto(`/dochazka?courseId=${seed.courses.open.id}&lessonDate=${seed.lessonDate}`);
    const { main } = seed.applications;
    await expect(row(page, `${main.lastName} ${main.firstName}`)).toBeVisible();
  });
});

test.describe('reception', () => {
  as('reception');

  test('finds the seeded wall visitor', async ({ page, seed }) => {
    await page.goto('/navstevnici');
    const response = page.waitForResponse((r) => r.url().includes('/navstevnici-hledat'));
    await page.getByPlaceholder('Hledat podle jména nebo emailu...').pressSequentially(seed.visitor.lastName);
    await response;
    await expect(row(page, seed.visitor.email)).toHaveCount(1);
  });
});
