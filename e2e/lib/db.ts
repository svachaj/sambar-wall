import path from 'node:path';
import dotenv from 'dotenv';
import sql from 'mssql';

// The app reads its settings from the repo-root .env; the e2e suite talks to the same database.
dotenv.config({ path: path.resolve(__dirname, '../../.env'), quiet: true });
dotenv.config({ path: path.resolve(__dirname, '../.env'), quiet: true });

let pool: sql.ConnectionPool | undefined;

export function dbName(): string {
  return process.env.DB_NAME ?? '';
}

// Mirrors the vendored go-mssqldb: for a named instance ("host\SQLEXPRESS") the real TCP port is
// resolved through SQL Server Browser and DB_PORT is ignored; otherwise host:DB_PORT is used.
function config(): sql.config {
  const [server, instanceName] = (process.env.DB_HOST ?? '').split('\\');
  return {
    server,
    ...(instanceName ? {} : { port: Number(process.env.DB_PORT || 1433) }),
    user: process.env.DB_USER,
    password: process.env.DB_PASSWORD,
    database: process.env.DB_NAME,
    options: { encrypt: false, trustServerCertificate: true, ...(instanceName ? { instanceName } : {}) },
    pool: { max: 4 },
  };
}

export async function getPool(): Promise<sql.ConnectionPool> {
  if (!pool) {
    pool = await new sql.ConnectionPool(config()).connect();
  }
  return pool;
}

export async function closePool(): Promise<void> {
  await pool?.close();
  pool = undefined;
}

/**
 * Runs a T-SQL statement with positional parameters bound as @p1, @p2, ... (same style as the Go code).
 */
export async function query<T = Record<string, unknown>>(text: string, ...params: unknown[]): Promise<T[]> {
  const request = (await getPool()).request();
  params.forEach((value, i) => request.input(`p${i + 1}`, value));
  const result = await request.query<T>(text);
  return result.recordset ?? [];
}

export async function queryOne<T = Record<string, unknown>>(text: string, ...params: unknown[]): Promise<T | undefined> {
  return (await query<T>(text, ...params))[0];
}
