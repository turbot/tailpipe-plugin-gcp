## Activity Examples

### Daily Activity Trends

Count events per day to identify activity trends over time.

```sql
select
  strftime(timestamp, '%Y-%m-%d') AS event_date,
  count(*) AS event_count
from
  gcp_audit_log
group by
  event_date
order by
  event_date asc;
```

```yaml
folder: Account
```

### Top 10 Events

List the 10 most frequently called events.

```sql
select
  service_name,
  method_name,
  count(*) as event_count
from
  gcp_audit_log
group by
  service_name,
  method_name
order by
  event_count desc
limit 10;
```

```yaml
folder: Account
```

### Top 10 Events (Exclude Read-Only)

List the top 10 most frequently called events, excluding read-only events.

```sql
select
  service_name,
  method_name,
  count(*) as event_count
from
  gcp_audit_log
where
  severity != 'INFO'
group by
  service_name,
  method_name
order by
  event_count desc
limit 10;
```

```yaml
folder: Account
```

### Top Events by Project

Count and group events by project ID, event source, and event name to analyze activity across projects.

```sql
select
  service_name,
  method_name,
  log_name as project_id,
  count(*) as event_count
from
  gcp_audit_log
group by
  service_name,
  method_name,
  log_name
order by
  event_count desc;
```

```yaml
folder: Account
```

### Top Error Codes

Identify the most frequent error codes.

```sql
select
  status.message as error_code,
  count(*) as event_count
from
  gcp_audit_log
where
  status.message is not null
group by
  status.message
order by
  event_count desc;
```

```yaml
folder: Account
```

## Detection Examples

### IAM Policy Changes

Detect IAM policy modifications.

```sql
select
  timestamp,
  service_name,
  method_name,
  authentication_info ->> 'principal_email' as user_email,
  resource_name
from
  gcp_audit_log
where
  service_name = 'iam.googleapis.com'
  and method_name ilike '%setiampolicy'
order by
  timestamp desc;
```

```yaml
folder: IAM
```

### Logging Disabled

Detect when logging was disabled for a project.

```sql
select
  timestamp,
  service_name,
  method_name,
  authentication_info ->> 'principal_email' as user_email,
  resource_name
from
  gcp_audit_log
where
  service_name = 'logging.googleapis.com'
  and method_name ilike '%DeleteSink'
order by
  timestamp desc;
```

```yaml
folder: Logging
```

### Unsuccessful Login Attempts

Find failed login attempts to detect potential unauthorized access.

```sql
select
  timestamp,
  method_name,
  authentication_info ->> 'principal_email' as user_email,
  resource_name
from
  gcp_audit_log
where
  method_name ilike 'google.iam.admin.v%.IAM.SignJwt'
  and status.code != '0'
order by
  timestamp desc;
```

```yaml
folder: IAM
```

### Activity in Unapproved Regions

Identify actions occurring in GCP regions outside an approved list.

```sql
select
  timestamp,
  service_name,
  method_name,
  authentication_info ->> 'principal_email' as user_email,
  resource_name
from
  gcp_audit_log
where
  resource_location ->> 'current_locations' not in ('us-central1', 'us-east1')
order by
  timestamp desc;
```

```yaml
folder: Account
```

## Operational Examples

### Firewall Rule Modifications

Track changes to firewall rules.

```sql
select
  timestamp,
  service_name,
  method_name,
  authentication_info ->> 'principal_email' as user_email,
  resource_name
from
  gcp_audit_log
where
  service_name = 'compute.googleapis.com'
  and method_name like '%firewalls.%'
order by
  timestamp desc;
```

```yaml
folder: Compute
```

### High Volume of Storage Access Requests

Detect unusually high access activity to Cloud Storage buckets and objects.

```sql
select
  authentication_info ->> 'principal_email' as user_email,
  count(*) as event_count,
  date_trunc('minute', timestamp) as event_minute
from
  gcp_audit_log
where
  service_name = 'storage.googleapis.com'
  and method_name like '%storage.objects.%'
group by
  user_email,
  event_minute
having
  count(*) > 100
order by
  event_count desc;
```

```yaml
folder: Storage
```

### Excessive Role Assignments

Identify IAM roles being assigned at an unusually high frequency.

```sql
select
  authentication_info ->> 'principal_email' as user_email,
  count(*) as event_count,
  date_trunc('hour', timestamp) as event_hour
from
  gcp_audit_log
where
  service_name = 'iam.googleapis.com'
  and method_name ilike '%setiampolicy'
group by
  user_email,
  event_hour
having
  count(*) > 10
order by
  event_hour desc,
  event_count desc;
```

```yaml
folder: IAM
```
