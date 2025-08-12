---
title: "Queries: gcp_requests_log - Example queries for Cloud Armor Request Logs"
description: "Example queries to analyze Cloud Armor request logs for security monitoring and threat detection."
---

# Queries: gcp_requests_log

This page provides example queries to help you analyze Cloud Armor request logs for security monitoring, threat detection, and performance analysis.

## Security Analysis

### Top blocked IP addresses

```sql
select
  http_request.remote_ip,
  count(*) as block_count,
  array_agg(distinct security_policy.rule_name) as rules_triggered
from
  gcp_requests_log
where
  security_policy.action = 'deny'
group by
  http_request.remote_ip
order by
  block_count desc
limit 20;
```

### Security policy effectiveness

```sql
select
  security_policy.rule_name,
  security_policy.rule_type,
  count(*) as total_requests,
  count(case when security_policy.action = 'deny' then 1 end) as blocked_requests,
  round(
    (count(case when security_policy.action = 'deny' then 1 end)::float / count(*)::float) * 100,
    2
  ) as block_rate_percent
from
  gcp_requests_log
where
  security_policy.rule_evaluated = true
group by
  security_policy.rule_name,
  security_policy.rule_type
order by
  blocked_requests desc;
```

### Threat intelligence summary

```sql
select
  threat_info.threat_type,
  threat_info.threat_category,
  threat_info.threat_severity,
  count(*) as occurrence_count,
  count(distinct http_request.remote_ip) as unique_ips
from
  gcp_requests_log
where
  threat_info.threat_type is not null
group by
  threat_info.threat_type,
  threat_info.threat_category,
  threat_info.threat_severity
order by
  occurrence_count desc;
```

## Performance Analysis

### Response time distribution

```sql
select
  case
    when cast(replace(http_request.latency, 's', '') as float) < 0.1 then 'Very Fast (< 100ms)'
    when cast(replace(http_request.latency, 's', '') as float) < 0.5 then 'Fast (100-500ms)'
    when cast(replace(http_request.latency, 's', '') as float) < 1.0 then 'Normal (500ms-1s)'
    when cast(replace(http_request.latency, 's', '') as float) < 5.0 then 'Slow (1-5s)'
    else 'Very Slow (> 5s)'
  end as latency_category,
  count(*) as request_count,
  round(
    (count(*)::float / sum(count(*)) over()) * 100,
    2
  ) as percentage
from
  gcp_requests_log
where
  http_request.latency is not null
group by
  latency_category
order by
  request_count desc;
```

### Cache performance by hour

```sql
select
  date_trunc('hour', timestamp) as hour,
  cache_info.cache_hit,
  count(*) as request_count,
  avg(cast(replace(http_request.latency, 's', '') as float)) as avg_latency_seconds
from
  gcp_requests_log
where
  cache_info.cache_lookup = true
group by
  hour,
  cache_info.cache_hit
order by
  hour desc,
  cache_info.cache_hit;
```

### Top slow endpoints

```sql
select
  http_request.url,
  http_request.method,
  count(*) as request_count,
  avg(cast(replace(http_request.latency, 's', '') as float)) as avg_latency_seconds,
  max(cast(replace(http_request.latency, 's', '') as float)) as max_latency_seconds
from
  gcp_requests_log
where
  http_request.latency is not null
group by
  http_request.url,
  http_request.method
having
  avg(cast(replace(http_request.latency, 's', '') as float)) > 1.0
order by
  avg_latency_seconds desc
limit 20;
```

## Traffic Analysis

### Request volume by hour

```sql
select
  date_trunc('hour', timestamp) as hour,
  count(*) as total_requests,
  count(case when security_policy.action = 'deny' then 1 end) as blocked_requests,
  count(case when threat_info.threat_type is not null then 1 end) as threat_requests
from
  gcp_requests_log
group by
  hour
order by
  hour desc
limit 24;
```

### HTTP status code distribution

```sql
select
  http_request.status,
  count(*) as request_count,
  round(
    (count(*)::float / sum(count(*)) over()) * 100,
    2
  ) as percentage
from
  gcp_requests_log
where
  http_request.status is not null
group by
  http_request.status
order by
  request_count desc;
```

### User agent analysis

```sql
select
  http_request.user_agent,
  count(*) as request_count,
  count(distinct http_request.remote_ip) as unique_ips
from
  gcp_requests_log
where
  http_request.user_agent is not null
group by
  http_request.user_agent
order by
  request_count desc
limit 20;
```

## Security Monitoring

### Recent security events

```sql
select
  timestamp,
  http_request.remote_ip,
  http_request.url,
  security_policy.action,
  security_policy.rule_name,
  threat_info.threat_type,
  threat_info.threat_severity
from
  gcp_requests_log
where
  security_policy.action = 'deny'
  or threat_info.threat_type is not null
order by
  timestamp desc
limit 100;
```

### Preview mode analysis

```sql
select
  security_policy.rule_name,
  count(*) as total_requests,
  count(case when security_policy.preview_mode = true then 1 end) as preview_requests,
  count(case when security_policy.preview_mode = false then 1 end) as enforced_requests
from
  gcp_requests_log
where
  security_policy.rule_evaluated = true
group by
  security_policy.rule_name
order by
  total_requests desc;
```

### Geographic threat analysis

```sql
select
  http_request.remote_ip,
  count(*) as request_count,
  array_agg(distinct threat_info.threat_type) as threat_types,
  array_agg(distinct security_policy.rule_name) as rules_triggered
from
  gcp_requests_log
where
  threat_info.threat_type is not null
  or security_policy.action = 'deny'
group by
  http_request.remote_ip
order by
  request_count desc
limit 50;
```

## Verbose Logging Analysis

### Request header analysis

```sql
select
  key as header_name,
  count(*) as occurrence_count,
  count(distinct value) as unique_values
from
  gcp_requests_log,
  jsonb_each_text(verbose_logging.request_headers) as headers(key, value)
where
  verbose_logging.request_headers is not null
group by
  key
order by
  occurrence_count desc;
```

### Large request analysis

```sql
select
  timestamp,
  http_request.remote_ip,
  http_request.url,
  http_request.request_size,
  http_request.response_size,
  verbose_logging.processing_time
from
  gcp_requests_log
where
  http_request.request_size > 1000000  -- 1MB
  or http_request.response_size > 1000000
order by
  timestamp desc
limit 50;
```
