package p_seer_gdelt

import (
	"time"

	"github.com/UniquityVentures/lago/components"
	"github.com/UniquityVentures/lago/getters"
	"github.com/UniquityVentures/lago/lago"
)

var gdeltSourceCreateFormDefaults = GDELTSource{
	Sort: gdeltSortDateDesc,
}

func gdeltSourceFormFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "seer_gdelt.GDELTSourceFormFields"},
		Children: []components.PageInterface{
			&components.ContainerError{
				Page:  components.Page{Key: "seer_gdelt.GDELTSourceForm.GDELTWorkerID"},
				Error: getters.Key[error]("$error.GDELTWorkerID"),
				Children: []components.PageInterface{
					&components.InputForeignKey[GDELTWorker]{
						Label:       "Worker",
						Name:        "GDELTWorkerID",
						Url:         lago.RoutePath("seer_gdelt.GDELTWorkerSelectRoute", nil),
						Display:     getters.Key[string]("$in.Name"),
						Placeholder: "Optional worker…",
						Required:    false,
						Getter:      getters.Association[GDELTWorker](getters.Deref(getters.Key[*uint]("$in.GDELTWorkerID"))),
						Classes:     "w-full max-w-xl",
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Query"),
				Children: []components.PageInterface{
					&components.InputText{
						Label:   "Keywords or phrase",
						Name:    "Query",
						Getter:  getters.Key[string]("$in.Query"),
						Classes: "w-full max-w-xl",
					},
				},
			},
			&components.ContainerRow{
				Page:    components.Page{Key: "seer_gdelt.GDELTSourceForm.RowDomainCountry"},
				Classes: "flex-col @lg:flex-row gap-3",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.Domain"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:   "Domain",
								Name:    "Domain",
								Getter:  getters.Key[string]("$in.Domain"),
								Classes: "w-full",
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.ActionCountry"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:   "Action country code (FIPS)",
								Name:    "ActionCountry",
								Getter:  getters.Key[string]("$in.ActionCountry"),
								Classes: "w-full",
							},
						},
					},
				},
			},
			&components.ContainerRow{
				Page:    components.Page{Key: "seer_gdelt.GDELTSourceForm.RowDatesMentions"},
				Classes: "flex-col @lg:flex-row gap-3",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.StartDate"),
						Children: []components.PageInterface{
							&components.InputDate{
								Label:   "Start date",
								Name:    "StartDate",
								Getter:  getters.Deref(getters.Key[*time.Time]("$in.StartDate")),
								Classes: "w-full",
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.EndDate"),
						Children: []components.PageInterface{
							&components.InputDate{
								Label:   "End date",
								Name:    "EndDate",
								Getter:  getters.Deref(getters.Key[*time.Time]("$in.EndDate")),
								Classes: "w-full",
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.MinMentions"),
						Children: []components.PageInterface{
							&components.InputNumber[uint]{
								Label:   "Minimum mentions",
								Name:    "MinMentions",
								Getter:  getters.Key[uint]("$in.MinMentions"),
								Classes: "w-full",
							},
						},
					},
				},
			},
			&components.ContainerRow{
				Page:    components.Page{Key: "seer_gdelt.GDELTSourceForm.RowSortMax"},
				Classes: "flex-col @lg:flex-row gap-3",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.Sort"),
						Children: []components.PageInterface{
							&components.InputSelect[string]{
								Label:   "Sort",
								Name:    "Sort",
								Choices: getters.Static(gdeltSortChoices),
								Getter:  gdeltSourceSortPairGetter(),
								Classes: "w-full",
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.MaxRecords"),
						Children: []components.PageInterface{
							&components.InputNumber[uint]{
								Label:   "Max records (0 = plugin default)",
								Name:    "MaxRecords",
								Getter:  getters.Key[uint]("$in.MaxRecords"),
								Classes: "w-full",
							},
						},
					},
				},
			},
			&components.ClientData{
				Page: components.Page{Key: "seer_gdelt.GDELTSourceForm.NLFilterBlock"},
				Data: "{ isBlacklist: false }",
				Init: "isBlacklist = $el.querySelector('[name=IsBlacklist]')?.checked ?? false",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.IsBlacklist"),
						Children: []components.PageInterface{
							&components.InputCheckbox{
								Page:    components.Page{Key: "seer_gdelt.GDELTSourceForm.IsBlacklist"},
								Label:   "Treat natural language filter as blacklist (off = whitelist)",
								Name:    "IsBlacklist",
								Getter:  getters.Key[bool]("$in.IsBlacklist"),
								XModel:  "isBlacklist",
								Classes: "w-full max-w-xl",
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.NaturalLanguageFilter"),
						Children: []components.PageInterface{
							&components.ClientIf{
								Condition: "!isBlacklist",
								Children: []components.PageInterface{
									&components.InputTextarea{
										Label:   "Natural language filter (whitelist: keep events that match)",
										Name:    "NaturalLanguageFilter",
										Rows:    4,
										Getter:  getters.Key[string]("$in.NaturalLanguageFilter"),
										Classes: "w-full max-w-xl",
									},
								},
							},
							&components.ClientIf{
								Condition: "isBlacklist",
								Children: []components.PageInterface{
									&components.InputTextarea{
										Label:   "Natural language filter (blacklist: drop events that match)",
										Name:    "NaturalLanguageFilter",
										Rows:    4,
										Getter:  getters.Key[string]("$in.NaturalLanguageFilter"),
										Classes: "w-full max-w-xl",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func registerGDELTSourceCreatePages() {
	createFormName := getters.Static("seer_gdelt.GDELTSourceCreateForm")

	lago.RegistryPage.Register("seer_gdelt.GDELTSourceCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_gdelt.Menu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      createFormName,
				ActionURL: lago.RoutePath("seer_gdelt.GDELTSourceCreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[GDELTSource]{
						Getter:   getters.Static(gdeltSourceCreateFormDefaults),
						Attr:     getters.FormBubbling(createFormName),
						Title:    "Create GDELT source",
						Subtitle: "BigQuery search parameters plus optional natural-language rules. At least one of keywords, domain, country, or dates is required.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							gdeltSourceFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Save source"},
						},
					},
				},
			},
		},
	})
}

func registerGDELTSourceUpdatePages() {
	updateFormName := getters.Static("seer_gdelt.GDELTSourceUpdateForm")
	deleteFormName := getters.Static("seer_gdelt.GDELTSourceDeleteForm")

	lago.RegistryPage.Register("seer_gdelt.GDELTSourceUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_gdelt.GDELTSourceDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: updateFormName,
				ActionURL: lago.RoutePath("seer_gdelt.GDELTSourceUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("gdeltSource.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[GDELTSource]{
						Getter:   getters.Key[GDELTSource]("gdeltSource"),
						Attr:     getters.FormBubbling(updateFormName),
						Title:    "Edit GDELT source",
						Subtitle: "Same filters as interactive search; worker runs use this saved configuration.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							gdeltSourceFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
								Children: []components.PageInterface{
									&components.ContainerRow{
										Classes: "flex justify-end gap-2",
										Children: []components.PageInterface{
											&components.ButtonSubmit{Label: "Save changes"},
											&components.ButtonModalForm{
												Label:       "Delete",
												Icon:        "trash",
												Name:        deleteFormName,
												Url:         lago.RoutePath("seer_gdelt.GDELTSourceDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("gdeltSource.ID"))}),
												FormPostURL: lago.RoutePath("seer_gdelt.GDELTSourceDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("gdeltSource.ID"))}),
												ModalUID:    "seer-gdelt-source-delete-modal",
												Classes:     "btn-error",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_gdelt.GDELTSourceDeleteForm", &components.Modal{
		UID: "seer-gdelt-source-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Delete GDELT source?",
				Message: "Removes this source. Events already stored keep their rows; the link to this source is cleared.",
				Attr:    getters.FormBubbling(deleteFormName),
			},
		},
	})
}
