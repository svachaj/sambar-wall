import fs from 'node:fs';
import { dbName, query, queryOne } from './db';
import {
  AUTH_DIR,
  E2E_COURSE_TIME_TO,
  E2E_COURSE_VALID_TO,
  E2E_DOMAIN,
  E2E_EMAIL_LIKE,
  ROLE_CODES,
  ROLE_EMAILS,
  ROLES,
  Role,
  SEED_FILE,
} from './constants';

export type SeedCourse = { id: number; timeFrom: string; timeTo: string; capacity: number };
export type SeedApplication = { id: number; firstName: string; lastName: string; personalId: string };
export type SeedData = {
  runId: string;
  users: Record<Role, { id: number; email: string }>;
  courses: { open: SeedCourse; full: SeedCourse };
  applications: { main: SeedApplication; deactivate: SeedApplication; full: SeedApplication };
  visitor: { email: string; firstName: string; lastName: string };
  // a Monday (the e2e courses run on Mondays) used for attendance
  lessonDate: string;
};

/**
 * The suite writes to a shared database, so refuse to touch anything unless the developer has
 * explicitly allowed the database the app is configured for (E2E_ALLOW_DB in e2e/.env).
 */
export function assertSafeDatabase(): void {
  const allowed = process.env.E2E_ALLOW_DB;
  if (!allowed) {
    throw new Error('E2E_ALLOW_DB is not set. Copy e2e/.env.example to e2e/.env and set it to the test database name.');
  }
  if (dbName() !== allowed) {
    throw new Error(`Refusing to run: DB_NAME is "${dbName()}" but E2E_ALLOW_DB is "${allowed}".`);
  }
}

/**
 * Deletes everything the suite may have created: rows owned by @e2e-sambar.test users and the e2e courses
 * (recognised by both their e2e creator and their 23:59/2099 marker). Safe to run repeatedly.
 */
export async function purge(): Promise<void> {
  await query(
    `SET XACT_ABORT ON;
    BEGIN TRAN;
    DECLARE @users TABLE (ID int);
    INSERT INTO @users SELECT ID FROM t_system_user WHERE UserName LIKE @p1;

    DECLARE @courses TABLE (ID int);
    INSERT INTO @courses SELECT ID FROM t_course
      WHERE ID_CreatedBy IN (SELECT ID FROM @users) AND ValidTo >= @p2 AND CAST(TimeTo AS time) = @p3;

    DECLARE @forms TABLE (ID int);
    INSERT INTO @forms SELECT ID FROM t_course_application_form
      WHERE ID_course IN (SELECT ID FROM @courses)
         OR ID_CreatedBy IN (SELECT ID FROM @users)
         OR ID_participant IN (SELECT ID FROM t_system_user_participant WHERE ID_ParentUser IN (SELECT ID FROM @users));

    DELETE FROM t_course_attendance WHERE ID_ApplicationForm IN (SELECT ID FROM @forms) OR ID_Course IN (SELECT ID FROM @courses);
    DELETE FROM t_course_application_form WHERE ID IN (SELECT ID FROM @forms);
    DELETE FROM t_system_user_participant WHERE ID_ParentUser IN (SELECT ID FROM @users);
    DELETE FROM t_course WHERE ID IN (SELECT ID FROM @courses);
    DELETE FROM t_system_wall_user WHERE Email LIKE @p1;
    DELETE FROM t_system_registration_code WHERE Email LIKE @p1;
    DELETE FROM t_system_role_user WHERE UserID IN (SELECT ID FROM @users);
    DELETE FROM t_system_user WHERE ID IN (SELECT ID FROM @users);
    COMMIT;`,
    E2E_EMAIL_LIKE,
    E2E_COURSE_VALID_TO,
    E2E_COURSE_TIME_TO,
  );
}

