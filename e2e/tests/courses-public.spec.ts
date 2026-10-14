import { test, expect, row } from '../lib/fixtures';
import { loginAs } from '../lib/auth';

test.describe('public course list', () => {
  test('lists the open e2e course with free places and the full one as full', async ({ page, seed }) => {
    const { open, full } = seed.courses;
    await page.goto('/');

    const openRow = row(page, `${open.timeFrom} - ${open.timeTo}`);
    await expect(openRow.getByRole('button', { name: /volných míst/ }).last()).toBeVisible();
    // 50 places minus the two seeded applications
    await expect(openRow.getByRole('button').last()).toContainText(`${open.capacity - 2} volných míst`);

    const fullRow = row(page, `${full.timeFrom} - ${full.timeTo}`);
    await expect(fullRow.getByText('Kurz naplněn').last()).toBeVisible();
    await expect(fullRow.getByRole('button')).toHaveCount(0);
  });

  test('/kurzy shows the same list as /', async ({ page, seed }) => {
    await page.goto('/kurzy');
    await expect(row(page, `${seed.courses.open.timeFrom} - ${seed.courses.open.timeTo}`)).toHaveCount(1);
  });

  test('signing up for a course as a visitor asks to log in and then opens the application form', async ({ page, seed }) => {
    const { open } = seed.courses;
    await page.goto('/');
    await row(page, `${open.timeFrom} - ${open.timeTo}`).getByRole('button', { name: /volných míst/ }).last().click();

    await expect(page.getByRole('heading', { name: 'Přihlášení' }).first()).toBeVisible();
    await loginAs(page, seed.users.user.email);

    await expect(page).toHaveURL(`/prihlaska/${open.id}`);
    await expect(page.getByRole('heading', { name: 'Přihláška na kurz' })).toBeVisible();
  });

  test('anonymous header offers login and the wall agreement', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByRole('button', { name: /Přihlášení/ })).toBeVisible();
    await page.locator('a[href="/souhlas-s-provoznim-radem"]').click();
    await expect(page.getByRole('heading', { name: 'Souhlas s provozním řádem stěny' })).toBeVisible();
  });
});

test.describe('error pages', () => {
  test('/404 renders the not-found page', async ({ page }) => {
    await page.goto('/404');
    await expect(page.getByRole('heading', { name: 'Page not found' })).toBeVisible();
  });

  test('an unknown URL renders the not-found page with status 404', async ({ page }) => {
    const response = await page.goto('/tahle-stranka-neexistuje');
    expect(response?.status()).toBe(404);
    await expect(page.getByText('404').first()).toBeVisible();
  });
});
