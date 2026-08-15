import { AnarvaClient, AnarvaError } from '../src/index.js';

const client = new AnarvaClient({
  apiKey: process.env.ANARVA_API_KEY,
});

async function main() {
  const operationId = 'op-101';
  try {
    console.log(`Polling status for operation ${operationId}...`);
    const operation = await client.operations.wait(operationId, {
      timeoutMs: 30000,
      intervalMs: 1000,
    });
    console.log('Operation Polling Complete:', operation);
  } catch (err) {
    if (err instanceof AnarvaError) {
      console.error(`Anarva Error [${err.code}]: ${err.message} (Request ID: ${err.requestId})`);
    } else {
      console.error('Unexpected Error:', err);
    }
  }
}

main();
