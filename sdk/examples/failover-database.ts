import { AnarvaClient, AnarvaError } from '../src/index.js';

const client = new AnarvaClient({
  apiKey: process.env.ANARVA_API_KEY,
});

async function main() {
  const dbId = 'anarva-rds-prod-01';
  try {
    console.log(`Triggering Controlled RDS Multi-AZ Failover for ${dbId}...`);
    const job = await client.databases.failover(dbId);
    console.log('Controlled Failover Initiated:', job);
    console.log(`Primary AZ swapped from ${job.previousPrimaryAz} to ${job.newPrimaryAz}`);
  } catch (err) {
    if (err instanceof AnarvaError) {
      console.error(`Anarva Error [${err.code}]: ${err.message} (Request ID: ${err.requestId})`);
    } else {
      console.error('Unexpected Error:', err);
    }
  }
}

main();
