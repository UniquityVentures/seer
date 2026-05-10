package main

import (
	"log"
	"log/slog"
	"net/http"
	_ "net/http/pprof"

	"github.com/UniquityVentures/lago/lago"
	"github.com/UniquityVentures/seer/plugins/p_seer_workerregistry"

	_ "github.com/UniquityVentures/lago/plugins/p_dashboard"
	_ "github.com/UniquityVentures/lago/plugins/p_filesystem"
	_ "github.com/UniquityVentures/lago/plugins/p_google_genai"
	_ "github.com/UniquityVentures/lago/plugins/p_pwa"
	_ "github.com/UniquityVentures/lago/plugins/p_users"
	_ "github.com/UniquityVentures/seer/plugins/p_seer_aisstream"
	_ "github.com/UniquityVentures/seer/plugins/p_seer_assistant"
	_ "github.com/UniquityVentures/seer/plugins/p_seer_dashboard"
	_ "github.com/UniquityVentures/seer/plugins/p_seer_deepsearch"
	_ "github.com/UniquityVentures/seer/plugins/p_seer_gdelt"
	_ "github.com/UniquityVentures/seer/plugins/p_seer_intel"
	_ "github.com/UniquityVentures/seer/plugins/p_seer_opensky"
	_ "github.com/UniquityVentures/seer/plugins/p_seer_reddit"
	_ "github.com/UniquityVentures/seer/plugins/p_seer_runners"
	_ "github.com/UniquityVentures/seer/plugins/p_seer_websites"
)

func main() {
	config, err := lago.LoadConfigFromFile("seer.toml")
	if err != nil {
		panic(err)
	}

	go func() {
		log.Fatal(http.ListenAndServe(":7777", nil))
	}()
	p_seer_workerregistry.BuildAllRegistries()
	if err := lago.Start(config); err != nil {
		slog.Error(err.Error())
	}
}
