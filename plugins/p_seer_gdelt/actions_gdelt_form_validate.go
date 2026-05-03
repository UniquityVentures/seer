package p_seer_gdelt

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/UniquityVentures/lago/components"
	"github.com/UniquityVentures/lago/getters"
	"github.com/UniquityVentures/lago/views"
)

type gdeltSourceFormValidate struct{}

func (gdeltSourceFormValidate) Patch(_ views.View, _ *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	if formErrors == nil {
		formErrors = map[string]error{}
	}
	q := strings.TrimSpace(stringField(formData, "Query"))
	dom := strings.TrimSpace(stringField(formData, "Domain"))
	ac := strings.TrimSpace(stringField(formData, "ActionCountry"))
	start := timePtrField(formData, "StartDate")
	end := timePtrField(formData, "EndDate")
	if q == "" && dom == "" && ac == "" && (start == nil || start.IsZero()) && (end == nil || end.IsZero()) {
		formErrors["Query"] = errors.New("add keyword, domain, action country, or a date range")
	}
	if mr := uintField(formData, "MaxRecords"); mr > maxGDELTMaxRecords {
		formErrors["MaxRecords"] = errors.New("max records must be between 1 and 250 (or 0 for default)")
	}
	sort := strings.TrimSpace(stringField(formData, "Sort"))
	if sort != "" && gdeltSortOrDefault(sort) != sort {
		formErrors["Sort"] = errors.New("invalid sort option")
	}
	if start != nil && end != nil && !start.IsZero() && !end.IsZero() && start.After(*end) {
		formErrors["EndDate"] = errors.New("end date must be on or after start date")
	}
	return formData, formErrors
}

func stringField(formData map[string]any, key string) string {
	v, ok := formData[key]
	if !ok || v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func uintField(formData map[string]any, key string) uint {
	v, ok := formData[key]
	if !ok || v == nil {
		return 0
	}
	switch x := v.(type) {
	case uint:
		return x
	case uint64:
		return uint(x)
	case int:
		if x < 0 {
			return 0
		}
		return uint(x)
	case int64:
		if x < 0 {
			return 0
		}
		return uint(x)
	default:
		return 0
	}
}

func timePtrField(formData map[string]any, key string) *time.Time {
	v, ok := formData[key]
	if !ok || v == nil {
		return nil
	}
	switch x := v.(type) {
	case *time.Time:
		return x
	case time.Time:
		if x.IsZero() {
			return nil
		}
		return &x
	default:
		return nil
	}
}

type gdeltWorkerValidate struct{}

func (gdeltWorkerValidate) Patch(_ views.View, r *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	if formErrors == nil {
		formErrors = map[string]error{}
	}
	name, _ := formData["Name"].(string)
	if strings.TrimSpace(name) == "" {
		formErrors["Name"] = errors.New("name is required")
	}
	durRaw, ok := formData["Duration"]
	if !ok {
		formErrors["Duration"] = errors.New("duration is required")
		return formData, formErrors
	}
	d, ok := durRaw.(*time.Duration)
	if !ok {
		formErrors["Duration"] = errors.New("invalid duration")
		return formData, formErrors
	}
	if d == nil || *d <= 0 {
		formErrors["Duration"] = errors.New("duration must be positive")
	}
	formData, formErrors = gdeltWorkerSourceIDsValidateAndFlatten(formData, formErrors)
	formErrors = validateGDELTWorkerSourceIDs(r, formData, formErrors)
	return formData, formErrors
}

func gdeltWorkerSourceIDsValidateAndFlatten(formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	raw, ok := formData["GDELTSourceIDs"]
	if !ok {
		return formData, formErrors
	}
	assoc, ok := raw.(components.AssociationIDs)
	if !ok {
		formErrors["GDELTSourceIDs"] = errors.New("invalid GDELT sources")
		delete(formData, "GDELTSourceIDs")
		return formData, formErrors
	}
	formData["GDELTSourceIDs"] = assoc.IDs
	return formData, formErrors
}

func validateGDELTWorkerSourceIDs(r *http.Request, formData map[string]any, formErrors map[string]error) map[string]error {
	ids, _ := formData["GDELTSourceIDs"].([]uint)
	if len(ids) == 0 || formErrors["GDELTSourceIDs"] != nil {
		return formErrors
	}
	db, err := getters.DBFromContext(r.Context())
	if err != nil {
		formErrors["GDELTSourceIDs"] = err
		return formErrors
	}
	query := db.WithContext(r.Context()).Model(&GDELTSource{}).Where("id IN ?", ids)
	if wk, ok := r.Context().Value("gdeltWorker").(GDELTWorker); ok && wk.ID != 0 {
		query = query.Where("gdelt_worker_id IS NULL OR gdelt_worker_id = ?", wk.ID)
	} else {
		query = query.Where("gdelt_worker_id IS NULL")
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		formErrors["GDELTSourceIDs"] = err
		return formErrors
	}
	if count != int64(len(ids)) {
		formErrors["GDELTSourceIDs"] = errors.New("select only GDELT sources without a worker")
	}
	return formErrors
}
