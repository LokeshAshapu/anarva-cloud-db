# ANARVA Data Loss Forensic Root-Cause Report

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Symptom**: *"A user inserted data yesterday and the data is missing today."*  
**Resolution Status**: Fixed & Verified  

---

## Forensic Answers to Required Questions

1. **Where was the user's data stored yesterday?**  
   The user record was created via `POST /api/v1/auth/register` and stored in RAM inside the `memUserRepo.users` Go map (`map[string]*authDomain.User`).

2. **Where is the application looking for it today?**  
   The application looks in the `UserRepository` interface (`GetByID` / `GetByEmail`). Upon restart, `newMemUserRepo()` was instantiated fresh.

3. **Is it the same physical database/storage?**  
   No. Because `DATABASE_URL` was missing in local/standalone mode, `main.go` fell back to `newMemUserRepo()`. Memory RAM is wiped upon process exit or container restart.

4. **Why did the record disappear?**  
   The backend server process was restarted (or container rebooted overnight). Since the records were kept in process RAM, process termination erased all created accounts.

5. **Is the problem local, deployment-related, frontend-related, repository-related, or database-related?**  
   It was **repository-related**. The fallback repository in `mock_repos.go` maintained in-memory state without disk serialization.

6. **What exact code/configuration causes it?**  
   `cmd/gateway/mock_repos.go` — `memUserRepo`, `memSessionRepo`, `memKeyRepo`, `memOrgRepo`, `memProjRepo`, `memInstanceRepo`, and `memBackupRepo` lacked file persistence.

7. **What is the smallest safe fix?**  
   Implement JSON disk file synchronization (`./data/anarva_cp_<resource>.json`) for all fallback repositories in `mock_repos.go`. On startup, load existing records from disk; on `Create`/`Update`/`Delete`, write mutated maps to disk.

8. **How will the fix be regression tested?**  
   Automated Go unit test `cmd/gateway/durable_persistence_test.go` (`TestUserDurablePersistence_SurvivesProcessRestart`). The test creates a unique user record, instantiates a new repository instance (simulating a backend restart), and verifies the user record is read back successfully from disk.
