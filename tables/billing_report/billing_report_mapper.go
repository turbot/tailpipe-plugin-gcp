package billing_report

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/tailpipe-plugin-sdk/mappers"
)

type BillingReportMapper struct{}

func NewBillingReportMapper() *BillingReportMapper {
	return &BillingReportMapper{}
}

func (m *BillingReportMapper) Identifier() string {
	return "gcp_billing_report_mapper"
}

func (m *BillingReportMapper) Map(_ context.Context, a any, _ ...mappers.MapOption[*BillingReport]) (*BillingReport, error) {
	var input []byte
	switch v := a.(type) {
	case []byte:
		input = v
	case string:
		input = []byte(v)
	default:
		return nil, fmt.Errorf("unable to map row, invalid type received")
	}

	var raw map[string]any
	if err := json.Unmarshal(input, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal billing report row to map: %w", err)
	}

	report := &BillingReport{}

	// Helper to parse string time fields
	parseTime := func(val any) *time.Time {
		if s, ok := val.(string); ok && s != "" {
			t, err := helpers.ParseTime(s)
			if err == nil {
				return &t
			}
		}
		return nil
	}

	// Top-level fields
	if v, ok := raw["billing_account_id"].(string); ok {
		report.BillingAccountID = &v
	}
	if v, ok := raw["invoice"].(map[string]any); ok {
		if s, ok := v["month"].(string); ok {
			report.InvoiceMonth = &s
		}
	}
	if v, ok := raw["invoice_publisher_type"].(string); ok {
		report.InvoicePublisherType = &v
	}
	if v, ok := raw["cost_type"].(string); ok {
		report.CostType = &v
	}
	if v, ok := raw["service"].(map[string]any); ok {
		if s, ok := v["id"].(string); ok {
			report.ServiceID = &s
		}
		if s, ok := v["description"].(string); ok {
			report.ServiceDescription = &s
		}
	}
	if v, ok := raw["sku"].(map[string]any); ok {
		if s, ok := v["id"].(string); ok {
			report.SkuID = &s
		}
		if s, ok := v["description"].(string); ok {
			report.SkuDescription = &s
		}
	}
	if v := parseTime(raw["usage_start_time"]); v != nil {
		report.UsageStartTime = v
	}
	if v := parseTime(raw["usage_end_time"]); v != nil {
		report.UsageEndTime = v
	}
	if v := parseTime(raw["export_time"]); v != nil {
		report.ExportTime = v
	}
	if v, ok := raw["cost"].(float64); ok {
		report.Cost = &v
	}
	if v, ok := raw["currency"].(string); ok {
		report.Currency = &v
	}
	if v, ok := raw["currency_conversion_rate"].(float64); ok {
		report.CurrencyConversionRate = &v
	}
	if v, ok := raw["transaction_type"].(string); ok {
		report.TransactionType = &v
	}
	if v, ok := raw["seller_name"].(string); ok {
		report.SellerName = &v
	}
	if v, ok := raw["cost_at_list"].(float64); ok {
		report.CostAtList = &v
	}

	// Usage
	if v, ok := raw["usage"].(map[string]any); ok {
		if f, ok := v["amount"].(float64); ok {
			report.UsageAmount = &f
		}
		if s, ok := v["unit"].(string); ok {
			report.UsageUnit = &s
		}
		if f, ok := v["amount_in_pricing_units"].(float64); ok {
			report.UsageAmountInPricingUnit = &f
		}
		if s, ok := v["pricing_unit"].(string); ok {
			report.UsagePricingUnit = &s
		}
	}

	// Project
	if v, ok := raw["project"].(map[string]any); ok {
		proj := &BillingProject{}
		if s, ok := v["id"].(string); ok {
			proj.ID = &s
		}
		if s, ok := v["number"].(string); ok {
			proj.Number = &s
		}
		if s, ok := v["name"].(string); ok {
			proj.Name = &s
		}
		if s, ok := v["ancestry_numbers"].(string); ok {
			proj.AncestryNumbers = &s
		}
		// Labels
		if arr, ok := v["labels"].([]any); ok {
			labels := make([]BillingLabel, 0, len(arr))
			for _, item := range arr {
				if m, ok := item.(map[string]any); ok {
					label := BillingLabel{}
					if k, ok := m["key"].(string); ok {
						label.Key = &k
					}
					if val, ok := m["value"].(string); ok {
						label.Value = &val
					}
					labels = append(labels, label)
				}
			}
			proj.Labels = &labels
		}
		// Ancestors
		if arr, ok := v["ancestors"].([]any); ok {
			ancestors := make([]BillingAncestor, 0, len(arr))
			for _, item := range arr {
				if m, ok := item.(map[string]any); ok {
					ancestor := BillingAncestor{}
					if rn, ok := m["resource_name"].(string); ok {
						ancestor.ResourceName = &rn
					}
					if dn, ok := m["display_name"].(string); ok {
						ancestor.DisplayName = &dn
					}
					ancestors = append(ancestors, ancestor)
				}
			}
			proj.Ancestors = &ancestors
		}
		report.Project = proj
	}

	// Labels
	if arr, ok := raw["labels"].([]any); ok {
		labels := make([]BillingLabel, 0, len(arr))
		for _, item := range arr {
			if m, ok := item.(map[string]any); ok {
				label := BillingLabel{}
				if k, ok := m["key"].(string); ok {
					label.Key = &k
				}
				if val, ok := m["value"].(string); ok {
					label.Value = &val
				}
				labels = append(labels, label)
			}
		}
		report.Labels = &labels
	}
	if arr, ok := raw["system_labels"].([]any); ok {
		labels := make([]BillingLabel, 0, len(arr))
		for _, item := range arr {
			if m, ok := item.(map[string]any); ok {
				label := BillingLabel{}
				if k, ok := m["key"].(string); ok {
					label.Key = &k
				}
				if val, ok := m["value"].(string); ok {
					label.Value = &val
				}
				labels = append(labels, label)
			}
		}
		report.SystemLabels = &labels
	}

	// Location
	if v, ok := raw["location"].(map[string]any); ok {
		loc := &BillingLocation{}
		if s, ok := v["location"].(string); ok {
			loc.Location = &s
		}
		if s, ok := v["country"].(string); ok {
			loc.Country = &s
		}
		if s, ok := v["region"].(string); ok {
			loc.Region = &s
		}
		if s, ok := v["zone"].(string); ok {
			loc.Zone = &s
		}
		report.Location = loc
	}

	// Credits
	if arr, ok := raw["credits"].([]any); ok {
		credits := make([]BillingCredit, 0, len(arr))
		for _, item := range arr {
			if m, ok := item.(map[string]any); ok {
				credit := BillingCredit{}
				if s, ok := m["id"].(string); ok {
					credit.ID = &s
				}
				if s, ok := m["full_name"].(string); ok {
					credit.FullName = &s
				}
				if s, ok := m["type"].(string); ok {
					credit.Type = &s
				}
				if s, ok := m["name"].(string); ok {
					credit.Name = &s
				}
				if f, ok := m["amount"].(float64); ok {
					credit.Amount = &f
				}
				credits = append(credits, credit)
			}
		}
		report.Credits = &credits
	}

	// AdjustmentInfo
	if v, ok := raw["adjustment_info"].(map[string]any); ok {
		adj := &BillingAdjustmentInfo{}
		if s, ok := v["id"].(string); ok {
			adj.ID = &s
		}
		if s, ok := v["description"].(string); ok {
			adj.Description = &s
		}
		if s, ok := v["type"].(string); ok {
			adj.Type = &s
		}
		if s, ok := v["mode"].(string); ok {
			adj.Mode = &s
		}
		report.AdjustmentInfo = adj
	}

	// Tags
	if arr, ok := raw["tags"].([]any); ok {
		tags := make([]BillingTag, 0, len(arr))
		for _, item := range arr {
			if m, ok := item.(map[string]any); ok {
				tag := BillingTag{}
				if k, ok := m["key"].(string); ok {
					tag.Key = &k
				}
				if v, ok := m["value"].(string); ok {
					tag.Value = &v
				}
				if ns, ok := m["namespace"].(string); ok {
					tag.Namespace = &ns
				}
				if inh, ok := m["inherited"].(bool); ok {
					tag.Inherited = &inh
				}
				tags = append(tags, tag)
			}
		}
		report.Tags = &tags
	}

	// Resource
	if v, ok := raw["resource"].(map[string]any); ok {
		res := &BillingResource{}
		if s, ok := v["global_name"].(string); ok {
			res.GlobalName = &s
		}
		if s, ok := v["name"].(string); ok {
			res.Name = &s
		}
		report.Resource = res
	}

	// Price
	if v, ok := raw["price"].(map[string]any); ok {
		price := &BillingPrice{}
		if f, ok := v["effective_price"].(float64); ok {
			price.EffectivePrice = &f
		}
		if f, ok := v["tier_start_amount"].(float64); ok {
			price.TierStartAmount = &f
		}
		if s, ok := v["unit"].(string); ok {
			price.Unit = &s
		}
		if f, ok := v["pricing_unit_quantity"].(float64); ok {
			price.PricingUnitQuantity = &f
		}
		report.Price = price
	}

	// Subscription
	if v, ok := raw["subscription"].(map[string]any); ok {
		sub := &BillingSubscription{}
		if s, ok := v["instance_id"].(string); ok {
			sub.InstanceID = &s
		}
		report.Subscription = sub
	}

	return report, nil
}