/** Counts e2e rows left in every table the suite touches; all zeros after a purge. */
export async function countLeftovers(): Promise<Record<string, number>> {
  const row = await queryOne<Record<string, number>>(
    `SELECT
      (SELECT COUNT(*) FROM t_system_user WHERE UserName LIKE @p1) AS users,
      (SELECT COUNT(*) FROM t_system_role_user WHERE UserID IN (SELECT ID FROM t_system_user WHERE UserName LIKE @p1)) AS roleUsers,
      (SELECT COUNT(*) FROM t_system_registration_code WHERE Email LIKE @p1) AS codes,
      (SELECT COUNT(*) FROM t_system_wall_user WHERE Email LIKE @p1) AS wallUsers,
      (SELECT COUNT(*) FROM t_course_application_form WHERE Email LIKE @p1) AS applications,
      (SELECT COUNT(*) FROM t_course WHERE ValidTo >= @p2 AND CAST(TimeTo AS time) = @p3) AS courses`,
    E2E_EMAIL_LIKE,
    E2E_COURSE_VALID_TO,
    E2E_COURSE_TIME_TO,
  );
  return row ?? {};
}

async function createUser(email: string, roleCode?: string): Promise<number> {
  const user = await queryOne<{ ID: number }>(
    `INSERT INTO t_system_user (Email, UserName, CreateDate, IsActivated, IsDeleted, IsEnabled)
     OUTPUT INSERTED.ID VALUES (@p1, @p1, GETDATE(), 1, 0, 1)`,
    email,
  );
  if (!user) throw new Error(`Could not create user ${email}`);
  if (roleCode) {
    const linked = await query<{ UserID: number }>(
      `INSERT INTO t_system_role_user (RoleID, UserID) OUTPUT INSERTED.UserID SELECT ID, @p2 FROM t_system_role WHERE Code = @p1`,
      roleCode,
      user.ID,
    );
    if (linked.length !== 1) throw new Error(`Role ${roleCode} not found in t_system_role`);
  }
  return user.ID;
}

/**
 * Courses cannot be created in the app and t_course has legacy NOT NULL columns, so an existing active
 * course is cloned (column list read from the schema) and only the fields the tests rely on are overridden.
 */
async function cloneCourse(templateId: number, creatorId: number, mondayId: number, timeFrom: string, capacity: number): Promise<number> {
  const columns = await query<{ name: string }>(
    `SELECT COLUMN_NAME AS name FROM INFORMATION_SCHEMA.COLUMNS
     WHERE TABLE_NAME = 't_course' AND COLUMNPROPERTY(OBJECT_ID('t_course'), COLUMN_NAME, 'IsIdentity') = 0
     ORDER BY ORDINAL_POSITION`,
  );

  const params: unknown[] = [templateId];
  const param = (value: unknown) => {
    params.push(value);
    return `@p${params.length}`;
  };
  const overrides: Record<string, string> = {
    Capacity: param(capacity),
    ID_dayOfCourse: param(mondayId),
    TimeFrom: param(timeFrom),
    TimeTo: param(E2E_COURSE_TIME_TO),
    ValidFrom: param('2000-01-01'),
    ValidTo: param(E2E_COURSE_VALID_TO),
    IsActive: '1',
    Price: '1',
    ID_CreatedBy: param(creatorId),
    ID_UpdatedBy: param(creatorId),
    CreatedDate: 'GETDATE()',
    UpdatedDate: 'GETDATE()',
    GID: 'NEWID()',
  };

  const names = columns.map((c) => `[${c.name}]`).join(', ');
  const values = columns.map((c) => overrides[c.name] ?? `[${c.name}]`).join(', ');
  const course = await queryOne<{ ID: number }>(
    `INSERT INTO t_course (${names}) OUTPUT INSERTED.ID SELECT ${values} FROM t_course WHERE ID = @p1`,
    ...params,
  );
  if (!course) throw new Error('Could not clone a course');
  return course.ID;
}

