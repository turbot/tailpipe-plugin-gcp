## Activity Examples

### Daily Cost Trends

Track daily GCP spending trends.

```sql
select
  date_trunc('day', usage_start_time) as usage_date,
  sum(cost) as daily_cost,
  currency
from
  gcp_billing_report
where
  usage_start_time >= current_date - interval '30 days'
group by
  usage_date,
  currency
order by
  usage_date desc;
```

```yaml
folder: Billing
```

### Top 10 Costly Services

Identify the most expensive GCP services over the past month.

```sql
select
  service_description as service,
  sum(cost) as total_cost,
  currency
from
  gcp_billing_report
where
  usage_start_time >= current_date - interval '1 month'
group by
  service,
  currency
order by
  total_cost desc
limit 10;
```

```yaml
folder: Billing
```

### Top 10 Spending Projects

Determine which GCP projects are generating the highest costs.

```sql
select
  project_id,
  project_name,
  sum(cost) as total_cost,
  currency
from
  gcp_billing_report
group by
  project_id,
  project_name,
  currency
order by
  total_cost desc
limit 10;
```

```yaml
folder: Billing
```

## Detection Examples

### Unusual Cost Spikes

Detect services with a sudden increase in costs.

```sql
with monthly_costs as (
  select
    date_trunc('month', usage_start_time) as month,
    service_description as service,
    sum(cost) as total_cost
  from
    gcp_billing_report
  where
    usage_start_time >= current_date - interval '2 months'
  group by
    month, service
)
select
  current_month.service,
  previous_month.total_cost as previous_cost,
  current_month.total_cost as current_cost,
  round(((current_month.total_cost - previous_month.total_cost) / previous_month.total_cost) * 100, 2) as percentage_increase
from
  monthly_costs current_month
join
  monthly_costs previous_month
  on current_month.service = previous_month.service
  and current_month.month = date_trunc('month', current_date)
  and previous_month.month = date_trunc('month', current_date) - interval '1 month'
where
  previous_month.total_cost > 0
  and ((current_month.total_cost - previous_month.total_cost) / previous_month.total_cost) > 0.2
order by
  percentage_increase desc;
```

```yaml
folder: Billing
```

### Top 10 Projects by Network Egress Usage

Find projects with high network egress usage.

```sql
select
  project_id,
  project_name,
  sum(usage.amount) as total_egress_gb,
  currency,
  sum(cost) as total_egress_cost
from
  gcp_billing_report
where
  sku_description ilike '%egress%'
  or sku_description ilike '%internet%'
group by
  project_id,
  project_name,
  currency
order by
  total_egress_gb desc
limit 10;
```

```yaml
folder: Billing
```

## Operational Examples

### Compute Engine Cost Breakdown by Machine Type

Analyze Compute Engine costs based on machine types.

```sql
select
  sku_description as machine_type,
  count(distinct project_id) as project_count,
  sum(cost) as total_cost,
  sum(usage.amount) as total_usage,
  usage.unit,
  currency
from
  gcp_billing_report
where
  usage_start_time >= current_date - interval '30 days'
  and service_description = 'Compute Engine'
  and sku_description ilike '%instance%'
group by
  machine_type,
  usage.unit,
  currency
order by
  total_cost desc;
```

```yaml
folder: Compute
```

### Top 10 Storage Costs by SKU

Identify expensive storage usage across different GCP storage services.

```sql
select
  project_id,
  service_description,
  sku_description,
  sum(cost) as total_cost,
  sum(usage.amount) as total_usage,
  usage.unit,
  currency
from
  gcp_billing_report
where
  usage_start_time >= current_date - interval '30 days'
  and (service_description ilike '%storage%'
    or service_description ilike '%filestore%'
    or service_description ilike '%disk%')
group by
  project_id,
  service_description,
  sku_description,
  usage.unit,
  currency
order by
  total_cost desc
limit 10;
```

```yaml
folder: Storage
```

## Volume Examples

