# Anarva Cloud Managed MySQL Database Platform Architecture (Phase 20)

## Overview
Anarva Cloud Phase 20 introduces **MySQL** as the second managed relational database engine alongside PostgreSQL under `internal/database/mysql/...`. It extracts common relational database abstractions in `internal/database/...` while maintaining strict engine-specific behavior for PostgreSQL and MySQL.

---

## Core Domain Models (`internal/database/mysql/domain/mysql.go`)
- **`MySQLInstance`**: Managed MySQL database metadata including CPU, Memory, Storage, Region, Network, Maintenance Window, Port (`3306`), and reality labels (`LOCAL_MYSQL (LIMITED_CAPABILITIES)`).
- **`MySQLVersion`**: Version configuration (`8.0`, `8.4`).
- **`MySQLDatabase`**: Individual database schemas with charset and collation support.
- **`MySQLUser` & `MySQLPrivilege`**: Database user accounts and privilege grants (`SELECT`, `INSERT`, `UPDATE`, `DELETE`, `CREATE`, `ALTER`, `INDEX`, `DROP`, `EXECUTE`).
- **`MySQLHealth`**: Active health probes (`mysqladmin ping`) and metrics (`activeConns`, `threadsRunning`, `bufferPoolUsage`, `uptimeSec`).
- **`MySQLBackup`**: Automated snapshot and manual backup management.

---

## Security & SQL Proxy Architecture (`internal/database/mysql/service/sql_service.go`)
- **Dangerous Query Filter**: Rejects administrative queries (`DROP DATABASE`, `SHUTDOWN`, `RESET MASTER`).
- **Execution Limits**: 5s statement timeouts, 1000-row result caps, and 2MB payload caps.
- **Network Isolation**: Default port `3306` ingress is restricted to `PRIVATE` VPC subnets.

---

## Provider Reality Labels
- **Local Execution**: `LOCAL_MYSQL (LIMITED_CAPABILITIES)`
- **Cloud Provider**: `MANAGED_MYSQL_CONNECTED`
- **Unconfigured**: `MANAGED_MYSQL_NOT_CONFIGURED`
