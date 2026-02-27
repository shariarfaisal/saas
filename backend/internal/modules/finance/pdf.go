package finance

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/munchies/platform/backend/internal/db/sqlc"
)

const invoiceHTMLTemplate = `<!DOCTYPE html>
<html>
<head><style>
body { font-family: Arial, sans-serif; margin: 40px; }
h1 { color: #333; }
table { width: 100%%; border-collapse: collapse; margin: 20px 0; }
th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
th { background-color: #f2f2f2; }
.total { font-weight: bold; font-size: 1.2em; }
.header { display: flex; justify-content: space-between; }
</style></head>
<body>
<h1>Invoice {{.InvoiceNumber}}</h1>
<p><strong>Period:</strong> {{.PeriodStart}} to {{.PeriodEnd}}</p>
<p><strong>Status:</strong> {{.Status}}</p>
<table>
<tr><th>Description</th><th>Amount</th></tr>
<tr><td>Gross Sales</td><td>{{.GrossSales}}</td></tr>
<tr><td>Item Discounts</td><td>-{{.ItemDiscounts}}</td></tr>
<tr><td>Vendor Promo Discounts</td><td>-{{.VendorPromoDiscounts}}</td></tr>
<tr><td>Net Sales</td><td>{{.NetSales}}</td></tr>
<tr><td>VAT Collected</td><td>{{.VatCollected}}</td></tr>
<tr><td>Commission ({{.CommissionRate}}%%)</td><td>-{{.CommissionAmount}}</td></tr>
<tr><td>Penalties</td><td>-{{.PenaltyAmount}}</td></tr>
<tr><td>Adjustments</td><td>{{.AdjustmentAmount}}</td></tr>
<tr class="total"><td>Net Payable</td><td>{{.NetPayable}}</td></tr>
</table>
<p><strong>Total Orders:</strong> {{.TotalOrders}} | <strong>Delivered:</strong> {{.DeliveredOrders}} | <strong>Cancelled:</strong> {{.CancelledOrders}} | <strong>Rejected:</strong> {{.RejectedOrders}}</p>
</body>
</html>`

// GenerateInvoicePDF generates a PDF representation of an invoice.
// For now, it returns the HTML as a simple PDF-like document.
// In production, this would use wkhtmltopdf or a headless browser.
func GenerateInvoicePDF(inv *sqlc.Invoice) ([]byte, error) {
	tmpl, err := template.New("invoice").Parse(invoiceHTMLTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	data := map[string]interface{}{
		"InvoiceNumber":        inv.InvoiceNumber,
		"PeriodStart":          inv.PeriodStart.Format("2006-01-02"),
		"PeriodEnd":            inv.PeriodEnd.Format("2006-01-02"),
		"Status":               inv.Status,
		"GrossSales":           inv.GrossSales.StringFixed(2),
		"ItemDiscounts":        inv.ItemDiscounts.StringFixed(2),
		"VendorPromoDiscounts": inv.VendorPromoDiscounts.StringFixed(2),
		"NetSales":             inv.NetSales.StringFixed(2),
		"VatCollected":         inv.VatCollected.StringFixed(2),
		"CommissionRate":       inv.CommissionRate.StringFixed(2),
		"CommissionAmount":     inv.CommissionAmount.StringFixed(2),
		"PenaltyAmount":        inv.PenaltyAmount.StringFixed(2),
		"AdjustmentAmount":     inv.AdjustmentAmount.StringFixed(2),
		"NetPayable":           inv.NetPayable.StringFixed(2),
		"TotalOrders":          inv.TotalOrders,
		"DeliveredOrders":      inv.DeliveredOrders,
		"CancelledOrders":      inv.CancelledOrders,
		"RejectedOrders":       inv.RejectedOrders,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	return buf.Bytes(), nil
}
