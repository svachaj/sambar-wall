import { closePool } from './lib/db';
import { assertSafeDatabase, purge, seed } from './lib/seed';

export default async function globalSetup(): Promise<void> {
  assertSafeDatabase();
  try {
    // leftovers of an interrupted run would break uniqueness assumptions
    await purge();
    await seed();
  } finally {
    await closePool();
  }
}
