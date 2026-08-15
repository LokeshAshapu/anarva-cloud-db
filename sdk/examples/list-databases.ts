import { AnarvaClient, AnarvaError } from '../src/index.js';

const client = new AnarvaClient({
  apiKey: process.env.ANARVA_API_KEY,
});

async function main() {
  try {
    console.log('Fetching Managed PostgreSQL Databases...');
    const databases = await client.databases.list();
    console.log('Databases:', JSON.stringify(databases, null, 2));
  } catch (err) {
    if (err instanceof AnarvaError) {
      console.error(`Anarva Error [${err.code}]: ${err.message} (Request ID: ${err.requestId})`);
    } else {
      console.error('Unexpected Error:', err);
    }
  }
}

main();
