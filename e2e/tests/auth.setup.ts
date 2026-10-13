import { test as setup } from '@playwright/test';
import { loginAs } from '../lib/auth';
import { ROLE_EMAILS, ROLES, storageStateFor } from '../lib/constants';

// Logs every seeded role in once through the real UI and stores the session for the specs.
for (const role of ROLES) {
  setup(`sign in as ${role}`, async ({ page }) => {
    await loginAs(page, ROLE_EMAILS[role]);
    await page.context().storageState({ path: storageStateFor(role) });
  });
}
