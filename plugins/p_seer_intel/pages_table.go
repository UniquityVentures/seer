package p_seer_intel

import (
	"context"
	"strings"
	"time"

	"github.com/UniquityVentures/lamu/components"
	"github.com/UniquityVentures/lamu/getters"
	"github.com/UniquityVentures/lamu/lamu"
)

func intelDateGetter(field string) getters.Getter[time.Time] {
	return func(ctx context.Context) (time.Time, error) {
		raw, err := getters.Key[string]("$get." + field)(ctx)
		if err != nil {
			return time.Time{}, nil
		}
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return time.Time{}, nil
		}
		t, err := time.Parse("2006-01-02", raw)
		if err != nil {
			return time.Time{}, nil
		}
		return t, nil
	}
}

func registerTablePages() {
	registerPluginPage("seer_intel.IntelFilter", &components.FormComponent[map[string]any]{
		Page:     components.Page{Key: "seer_intel.FilterForm"},
		Attr:     getters.FormBoostedGet(lamu.RoutePath("seer_intel.DefaultRoute", nil)),
		Title:    "Filter Intel",
		Classes:  "@container rounded-box border border-base-300 bg-base-100 p-2",
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Page:    components.Page{Key: "seer_intel.FilterForm.Title"},
				Label:   "Title",
				Name:    "Title",
				Getter:  getters.Key[string]("$get.Title"),
				Classes: "w-full",
			},
			&components.InputText{
				Page:    components.Page{Key: "seer_intel.FilterForm.Summary"},
				Label:   "Description",
				Name:    "Summary",
				Getter:  getters.Key[string]("$get.Summary"),
				Classes: "w-full",
			},
			&components.InputDate{
				Page:    components.Page{Key: "seer_intel.FilterForm.CreateDateStart"},
				Label:   "Create Date Start",
				Name:    "create_date_start",
				Getter:  intelDateGetter("create_date_start"),
				Classes: "w-full",
			},
			&components.InputDate{
				Page:    components.Page{Key: "seer_intel.FilterForm.CreateDateEnd"},
				Label:   "Create Date End",
				Name:    "create_date_end",
				Getter:  intelDateGetter("create_date_end"),
				Classes: "w-full",
			},
			&components.InputText{
				Page:    components.Page{Key: "seer_intel.FilterForm.EmbeddingSearch"},
				Label:   "Embedding Search (AI Similarity)",
				Name:    "embedding_search",
				Getter:  getters.Key[string]("$get.embedding_search"),
				Classes: "w-full",
			},
		},
		ChildrenAction: []components.PageInterface{
			&components.ContainerRow{
				Page:    components.Page{Key: "seer_intel.FilterForm.Actions"},
				Classes: "flex gap-2",
				Children: []components.PageInterface{
					&components.ButtonSubmit{Label: "Apply Filters"},
					&components.ButtonClear{Label: "Clear"},
				},
			},
		},
	})

	registerTablePagesList()
}

func registerTablePagesList() {
	registerPluginPage("seer_intel.IntelTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lamu.DynamicPage{Name: "seer_intel.IntelMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Intel]{
				Page:     components.Page{Key: "seer_intel.IntelTableBody"},
				UID:      "seer-intel-table",
				Classes:  "w-full",
				Data:     getters.Key[components.ObjectList[Intel]]("intels"),
				Title:    "Intel",
				Subtitle: "Ingested intelligence streams",
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lamu.DynamicPage{Name: "seer_intel.IntelFilter"}},
				},
				RowAttr: getters.RowAttrNavigate(
					lamu.RoutePath("seer_intel.DetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "Title",
						Name:  "Title",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Title")},
						},
					},
					{
						Label: "Kind",
						Name:  "Kind",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Kind")},
						},
					},
					{
						Label: "Datetime",
						Name:  "Datetime",
						Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.Datetime")},
						},
					},
				},
			},
		},
	})
}
