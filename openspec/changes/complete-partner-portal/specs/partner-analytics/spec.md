## ADDED Requirements

### Requirement: Sales Report from API
The sales report section SHALL fetch data from `GET /partner/analytics/sales?restaurant_id=:id&from=:date&to=:date&group_by=day|week|month` and populate the revenue vs net-payable bar chart and summary figures. The existing date range picker and group-by controls SHALL drive the query parameters. The page currently uses hardcoded Recharts data.

#### Scenario: Partner views daily sales report
- **WHEN** the partner selects a date range and "Group by Day"
- **THEN** the API is called with the correct params
- **AND** the bar chart updates with real revenue figures

#### Scenario: Partner changes grouping to weekly
- **WHEN** the partner changes group-by to "Week"
- **THEN** the query refetches with `group_by=week`
- **AND** the chart re-renders with weekly aggregates

### Requirement: Top-Selling Products from API
The top-selling products table SHALL fetch data from `GET /partner/analytics/top-products?restaurant_id=:id&from=:date&to=:date`. The table currently shows hardcoded mock rows.

#### Scenario: Top products table shows real rankings
- **WHEN** the partner opens the analytics page
- **THEN** products are listed by order count descending with real revenue and growth figures from the API

### Requirement: Peak Hours Heatmap from API
The peak hours heatmap SHALL fetch hourly order distribution from `GET /partner/analytics/peak-hours?restaurant_id=:id&from=:date&to=:date`. The 7-day × 15-hour grid SHALL be colored by relative order volume from the API response.

#### Scenario: Heatmap renders real intensity data
- **WHEN** the analytics page loads
- **THEN** the heatmap grid cells are colored based on actual hourly order counts

### Requirement: Order Status Breakdown Chart from API
The order status breakdown pie chart SHALL fetch counts from `GET /partner/analytics/order-breakdown?restaurant_id=:id&from=:date&to=:date` returning delivered, cancelled, rejected, and in-progress counts.

#### Scenario: Pie chart shows real status distribution
- **WHEN** the analytics page loads
- **THEN** the pie chart slices reflect real counts from the API

### Requirement: Rider Performance Table from API
The rider performance table SHALL fetch data from `GET /partner/analytics/rider-performance?restaurant_id=:id&from=:date&to=:date` listing each rider with total deliveries, average delivery time, rating, and completion rate.

#### Scenario: Rider performance table shows real data
- **WHEN** the analytics page loads
- **THEN** the table rows are populated with real per-rider metrics

### Requirement: CSV Export
The "Export CSV" button SHALL call `GET /partner/analytics/sales/export?restaurant_id=:id&from=:date&to=:date&format=csv` and trigger a file download. The button currently has no action.

#### Scenario: Partner exports analytics as CSV
- **WHEN** the partner clicks "Export CSV"
- **THEN** a download request is made and the browser downloads the CSV file
- **AND** a loading state is shown on the button during the request
