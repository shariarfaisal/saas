package catalog

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/modules/tenant"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/pagination"
	"github.com/munchies/platform/backend/internal/pkg/respond"
	"time"
)

// Handler handles catalog HTTP requests.
type Handler struct {
	svc *Service
}

// NewHandler creates a new catalog handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ListCategories handles GET /partner/restaurants/{id}/categories
func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	restaurantID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid restaurant id"))
		return
	}
	cats, err := h.svc.ListCategories(r.Context(), restaurantID, t.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, cats)
}

// CreateCategory handles POST /partner/restaurants/{id}/categories
func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	restaurantID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid restaurant id"))
		return
	}

	var req struct {
		Name                 string  `json:"name"`
		Description          *string `json:"description"`
		ImageUrl             *string `json:"image_url"`
		IconUrl              *string `json:"icon_url"`
		ExtraPrepTimeMinutes int32   `json:"extra_prep_time_minutes"`
		IsTobacco            bool    `json:"is_tobacco"`
		IsActive             bool    `json:"is_active"`
		SortOrder            int32   `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	if req.Name == "" {
		respond.Error(w, apperror.BadRequest("name is required"))
		return
	}

	cat, err := h.svc.CreateCategory(r.Context(), t.ID, CreateCategoryRequest{
		RestaurantID:         pgtype.UUID{Bytes: restaurantID, Valid: true},
		Name:                 req.Name,
		Description:          req.Description,
		ImageUrl:             req.ImageUrl,
		IconUrl:              req.IconUrl,
		ExtraPrepTimeMinutes: req.ExtraPrepTimeMinutes,
		IsTobacco:            req.IsTobacco,
		IsActive:             req.IsActive,
		SortOrder:            req.SortOrder,
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusCreated, cat)
}

// UpdateCategory handles PUT /partner/restaurants/{id}/categories/{cat_id}
func (h *Handler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	catID, err := uuid.Parse(chi.URLParam(r, "cat_id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid category id"))
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		ImageUrl    *string `json:"image_url"`
		SortOrder   *int32  `json:"sort_order"`
		IsActive    *bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	cat, err := h.svc.UpdateCategory(r.Context(), catID, t.ID, UpdateCategoryRequest{
		Name:        req.Name,
		Description: req.Description,
		ImageUrl:    req.ImageUrl,
		SortOrder:   req.SortOrder,
		IsActive:    req.IsActive,
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, cat)
}

// DeleteCategory handles DELETE /partner/restaurants/{id}/categories/{cat_id}
func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	catID, err := uuid.Parse(chi.URLParam(r, "cat_id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid category id"))
		return
	}
	if err := h.svc.DeleteCategory(r.Context(), catID, t.ID); err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ReorderCategories handles PATCH /partner/restaurants/{id}/categories/reorder
func (h *Handler) ReorderCategories(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	var req []ReorderCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	if err := h.svc.ReorderCategories(r.Context(), t.ID, req); err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListProducts handles GET /partner/restaurants/{id}/products
func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	restaurantID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid restaurant id"))
		return
	}
	page, perPage := parsePagination(r)
	items, meta, err := h.svc.ListProducts(r.Context(), restaurantID, t.ID, page, perPage)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, pagination.PagedResponse{Data: items, Meta: meta})
}

// CreateProduct handles POST /partner/restaurants/{id}/products
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	restaurantID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid restaurant id"))
		return
	}

	var req struct {
		CategoryID   *string          `json:"category_id"`
		Name         string           `json:"name"`
		Description  *string          `json:"description"`
		BasePrice    pgtype.Numeric   `json:"base_price"`
		VatRate      pgtype.Numeric   `json:"vat_rate"`
		Availability sqlc.ProductAvail `json:"availability"`
		Images       []string         `json:"images"`
		Tags         []string         `json:"tags"`
		IsFeatured   bool             `json:"is_featured"`
		IsInvTracked bool             `json:"is_inv_tracked"`
		SortOrder    int32            `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	if req.Name == "" {
		respond.Error(w, apperror.BadRequest("name is required"))
		return
	}

	catID := pgtype.UUID{}
	if req.CategoryID != nil {
		parsed, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			respond.Error(w, apperror.BadRequest("invalid category_id"))
			return
		}
		catID = pgtype.UUID{Bytes: parsed, Valid: true}
	}

	prod, err := h.svc.CreateProduct(r.Context(), t.ID, CreateProductRequest{
		RestaurantID: restaurantID,
		CategoryID:   catID,
		Name:         req.Name,
		Description:  req.Description,
		BasePrice:    req.BasePrice,
		VatRate:      req.VatRate,
		Availability: req.Availability,
		Images:       req.Images,
		Tags:         req.Tags,
		IsFeatured:   req.IsFeatured,
		IsInvTracked: req.IsInvTracked,
		SortOrder:    req.SortOrder,
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusCreated, prod)
}

