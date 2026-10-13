import crypto from 'node:crypto';
import { expect, Page } from '@playwright/test';
import { query, queryOne } from './db';

/** Latest code the app stored for this email (the mock email service drops the actual email). */
export async function latestCode(email: string): Promise<string> {
  const row = await queryOne<{ Code: string }>(
    `SELECT TOP 1 Code FROM t_system_registration_code WHERE Email = @p1 ORDER BY ID DESC`,
    email.toLowerCase(),
  );
  if (!row) throw new Error(`No verification code stored for ${email}`);
  return row.Code;
}

/**
 * Inserts a code directly, e.g. to build a magic link or to simulate an expired code. Like the app,
 * CreateDate comes from the database clock.
 */
export async function insertCode(email: string, code: string, minutesAgo = 0): Promise<void> {
  await query(
    `INSERT INTO t_system_registration_code (Email, Code, CreateDate) VALUES (@p1, @p2, DATEADD(second, -@p3, SYSDATETIME()))`,
    email.toLowerCase(),
    code,
    Math.round(minutesAgo * 60),
  );
}

/** Same format as utils.Encrypt: hex(nonce | AES-GCM ciphertext | tag) keyed by APP_CRYPTO_KEY. */
export function magicLinkPath(email: string, code: string): string {
  const key = Buffer.from(process.env.APP_CRYPTO_KEY ?? '');
  const nonce = crypto.randomBytes(12);
  const cipher = crypto.createCipheriv(`aes-${key.length * 8}-gcm` as crypto.CipherGCMTypes, key, nonce);
  const encrypted = Buffer.concat([cipher.update(`${email};${code}`, 'utf8'), cipher.final(), cipher.getAuthTag()]);
  return `/sign-me-in?c=${Buffer.concat([nonce, encrypted]).toString('hex')}`;
}

/** Step 1 of the login form: request a code for the email. */
export async function requestCode(page: Page, email: string): Promise<void> {
  await page.goto('/prihlaseni');
  await page.getByLabel('Zadej email pro přihlášení').fill(email);
  await page.getByRole('button', { name: 'Pokračovat' }).click();
  await expect(page.getByRole('heading', { name: 'Dokončení přihlášení' })).toBeVisible();
}

/** Step 2 of the login form: submit a code. */
export async function submitCode(page: Page, code: string): Promise<void> {
  await page.getByLabel('Ověřovací kód z emailu').fill(code);
  await page.getByRole('button', { name: 'Přihlásit se' }).click();
}

/** Full login through the real UI; the code is read from the database. Resolves once the app redirected away. */
export async function loginAs(page: Page, email: string): Promise<void> {
  await requestCode(page, email);
  await submitCode(page, await latestCode(email));
  await page.waitForURL((url) => url.pathname !== '/prihlaseni');
}
