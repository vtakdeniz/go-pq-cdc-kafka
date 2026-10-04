# CDC metadata demo (LSN + TransactionID)

Local playground for the `cdc.Message.LSN` and `cdc.Message.TransactionID` fields.

## What you get

| Service | URL |
|---------|-----|
| Web UI (insert FK pair) | http://localhost:3000 |
| Kafka UI | http://localhost:8085 |
| Postgres | `localhost:5432` (`cdc_user` / `cdc_pass` / `cdc_db`) |
| Kafka | `localhost:19092` |
| Topic | `cdc.demo` |

Press **Insert FK pair** → one transaction inserts `parents` + `children` → CDC publishes two Kafka records with the same `txid` header and increasing `lsn`.

## Run

Needs Docker and Go 1.25+.

```bash
cd example/metadata-demo
docker compose up -d
go mod tidy

# terminal 1
go run ./cmd/connector

# terminal 2
go run ./cmd/web
```

1. Open http://localhost:3000 and click **Insert FK pair**
2. Open http://localhost:8085 → Topics → **cdc.demo** → Messages
3. Confirm both messages share `transactionId` / header `txid`, and `lsn` increases

## Notes

- Both tables map to the same topic; the Kafka key is the transaction id so they share a partition.
- This demo depends on the temporary `replace` of `go-pq-cdc` to Trendyol/go-pq-cdc#182 (for `ListenerContext.Xid`).
- If Postgres was started before `init.sql` existed, recreate volumes: `docker compose down -v && docker compose up -d`
- If Kafka UI stays offline on Linux Docker, switch the `kafka-ui` service to `network_mode: host`, set `SERVER_PORT=8085`, and `KAFKA_CLUSTERS_0_BOOTSTRAPSERVERS=127.0.0.1:19092` (remove the `ports:` mapping). Docker Desktop on macOS should use the default bridge settings above.
