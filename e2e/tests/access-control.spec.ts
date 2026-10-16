import { test, expect, signedInAs } from '../lib/fixtures';
import { queryOne } from '../lib/db';

// Pages each role must NOT reach; the app answers with a redirect to the login page.
const forbidden = {
  anonymous: ['/moje-prihlasky', '/ucet', '/prihlasky', '/prihlasky-opakovat', '/dochazka', '/navstevnici'],
  user: ['/prihlasky', '/prihlasky-opakovat', '/dochazka', '/navstevnici'],
  reception: ['/prihlasky', '/dochazka'],
  instructor: ['/prihlasky', '/navstevnici'],
} as const;

const allowed = {
  user: ['/moje-prihlasky', '/ucet'],
  reception: ['/navstevnici'],
  instructor: ['/dochazka'],
  admin: ['/prihlasky', '/prihlasky-opakovat', '/dochazka', '/navstevnici', '/moje-prihlasky'],
} as const;

for (const [role, paths] of Object.entries(forbidden)) {
  test.describe(`${role} is kept out`, () => {
    if (role !== 'anonymous') signedInAs(role as 'user' | 'reception' | 'instructor');

    for (const path of paths) {
      test(path, async ({ page }) => {
        await page.goto(path);
        await expect(page).toHaveURL('/prihlaseni');
      });
    }
  });
}

for (const [role, paths] of Object.entries(allowed)) {
  test.describe(`${role} can open`, () => {
    signedInAs(role as keyof typeof allowed);

    for (const path of paths) {
      test(path, async ({ page }) => {
        const response = await page.goto(path);
        expect(response?.status()).toBe(200);
        await expect(page).toHaveURL(path);
      });
    }
  });
}

test.describe('write endpoints check the role too', () => {
  signedInAs('user');

  test('a parent cannot mark an application as paid', async ({ page, seed }) => {
    const { main } = seed.applications;
    const before = await queryOne<{ Paid: boolean }>(`SELECT Paid FROM t_course_application_form WHERE ID = @p1`, main.id);

    await page.goto('/moje-prihlasky');
    const response = await page.request.put(`/prihlaska/${main.id}?paid=${!before!.Paid}`, { maxRedirects: 0 });
    expect(response.status()).toBe(302);
    expect(response.headers()['location']).toBe('/prihlaseni');

    const after = await queryOne<{ Paid: boolean }>(`SELECT Paid FROM t_course_application_form WHERE ID = @p1`, main.id);
    expect(after?.Paid).toBe(before!.Paid);
  });

  test('a parent cannot write attendance', async ({ page, seed }) => {
    const response = await page.request.put(
      `/dochazka/${seed.applications.main.id}?status=present&courseId=${seed.courses.open.id}&lessonDate=${seed.lessonDate}`,
      { maxRedirects: 0 },
    );
    expect(response.status()).toBe(302);
    const record = await queryOne(
      `SELECT ID FROM t_course_attendance WHERE ID_ApplicationForm = @p1 AND ID_CreatedBy = @p2`,
      seed.applications.main.id,
      seed.users.user.id,
    );
    expect(record).toBeUndefined();
  });

  test('a parent cannot edit someone else\'s application', async ({ page, seed }) => {
    const response = await page.request.post('/prihlaska-edit', {
      form: { applicationFormId: String(seed.applications.main.id), firstName: 'X', lastName: 'Y', personalId: '1500000000', phone: '1', parentName: 'Z' },
      maxRedirects: 0,
    });
    expect(response.status()).toBe(302);
  });
});

test.describe('menu matches the role', () => {
  test.describe('parent', () => {
    signedInAs('user');
    test('sees only "Moje přihlášky"', async ({ page }) => {
      await page.goto('/moje-prihlasky');
      await expect(page.getByRole('link', { name: 'Moje přihlášky' })).toBeVisible();
      await expect(page.getByRole('link', { name: 'Všechny přihlášky' })).toHaveCount(0);
      await expect(page.getByRole('link', { name: 'Docházka' })).toHaveCount(0);
      await expect(page.getByRole('link', { name: 'Návštěvníci' })).toHaveCount(0);
    });
  });

  test.describe('reception', () => {
    signedInAs('reception');
    test('sees "Návštěvníci" but no admin items', async ({ page }) => {
      await page.goto('/navstevnici');
      await expect(page.getByRole('link', { name: 'Návštěvníci' })).toBeVisible();
      await expect(page.getByRole('link', { name: 'Všechny přihlášky' })).toHaveCount(0);
      await expect(page.getByRole('link', { name: 'Docházka' })).toHaveCount(0);
    });
  });
});
