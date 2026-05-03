package p_seer_gdelt

import (
	"context"
	"time"

	"github.com/UniquityVentures/lago/components"
	"github.com/UniquityVentures/lago/getters"
	"github.com/UniquityVentures/lago/lago"
	"github.com/UniquityVentures/seer/plugins/p_seer_workerregistry"
	"gorm.io/gorm"
)

func gdeltWorkerFormFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "seer_gdelt.GDELTWorkerFormFields"},
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.Key[error]("$error.Name"),
				Children: []components.PageInterface{
					&components.InputText{
						Label:    "Name",
						Name:     "Name",
						Required: true,
						Getter:   getters.Key[string]("$in.Name"),
						Classes:  "w-full max-w-xl",
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Duration"),
				Children: []components.PageInterface{
					&components.InputDuration{
						Label:    "Duration",
						Name:     "Duration",
						Required: true,
						Getter:   getters.Ref(getters.Key[time.Duration]("$in.Duration")),
						Classes:  "w-full max-w-xl",
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.GDELTSourceIDs"),
				Children: []components.PageInterface{
					&components.InputManyToMany[GDELTSource]{
						Label:       "GDELT sources without worker",
						Name:        "GDELTSourceIDs",
						Getter:      gdeltSourcesForCurrentWorker,
						Url:         lago.RoutePath("seer_gdelt.GDELTSourceUnsetSelectRoute", nil),
						Display:     gdeltSourceSelectionDisplayFromIn,
						Placeholder: "Select unassigned sources…",
						Classes:     "w-full max-w-xl",
					},
				},
			},
		},
	}
}

func gdeltSourcesForCurrentWorker(ctx context.Context) ([]GDELTSource, error) {
	id, err := getters.Key[uint]("$in.ID")(ctx)
	if err != nil || id == 0 {
		return nil, err
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return gorm.G[GDELTSource](db).Where("gdelt_worker_id = ?", id).Order("id DESC").Find(ctx)
}

func gdeltWorkerDetailWorkerPoolActionsGetter() getters.Getter[components.PageInterface] {
	return func(ctx context.Context) (components.PageInterface, error) {
		id, err := getters.Key[uint]("$in.ID")(ctx)
		if err != nil {
			return nil, err
		}
		if GDELTWorkerPoolIsRunning(id) {
			return &components.ContainerRow{
				Page:    components.Page{Key: "seer_gdelt.GDELTWorkerDetailWorkerPoolActions"},
				Classes: "flex flex-wrap gap-2 items-center mt-2",
				Children: []components.PageInterface{
					&components.ButtonPost{
						Label: "Stop worker pool",
						URL: lago.RoutePath("seer_gdelt.GDELTWorkerPoolStopRoute", map[string]getters.Getter[any]{
							"id": getters.Any(getters.Key[uint]("$in.ID")),
						}),
						Icon:    "stop",
						Classes: "btn-outline btn-error btn-sm",
					},
				},
			}, nil
		}
		return &components.ContainerRow{
			Page:    components.Page{Key: "seer_gdelt.GDELTWorkerDetailWorkerPoolActions"},
			Classes: "flex flex-wrap gap-2 items-center mt-2",
			Children: []components.PageInterface{
				&components.ButtonPost{
					Label: "Start worker pool",
					URL: lago.RoutePath("seer_gdelt.GDELTWorkerPoolStartRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$in.ID")),
					}),
					Icon:    "play",
					Classes: "btn-outline btn-success btn-sm",
				},
			},
		}, nil
	}
}

