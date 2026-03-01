## ADDED Requirements

### Requirement: Finance Summary from API
The finance overview page SHALL fetch the current period summary from `GET /partner/finance/summary` and display current period net payable, YTD gross sales, YTD commission deducted, and YTD net payable. The page currently uses `useState` with hardcoded figures.

#### Scenario: Finance summary loads real data
- **WHEN** a partner navigates to the finance page
- **THEN** the summary cards show real figures fetched from the API
- **AND** an outstanding balance alert is shown if net payable is overdue

### Requirement: Invoice List from API
The invoice list page SHALL fetch paginated invoices from `GET /partner/finance/invoices?restaurant_id=:id&page=:n` and display them with status badges (draft, finalized, paid, overdue). The page currently uses `useState(mockInvoices)`.

#### Scenario: Invoice list loads from API
- **WHEN** a partner navigates to the invoices page
- **THEN** invoices are fetched and displayed with correct status badges and amounts

#### Scenario: Partner filters invoices by status
- **WHEN** the partner selects a status filter
- **THEN** the list refetches with the status query param and shows filtered results

### Requirement: Invoice Detail from API
The invoice detail page (`/finance/invoices/[id]`) SHALL fetch full invoice data from `GET /partner/finance/invoices/:id` and render the complete financial breakdown: gross sales, item discounts, vendor promo discounts, net sales, commission, penalty deductions, adjustments, VAT collected, delivery revenue share, and net payable.

#### Scenario: Invoice detail page shows full breakdown
- **WHEN** the partner opens an invoice
- **THEN** all line items in the financial breakdown are populated from the API response

### Requirement: Invoice PDF Download
The "Download PDF" button on the invoice detail and list pages SHALL call `GET /partner/finance/invoices/:id/pdf` and trigger a file download. The button currently calls a no-op mock handler.

#### Scenario: Partner downloads invoice PDF
- **WHEN** the partner clicks "Download PDF"
- **THEN** a request is made to the PDF endpoint
- **AND** the browser triggers a file download with the PDF attachment
- **AND** a loading spinner is shown on the button while the request is in flight

### Requirement: Payment History from API
The payment history page SHALL fetch paginated payment transactions from `GET /partner/finance/payments?restaurant_id=:id` and display them with payment method, amount, status (completed / pending / failed), and date. The page currently uses `useState(mockPayments)`.

#### Scenario: Payment history loads from API
- **WHEN** a partner navigates to the payment history page
- **THEN** transactions are fetched and displayed in the table with correct status badges