// GetProduct handles GET /partner/products/{id}
func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid product id"))
		return
	}
	prod, err := h.svc.GetProduct(r.Context(), id, t.ID)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, prod)
}

// UpdateProduct handles PUT /partner/products/{id}
func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid product id"))
		return
	}

	var req struct {
		CategoryID  *string        `json:"category_id"`
		Name        *string        `json:"name"`
		Description *string        `json:"description"`
		BasePrice   pgtype.Numeric `json:"base_price"`
		VatRate     pgtype.Numeric `json:"vat_rate"`
		Images      []string       `json:"images"`
		Tags        []string       `json:"tags"`
		IsFeatured  *bool          `json:"is_featured"`
		SortOrder   *int32         `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	catID := pgtype.UUID{}
	if req.CategoryID != nil {
		parsed, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			respond.Error(w, apperror.BadRequest("invalid category_id"))
			return
		}
		catID = pgtype.UUID{Bytes: parsed, Valid: true}
	}

	prod, err := h.svc.UpdateProduct(r.Context(), id, t.ID, UpdateProductRequest{
		Name:        req.Name,
		Description: req.Description,
		CategoryID:  catID,
		BasePrice:   req.BasePrice,
		VatRate:     req.VatRate,
		Images:      req.Images,
		Tags:        req.Tags,
		IsFeatured:  req.IsFeatured,
		SortOrder:   req.SortOrder,
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, prod)
}

// UpdateProductAvailability handles PATCH /partner/products/{id}/availability
func (h *Handler) UpdateProductAvailability(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid product id"))
		return
	}

	var req struct {
		Availability sqlc.ProductAvail `json:"availability"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	prod, err := h.svc.UpdateProductAvailability(r.Context(), id, t.ID, req.Availability)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, prod)
}

// DeleteProduct handles DELETE /partner/products/{id}
func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid product id"))
		return
	}
	if err := h.svc.DeleteProduct(r.Context(), id, t.ID); err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UpsertDiscount handles POST /partner/products/{id}/discount
