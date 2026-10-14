import path from 'node:path';

export const PORT = Number(process.env.E2E_PORT || 5599);
export const BASE_URL = `http://localhost:${PORT}`;

// Every account/visitor the suite creates uses this reserved (never deliverable) domain,
// and purging only ever touches rows tied to it.
export const E2E_DOMAIN = 'e2e-sambar.test';
export const E2E_EMAIL_LIKE = `%@${E2E_DOMAIN}`;

export const ROLES = ['admin', 'instructor', 'reception', 'user'] as const;
export type Role = (typeof ROLES)[number];

export const ROLE_EMAILS: Record<Role, string> = {
  admin: `e2e-admin@${E2E_DOMAIN}`,
  instructor: `e2e-instructor@${E2E_DOMAIN}`,
  reception: `e2e-reception@${E2E_DOMAIN}`,
  user: `e2e-user@${E2E_DOMAIN}`,
};

// t_system_role.Code values (modules/constants/roles.go)
export const ROLE_CODES: Partial<Record<Role, string>> = {
  admin: 'SADM',
  instructor: 'instruktor',
  reception: 'recepce',
};

export const AUTH_DIR = path.resolve(__dirname, '../.auth');
export const SEED_FILE = path.join(AUTH_DIR, 'seed.json');
export const storageStateFor = (role: Role) => path.join(AUTH_DIR, `${role}.json`);

// Marker of the e2e courses: they end at 23:59 and are valid until 2099, which no real course is.
export const E2E_COURSE_TIME_TO = '23:59';
export const E2E_COURSE_VALID_TO = '2099-12-31';
