# Example Queries for gcp_billing_report

This page provides example SQL queries for analyzing Google Cloud billing data using the `gcp_billing_report` table.

---

## 1. Total Cost by Month

Summarize total GCP spend by invoice month.

```sql
SELECT
  invoice_month,
  SUM(cost) AS total_cost
FROM gcp_billing_report
GROUP BY invoice_month
ORDER BY invoice_month DESC;
```

---

## 2. Top Services by Cost (Last 90 Days)

Identify the top 10 GCP services by cost in the last 90 days.

```sql
SELECT
  service_description,
  SUM(cost) AS total_cost
FROM gcp_billing_report
WHERE usage_start_time >= DATE_SUB(CURRENT_DATE(), INTERVAL 90 DAY)
GROUP BY service_description
ORDER BY total_cost DESC
LIMIT 10;
```

---

## 3. Credits Applied by Type

Show total credits applied, grouped by credit type.

```sql
SELECT
  c.type AS credit_type,
  SUM(c.amount) AS total_credits
FROM gcp_billing_report, UNNEST(credits) AS c
GROUP BY credit_type
ORDER BY total_credits DESC;
```

---

## 4. Project-Level Spend (Current Year)

Break down spend by project for the current year.

```sql
SELECT
  project.id AS project_id,
  project.name AS project_name,
  SUM(cost) AS total_cost
FROM gcp_billing_report
WHERE EXTRACT(YEAR FROM usage_start_time) = EXTRACT(YEAR FROM CURRENT_DATE())
GROUP BY project_id, project_name
ORDER BY total_cost DESC;
```

---

## 5. Cost by SKU and Region

Analyze cost by SKU and region.

```sql
SELECT
  sku_description,
  location.region,
  SUM(cost) AS total_cost
FROM gcp_billing_report
GROUP BY sku_description, location.region
ORDER BY total_cost DESC;
```

---

## 6. Unused Commitments

Find commitment charges with zero usage.

```sql
SELECT
  subscription.id AS commitment_id,
  SUM(cost) AS total_commitment_cost,
  SUM(usage_amount) AS total_usage
FROM gcp_billing_report
WHERE subscription.id IS NOT NULL
GROUP BY commitment_id
HAVING total_usage = 0
ORDER BY total_commitment_cost DESC;
``` 