func registerGDELTWorkerPages() {
	createName := getters.Static("seer_gdelt.GDELTWorkerCreateForm")
	updateName := getters.Static("seer_gdelt.GDELTWorkerUpdateForm")
	deleteName := getters.Static("seer_gdelt.GDELTWorkerDeleteForm")

	lago.RegistryPage.Register("seer_gdelt.GDELTWorkerTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_gdelt.Menu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[GDELTWorker]{
				Page:    components.Page{Key: "seer_gdelt.GDELTWorkerTableBody"},
				UID:     "seer-gdelt-workers-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[GDELTWorker]]("gdeltWorkers"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("seer_gdelt.GDELTWorkerCreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("seer_gdelt.GDELTWorkerDetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Name")},
						},
					},
					{
						Label: "Duration",
						Children: []components.PageInterface{
							&components.FieldDuration{
								Getter: getters.Ref(getters.Key[time.Duration]("$row.Duration")),
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_gdelt.GDELTWorkerSelectionTable", &components.Modal{
		UID: "gdelt-worker-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[GDELTWorker]{
				Page:    components.Page{Key: "seer_gdelt.GDELTWorkerSelectionTableBody"},
				UID:     "gdelt-worker-selection-table",
				Title:   "Select worker",
				Data:    getters.Key[components.ObjectList[GDELTWorker]]("gdeltWorkers"),
				RowAttr: getters.RowAttrSelect("GDELTWorkerID", getters.Key[uint]("$row.ID"), getters.Key[string]("$row.Name")),
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Name")},
						},
					},
					{
						Label: "Duration",
						Children: []components.PageInterface{
							&components.FieldDuration{
								Getter: getters.Ref(getters.Key[time.Duration]("$row.Duration")),
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_gdelt.GDELTWorkerDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_gdelt.GDELTWorkerDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[GDELTWorker]{
				Getter: getters.Key[GDELTWorker]("gdeltWorker"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "seer_gdelt.GDELTWorkerDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
							&components.LabelInline{
								Title: "Duration",
								Children: []components.PageInterface{
									&components.FieldDuration{
										Getter: getters.Ref(getters.Key[time.Duration]("gdeltWorker.Duration")),
									},
								},
							},
							&components.LabelNewline{
								Title: "Assigned sources",
								Children: []components.PageInterface{
									&components.FieldManyToMany[GDELTSource]{
										Getter:  gdeltSourcesForCurrentWorker,
										Display: gdeltSourceSelectionDisplayFromIn,
										Link: lago.RoutePath("seer_gdelt.GDELTSourceDetailRoute", map[string]getters.Getter[any]{
											"id": getters.Any(getters.Key[uint]("$in.ID")),
										}),
										Classes: "w-full max-w-xl",
									},
								},
							},
							&components.GetterPage{Getter: gdeltWorkerDetailWorkerPoolActionsGetter()},
							p_seer_workerregistry.WorkerRunLogsBlock("seer-gdelt-worker-run-logs"),
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_gdelt.GDELTWorkerCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_gdelt.Menu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      createName,
				ActionURL: lago.RoutePath("seer_gdelt.GDELTWorkerCreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[GDELTWorker]{
						Getter:   getters.Static(GDELTWorker{Name: "", Duration: time.Hour}),
						Attr:     getters.FormBubbling(createName),
						Title:    "Create worker",
						Subtitle: "Workers run on a schedule and pull BigQuery for each attached GDELT source.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							gdeltWorkerFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Save worker"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_gdelt.GDELTWorkerUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_gdelt.GDELTWorkerDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: updateName,
				ActionURL: lago.RoutePath("seer_gdelt.GDELTWorkerUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("gdeltWorker.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[GDELTWorker]{
						Getter:   getters.Key[GDELTWorker]("gdeltWorker"),
						Attr:     getters.FormBubbling(updateName),
						Title:    "Edit worker",
						Subtitle: "Go duration syntax: 30s, 5m, 1h30m.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							gdeltWorkerFormFields(),
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
												Name:        deleteName,
												Url:         lago.RoutePath("seer_gdelt.GDELTWorkerDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("gdeltWorker.ID"))}),
												FormPostURL: lago.RoutePath("seer_gdelt.GDELTWorkerDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("gdeltWorker.ID"))}),
												ModalUID:    "seer-gdelt-worker-delete-modal",
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

	lago.RegistryPage.Register("seer_gdelt.GDELTWorkerDeleteForm", &components.Modal{
		UID: "seer-gdelt-worker-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Delete worker?",
				Message: "Sources attached to this worker will be unassigned (not deleted).",
				Attr:    getters.FormBubbling(deleteName),
			},
		},
	})
}