async function createApplication(courseId: number, parentId: number, parentEmail: string, app: Omit<SeedApplication, 'id'>): Promise<number> {
  const participant = await queryOne<{ ID: number }>(
    `INSERT INTO t_system_user_participant (FirstName, LastName, BirthYear, ID_ParentUser, UpdatedDate, CreatedDate, ID_UpdatedBy, ID_CreatedBy, GID, ID_experiences, IsActive)
     OUTPUT INSERTED.ID VALUES (@p1, @p2, @p3, @p4, GETDATE(), GETDATE(), @p4, @p4, NEWID(), 1, 1)`,
    app.firstName,
    app.lastName,
    2000 + Number(app.personalId.slice(0, 2)),
    parentId,
  );
  const form = await queryOne<{ ID: number }>(
    `INSERT INTO t_course_application_form (ID_course, ID_participant, PersonalId, ParentName, Phone, Email, HealthState,
       GDPR_confirmed, Rules_confirmed, UpdatedDate, CreatedDate, ID_UpdatedBy, ID_CreatedBy, GID, EmailSent, Paid, IsActive)
     OUTPUT INSERTED.ID VALUES (@p1, @p2, @p3, 'E2E Rodič', '777000111', @p4, 'zdravý', 1, 1, GETDATE(), GETDATE(), @p5, @p5, NEWID(), 1, 0, 1)`,
    courseId,
    participant!.ID,
    app.personalId,
    parentEmail,
    parentId,
  );
  return form!.ID;
}

/** Random 10-digit Czech personal ID of a child born in 2015 (only the length and the first two digits matter to the app). */
export function personalId(): string {
  return '15' + String(Math.floor(Math.random() * 1e8)).padStart(8, '0');
}

export function isoDate(d: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

function lastMonday(): string {
  const d = new Date();
  d.setDate(d.getDate() - ((d.getDay() + 6) % 7));
  return isoDate(d);
}

export async function seed(): Promise<SeedData> {
  const runId = Date.now().toString(36);

  const users = {} as SeedData['users'];
  for (const role of ROLES) {
    users[role] = { id: await createUser(ROLE_EMAILS[role], ROLE_CODES[role]), email: ROLE_EMAILS[role] };
  }

  const template = await queryOne<{ ID: number }>(
    `SELECT TOP 1 ID FROM t_course WHERE IsActive = 1 AND ValidFrom <= GETDATE() AND ValidTo >= GETDATE() ORDER BY ID DESC`,
  );
  if (!template) throw new Error('No active course to use as a template for the e2e courses');
  const monday = await queryOne<{ ID: number }>(`SELECT ID FROM t_course_day WHERE Code = 'CD1'`);
  if (!monday) throw new Error('Course day CD1 (Monday) not found');

  const open: SeedCourse = { id: 0, timeFrom: '23:11', timeTo: E2E_COURSE_TIME_TO, capacity: 50 };
  const full: SeedCourse = { id: 0, timeFrom: '23:12', timeTo: E2E_COURSE_TIME_TO, capacity: 1 };
  open.id = await cloneCourse(template.ID, users.admin.id, monday.ID, open.timeFrom, open.capacity);
  full.id = await cloneCourse(template.ID, users.admin.id, monday.ID, full.timeFrom, full.capacity);

  const app = (lastName: string): Omit<SeedApplication, 'id'> => ({ firstName: 'E2e', lastName: `${lastName}${runId}`, personalId: personalId() });
  const main = app('Hlavni');
  const deactivate = app('Neaktivni');
  const fullApp = app('Plny');
  const applications: SeedData['applications'] = {
    main: { ...main, id: await createApplication(open.id, users.user.id, users.user.email, main) },
    deactivate: { ...deactivate, id: await createApplication(open.id, users.user.id, users.user.email, deactivate) },
    // fills the capacity-1 course
    full: { ...fullApp, id: await createApplication(full.id, users.user.id, users.user.email, fullApp) },
  };

  const visitor = { email: `e2e-visitor@${E2E_DOMAIN}`, firstName: 'Návštěvník', lastName: `Stena${runId}` };
  await query(
    `INSERT INTO t_system_wall_user (Email, FirstName, LastName, BirthDate, IsEnabled, CreateDate, GDPR_confirmed, Rules_confirmed, commercial_confirmed)
     VALUES (@p1, @p2, @p3, '1990-02-01', 1, GETDATE(), 1, 1, 1)`,
    visitor.email,
    visitor.firstName,
    visitor.lastName,
  );

  const data: SeedData = { runId, users, courses: { open, full }, applications, visitor, lessonDate: lastMonday() };
  fs.mkdirSync(AUTH_DIR, { recursive: true });
  fs.writeFileSync(SEED_FILE, JSON.stringify(data, null, 2));
  return data;
}

export function readSeed(): SeedData {
  return JSON.parse(fs.readFileSync(SEED_FILE, 'utf8')) as SeedData;
}
