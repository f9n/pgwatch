---
title: Agent Mode and Maintenance Commands
---

These features were added for distributed pgwatch setups, allowing:

- pgwatch instances running in agent mode to skip maintenance tasks
- Maintenance tasks to be executed via separate commands

## Agent Mode

Agent mode disables maintenance tasks (partition deletion and unique sources maintenance).

### Usage

```bash
# Run with agent mode
pgwatch --sources=sources.yaml --sink=postgresql://sink-db/metrics --agent-mode

# Using environment variable
export PW_AGENT_MODE=true
pgwatch --sources=sources.yaml --sink=postgresql://sink-db/metrics
```

When agent mode is enabled:
- `deleteOldPartitions()` will not run
- `maintainUniqueSources()` will not run
- Log message "agent mode enabled, maintenance tasks disabled" will appear

## Maintenance Commands

Maintenance tasks can now be executed via separate commands.

### Partition Cleanup

Cleans up old partitions:

```bash
# Basic usage (uses global retention setting)
pgwatch --sink=postgresql://sink-db/metrics maintenance cleanup

# Custom retention period (30 days)
pgwatch --sink=postgresql://sink-db/metrics maintenance cleanup --retention-days 30

# Dry run (only shows what would be deleted)
pgwatch --sink=postgresql://sink-db/metrics maintenance cleanup --dry-run

# Both custom retention and dry run
pgwatch --sink=postgresql://sink-db/metrics maintenance cleanup --retention-days 7 --dry-run
```

### Unique Sources Maintenance

Updates the unique sources listing used for Grafana dropdowns to prevent duplicate entries:

```bash
pgwatch --sink=postgresql://sink-db/metrics maintenance maintain-unique-sources
```

## Example Distributed Setup

### 1. Metric Collection Agents (on each VM)

```bash
# On VM1
pgwatch --sources=vm1-sources.yaml --sink=postgresql://central-sink/metrics --agent-mode

# On VM2  
pgwatch --sources=vm2-sources.yaml --sink=postgresql://central-sink/metrics --agent-mode

# On VM3
pgwatch --sources=vm3-sources.yaml --sink=postgresql://central-sink/metrics --agent-mode
```

### 2. Central Maintenance (single location, via cronjob)

```bash
# Daily partition cleanup at 02:00
0 2 * * * /usr/local/bin/pgwatch --sink=postgresql://central-sink/metrics maintenance cleanup

# Daily unique sources maintenance at 03:00
0 3 * * * /usr/local/bin/pgwatch --sink=postgresql://central-sink/metrics maintenance maintain-unique-sources
```

## Benefits

1. **Performance**: Maintenance tasks run only from a single instance
2. **Scalability**: Each agent focuses solely on metric collection
3. **Flexibility**: Centralized control over maintenance scheduling
4. **No Lock Contention**: Optimized advisory lock usage

## Notes

- Maintenance commands only work with PostgreSQL sinks
- MultiWriter (multiple sinks) is not supported
- Dry-run is not fully supported for TimescaleDB
- Advisory locks ensure only one maintenance operation runs at a time

## Troubleshooting

### Common Issues

**Error: "maintenance cleanup is only supported for single PostgreSQL sinks"**
- You're using multiple sinks or non-PostgreSQL sinks
- Solution: Use a single PostgreSQL sink for maintenance commands

**Error: "another instance is already running maintenance"**
- Another pgwatch instance is already running maintenance
- Solution: Wait for the other instance to complete, or check if it's stuck

**Error: "retention period must be greater than 0"**
- No retention period specified or global retention is 0
- Solution: Use `--retention-days` flag or set global `--retention` parameter

### Monitoring

You can monitor maintenance operations via PostgreSQL logs:

```sql
-- Check current advisory locks
SELECT * FROM pg_locks WHERE locktype = 'advisory';

-- Check partition cleanup history (look for DROP TABLE operations)
SELECT * FROM pg_stat_activity WHERE query LIKE '%DROP TABLE%';
``` 