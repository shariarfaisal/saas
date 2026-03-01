## ADDED Requirements

### Requirement: Category CRUD with Real API
Menu categories SHALL be fetched from `GET /partner/restaurants/:id/categories`, created via `POST /partner/restaurants/:id/categories`, updated via `PUT /partner/categories/:id`, deleted via `DELETE /partner/categories/:id`, and reordered via `PUT /partner/restaurants/:id/categories/reorder`. The left-panel category list currently uses `useState(mockCategories)` and SHALL be replaced with a `useQuery` hook.

#### Scenario: Category list loads from API
- **WHEN** a partner navigates to the menu page
- **THEN** categories are fetched from the API and displayed in the left panel

#### Scenario: Partner creates a new category
- **WHEN** the partner submits the add-category form
- **THEN** `POST /partner/restaurants/:id/categories` is called
- **AND** the new category appears in the panel without page reload

#### Scenario: Partner reorders categories via drag-drop
- **WHEN** the partner drags a category to a new position
- **THEN** `PUT /partner/restaurants/:id/categories/reorder` is called with the new sort order array
- **AND** the panel reflects the new order persistently

### Requirement: Product CRUD with Real API
Products SHALL be fetched by category via `GET /partner/restaurants/:id/products?category_id=:catId`, created via `POST /partner/restaurants/:id/products`, updated via `PUT /partner/products/:id`, deleted via `DELETE /partner/products/:id`, and availability toggled via `PATCH /partner/products/:id/availability`. The product grid currently uses `useState(mockProducts)` and SHALL be replaced.

#### Scenario: Products load per selected category
- **WHEN** the partner selects a category from the left panel
- **THEN** products for that category are fetched and displayed in the right grid

#### Scenario: Partner toggles product availability
- **WHEN** the partner clicks the availability toggle on a product card
- **THEN** `PATCH /partner/products/:id/availability` is called
- **AND** the toggle visually reflects the new state

### Requirement: Product Variant Builder — Functional
The product create/edit sheet SHALL support adding, editing, and removing price variants (e.g., Small / Medium / Large each with their own price). Variant entries SHALL be stored as a `variants` JSON array in the product payload. The current "Add Variant" button exists but has no add/remove logic.

#### Scenario: Partner adds a variant to a product
- **WHEN** the partner clicks "Add Variant" and fills in name and price
- **THEN** a new variant row appears in the form
- **AND** on save, the variant is included in the `POST /PUT /partner/products` payload

#### Scenario: Partner removes a variant
- **WHEN** the partner clicks the remove icon on a variant row
- **THEN** the row is deleted from the form
- **AND** on save, the removed variant is absent from the payload

### Requirement: Product Addon Builder — Functional
The product create/edit sheet SHALL support adding, editing, and removing add-on groups (e.g., "Extra Toppings" with individual add-on items and prices). The current "Add Addon" button exists but has no logic.

#### Scenario: Partner adds an addon group
- **WHEN** the partner fills in addon group name and adds items with prices
- **THEN** the addon group is included in the product save payload

### Requirement: Bulk CSV Product Import
The bulk CSV upload modal SHALL validate the uploaded file client-side (headers, required fields, price format) before uploading to `POST /partner/restaurants/:id/products/bulk-upload`. Validation errors SHALL be displayed per row in a preview table before submission.

#### Scenario: Partner uploads a valid CSV
- **WHEN** the partner selects a CSV with correct format and clicks Upload
- **THEN** client-side validation passes and the file is POSTed to the bulk-upload endpoint
- **AND** a success toast shows the count of products imported

#### Scenario: CSV has validation errors
- **WHEN** the partner uploads a CSV with missing required fields
- **THEN** a row-by-row error table is shown before upload
- **AND** the Upload button is disabled until errors are resolved

### Requirement: Product Image Upload
The product create/edit sheet image field SHALL upload selected images to `POST /partner/media/upload` (multipart/form-data) and store the returned CDN URL in the product payload. The current file input has no upload logic.

#### Scenario: Partner uploads a product image
- **WHEN** the partner selects an image file in the product form
- **THEN** the file is uploaded to the media API
- **AND** the returned CDN URL is stored and a preview is shown in the form
