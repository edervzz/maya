package hcons

const MAYA_LOCKS_TABLE = "_MayaLocks"

const MigrationsTable = `CREATE TABLE _MayaMigrationsHistory
    (
        id varchar(100) NOT NULL,
        version varchar(30) NOT NULL,
        PRIMARY KEY (id)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Migration history tracker.';
    `

const LockTable = `CREATE TABLE _MayaLocks 
    (
        id varchar(100) NOT NULL,
        entity varchar(100) NOT NULL,
        created_at DATETIME NOT NULL,
        info varchar(100) NULL,
        PRIMARY KEY (id)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Generic Lock table. Maya uses on mutations.';
    `
