import { AnarvaClient, AnarvaError } from '../src/index.js';

const client = new AnarvaClient({
  apiKey: process.env.ANARVA_API_KEY,
});

async function main() {
  try {
    console.log('Provisioning new Managed PostgreSQL Instance...');
    const database = await client.databases.create({
      name: 'analytics-postgres-db',
      projectId: 'proj-default',
      engine: 'POSTGRESQL',
      storageGb: 50,
      acuUnits: 2.0,
      multiAz: true,
    });
    console.log('Database Created Successfully:', database);
  } catch (err) {
    if (err instanceof AnarvaError) {
      console.error(`Anarva Error [${err.code}]: ${err.message} (Request ID: ${err.requestId})`);
    } else {
      console.error('Unexpected Error:', err);
    }
  }
}

main();
