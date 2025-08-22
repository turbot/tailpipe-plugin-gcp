---
title: "Queries: gcp_requests_log - Example queries for Cloud Armor Request Logs"
description: "Example queries to analyze Cloud Armor request logs for security monitoring and threat detection."
---

# Queries: gcp_requests_log

This page provides example queries to help you analyze Cloud Armor request logs for security monitoring, threat detection, and performance analysis.

## Security Analysis Examples

### Top blocked IP addresses

Identify the IP addresses with the highest number of blocked requests by Cloud Armor security policies.

```sql
select
  json_extract(http_request, '$.remote_ip') as remote_ip,
  count(*) as block_count
from
  gcp_requests_log
where
  json_extract(enforced_security_policy, '$.outcome') = '"DENY"'
group by
  json_extract(http_request, '$.remote_ip')
order by
  block_count desc
limit 20;
```

```yaml
folder: Security
```

### Security policy effectiveness

Analyze the effectiveness of security policies by calculating block rates for each rule.

```sql
select
  json_extract(enforced_security_policy, '$.name') as rule_name,
  json_extract(enforced_security_policy, '$.configured_action') as action_type,
  count(*) as total_requests,
  count(case when json_extract(enforced_security_policy, '$.outcome') = '"DENY"' then 1 end) as blocked_requests,
  round(
    (count(case when json_extract(enforced_security_policy, '$.outcome') = '"DENY"' then 1 end)::float / count(*)::float) * 100,
    2
  ) as block_rate_percent
from
  gcp_requests_log
where
  enforced_security_policy is not null
group by
  json_extract(enforced_security_policy, '$.name'),
  json_extract(enforced_security_policy, '$.configured_action')
order by
  blocked_requests desc;
```

```yaml
folder: Security
```

### Geographic threat analysis

Analyze threats by geographic location based on IP addresses.

```sql
select
  json_extract(http_request, '$.remote_ip') as remote_ip,
  count(*) as request_count,
  json_extract(enforced_security_policy, '$.outcome') as security_outcome
from
  gcp_requests_log
where
  json_extract(enforced_security_policy, '$.outcome') = 'DENY'
group by
  json_extract(http_request, '$.remote_ip'),
  json_extract(enforced_security_policy, '$.outcome')
order by
  request_count desc
limit 50;
```

```yaml
folder: Security
```

## Performance Analysis Examples

### Response time distribution

Analyze the distribution of response times to identify performance patterns.

```sql
select
  json_extract(http_request, '$.latency') as latency_format,
  count(*) as request_count,
  round(
    (count(*)::float / sum(count(*)) over()) * 100,
    2
  ) as percentage
from
  gcp_requests_log
where
  json_extract(http_request, '$.latency') is not null
group by
  json_extract(http_request, '$.latency')
order by
  request_count desc
limit 10;
```

```yaml
folder: Performance
```

### Cache performance by hour

Monitor cache lookup patterns over time.

```sql
select
  date_trunc('hour', timestamp) as hour,
  json_extract(http_request, '$.cache_lookup') as cache_lookup,
  count(*) as request_count
from
  gcp_requests_log
where
  json_extract(http_request, '$.cache_lookup') = 'true'
group by
  hour,
  json_extract(http_request, '$.cache_lookup')
order by
  hour desc,
  json_extract(http_request, '$.cache_lookup');
```

```yaml
folder: Performance
```

### Top requested endpoints

Identify the most frequently requested endpoints.

```sql
select
  json_extract(http_request, '$.request_url') as url,
  json_extract(http_request, '$.request_method') as method,
  count(*) as request_count
from
  gcp_requests_log
where
  json_extract(http_request, '$.request_url') is not null
group by
  json_extract(http_request, '$.request_url'),
  json_extract(http_request, '$.request_method')
order by
  request_count desc
limit 20;
```

```yaml
folder: Performance
```

## Traffic Analysis Examples

### Request volume by hour

Analyze request volume patterns over time to identify traffic trends.

```sql
select
  date_trunc('hour', timestamp) as hour,
  count(*) as total_requests,
  count(case when json_extract(enforced_security_policy, '$.outcome') = '"DENY"' then 1 end) as blocked_requests
from
  gcp_requests_log
group by
  hour
order by
  hour desc
limit 24;
```

```yaml
folder: Traffic
```

### HTTP status code distribution

Analyze the distribution of HTTP status codes to understand request outcomes.

```sql
select
  json_extract(http_request, '$.status') as status_code,
  count(*) as request_count,
  round(
    (count(*)::float / sum(count(*)) over()) * 100,
    2
  ) as percentage
from
  gcp_requests_log
where
  json_extract(http_request, '$.status') is not null
group by
  json_extract(http_request, '$.status')
order by
  request_count desc;
```

```yaml
folder: Traffic
```

### User agent analysis

Analyze user agents to identify different client types and potential threats.

```sql
select
  json_extract(http_request, '$.user_agent') as user_agent,
  count(*) as request_count,
  count(distinct json_extract(http_request, '$.remote_ip')) as unique_ips
from
  gcp_requests_log
where
  json_extract(http_request, '$.user_agent') is not null
group by
  json_extract(http_request, '$.user_agent')
order by
  request_count desc
limit 20;
```

```yaml
folder: Traffic
```

## Security Monitoring Examples

### Recent security events

Monitor recent security events including blocked requests and policy violations.

```sql
select
  timestamp,
  json_extract(http_request, '$.remote_ip') as remote_ip,
  json_extract(http_request, '$.request_url') as url,
  json_extract(enforced_security_policy, '$.outcome') as security_outcome,
  json_extract(enforced_security_policy, '$.name') as policy_name
from
  gcp_requests_log
where
  json_extract(enforced_security_policy, '$.outcome') = '"DENY"'
order by
  timestamp desc
limit 100;
```

```yaml
folder: Security
```

### Preview mode analysis

Analyze security policies in preview mode to understand their impact before enforcement.

```sql
select
  json_extract(preview_security_policy, '$.name') as rule_name,
  count(*) as total_requests,
  count(case when json_extract(preview_security_policy, '$.outcome') = 'DENY' then 1 end) as would_block_requests
from
  gcp_requests_log
where
  preview_security_policy is not null
group by
  json_extract(preview_security_policy, '$.name')
order by
  total_requests desc;
```

```yaml
folder: Security
```

### Large request analysis

Identify and analyze large requests that may indicate potential attacks or unusual traffic.

```sql
select
  timestamp,
  json_extract(http_request, '$.remote_ip') as remote_ip,
  json_extract(http_request, '$.request_url') as url,
  json_extract(http_request, '$.request_size') as request_size,
  json_extract(http_request, '$.response_size') as response_size
from
  gcp_requests_log
where
  cast(json_extract(http_request, '$.request_size') as integer) > 1000000  -- 1MB
  or cast(json_extract(http_request, '$.response_size') as integer) > 1000000
order by
  timestamp desc
limit 50;
```

```yaml
folder: Security
```
