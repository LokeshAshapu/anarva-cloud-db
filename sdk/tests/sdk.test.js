import { createServer } from 'node:http';
import { test, describe, before, after } from 'node:test';
import assert from 'node:assert';
import { AnarvaClient, AnarvaError } from '../src/index.js';
describe('@anarva/sdk Unit & Integration Tests', () => {
    let server;
    let serverUrl;
    before((_, done) => {
        server = createServer((req, res) => {
            assert.strictEqual(req.headers['authorization'], 'Bearer anarva_live_testsecret123');
            if (req.url === '/api/v1/resources?resourceType=RDS') {
                res.writeHead(200, { 'Content-Type': 'application/json', 'x-request-id': 'req_rds_101' });
                res.end(JSON.stringify({
                    data: [
                        {
                            id: 'res-rds-postgres-01',
                            name: 'anarva-rds-prod-01',
                            engine: 'POSTGRESQL',
                            status: 'AVAILABLE',
                            multiAz: true,
                        },
                    ],
                    requestId: 'req_rds_101',
                }));
                return;
            }
            if (req.url === '/api/v1/operations/op-101') {
                res.writeHead(200, { 'Content-Type': 'application/json' });
                res.end(JSON.stringify({
                    data: {
                        id: 'op-101',
                        status: 'COMPLETED',
                        resourceId: 'res-rds-postgres-01',
                    },
                }));
                return;
            }
            if (req.url === '/api/v1/databases/db-missing') {
                res.writeHead(404, { 'Content-Type': 'application/json', 'x-request-id': 'req_err_404' });
                res.end(JSON.stringify({
                    error: {
                        code: 'RESOURCE_NOT_FOUND',
                        message: 'Database db-missing not found',
                        requestId: 'req_err_404',
                    },
                }));
                return;
            }
            res.writeHead(400);
            res.end();
        });
        server.listen(0, '127.0.0.1', () => {
            const addr = server.address();
            serverUrl = `http://127.0.0.1:${addr.port}`;
            done();
        });
    });
    after((_, done) => {
        server.close(done);
    });
    test('AnarvaError redacts API key secrets', () => {
        const err = new AnarvaError({
            code: 'AUTH_FAILED',
            message: 'Failed auth for anarva_live_999888777666',
            requestId: 'req_redact_01',
        });
        assert.strictEqual(err.message.includes('anarva_live_999888777666'), false);
        assert.strictEqual(err.message.includes('[REDACTED_API_KEY]'), true);
    });
    test('client.databases.list() retrieves data successfully', async () => {
        const client = new AnarvaClient({
            apiKey: 'anarva_live_testsecret123',
            apiUrl: serverUrl,
        });
        const dbs = await client.databases.list();
        assert.strictEqual(dbs.length, 1);
        assert.strictEqual(dbs[0].id, 'res-rds-postgres-01');
        assert.strictEqual(dbs[0].multiAz, true);
    });
    test('client.databases.get() handles 404 AnarvaError cleanly', async () => {
        const client = new AnarvaClient({
            apiKey: 'anarva_live_testsecret123',
            apiUrl: serverUrl,
        });
        try {
            await client.databases.get('db-missing');
            assert.fail('Should have thrown AnarvaError');
        }
        catch (err) {
            assert.strictEqual(err instanceof AnarvaError, true);
            assert.strictEqual(err.code, 'RESOURCE_NOT_FOUND');
            assert.strictEqual(err.requestId, 'req_err_404');
        }
    });
    test('client.operations.wait() polls until COMPLETED', async () => {
        const client = new AnarvaClient({
            apiKey: 'anarva_live_testsecret123',
            apiUrl: serverUrl,
        });
        const op = await client.operations.wait('op-101', { timeoutMs: 5000, intervalMs: 100 });
        assert.strictEqual(op.status, 'COMPLETED');
        assert.strictEqual(op.resourceId, 'res-rds-postgres-01');
    });
});
