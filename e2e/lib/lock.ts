import os from 'node:os';
import sql from 'mssql';
import { config } from './db';

// Only one e2e run (local or deployed, from any machine) may use the shared test DB at a time, because every
// run seeds and purges the same rows. The lock is a SQL Server application lock owned by a transaction on a
// dedicated connection: SQL Server frees it automatically if the run crashes or loses its connection.
const RESOURCE = 'sambar-wall-e2e';
const POLL_MS = 3_000;

export type DbLock = { release: () => Promise<void> };

export async function acquireDbLock(waitMinutes = Number(process.env.E2E_LOCK_WAIT_MINUTES || 5)): Promise<DbLock> {
  const pool = await new sql.ConnectionPool({ ...config(), pool: { min: 1, max: 1 } }).connect();
  const transaction = new sql.Transaction(pool);
  await transaction.begin();

  const deadline = Date.now() + waitMinutes * 60_000;
  let announced = false;
  for (;;) {
    const result = await new sql.Request(transaction)
      .input('resource', RESOURCE)
      .output('status', sql.Int)
      .query(`DECLARE @r int; EXEC @r = sp_getapplock @Resource = @resource, @LockMode = 'Exclusive', @LockOwner = 'Transaction', @LockTimeout = 0; SET @status = @r;`);
    if ((result.output.status as number) >= 0) break;

    if (Date.now() > deadline) {
      await transaction.rollback().catch(() => undefined);
      await pool.close();
      throw new Error(
        `Another e2e run is using the shared test DB (lock "${RESOURCE}"), still busy after ${waitMinutes} min. ` +
          'Wait for it to finish, or raise E2E_LOCK_WAIT_MINUTES.',
      );
    }
    if (!announced) {
      console.log(`Another e2e run is using the shared test DB; waiting up to ${waitMinutes} min for it to finish...`);
      announced = true;
    }
    await new Promise((resolve) => setTimeout(resolve, POLL_MS));
  }
  if (announced) console.log('Lock acquired, starting.');
  console.log(`e2e DB lock held by ${os.hostname()} (pid ${process.pid}).`);

  return {
    release: async () => {
      // ending the transaction releases the application lock
      await transaction.commit().catch(() => undefined);
      await pool.close();
    },
  };
}
