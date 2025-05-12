package billing_report

import (
	"time"

	"github.com/turbot/tailpipe-plugin-sdk/schema"
)

type BillingReport struct {
	schema.CommonFields

	BillingAccountID     *string    `json:"billing_account_id,omitempty" parquet:"name=billing_account_id"`
	InvoiceMonth         *string    `json:"invoice_month,omitempty" parquet:"name=invoice_month"`
	InvoicePublisherType *string    `json:"invoice_publisher_type,omitempty" parquet:"name=invoice_publisher_type"`
	CostType             *string    `json:"cost_type,omitempty" parquet:"name=cost_type"`
	ServiceID            *string    `json:"service_id,omitempty" parquet:"name=service_id"`
	ServiceDescription   *string    `json:"service_description,omitempty" parquet:"name=service_description"`
	SkuID                *string    `json:"sku_id,omitempty" parquet:"name=sku_id"`
	SkuDescription       *string    `json:"sku_description,omitempty" parquet:"name=sku_description"`
	UsageStartTime       *time.Time `json:"usage_start_time,omitempty" parquet:"name=usage_start_time"`
	UsageEndTime         *time.Time `json:"usage_end_time,omitempty" parquet:"name=usage_end_time"`

	Project                  *BillingProject        `json:"project,omitempty" parquet:"name=project"`
	Labels                   *[]BillingLabel        `json:"labels,omitempty" parquet:"name=labels"`
	SystemLabels             *[]BillingLabel        `json:"system_labels,omitempty" parquet:"name=system_labels"`
	Location                 *BillingLocation       `json:"location,omitempty" parquet:"name=location"`
	Cost                     *float64               `json:"cost,omitempty" parquet:"name=cost"`
	Currency                 *string                `json:"currency,omitempty" parquet:"name=currency"`
	CurrencyConversionRate   *float64               `json:"currency_conversion_rate,omitempty" parquet:"name=currency_conversion_rate"`
	UsageAmount              *float64               `json:"usage_amount,omitempty" parquet:"name=usage_amount"`
	UsageUnit                *string                `json:"usage_unit,omitempty" parquet:"name=usage_unit"`
	UsageAmountInPricingUnit *float64               `json:"usage_amount_in_pricing_units,omitempty" parquet:"name=usage_amount_in_pricing_units"`
	UsagePricingUnit         *string                `json:"usage_pricing_unit,omitempty" parquet:"name=usage_pricing_unit"`
	Credits                  *[]BillingCredit       `json:"credits,omitempty" parquet:"name=credits"`
	AdjustmentInfo           *BillingAdjustmentInfo `json:"adjustment_info,omitempty" parquet:"name=adjustment_info"`
	ExportTime               *time.Time             `json:"export_time,omitempty" parquet:"name=export_time"`
	Tags                     *[]BillingTag          `json:"tags,omitempty" parquet:"name=tags"`
	CostAtList               *float64               `json:"cost_at_list,omitempty" parquet:"name=cost_at_list"`
	TransactionType          *string                `json:"transaction_type,omitempty" parquet:"name=transaction_type"`
	SellerName               *string                `json:"seller_name,omitempty" parquet:"name=seller_name"`

	// Detailed export fields
	Resource     *BillingResource     `json:"resource,omitempty" parquet:"name=resource"`
	Price        *BillingPrice        `json:"price,omitempty" parquet:"name=price"`
	Subscription *BillingSubscription `json:"subscription,omitempty" parquet:"name=subscription"`
}

type BillingProject struct {
	ID              *string            `json:"id,omitempty" parquet:"name=id"`
	Number          *string            `json:"number,omitempty" parquet:"name=number"`
	Name            *string            `json:"name,omitempty" parquet:"name=name"`
	AncestryNumbers *string            `json:"ancestry_numbers,omitempty" parquet:"name=ancestry_numbers"`
	Ancestors       *[]BillingAncestor `json:"ancestors,omitempty" parquet:"name=ancestors"`
	Labels          *[]BillingLabel    `json:"labels,omitempty" parquet:"name=labels"`
}

type BillingAncestor struct {
	ResourceName *string `json:"resource_name,omitempty" parquet:"name=resource_name"`
	DisplayName  *string `json:"display_name,omitempty" parquet:"name=display_name"`
}

type BillingLabel struct {
	Key   *string `json:"key,omitempty" parquet:"name=key"`
	Value *string `json:"value,omitempty" parquet:"name=value"`
}

type BillingLocation struct {
	Location *string `json:"location,omitempty" parquet:"name=location"`
	Country  *string `json:"country,omitempty" parquet:"name=country"`
	Region   *string `json:"region,omitempty" parquet:"name=region"`
	Zone     *string `json:"zone,omitempty" parquet:"name=zone"`
}

type BillingCredit struct {
	ID       *string  `json:"id,omitempty" parquet:"name=id"`
	FullName *string  `json:"full_name,omitempty" parquet:"name=full_name"`
	Type     *string  `json:"type,omitempty" parquet:"name=type"`
	Name     *string  `json:"name,omitempty" parquet:"name=name"`
	Amount   *float64 `json:"amount,omitempty" parquet:"name=amount"`
}

type BillingAdjustmentInfo struct {
	ID          *string `json:"id,omitempty" parquet:"name=id"`
	Description *string `json:"description,omitempty" parquet:"name=description"`
	Type        *string `json:"type,omitempty" parquet:"name=type"`
	Mode        *string `json:"mode,omitempty" parquet:"name=mode"`
}

type BillingTag struct {
	Key       *string `json:"key,omitempty" parquet:"name=key"`
	Value     *string `json:"value,omitempty" parquet:"name=value"`
	Namespace *string `json:"namespace,omitempty" parquet:"name=namespace"`
	Inherited *bool   `json:"inherited,omitempty" parquet:"name=inherited"`
}

type BillingResource struct {
	GlobalName *string `json:"global_name,omitempty" parquet:"name=global_name"`
	Name       *string `json:"name,omitempty" parquet:"name=name"`
}

type BillingPrice struct {
	EffectivePrice      *float64 `json:"effective_price,omitempty" parquet:"name=effective_price"`
	TierStartAmount     *float64 `json:"tier_start_amount,omitempty" parquet:"name=tier_start_amount"`
	Unit                *string  `json:"unit,omitempty" parquet:"name=unit"`
	PricingUnitQuantity *float64 `json:"pricing_unit_quantity,omitempty" parquet:"name=pricing_unit_quantity"`
}

type BillingSubscription struct {
	InstanceID *string `json:"instance_id,omitempty" parquet:"name=instance_id"`
}
