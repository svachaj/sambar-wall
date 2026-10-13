import fs from 'node:fs';
import { Download, Page } from '@playwright/test';
import { test, expect, row, signedInAs, tick, toast } from '../lib/fixtures';
import { queryOne } from '../lib/db';

const application = (id: number) =>
  queryOne<{ Paid: boolean; IsActive: boolean; HealthState: string; Phone: string }>(
    `SELECT Paid, IsActive, HealthState, Phone FROM t_course_application_form WHERE ID = @p1`,
    id,
  );

async function search(page: Page, text: string) {
  const response = page.waitForResponse((r) => r.url().includes('/prihlasky-hledat'));
  await page.getByPlaceholder('Hledat...').pressSequentially(text);
  await response;
}

async function readDownload(download: Download): Promise<Buffer> {
  return fs.readFileSync(await download.path());
}

test.describe('admin: all applications', () => {
  signedInAs('admin');

  test('lists and searches applications', async ({ page, seed }) => {
    const { main, deactivate } = seed.applications;
    await page.goto('/prihlasky');
    await expect(page.getByRole('heading', { name: 'Všechny přihlášky' })).toBeVisible();

    await search(page, main.lastName);
    await expect(row(page, main.lastName)).toBeVisible();
    await expect(row(page, deactivate.lastName)).toHaveCount(0);
    await expect(page).toHaveURL(`/prihlasky?search=${main.lastName}`);
  });

  test('searching by application number finds exactly that application', async ({ page, seed }) => {
    await page.goto(`/prihlasky?search=${seed.applications.main.id}`);
    await expect(row(page, seed.applications.main.lastName)).toBeVisible();
  });

  test('toggles the paid flag', async ({ page, seed }) => {
    const { main } = seed.applications;
    await page.goto(`/prihlasky?search=${main.lastName}`);
    const paid = row(page, main.lastName).locator('td:visible [hx-put]');

    await expect(paid).toHaveText('Ne');
    await paid.click();
    await expect(paid).toHaveText('ANO');
    await expect(toast(page, 'Úspěšně nastaveno.')).toBeVisible();
    expect((await application(main.id))?.Paid).toBe(true);

    await paid.click();
    await expect(paid).toHaveText('Ne');
    await expect.poll(async () => (await application(main.id))?.Paid).toBe(false);
  });

  test('edits an application', async ({ page, seed }) => {
    const { main } = seed.applications;
    await page.goto(`/prihlasky?search=${main.lastName}`);
    await row(page, main.lastName).getByRole('link', { name: `${main.lastName} ${main.firstName}` }).last().click();

    await expect(page.getByRole('heading', { name: 'Editace přihlášky' })).toBeVisible();
    await expect(page.getByLabel('Rodné číslo účastníka')).toHaveValue(main.personalId);
    await page.getByLabel('Zdravotní stav').fill('upraveno v e2e');
    await page.getByLabel('Telefonní číslo zákonného zástupce').fill('606000111');
    await page.getByRole('button', { name: 'Uložit' }).click();

    await expect(toast(page, 'Úspěšně aktualizováno.')).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Všechny přihlášky' })).toBeVisible();
    expect(await application(main.id)).toMatchObject({ HealthState: 'upraveno v e2e', Phone: '606000111', IsActive: true });
  });

  test('cancelling an edit keeps the application unchanged', async ({ page, seed }) => {
    const { main } = seed.applications;
    const before = await application(main.id);
    await page.goto(`/prihlaska-edit/${main.id}`);
    await page.getByLabel('Zdravotní stav').fill('tohle se neuloží');
    await page.getByRole('button', { name: 'Zrušit' }).click();

    await expect(page.getByRole('heading', { name: 'Všechny přihlášky' })).toBeVisible();
    expect(await application(main.id)).toEqual(before);
  });

  test('deactivating an application moves it to the inactive filter', async ({ page, seed }) => {
    const { deactivate } = seed.applications;
    await page.goto(`/prihlaska-edit/${deactivate.id}`);
    await tick(page.getByLabel('Aktivní'), false);
    await page.getByRole('button', { name: 'Uložit' }).click();
    await expect(toast(page, 'Úspěšně aktualizováno.')).toBeVisible();
    expect((await application(deactivate.id))?.IsActive).toBe(false);

    await page.goto(`/prihlasky?search=${deactivate.lastName}`);
    await expect(page.getByText('Nemáte žádné přihlášky.')).toBeVisible();

    const response = page.waitForResponse((r) => r.url().includes('/prihlasky-hledat'));
    await page.locator('select[name="status"]').selectOption('inactive');
    await response;
    await expect(row(page, deactivate.lastName).getByText('NEAKTIVNÍ', { exact: true })).toBeVisible();
  });

  test('exports applications to Excel', async ({ page }) => {
    await page.goto('/prihlasky');
    const downloading = page.waitForEvent('download');
    await page.getByRole('button', { name: 'Export vše' }).click();
    const download = await downloading;

    expect(download.suggestedFilename()).toBe('prihlasky-vse.xlsx');
    const file = await readDownload(download);
    expect(file.subarray(0, 2).toString()).toBe('PK'); // xlsx is a zip archive
  });

  test('exports emails that agreed to newsletters', async ({ page, seed }) => {
    await page.goto('/prihlasky');
    const downloading = page.waitForEvent('download');
    await page.getByTitle('Stáhnout emaily pro komerční sdělení').click();
    const download = await downloading;

    expect(download.suggestedFilename()).toBe('emaily-pro-komercni-sdeleni.txt');
    expect((await readDownload(download)).toString().split(';')).toContain(seed.visitor.email);
  });

  test('the newsletter export also works from a nested page', async ({ page, seed }) => {
    test.fail(true, 'Known bug: the header form uses the relative hx-get "commercial-emails-export-init", which resolves to /prihlaska-edit/commercial-emails-export-init here.');
    await page.goto(`/prihlaska-edit/${seed.applications.main.id}`);
    const downloading = page.waitForEvent('download', { timeout: 5_000 });
    await page.getByTitle('Stáhnout emaily pro komerční sdělení').click();
    await downloading;
  });

  test('the bulk "send again" page renders (the action itself is never run against the shared DB)', async ({ page }) => {
    // POST /prihlasky-opakovat re-creates applications for every "will continue" record in the database.
    await page.goto('/prihlasky-opakovat');
    await expect(page.getByRole('button', { name: 'Poslat hromadně přihlášky znovu na pololetí' })).toBeVisible();
  });
});
