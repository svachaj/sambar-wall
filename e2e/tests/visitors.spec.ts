import { test, expect, row, signedInAs } from '../lib/fixtures';

test.describe('wall visitors', () => {
  signedInAs('reception');

  test('reception finds a visitor by name', async ({ page, seed }) => {
    const { visitor } = seed;
    await page.goto('/navstevnici');
    await expect(page.getByText('Žádní návštěvníci nenalezeni')).toBeVisible();

    const response = page.waitForResponse((r) => r.url().includes('/navstevnici-hledat'));
    await page.getByPlaceholder('Hledat podle jména nebo emailu...').pressSequentially(visitor.lastName);
    await response;

    const found = row(page, visitor.email);
    await expect(found).toHaveCount(1);
    await expect(found.getByText(visitor.lastName, { exact: true })).toBeVisible();
  });

  test('search ignores diacritics', async ({ page, seed }) => {
    await page.goto('/navstevnici');
    const response = page.waitForResponse((r) => r.url().includes('/navstevnici-hledat'));
    // seeded first name is "Návštěvník"
    await page.getByPlaceholder('Hledat podle jména nebo emailu...').pressSequentially(`navstevnik`);
    await response;
    await expect(row(page, seed.visitor.email)).toHaveCount(1);
  });
});
