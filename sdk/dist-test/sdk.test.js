"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const node_http_1 = require("node:http");
const node_test_1 = require("node:test");
const node_assert_1 = __importDefault(require("node:assert"));
const index_js_1 = require("../src/index.js");
(0, node_test_1.describe)('@anarva/sdk Unit & Integration Tests', () => {
    let server;
    let serverUrl;
    (0, node_test_1.before)((_, done) => {
        server = (0, node_http_1.createServer)((req, res) => {
            node_assert_1.default.strictEqual(req.headers['authorization'], 'Bearer anarva_live_testsecret123');
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
    (0, node_test_1.after)((_, done) => {
        server.close(done);
    });
    (0, node_test_1.test)('AnarvaError redacts API key secrets', () => {
        const err = new index_js_1.AnarvaError({
            code: 'AUTH_FAILED',
            message: 'Failed auth for anarva_live_999888777666',
            requestId: 'req_redact_01',
        });
        node_assert_1.default.strictEqual(err.message.includes('anarva_live_999888777666'), false);
        node_assert_1.default.strictEqual(err.message.includes('[REDACTED_API_KEY]'), true);
    });
    (0, node_test_1.test)('client.databases.list() retrieves data successfully', async () => {
        const client = new index_js_1.AnarvaClient({
            apiKey: 'anarva_live_testsecret123',
            apiUrl: serverUrl,
        });
        const dbs = await client.databases.list();
        node_assert_1.default.strictEqual(dbs.length, 1);
        node_assert_1.default.strictEqual(dbs[0].id, 'res-rds-postgres-01');
        node_assert_1.default.strictEqual(dbs[0].multiAz, true);
    });
    (0, node_test_1.test)('client.databases.get() handles 404 AnarvaError cleanly', async () => {
        const client = new index_js_1.AnarvaClient({
            apiKey: 'anarva_live_testsecret123',
            apiUrl: serverUrl,
        });
        try {
            await client.databases.get('db-missing');
            node_assert_1.default.fail('Should have thrown AnarvaError');
        }
        catch (err) {
            node_assert_1.default.strictEqual(err instanceof index_js_1.AnarvaError, true);
            node_assert_1.default.strictEqual(err.code, 'RESOURCE_NOT_FOUND');
            node_assert_1.default.strictEqual(err.requestId, 'req_err_404');
        }
    });
    (0, node_test_1.test)('client.operations.wait() polls until COMPLETED', async () => {
        const client = new index_js_1.AnarvaClient({
            apiKey: 'anarva_live_testsecret123',
            apiUrl: serverUrl,
        });
        const op = await client.operations.wait('op-101', { timeoutMs: 5000, intervalMs: 100 });
        node_assert_1.default.strictEqual(op.status, 'COMPLETED');
        node_assert_1.default.strictEqual(op.resourceId, 'res-rds-postgres-01');
    });
});
