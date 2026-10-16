import { test, expect, toast } from '../lib/fixtures';
import { insertCode, latestCode, loginAs, magicLinkPath, requestCode, submitCode } from '../lib/auth';
import { E2E_DOMAIN } from '../lib/constants';
import { queryOne } from '../lib/db';

const newEmail = (runId: string) => `e2e-${runId}-${Math.random().toString(36).slice(2, 8)}@${E2E_DOMAIN}`;
const signOutButton = (page: import('@playwright/test').Page) => page.locator('button[hx-get="/sign-out"]');

test.describe('login with an email code', () => {
  test('a new email gets an account and is signed in', async ({ page, seed }) => {
    const email = newEmail(seed.runId);

    await requestCode(page, email);
    await expect(toast(page, `Na zadaný email ${email} byl odeslán ověřovací kód.`)).toBeVisible();
    await submitCode(page, await latestCode(email));

    await expect(page).toHaveURL('/');
    await expect(page.getByRole('link', { name: 'Moje přihlášky' })).toBeVisible();
    const user = await queryOne(`SELECT ID FROM t_system_user WHERE UserName = @p1 AND IsEnabled = 1`, email);
    expect(user).toBeTruthy();
  });

  test('an invalid email is rejected in step 1', async ({ page }) => {
    await page.goto('/prihlaseni');
    await page.getByLabel('Zadej email pro přihlášení').fill('neni-email');
    await page.getByRole('button', { name: 'Pokračovat' }).click();
    await expect(page.getByText('Neplatný email')).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Dokončení přihlášení' })).toHaveCount(0);
  });

  test('a wrong code is rejected', async ({ page, seed }) => {
    await requestCode(page, newEmail(seed.runId));
    await submitCode(page, '00000');
    await expect(page.getByText('Neplatný kód').first()).toBeVisible();
    await expect(page).toHaveURL('/prihlaseni');
  });

  test('a code older than 10 minutes is rejected', async ({ page, seed }) => {
    const email = newEmail(seed.runId);
    await requestCode(page, email);
    await insertCode(email, '55555', 11);
    await submitCode(page, '55555');
    await expect(page.getByText('Neplatný kód').first()).toBeVisible();
    await expect(page).toHaveURL('/prihlaseni');
  });

  test('a code issued 9 minutes ago still works', async ({ page, seed }) => {
    const email = newEmail(seed.runId);
    await requestCode(page, email);
    await insertCode(email, '44444', 9);
    await submitCode(page, '44444');
    await expect(page).toHaveURL('/');
  });

  test('a code can be used only once', async ({ page, browser, seed }) => {
    const email = newEmail(seed.runId);
    await requestCode(page, email);
    const code = await latestCode(email);
    await submitCode(page, code);
    await expect(page).toHaveURL('/');

    const other = await browser.newPage();
    await requestCode(other, email);
    await submitCode(other, code);
    await expect(other.getByText('Neplatný kód').first()).toBeVisible();
    await other.close();
  });

  test('admins land on the home page with the admin menu', async ({ page, seed }) => {
    await loginAs(page, seed.users.admin.email);
    await expect(page).toHaveURL('/');
    await expect(page.getByRole('link', { name: 'Všechny přihlášky' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Docházka' })).toBeVisible();
  });

  test('instructors land on the attendance page without "Moje přihlášky"', async ({ page, seed }) => {
    await loginAs(page, seed.users.instructor.email);
    await expect(page).toHaveURL(/\/dochazka$/);
    await expect(page.getByRole('heading', { name: 'Docházka dětí' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Moje přihlášky' })).toHaveCount(0);
  });

  test('a protected page asks for login and returns to it afterwards', async ({ page, seed }) => {
    await page.goto('/moje-prihlasky');
    await expect(page).toHaveURL('/prihlaseni');
    await loginAs(page, seed.users.user.email);
    await expect(page).toHaveURL('/moje-prihlasky');
    await expect(page.getByRole('heading', { name: 'Moje přihlášky' })).toBeVisible();
  });

  test('sign-out ends the session', async ({ page, seed }) => {
    await loginAs(page, seed.users.user.email);
    await signOutButton(page).click();
    await expect(page.getByRole('button', { name: /Přihlášení/ })).toBeVisible();

    await page.goto('/moje-prihlasky');
    await expect(page).toHaveURL('/prihlaseni');
  });

  test('the account page shows the signed-in email', async ({ page, seed }) => {
    await loginAs(page, seed.users.user.email);
    await page.goto('/ucet');
    await expect(page.getByText(seed.users.user.email)).toBeVisible();
  });
});

test.describe('login with a magic link', () => {
  test('opening the link signs the user in', async ({ page, seed }) => {
    const email = newEmail(seed.runId);
    await insertCode(email, '24680');

    await page.goto(magicLinkPath(email, '24680'));

    await expect(page).toHaveURL('/kurzy');
    await expect(page.getByRole('link', { name: 'Moje přihlášky' })).toBeVisible();
  });

  test('a plain GET of the link (e.g. an email scanner) does not use up the code', async ({ page, request, seed }) => {
    const email = newEmail(seed.runId);
    await insertCode(email, '13579');
    const link = magicLinkPath(email, '13579');

    const prefetch = await request.get(link);
    expect(prefetch.ok()).toBeTruthy();

    await page.goto(link);
    await expect(page).toHaveURL('/kurzy');
    await expect(page.getByRole('link', { name: 'Moje přihlášky' })).toBeVisible();
  });

  test('a used link sends the user to login with an "expired" message', async ({ page, browser, seed }) => {
    const email = newEmail(seed.runId);
    await insertCode(email, '11223');
    const link = magicLinkPath(email, '11223');
    await page.goto(link);
    await expect(page).toHaveURL('/kurzy');

    const other = await browser.newPage();
    await other.goto(link);
    await expect(other).toHaveURL('/prihlaseni?expired=true');
    await expect(toast(other, 'Kód již vypršel, zadejte svůj email znovu.')).toBeVisible();
    await other.close();
  });

  test('a tampered link is rejected', async ({ page }) => {
    await page.goto('/sign-me-in?c=deadbeef');
    await expect(page).toHaveURL('/prihlaseni?expired=true');
  });
});