### Sustained Use Discount Utilization

Calculate the percentage of Compute Engine usage covered by sustained use discounts.

```sql
with total_compute as (
  select
    sum(cost) as total_compute_cost
  from
    gcp_billing_report
  where
    usage_start_time >= current_date - interval '30 days'
    and service_description = 'Compute Engine'
),
sustained_use as (
  select
    sum(cost) as sustained_use_cost
  from
    gcp_billing_report
  where
    usage_start_time >= current_date - interval '30 days'
    and service_description = 'Compute Engine'
    and sku_description ilike '%sustained use%'
)
select
  t.total_compute_cost,
  s.sustained_use_cost,
  round((abs(s.sustained_use_cost) / t.total_compute_cost) * 100, 2) as sustained_use_coverage_percent
from
  total_compute t,
  sustained_use s;
```

```yaml
folder: Compute
```

### Top 10 High-Volume Service Usage

Detect GCP services generating high volume usage.

```sql
select
  service_description as service,
  sum(usage.amount) as total_usage,
  usage.unit,
  sum(cost) as total_cost,
  currency
from
  gcp_billing_report
where
  usage_start_time >= current_date - interval '30 days'
  and usage.amount is not null
group by
  service,
  usage.unit,
  currency
order by
  total_usage desc
limit 10;
```

```yaml
folder: Billing
```

## Baseline Examples

### Cost Breakdown by GCP Region

Compare spending across GCP regions.

```sql
select
  location.region as region,
  sum(cost) as total_cost,
  currency,
  count(*) as line_items
from
  gcp_billing_report
where
  location.region is not null
group by
  region,
  currency
order by
  total_cost desc;
```

```yaml
folder: Billing
```

### Services with Unexpected Costs

Identify services that usually have low costs but show unexpected spending increases.

```sql
with avg_service_cost as (
  select
    service_description as service,
    avg(cost) as avg_monthly_cost
  from
    gcp_billing_report
  where
    usage_start_time >= current_date - interval '6 months'
  group by
    service
)
select
  c.service,
  c.total_cost,
  a.avg_monthly_cost,
  round(((c.total_cost - a.avg_monthly_cost) / a.avg_monthly_cost) * 100, 2) as percentage_increase
from (
  select
    service_description as service,
    sum(cost) as total_cost
  from
    gcp_billing_report
  where
    usage_start_time >= current_date - interval '1 month'
  group by
    service
) c
join avg_service_cost a on c.service = a.service
where
  a.avg_monthly_cost > 0
  and ((c.total_cost - a.avg_monthly_cost) / a.avg_monthly_cost) > 0.2
order by
  percentage_increase desc;
```

```yaml
folder: Billing
```

### Cost Comparison Across Billing Periods

Compare costs between the current and previous billing period.

```sql
with current_period as (
  select
    service_description as service,
    sum(cost) as cost,
    currency
  from
    gcp_billing_report
  where
    invoice_month = (
      select max(invoice_month)
      from gcp_billing_report
    )
  group by
    service_description,
    currency
),
previous_period as (
  select
    service_description as service,
    sum(cost) as cost,
    currency
  from
    gcp_billing_report
  where
    invoice_month = (
      select max(invoice_month)
      from gcp_billing_report
      where invoice_month < (
        select max(invoice_month)
        from gcp_billing_report
      )
    )
  group by
    service_description,
    currency
)
select
  current_period.service,
  current_period.cost as current_cost,
  previous_period.cost as previous_cost,
  (current_period.cost - coalesce(previous_period.cost, 0)) as cost_difference,
  case
    when previous_period.cost > 0 then
      ((current_period.cost - previous_period.cost) / previous_period.cost) * 100
    else
      null
  end as percentage_change,
  current_period.currency
from
  current_period
  left join previous_period on (
    current_period.service = previous_period.service
    and current_period.currency = previous_period.currency
  )
order by
  current_cost desc;
```

```yaml
folder: Billing
```
