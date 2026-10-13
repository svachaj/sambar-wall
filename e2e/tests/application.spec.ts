import { Page } from '@playwright/test';
import { test, expect, row, signedInAs, tick } from '../lib/fixtures';
import { query } from '../lib/db';
import { personalId } from '../lib/seed';

type Child = { firstName: string; lastName: string; personalId: string };

async function fillApplication(page: Page, child: Child) {
  await page.getByLabel('Jméno (koho přihlašuji)').fill(child.firstName);
  await page.getByLabel('Příjmení (koho přihlašuji)').fill(child.lastName);
  await page.getByLabel('Rodné číslo (koho přihlašuji)').fill(child.personalId);
  await page.getByLabel('Zdravotní stav').fill('alergie na pyl');
  await page.getByLabel('Jméno a příjmení zákonného zástupce').fill('Rodič Testovací');
  await page.getByLabel('Telefonní číslo zákonného zástupce').fill('777123456');
  await tick(page.getByLabel('Souhlasím se zpracováním osobních údajů'));
  await tick(page.getByLabel('Souhlasím s provozním řádem stěny'));
}

const submit = (page: Page) => page.getByRole('button', { name: 'Dokončit' }).click();

test.describe('course application', () => {
  signedInAs('user');

  test('a parent signs a child up and sees the application in "Moje přihlášky"', async ({ page, seed }) => {
    const { open } = seed.courses;
    const child = { firstName: 'Nový', lastName: `Prihlaska${seed.runId}`, personalId: personalId() };

    await page.goto(`/prihlaska/${open.id}`);
    await expect(page.getByRole('heading', { name: 'Přihláška na kurz' })).toBeVisible();
    await expect(page.getByText(`(${open.timeFrom} - ${open.timeTo})`)).toBeVisible();
    await fillApplication(page, child);
    await submit(page);

    await expect(page.getByRole('heading', { name: 'Přihláška byla úspěšně odeslána!' })).toBeVisible();

    const [saved] = await query<{ PersonalId: string; Phone: string; HealthState: string; EmailSent: boolean; Paid: boolean; IsActive: boolean; BirthYear: number; ID_ParentUser: number }>(
      `SELECT tcaf.PersonalId, tcaf.Phone, tcaf.HealthState, tcaf.EmailSent, tcaf.Paid, tcaf.IsActive, tsup.BirthYear, tsup.ID_ParentUser
       FROM t_course_application_form tcaf JOIN t_system_user_participant tsup ON tsup.ID = tcaf.ID_participant
       WHERE tcaf.ID_course = @p1 AND tsup.LastName = @p2`,
      open.id,
      child.lastName,
    );
    expect(saved).toMatchObject({
      PersonalId: child.personalId,
      Phone: '777123456',
      HealthState: 'alergie na pyl',
      EmailSent: true,
      Paid: false,
      IsActive: true,
      BirthYear: 2015,
      ID_ParentUser: seed.users.user.id,
    });

    await page.getByRole('link', { name: 'Přehled mých přihlášek' }).click();
    const application = row(page, `${child.lastName} ${child.firstName}`);
    await expect(application).toBeVisible();
    await expect(application.getByText('Ne', { exact: true }).last()).toBeVisible();
  });

  test('required fields and the personal ID length are validated', async ({ page, seed }) => {
    await page.goto(`/prihlaska/${seed.courses.open.id}`);
    await page.getByLabel('Rodné číslo (koho přihlašuji)').fill('1234');
    await submit(page);

    await expect(page.getByText('Povinné pole').first()).toBeVisible();
    await expect(page.getByText('Minimální délka je 9')).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Přihláška byla úspěšně odeslána!' })).toHaveCount(0);
  });

  test('the same child cannot be signed up for the same course twice', async ({ page, seed }) => {
    const existing = seed.applications.main;
    await page.goto(`/prihlaska/${seed.courses.open.id}`);
    await fillApplication(page, { ...existing, firstName: 'Duplicitní' });
    await submit(page);

    await expect(page.getByText('Tento účastník je již na tento kurz přihlášen.')).toBeVisible();
    const rows = await query(`SELECT ID FROM t_course_application_form WHERE ID_course = @p1 AND PersonalId = @p2`, seed.courses.open.id, existing.personalId);
    expect(rows).toHaveLength(1);
  });

  test('a full course cannot be applied for', async ({ page, seed }) => {
    await page.goto(`/prihlaska/${seed.courses.full.id}`);
    await expect(page.getByText('Kapacita kurzu se již bohužel vyčerpala. Zkuste prosím jiný kurz.')).toBeVisible();
    await expect(page.getByLabel('Jméno (koho přihlašuji)')).toHaveCount(0);
  });

  test('an unknown course shows a friendly message', async ({ page }) => {
    test.fail(true, 'Known bug: CheckCourseCapacity errors on a missing course, so the generic error page is shown instead of "Požadovaný kurz nebyl nalezen".');
    await page.goto('/prihlaska/999999999');
    await expect(page.getByText('Požadovaný kurz nebyl nalezen', { exact: false })).toBeVisible({ timeout: 3_000 });
  });
});

test('the application form requires login', async ({ page, seed }) => {
  await page.goto(`/prihlaska/${seed.courses.open.id}`);
  await expect(page).toHaveURL('/prihlaseni');
});
