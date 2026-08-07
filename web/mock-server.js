const http = require('http');

const PORT = 8080;

// Prevent server from crashing on aborted connections or socket resets
process.on('uncaughtException', (err) => {
  console.error('[Mock API Server] Uncaught Exception:', err);
});

process.on('unhandledRejection', (reason, promise) => {
  console.error('[Mock API Server] Unhandled Rejection:', reason);
});


// Mock database data for the SQL console
const mockData = {
  users: {
    columns: [
      { name: 'id', type: 'UUID' },
      { name: 'email', type: 'VARCHAR' },
      { name: 'full_name', type: 'VARCHAR' },
      { name: 'role', type: 'VARCHAR' },
      { name: 'status', type: 'VARCHAR' }
    ],
    rows: [
      { id: 'usr-87a1d9b3', email: 'admin@anarva.io', full_name: 'Anarva Admin', role: 'owner', status: 'active' },
      { id: 'usr-92c4b8e1', email: 'lokesh@anarva.io', full_name: 'Lokesh Ashapu', role: 'admin', status: 'active' },
      { id: 'usr-11f8e2d4', email: 'developer@anarva.io', full_name: 'Dev Team', role: 'developer', status: 'active' },
      { id: 'usr-55a7c3e9', email: 'security@anarva.io', full_name: 'Security Lead', role: 'auditor', status: 'active' }
    ]
  },
  databases: {
    columns: [
      { name: 'id', type: 'UUID' },
      { name: 'name', type: 'VARCHAR' },
      { name: 'engine', type: 'VARCHAR' },
      { name: 'status', type: 'VARCHAR' },
      { name: 'port', type: 'INTEGER' }
    ],
    rows: [
      { id: 'db-uuid-1', name: 'Primary Application Database', engine: 'postgres', status: 'RUNNING', port: 15432 },
      { id: 'db-uuid-2', name: 'Analytics Data Warehouse', engine: 'postgres', status: 'RUNNING', port: 15433 }
    ]
  },
  metrics: {
    columns: [
      { name: 'metric_name', type: 'VARCHAR' },
      { name: 'value', type: 'DOUBLE' },
      { name: 'timestamp', type: 'TIMESTAMP' }
    ],
    rows: [
      { metric_name: 'cpu_usage_percent', value: 12.4, timestamp: new Date().toISOString() },
      { metric_name: 'memory_usage_bytes', value: 2576980377, timestamp: new Date().toISOString() },
      { metric_name: 'query_latency_ms', value: 1.42, timestamp: new Date().toISOString() }
    ]
  }
};

const server = http.createServer((req, res) => {
  // CORS Headers
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type, Authorization');

  // Handle CORS preflight
  if (req.method === 'OPTIONS') {
    res.writeHead(204);
    res.end();
    return;
  }

  const url = new URL(req.url, `http://${req.headers.host || 'localhost'}`);
  const path = url.pathname;

  if (req.method === 'POST' && path === '/api/v1/auth/login') {
    let body = '';
    req.on('data', chunk => { body += chunk.toString(); });
    req.on('end', () => {
      try {
        const { email, password } = JSON.parse(body);
        console.log(`[Mock Auth] Login attempt for: ${email}`);
        
        // Respond with success
        res.writeHead(200, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({
          access_token: 'mock-jwt-token-for-local-dev',
          user: {
            id: 'usr-87a1d9b3',
            email: email,
            full_name: 'Anarva Admin',
            role: 'owner',
            status: 'active'
          }
        }));
      } catch (e) {
        res.writeHead(400, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ message: 'Invalid request payload' }));
      }
    });
  } else if (req.method === 'POST' && path === '/api/v1/query') {
    let body = '';
    req.on('data', chunk => { body += chunk.toString(); });
    req.on('end', () => {
      try {
        const { database_id, sql } = JSON.parse(body);
        console.log(`[Mock Query] Executing SQL on db ${database_id}: ${sql}`);

        const normalizedSql = sql.trim().toLowerCase();
        let selectedTable = null;

        if (normalizedSql.includes('from users')) {
          selectedTable = 'users';
        } else if (normalizedSql.includes('from databases') || normalizedSql.includes('from database')) {
          selectedTable = 'databases';
        } else if (normalizedSql.includes('from metrics')) {
          selectedTable = 'metrics';
        }

        const startTime = process.hrtime();

        setTimeout(() => {
          const diff = process.hrtime(startTime);
          const executionTimeMs = (diff[0] * 1000) + (diff[1] / 1000000);

          if (selectedTable && mockData[selectedTable]) {
            res.writeHead(200, { 'Content-Type': 'application/json' });
            res.end(JSON.stringify({
              columns: mockData[selectedTable].columns,
              rows: mockData[selectedTable].rows,
              rows_affected: mockData[selectedTable].rows.length,
              execution_time_ms: executionTimeMs
            }));
          } else {
            // General query handler
            const isSelect = normalizedSql.startsWith('select');
            res.writeHead(200, { 'Content-Type': 'application/json' });
            res.end(JSON.stringify({
              columns: isSelect ? [{ name: 'result', type: 'VARCHAR' }] : [],
              rows: isSelect ? [{ result: 'Success' }] : [],
              rows_affected: isSelect ? 1 : 0,
              execution_time_ms: executionTimeMs
            }));
          }
        }, 150); // Simulate network/DB latency
      } catch (e) {
        res.writeHead(400, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ message: 'Invalid query execution payload' }));
      }
    });
  } else {
    res.writeHead(404, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({ message: 'Mock endpoint not found' }));
  }
});

server.listen(PORT, () => {
  console.log(`[Mock API Server] Listening on http://localhost:${PORT}`);
});
