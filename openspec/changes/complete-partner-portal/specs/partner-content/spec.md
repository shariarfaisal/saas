## ADDED Requirements

### Requirement: Banner CRUD with API and Image Upload
The banners page SHALL fetch banners from `GET /partner/content/banners?restaurant_id=:id`, create via `POST /partner/content/banners`, update via `PUT /partner/content/banners/:id`, delete via `DELETE /partner/content/banners/:id`, and reorder via `PUT /partner/content/banners/reorder`. Banner images SHALL be uploaded to `POST /partner/media/upload` before form submission. The page currently uses `useState(mockBanners)` with no persistence.

#### Scenario: Banner list loads from API
- **WHEN** a partner navigates to the banners page
- **THEN** banners are fetched and listed with image previews and status badges

#### Scenario: Partner adds a new banner with image
- **WHEN** the partner selects an image, fills in title, link config, date range, and submits
- **THEN** the image is uploaded first via the media API
- **AND** `POST /partner/content/banners` is called with the CDN URL in the payload
- **AND** the new banner appears in the list

#### Scenario: Partner reorders banners via drag-drop
- **WHEN** the partner drags a banner to a new sort position
- **THEN** `PUT /partner/content/banners/reorder` is called with the new order array

#### Scenario: Partner toggles banner active/inactive
- **WHEN** the partner clicks the active toggle on a banner
- **THEN** `PUT /partner/content/banners/:id` is called with the updated `is_active` value

### Requirement: Homepage Sections CRUD with API
The sections page SHALL fetch sections from `GET /partner/content/sections`, create via `POST /partner/content/sections`, update via `PUT /partner/content/sections/:id`, delete via `DELETE /partner/content/sections/:id`, and reorder via `PUT /partner/content/sections/reorder`. The page currently uses `useState(mockSections)`.

#### Scenario: Sections list loads from API
- **WHEN** a partner navigates to the sections page
- **THEN** homepage sections are fetched and displayed with type labels and item counts

#### Scenario: Partner creates a Featured Restaurants section
- **WHEN** the partner fills in section type and display order and submits
- **THEN** `POST /partner/content/sections` is called
- **AND** the new section appears in the draggable list

### Requirement: Stories CRUD with API and Media Upload
The stories page SHALL fetch stories from `GET /partner/content/stories?restaurant_id=:id`, create via `POST /partner/content/stories`, delete via `DELETE /partner/content/stories/:id`, and toggle active via `PUT /partner/content/stories/:id`. Story media (image or video) SHALL be uploaded via `POST /partner/media/upload`. The page currently uses `useState(mockStories)`.

#### Scenario: Stories load from API
- **WHEN** a partner navigates to the stories page
- **THEN** stories are fetched and displayed as cards with media thumbnails and expiry dates

#### Scenario: Partner adds a story with image
- **WHEN** the partner uploads an image, sets expiry date, and submits
- **THEN** the image is uploaded via the media API
- **AND** `POST /partner/content/stories` is called with the media URL
- **AND** the new story card appears in the grid
