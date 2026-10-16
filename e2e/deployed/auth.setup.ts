import { test as setup, expect } from '@playwright/test';
import { loginWithoutEmail } from '../lib/auth';
import { deployedStateFor, ROLE_EMAILS, ROLES } from '../lib/constants';

// Logs every seeded role in on the deployed site without sending an email. A successful login also proves
// that the deployed site uses the database from .env.
for (const role of ROLES) {
  setup(`sign in as ${role} on the deployed site`, async ({ request }) => {
    const response = await loginWithoutEmail(request, ROLE_EMAILS[role]);
    expect(response.ok()).toBeTruthy();
    expect(response.headers()['hx-redirect'], 'login failed: is the deployed site using the DB from .env?').toBeTruthy();
    await request.storageState({ path: deployedStateFor(role) });
  });
}
