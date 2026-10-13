import { test as base, expect, Locator, Page } from '@playwright/test';
import { readSeed, SeedData } from './seed';
import { Role, storageStateFor } from './constants';

type Fixtures = { seed: SeedData };

export const test = base.extend<Fixtures>({
  // The application form page loads the Meta pixel; keep tests off third-party networks.
  context: async ({ context }, use) => {
    await context.route(/^https:\/\/(connect|www)\.facebook\.(net|com)\//, (route) => route.abort());
    await use(context);
  },
  seed: async ({}, use) => {
    await use(readSeed());
  },
});

export { expect };

/** Use inside a describe block to run its tests signed in as the role (sessions are created by auth.setup.ts). */
export function signedInAs(role: Role) {
  test.use({ storageState: storageStateFor(role) });
}

/** Checkboxes re-render via htmx validation after a click, so assert the state instead of using check(). */
export async function tick(checkbox: Locator, checked = true): Promise<void> {
  if ((await checkbox.isChecked()) !== checked) {
    await checkbox.click();
  }
  await expect(checkbox).toBeChecked({ checked });
  await htmxSettled(checkbox.page());
}

/** Waits until htmx has finished settling swapped content (its triggers are attached only then). */
export async function htmxSettled(page: Page): Promise<void> {
  await expect(page.locator('.htmx-added, .htmx-settling, .htmx-swapping')).toHaveCount(0);
}

/** Table row (desktop layout) containing the text. */
export function row(page: Page, text: string): Locator {
  return page.locator('tbody tr').filter({ hasText: text });
}

/** Toast with the given text (an older toast may still be on the page, all of them use id="toast-info"). */
export function toast(page: Page, text: string): Locator {
  return page.locator('#toast-info').filter({ hasText: text });
}
