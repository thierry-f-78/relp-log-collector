# RELP Log Collector

Welcome to the **RELP Log Collector** repository! This project addresses the need for a lightweight, reliable, and open-source solution for collecting, parsing, and indexing logs using the RELP protocol.

## Why This Project?

Existing solutions for RELP log collection often come with significant limitations:
- They are tied to companies that may change the license at any time.
- They are overly complex and resource-intensive.

Our program, written in Go, offers a GPL-licensed alternative that ensures the license cannot be changed without the consent of all contributors.

## Key Features

- **Protocol Support**: Native support for the RELP protocol, which is integrated into all Linux/rsyslog systems.
- **Secure and Reliable**: Secure log transmission via TLS, ensuring authentication and data integrity with RELP's acknowledgment mechanisms.
- **Database Integration**: Currently supports ClickHouse and PostgreSQL databases for log storage, ensuring efficient and scalable indexing. Support also Sqlite, but this base is not recommended for high quantity logs.
- **Persistence**: Logs are written to disk and acknowledged before dispatching, providing resilience in case of database failures.
- **High Availability**: Designed to be resilient with high availability setups. It can integrate with load balancers to distribute the load, and in case of node failure, connections will switch to another node. ClickHouse ensures load distribution and high availability for data acquisition.
- **Extensible**: Features a plugin system written in Go, allowing easy extension of its functionalities.

## How It Works

1. **Log Collection**: The program collects logs via the RELP protocol, which is secure and ensures no data loss, though duplicates may occur.
2. **Log Parsing and Indexing**: Logs are parsed and indexed for efficient storage and retrieval.
3. **Data Storage**: Logs are securely stored in ClickHouse, a fast and scalable database system.
4. **High Availability Setup**: Integrates with load balancers and high availability setups to ensure continuous operation even if some nodes fail.

## Installation

To install the RELP Log Collector, follow these steps:

1. Clone the repository:
    ```sh
    git clone https://github.com/yourusername/relp-log-collector.git
    cd relp-log-collector
    ```
2. Build the project:
    ```sh
    go build
    ```
3. Configure the program (edit the `config.yaml` file to suit your setup).

4. Run the program:
    ```sh
    ./relp-log-collector
    ```

## Database

ClickHouse is used as the database for the program. It require at least two tables. The actual data is stored in the `logs`, the second is for storing errors like unparsable log. there are the schema of these two tables:

```sql
create table syslog (
	"date"       DateTime64,
	"facility"   Int32,
	"severity"   Int32,
	"hostname"   String,
	"process"    String,
	"pid"        Int32,
	"data"       String
)
engine = MergeTree
partition by toDate("date")
order by
	("date");

create table errors (
	"date"       DateTime64,
	"data"       String
)
engine = MergeTree
partition by toDate("date")
order by
	("date");
```

For postgreSQL:

```sql
CREATE TABLE syslog (
   "date"       TIMESTAMP,
   "facility"   INTEGER,
   "severity"   INTEGER,
   "hostname"   VARCHAR,
   "process"    VARCHAR,
   "pid"        INTEGER,
   "data"       TEXT
);

CREATE TABLE errors (
   "date"       TIMESTAMP,
   "data"       TEXT
);
```

For Sqlite:

```sql
CREATE TABLE syslog (
   "date"       DATETIME,
   "facility"   INTEGER,
   "severity"   INTEGER,
   "hostname"   TEXT,
   "process"    TEXT,
   "pid"        INTEGER,
   "data"       TEXT
);

CREATE TABLE errors (
   "date"       DATETIME,
   "data"       TEXT
);

```

---

Feel free to customize this README further to better match your project's specifics and your preferences.
