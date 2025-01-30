## Activity Examples

### Daily activity trends

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

### Top 10 events

List the 10 most frequently called events.

```sql
select
  service_name as event_source,
  method_name as event_name,
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

### Top 10 events (exclude read-only)

List the top 10 most frequently called events, excluding read-only events.

```sql
select
  service_name as event_source,
  method_name as event_name,
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

### Top events by project

Count and group events by project ID, event source, and event name to analyze activity across projects.

```sql
select
  service_name as event_source,
  method_name as event_name,
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

### Top error codes

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

## Detection Examples

### IAM policy changes

Detect IAM policy modifications.

```sql
select
  timestamp,
  service_name,
  method_name,
  authentication_info->>'principal_email' as user_email,
  resource_name
from
  gcp_audit_log
where
  service_name = 'iam.googleapis.com'
  and method_name ilike '%setiampolicy'
order by
  timestamp desc;
```

### Logging disabled

Detect when logging was disabled for a project.

```sql
select
  timestamp,
  service_name,
  method_name,
  authentication_info->>'principal_email' as user_email,
  resource_name
from
  gcp_audit_log
where
  service_name = 'logging.googleapis.com'
  and method_name = 'DeleteSink'
order by
  timestamp desc;
```

### Unsuccessful login attempts

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
  method_name = 'google.iam.admin.v%.SignIn'
  and status.code != '0'
order by
  timestamp desc;
```

### Activity in unapproved regions

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

## Operational Examples

### Firewall rule modifications

Track changes to firewall rules.

```sql
select
  timestamp,
  service_name,
  method_name,
  authentication_info->>'principal_email' as user_email,
  resource_name
from
  gcp_audit_log
where
  service_name = 'compute.googleapis.com'
  and method_name like '%firewalls.%'
order by
  timestamp desc;
```

### High volume of storage access requests

Detect unusually high access activity to Cloud Storage buckets and objects.

```sql
select
  authentication_info->>'principal_email' as user_email,
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

### Excessive role assignments

Identify IAM roles being assigned at an unusually high frequency.

```sql
select
  authentication_info->>'principal_email' as user_email,
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

