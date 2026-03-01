## ADDED Requirements

### Requirement: Rider List from API
The rider list page SHALL fetch riders from `GET /partner/riders?restaurant_id=:id` and display name, hub, status badge (active / on_delivery / inactive), last known location, today's order count, and an availability toggle. The page currently uses `useState(mockRiders)`.

#### Scenario: Rider list loads from API
- **WHEN** a partner navigates to the riders page
- **THEN** riders are fetched and displayed with real status and location data

#### Scenario: Partner toggles rider availability
- **WHEN** the partner clicks the availability toggle
- **THEN** `PATCH /partner/riders/:id/availability` is called
- **AND** the badge updates to reflect the new status

### Requirement: Rider Create and Edit via API
The rider create form SHALL submit to `POST /partner/riders` and the edit form SHALL submit to `PUT /partner/riders/:id`. Fields include name, phone, email, hub assignment, vehicle type, and license plate. The forms currently have no API calls.

#### Scenario: Partner creates a new rider
- **WHEN** the partner fills in all required fields and submits
- **THEN** `POST /partner/riders` is called
- **AND** the new rider appears in the list

### Requirement: Rider Detail — Stats, Earnings, Attendance, Penalties from API
The rider detail page SHALL fetch full rider data from `GET /partner/riders/:id` and populate: stats (total deliveries, monthly deliveries, avg delivery time, rating, completion rate), earnings (today, this week, this month from `GET /partner/riders/:id/earnings`), penalties (list from `GET /partner/riders/:id/penalties`), and attendance calendar (30-day grid from `GET /partner/riders/:id/attendance`). All sections currently display hardcoded mock values.

#### Scenario: Rider detail shows real stats and earnings
- **WHEN** the partner opens a rider detail page
- **THEN** all stats, earnings, penalties, and attendance data are fetched from the API and displayed correctly

#### Scenario: Rider detail shows attendance calendar
- **WHEN** the attendance section loads
- **THEN** each day in the 30-day grid is colored correctly (present / absent / late) based on API data

### Requirement: Rider Search and Filter
The rider list SHALL support search by name or hub via a query parameter `q=` sent to the list endpoint. The existing search input SHALL debounce input and refetch.

#### Scenario: Partner searches for a rider by name
- **WHEN** the partner types a name in the search box
- **THEN** the list refetches with `q=<name>` after a 300 ms debounce
- **AND** only matching riders are shown