func (h *Handler) UpsertDiscount(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	productID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid product id"))
		return
	}

	var req struct {
		RestaurantID   string             `json:"restaurant_id"`
		DiscountType   sqlc.DiscountType  `json:"discount_type"`
		Amount         pgtype.Numeric     `json:"amount"`
		MaxDiscountCap pgtype.Numeric     `json:"max_discount_cap"`
		StartsAt       time.Time          `json:"starts_at"`
		EndsAt         pgtype.Timestamptz `json:"ends_at"`
		IsActive       bool               `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	restaurantID, err := uuid.Parse(req.RestaurantID)
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid restaurant_id"))
		return
	}

	d, err := h.svc.UpsertProductDiscount(r.Context(), t.ID, UpsertDiscountRequest{
		ProductID:      productID,
		RestaurantID:   restaurantID,
		DiscountType:   req.DiscountType,
		Amount:         req.Amount,
		MaxDiscountCap: req.MaxDiscountCap,
		StartsAt:       req.StartsAt,
		EndsAt:         req.EndsAt,
		IsActive:       req.IsActive,
	})
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, d)
}

// DeactivateDiscount handles DELETE /partner/products/{id}/discount
func (h *Handler) DeactivateDiscount(w http.ResponseWriter, r *http.Request) {
	productID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid product id"))
		return
	}
	if err := h.svc.DeactivateDiscount(r.Context(), productID); err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// BulkUpload handles POST /partner/products/bulk-upload
// Accepts a multipart CSV with columns: category_name,name,description,base_price,availability
func (h *Handler) BulkUpload(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respond.Error(w, apperror.BadRequest("invalid multipart form"))
		return
	}

	restaurantIDStr := r.FormValue("restaurant_id")
	restaurantID, err := uuid.Parse(restaurantIDStr)
	if err != nil {
		respond.Error(w, apperror.BadRequest("restaurant_id is required"))
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		respond.Error(w, apperror.BadRequest("file is required"))
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid CSV format"))
		return
	}

	if len(records) < 2 {
		respond.Error(w, apperror.BadRequest("CSV must have header row and at least one data row"))
		return
	}

	type rowError struct {
		Row     int    `json:"row"`
		Message string `json:"message"`
	}

	catCache := make(map[string]uuid.UUID)
	created := 0
	var rowErrors []rowError

	for i, row := range records[1:] {
		rowNum := i + 2 // 1-indexed, account for header
		if len(row) < 5 {
			rowErrors = append(rowErrors, rowError{Row: rowNum, Message: "insufficient columns (need 5)"})
			continue
		}
		catName := strings.TrimSpace(row[0])
		name := strings.TrimSpace(row[1])
		desc := strings.TrimSpace(row[2])
		basePriceStr := strings.TrimSpace(row[3])
		availStr := strings.TrimSpace(row[4])

		if name == "" {
			rowErrors = append(rowErrors, rowError{Row: rowNum, Message: "name is empty"})
			continue
		}

		var catID pgtype.UUID
		if catName != "" {
			if id, ok := catCache[catName]; ok {
				catID = pgtype.UUID{Bytes: id, Valid: true}
			} else {
				cat, err := h.svc.CreateCategory(r.Context(), t.ID, CreateCategoryRequest{
					RestaurantID: pgtype.UUID{Bytes: restaurantID, Valid: true},
					Name:         catName,
					IsActive:     true,
				})
				if err == nil {
					catCache[catName] = cat.ID
					catID = pgtype.UUID{Bytes: cat.ID, Valid: true}
				}
			}
		}

		avail := sqlc.ProductAvailAvailable
		if availStr != "" {
			avail = sqlc.ProductAvail(availStr)
		}

		var descPtr *string
		if desc != "" {
			descPtr = &desc
		}

		// Parse base price: expected as a decimal string (e.g. "9.99")
		var basePrice pgtype.Numeric
		if basePriceStr != "" {
			if err := basePrice.Scan(basePriceStr); err != nil {
				rowErrors = append(rowErrors, rowError{Row: rowNum, Message: "invalid base_price: " + basePriceStr})
				continue
			}
		}

		_, err := h.svc.CreateProduct(r.Context(), t.ID, CreateProductRequest{
			RestaurantID: restaurantID,
			CategoryID:   catID,
			Name:         name,
			Description:  descPtr,
			BasePrice:    basePrice,
			Availability: avail,
			Images:       []string{},
			Tags:         []string{},
		})
		if err == nil {
			created++
		} else {
			rowErrors = append(rowErrors, rowError{Row: rowNum, Message: err.Error()})
		}
	}

	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"created": created,
		"errors":  rowErrors,
	})
}

// DuplicateMenu handles POST /partner/restaurants/{id}/menu/duplicate
func (h *Handler) DuplicateMenu(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		respond.Error(w, apperror.NotFound("tenant"))
		return
	}
	srcID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid restaurant id"))
		return
	}

	var req struct {
		TargetRestaurantID string `json:"target_restaurant_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	dstID, err := uuid.Parse(req.TargetRestaurantID)
	if err != nil {
		respond.Error(w, apperror.BadRequest("invalid target_restaurant_id"))
		return
	}

	if err := h.svc.DuplicateMenu(r.Context(), srcID, dstID, t.ID); err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, map[string]string{"status": "duplicated"})
}

func parsePagination(r *http.Request) (page, perPage int) {
	q := r.URL.Query()
	page, _ = strconv.Atoi(q.Get("page"))
	perPage, _ = strconv.Atoi(q.Get("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = pagination.DefaultPageSize
	}
	return page, perPage
}

func toAppError(err error) *apperror.AppError {
	if e, ok := err.(*apperror.AppError); ok {
		return e
	}
	return apperror.Internal("unexpected error", err)
}
