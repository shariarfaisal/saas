package finance

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/jackc/pgx/v5/pgtype"
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
		"PeriodStart":          formatPgDate(inv.PeriodStart),
		"PeriodEnd":            formatPgDate(inv.PeriodEnd),
		"Status":               inv.Status,
		"GrossSales":           formatPgNumeric(inv.GrossSales),
		"ItemDiscounts":        formatPgNumeric(inv.ItemDiscounts),
		"VendorPromoDiscounts": formatPgNumeric(inv.VendorPromoDiscounts),
		"NetSales":             formatPgNumeric(inv.NetSales),
		"VatCollected":         formatPgNumeric(inv.VatCollected),
		"CommissionRate":       formatPgNumeric(inv.CommissionRate),
		"CommissionAmount":     formatPgNumeric(inv.CommissionAmount),
		"PenaltyAmount":        formatPgNumeric(inv.PenaltyAmount),
		"AdjustmentAmount":     formatPgNumeric(inv.AdjustmentAmount),
		"NetPayable":           formatPgNumeric(inv.NetPayable),
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

func formatPgDate(d pgtype.Date) string {
	if !d.Valid {
		return ""
	}
	return d.Time.Format("2006-01-02")
}

func formatPgNumeric(n pgtype.Numeric) string {
	if !n.Valid {
		return "0.00"
	}
	f, err := n.Float64Value()
	if err != nil {
		return "0.00"
	}
	return fmt.Sprintf("%.2f", f.Float64)
}
