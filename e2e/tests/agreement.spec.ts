import { Page } from '@playwright/test';
import { test, expect, htmxSettled, tick, toast } from '../lib/fixtures';
import { insertCode, latestCode } from '../lib/auth';
import { E2E_DOMAIN } from '../lib/constants';
import { queryOne } from '../lib/db';

const newEmail = (runId: string) => `e2e-wall-${runId}-${Math.random().toString(36).slice(2, 8)}@${E2E_DOMAIN}`;

async function startAgreement(page: Page, email: string) {
  await page.goto('/souhlas-s-provoznim-radem');
  await page.getByLabel('Nejprve zadej svůj email, prosím').fill(email);
  await page.getByRole('button', { name: 'Pokračovat' }).click();
  await expect(page.getByText('Na zadaný email byl odeslán ověřovací kód')).toBeVisible();
  await htmxSettled(page);
}

async function fillPersonalDetails(page: Page, lastName: string) {
  await page.getByLabel('Jméno', { exact: true }).fill('Lezec');
  await page.getByLabel('Příjmení', { exact: true }).fill(lastName);
  await page.getByLabel('Datum narození (den.měsíc.rok)').fill('1.2.1990');
}

const wallUser = (email: string) =>
  queryOne<{ FirstName: string; LastName: string; BirthDate: Date; commercial_confirmed: boolean; GDPR_confirmed: boolean; Rules_confirmed: boolean }>(
    `SELECT FirstName, LastName, BirthDate, commercial_confirmed, GDPR_confirmed, Rules_confirmed FROM t_system_wall_user WHERE Email = @p1`,
    email,
  );

