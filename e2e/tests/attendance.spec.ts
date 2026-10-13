import { Page } from '@playwright/test';
import { test, expect, row, signedInAs, toast } from '../lib/fixtures';
import { storageStateFor } from '../lib/constants';
import { queryOne } from '../lib/db';

const attendance = (applicationId: number, lessonDate: string) =>
  queryOne<{ Present: boolean; IsActive: boolean }>(
    `SELECT TOP 1 Present, IsActive FROM t_course_attendance
     WHERE ID_ApplicationForm = @p1 AND LessonDate = CAST(@p2 AS date) ORDER BY ID DESC`,
    applicationId,
    lessonDate,
  );

// the attendance buttons of the desktop table cell
const buttons = (page: Page, name: string) => row(page, name).locator('td:visible').last();

test.describe('attendance', () => {
  signedInAs('instructor');

  test('picking a course shows its participants', async ({ page, seed }) => {
    const { open } = seed.courses;
    await page.goto(`/dochazka?lessonDate=${seed.lessonDate}`);
    await page.getByLabel('Kurz').selectOption(String(open.id));

    await expect(page).toHaveURL(new RegExp(`courseId=${open.id}`));
    await expect(row(page, `${seed.applications.main.lastName} ${seed.applications.main.firstName}`)).toBeVisible();
    await expect(page.getByText('Pozor:')).toHaveCount(0); // the course runs on the chosen weekday
  });

  test('marks a child present, absent and clears it again; the parent sees the history', async ({ page, browser, seed }) => {
    const { main } = seed.applications;
    const name = `${main.lastName} ${main.firstName}`;
    await page.goto(`/dochazka?courseId=${seed.courses.open.id}&lessonDate=${seed.lessonDate}`);

    await buttons(page, name).getByRole('button', { name: 'Byl', exact: true }).click();
    await expect(toast(page, 'Úspěšně nastaveno.')).toBeVisible();
    await expect(buttons(page, name).getByRole('button', { name: 'Byl', exact: true })).toHaveCount(0); // now the active state
    await expect.poll(() => attendance(main.id, seed.lessonDate)).toMatchObject({ Present: true, IsActive: true });

    const parent = await browser.newContext({ storageState: storageStateFor('user') });
    const parentPage = await parent.newPage();
    await parentPage.goto('/moje-prihlasky');
    const [y, m, d] = seed.lessonDate.split('-').map(Number);
    const history = row(parentPage, `${d}.${m}.${y}`).filter({ hasText: name });
    await expect(history.getByText('Byl', { exact: true })).toBeVisible();
    await parent.close();

    await buttons(page, name).getByRole('button', { name: 'Nebyl', exact: true }).click();
    await expect(buttons(page, name).getByRole('button', { name: 'Nebyl', exact: true })).toHaveCount(0);
    await expect.poll(() => attendance(main.id, seed.lessonDate)).toMatchObject({ Present: false, IsActive: true });

    await buttons(page, name).getByRole('button', { name: '—', exact: true }).click();
    await expect(buttons(page, name).getByRole('button', { name: '—', exact: true })).toHaveCount(0);
    await expect.poll(() => attendance(main.id, seed.lessonDate)).toMatchObject({ IsActive: false });
  });

  test('warns when the course does not run on the chosen date', async ({ page, seed }) => {
    const tuesday = new Date(`${seed.lessonDate}T12:00:00`);
    tuesday.setDate(tuesday.getDate() + 1);
    const iso = tuesday.toISOString().slice(0, 10);
    await page.goto(`/dochazka?courseId=${seed.courses.open.id}&lessonDate=${iso}`);
    await expect(page.getByText('Pozor:')).toBeVisible();
  });

  test('exports the lesson and the whole period to Excel', async ({ page, seed }) => {
    await page.goto(`/dochazka?courseId=${seed.courses.open.id}&lessonDate=${seed.lessonDate}`);

    let downloading = page.waitForEvent('download');
    await page.getByRole('link', { name: 'Export lekce XLSX' }).click();
    let download = await downloading;
    expect(download.suggestedFilename()).toBe(`dochazka_${seed.lessonDate}.xlsx`);

    downloading = page.waitForEvent('download');
    await page.getByRole('button', { name: 'Export vše XLSX' }).click();
    download = await downloading;
    expect(download.suggestedFilename()).toMatch(/^dochazka_\d{4}-\d{2}-\d{2}_\d{4}-\d{2}-\d{2}\.xlsx$/);
  });
});
