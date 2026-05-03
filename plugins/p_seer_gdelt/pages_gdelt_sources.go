package p_seer_gdelt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/UniquityVentures/lago/components"
	"github.com/UniquityVentures/lago/getters"
	"github.com/UniquityVentures/lago/lago"
)

func gdeltSourceLabelFromParts(query, domain, actionCountry string) string {
	q := strings.TrimSpace(query)
	d := strings.TrimSpace(domain)
	c := strings.TrimSpace(actionCountry)
	switch {
	case q != "" && d != "":
		return fmt.Sprintf("%s · %s", truncateRunes(q, 40), d)
	case q != "":
		return truncateRunes(q, 56)
	case d != "":
		return "domain:" + d
	case c != "":
		return "country:" + c
	default:
		return "GDELT source"
	}
}

func truncateRunes(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

func gdeltSourceSelectionDisplayFromIn(ctx context.Context) (string, error) {
	q, _ := getters.Key[string]("$in.Query")(ctx)
	d, _ := getters.Key[string]("$in.Domain")(ctx)
	c, _ := getters.Key[string]("$in.ActionCountry")(ctx)
	return gdeltSourceLabelFromParts(q, d, c), nil
}

func gdeltSourceSelectionDisplayFromRow(ctx context.Context) (string, error) {
	rowAny := ctx.Value("$row")
	m, ok := rowAny.(map[string]any)
	if !ok {
		return "—", nil
	}
	q, _ := m["Query"].(string)
	d, _ := m["Domain"].(string)
	c, _ := m["ActionCountry"].(string)
	return gdeltSourceLabelFromParts(q, d, c), nil
}

func gdeltSourceDetailWorkerLabel(ctx context.Context) (string, error) {
	src, err := getters.Key[GDELTSource]("gdeltSource")(ctx)
	if err != nil {
		return "", err
	}
	if src.GDELTWorkerID == nil || *src.GDELTWorkerID == 0 {
		return "—", nil
	}
	if src.GDELTWorker != nil {
		return src.GDELTWorker.Name, nil
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return "", err
	}
	var w GDELTWorker
	if err := db.WithContext(ctx).Where("id = ?", *src.GDELTWorkerID).Take(&w).Error; err != nil {
		return fmt.Sprintf("id %d", *src.GDELTWorkerID), nil
	}
	return w.Name, nil
}

func gdeltSourceNLFilterDisplay(ctx context.Context) (string, error) {
	src, err := getters.Key[GDELTSource]("gdeltSource")(ctx)
	if err != nil {
		return "", err
	}
	t := strings.TrimSpace(src.NaturalLanguageFilter)
	if t == "" {
		return "—", nil
	}
	if src.IsBlacklist {
		return fmt.Sprintf("%s (blacklist)", t), nil
	}
	return fmt.Sprintf("%s (whitelist)", t), nil
}

func gdeltSourceStartDateDisplay(ctx context.Context) (string, error) {
	return gdeltSourceDatePtrFmt(ctx, "$in.StartDate")
}

func gdeltSourceEndDateDisplay(ctx context.Context) (string, error) {
	return gdeltSourceDatePtrFmt(ctx, "$in.EndDate")
}

func gdeltSourceDatePtrFmt(ctx context.Context, key string) (string, error) {
	t, err := getters.Key[*time.Time](key)(ctx)
	if err != nil || t == nil || t.IsZero() {
		return "—", nil
	}
	return t.UTC().Format(time.DateOnly), nil
}

func registerGDELTSourcesPages() {
	lago.RegistryPage.Register("seer_gdelt.GDELTSourceTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_gdelt.Menu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[GDELTSource]{
				Page:    components.Page{Key: "seer_gdelt.GDELTSourceTableBody"},
				UID:     "seer-gdelt-sources-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[GDELTSource]]("gdeltSources"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("seer_gdelt.GDELTSourceCreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("seer_gdelt.GDELTSourceDetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "Search",
						Children: []components.PageInterface{
							&components.FieldText{Getter: gdeltSourceSelectionDisplayFromRow},
						},
					},
					{
						Label: "Domain",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Domain")},
						},
					},
					{
						Label: "Action country",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.ActionCountry")},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_gdelt.GDELTSourceUnsetSelectionTable", &components.Modal{
		UID: "gdelt-source-unset-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[GDELTSource]{
				Page:  components.Page{Key: "seer_gdelt.GDELTSourceUnsetSelectionTableBody"},
				UID:   "gdelt-source-unset-selection-table",
				Title: "Select GDELT sources without worker",
				Data:  getters.Key[components.ObjectList[GDELTSource]]("gdeltSources"),
				RowAttr: getters.RowAttrSelectMulti(
					getters.IfOrElse(
						getters.Key[string]("$get.target_input"),
						getters.Static("GDELTSourceIDs"),
					),
					getters.Key[uint]("$row.ID"),
					gdeltSourceSelectionDisplayFromRow,
				),
				Columns: []components.TableColumn{
					{
						Label: "Source",
						Children: []components.PageInterface{
							&components.FieldText{Getter: gdeltSourceSelectionDisplayFromRow},
						},
					},
					{
						Label: "Domain",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Domain")},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_gdelt.GDELTSourceDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_gdelt.GDELTSourceDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[GDELTSource]{
				Getter: getters.Key[GDELTSource]("gdeltSource"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "seer_gdelt.GDELTSourceDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{
								Getter: getters.Format("Source #%d", getters.Any(getters.Key[uint]("$in.ID"))),
							},
							&components.LabelInline{
								Title: "Worker",
								Children: []components.PageInterface{
									&components.FieldText{Getter: gdeltSourceDetailWorkerLabel},
								},
							},
							&components.LabelInline{
								Title: "Keywords",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Query")},
								},
							},
							&components.LabelInline{
								Title: "Domain",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Domain")},
								},
							},
							&components.LabelInline{
								Title: "Action country",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.ActionCountry")},
								},
							},
							&components.LabelInline{
								Title: "Start date",
								Children: []components.PageInterface{
									&components.FieldText{Getter: gdeltSourceStartDateDisplay},
								},
							},
							&components.LabelInline{
								Title: "End date",
								Children: []components.PageInterface{
									&components.FieldText{Getter: gdeltSourceEndDateDisplay},
								},
							},
							&components.LabelInline{
								Title: "Min mentions",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$in.MinMentions")))},
								},
							},
							&components.LabelInline{
								Title: "Max records",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$in.MaxRecords")))},
								},
							},
							&components.LabelInline{
								Title: "Sort",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Sort")},
								},
							},
							&components.LabelInline{
								Title: "Natural language filter",
								Children: []components.PageInterface{
									&components.FieldText{Getter: gdeltSourceNLFilterDisplay},
								},
							},
						},
					},
				},
			},
		},
	})
}
