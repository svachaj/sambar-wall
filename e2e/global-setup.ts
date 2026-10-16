import { closePool } from './lib/db';
import { acquireDbLock } from './lib/lock';
import { assertSafeDatabase, countLeftovers, purge, seed } from './lib/seed';

/**
 * Takes the shared-DB lock, purges leftovers and seeds. The returned function is Playwright's global teardown:
 * it purges again and releases the lock only afterwards, so the next run can never see this run's data.
 */
export default async function globalSetup(): Promise<() => Promise<void>> {
  assertSafeDatabase();
  const lock = await acquireDbLock();
  try {
    // leftovers of an interrupted run would break uniqueness assumptions
    await purge();
    await seed();
  } catch (error) {
    await lock.release();
    throw error;
  } finally {
    await closePool();
  }

  return async () => {
    try {
      if (process.env.E2E_KEEP_DATA === '1') {
        console.log('E2E_KEEP_DATA=1: leaving e2e data in the database for inspection.');
        return;
      }
      await purge();
      const leftovers = Object.entries(await countLeftovers()).filter(([, n]) => n > 0);
      if (leftovers.length > 0) {
        throw new Error(`E2E data left in the database after purge: ${JSON.stringify(Object.fromEntries(leftovers))}`);
      }
    } finally {
      await closePool();
      await lock.release();
    }
  };
}