test.describe('wall rules agreement', () => {
  test('a visitor confirms the agreement with the emailed code', async ({ page, seed }) => {
    const email = newEmail(seed.runId);
    const lastName = `Souhlas${seed.runId}`;
    await startAgreement(page, email);

    await fillPersonalDetails(page, lastName);
    await page.getByLabel('Ověřovací kód z emailu').fill(await latestCode(email));
    await tick(page.getByLabel('Souhlasím s provozním řádem stěny'));
    await tick(page.getByLabel('Souhlasím se zpracováním osobních údajů'));
    await page.getByRole('button', { name: 'Dokončit' }).click();

    await expect(toast(page, 'Souhlas s provozním řádem byl úspěšně dokončen.')).toBeVisible();
    // the form resets for the next visitor
    await expect(page.getByLabel('Nejprve zadej svůj email, prosím')).toBeVisible();

    const saved = await wallUser(email);
    expect(saved).toMatchObject({ FirstName: 'Lezec', LastName: lastName, GDPR_confirmed: true, Rules_confirmed: true, commercial_confirmed: false });
    expect(saved!.BirthDate.toISOString().slice(0, 10)).toBe('1990-02-01');
  });

  test('the optional newsletter consent is stored', async ({ page, seed }) => {
    const email = newEmail(seed.runId);
    await startAgreement(page, email);

    await fillPersonalDetails(page, `Novinky${seed.runId}`);
    await page.getByLabel('Ověřovací kód z emailu').fill(await latestCode(email));
    await tick(page.getByLabel('Souhlasím s provozním řádem stěny'));
    await tick(page.getByLabel('Souhlasím se zpracováním osobních údajů'));
    await tick(page.getByLabel('Chci být v obraze – novinky ze stěny i světa lezení'));
    await page.getByRole('button', { name: 'Dokončit' }).click();

    await expect(toast(page, 'Souhlas s provozním řádem byl úspěšně dokončen.')).toBeVisible();
    expect((await wallUser(email))?.commercial_confirmed).toBe(true);
  });

  test('a wrong code is rejected and nothing is stored', async ({ page, seed }) => {
    const email = newEmail(seed.runId);
    await startAgreement(page, email);

    await fillPersonalDetails(page, `Spatny${seed.runId}`);
    await page.getByLabel('Ověřovací kód z emailu').fill('0000');
    await tick(page.getByLabel('Souhlasím s provozním řádem stěny'));
    await tick(page.getByLabel('Souhlasím se zpracováním osobních údajů'));
    await page.getByRole('button', { name: 'Dokončit' }).click();

    await expect(page.getByText('Chybný ověřovací kód').first()).toBeVisible();
    expect(await wallUser(email)).toBeUndefined();
  });

  test('a code older than 10 minutes is rejected', async ({ page, seed }) => {
    const email = newEmail(seed.runId);
    await startAgreement(page, email);
    await insertCode(email, '4321', 11);

    await fillPersonalDetails(page, `Stary${seed.runId}`);
    await page.getByLabel('Ověřovací kód z emailu').fill('4321');
    await tick(page.getByLabel('Souhlasím s provozním řádem stěny'));
    await tick(page.getByLabel('Souhlasím se zpracováním osobních údajů'));
    await page.getByRole('button', { name: 'Dokončit' }).click();

    await expect(page.getByText('Chybný ověřovací kód').first()).toBeVisible();
    expect(await wallUser(email)).toBeUndefined();
  });

  test('both mandatory consents are required', async ({ page, seed }) => {
    const email = newEmail(seed.runId);
    await startAgreement(page, email);

    await fillPersonalDetails(page, `Bez${seed.runId}`);
    await page.getByLabel('Ověřovací kód z emailu').fill(await latestCode(email));
    await page.getByRole('button', { name: 'Dokončit' }).click();

    await expect(page.getByText('Musíte souhlasit s provozním řádem stěny')).toBeVisible();
    await expect(page.getByText('Musíte souhlasit se zpracováním osobních údajů')).toBeVisible();
    expect(await wallUser(email)).toBeUndefined();
  });

  test('invalid values are rejected on submit', async ({ page, seed }) => {
    await page.goto('/souhlas-s-provoznim-radem');
    await page.getByLabel('Nejprve zadej svůj email, prosím').fill('spatny-email');
    await page.getByRole('button', { name: 'Pokračovat' }).click();
    await expect(page.getByText('Neplatný email')).toBeVisible();

    await startAgreement(page, newEmail(seed.runId));
    await page.getByLabel('Datum narození (den.měsíc.rok)').fill('1990');
    await page.getByRole('button', { name: 'Dokončit' }).click();
    await expect(page.getByText('Neplatný datum - akceptovaný formát je: den.měsíc.rok')).toBeVisible();
  });

  test('a field is validated as soon as the visitor leaves it', async ({ page, seed }) => {
    // POST /validate-form-field re-renders the field; it used to return 500 after the Echo v4.16 upgrade
    const validated = () => page.waitForResponse((r) => r.url().endsWith('/validate-form-field'));

    await page.goto('/souhlas-s-provoznim-radem');
    const emailInput = page.getByLabel('Nejprve zadej svůj email, prosím');
    let response = validated();
    await emailInput.fill('spatny-email');
    await emailInput.blur();
    expect((await response).status()).toBe(200);
    await expect(page.getByText('Neplatný email')).toBeVisible();
    await htmxSettled(page);

    response = validated();
    await page.getByLabel('Nejprve zadej svůj email, prosím').fill(newEmail(seed.runId));
    await page.getByLabel('Nejprve zadej svůj email, prosím').blur();
    await response;
    await expect(page.getByText('Neplatný email')).toHaveCount(0);

    await startAgreement(page, newEmail(seed.runId));
    const birthDate = page.getByLabel('Datum narození (den.měsíc.rok)');
    await birthDate.fill('1990');
    await birthDate.blur();
    await expect(page.getByText('Neplatný datum - akceptovaný formát je: den.měsíc.rok')).toBeVisible();

    // unticking a mandatory consent shows its message right away
    const rules = page.getByLabel('Souhlasím s provozním řádem stěny');
    await tick(rules);
    await tick(page.getByLabel('Souhlasím s provozním řádem stěny'), false);
    await expect(page.getByText('Musíte souhlasit s provozním řádem stěny')).toBeVisible();
  });
});
