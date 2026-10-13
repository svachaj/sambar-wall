import { closePool } from './lib/db';
import { assertSafeDatabase, countLeftovers, purge } from './lib/seed';

export default async function globalTeardown(): Promise<void> {
  if (process.env.E2E_KEEP_DATA === '1') {
    console.log('E2E_KEEP_DATA=1: leaving e2e data in the database for inspection.');
    return;
  }
  assertSafeDatabase();
  try {
    await purge();
    const leftovers = Object.entries(await countLeftovers()).filter(([, n]) => n > 0);
    if (leftovers.length > 0) {
      throw new Error(`E2E data left in the database after purge: ${JSON.stringify(Object.fromEntries(leftovers))}`);
    }
  } finally {
    await closePool();
  }
}
