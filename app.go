package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// App struct
type App struct {
	ctx context.Context
	cancel context.CancelFunc
}

// NewApp creates a new App application struct
func NewApp() *App {
	ctx, cancel := context.WithCancel(context.Background())
	return &App{ctx, cancel}
}

func (a* App) shutdown(ctx context.Context) {
	a.cancel() // Signal goroutine to stop
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func defaultSettings() Settings {
	return Settings{
		GBR:               true,
		MinutesBeforeCloud: 1,
		LTECategory:       1,
		IoT:               false,
		AVRGaming:         false,
		Healthcare:        false,
		Industry40:        false,
		IoTDevices:        false,
		PublicSafety:      false,
		SmartCityHome:     false,
		SmartTransport:    false,
		Smartphone:        true,
		ConnTo5G:          false,
	}
}

// Updates (a lot of them)
//lint:ignore
func loadSettings(a *App) (Settings, error) {
	fmt.Println("loadSettings")
	configDir, _ := os.UserConfigDir()
	filePath := filepath.Join(configDir, ".cloud-based-network-slicing-prototype", "settings.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		makeInitialConfigs(a)
		return defaultSettings(), nil
	}
	var settings Settings
	er := json.Unmarshal(data, &settings)
	if er != nil {
		// Handle some serious consequenses
		a.showJS(true, "Error loading settings: ")
		return Settings{}, er
	}

	return settings, nil
}

func makeInitialConfigs(a *App) {
	// fmt.Println("Making initial configs")
    configDir, _ := os.UserConfigDir()
	dataDir := filepath.Join(configDir, ".cloud-based-network-slicing-prototype")
    os.MkdirAll(dataDir, os.ModePerm)

    filePath := filepath.Join(dataDir, "settings.json")
	jsonForm, _ := json.MarshalIndent(defaultSettings(), "", "  ")
	err := os.WriteFile(filePath, jsonForm, 0644)
	if err != nil {
	    // Update
		a.showJS(true, "Error creating initial settings")
	}
}

/*
func initialSettingsHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("initialSettingsHandler")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	settings, err := loadSettings()
	if err != nil {
		http.Error(w, "Operation failed", http.StatusInternalServerError)
		return
	}

	data, _ := json.MarshalIndent(settings, "", "  ")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
*/

func validateSingleTrue(settings Settings) error {
	trueCount := 0

	fields := map[string]bool{
		"IoT":            settings.IoT,
		"AVRGaming":      settings.AVRGaming,
		"Healthcare":     settings.Healthcare,
		"Industry40":     settings.Industry40,
		"IoTDevices":     settings.IoTDevices,
		"PublicSafety":   settings.PublicSafety,
		"SmartCityHome":  settings.SmartCityHome,
		"SmartTransport": settings.SmartTransport,
		"Smartphone":     settings.Smartphone,
	}

	for _, v := range fields {
		if v {
			trueCount++
		}
	}

	if trueCount != 1 {
		return errors.New("exactly one of the conditional fields must be true")
	}
	return nil
}

func (a *App) GetInitialSettings() string{
	// a.showJS(false)
	sett, _ := loadSettings(a)
	jsonData, _ := json.Marshal(&sett)
	fmt.Println(string(jsonData))
	return string(jsonData)
}

/*
func changeSettingsHandler(w http.ResponseWriter, r *http.Request) {

	// fmt.Println("changeSettingsHandler")
	var wg sync.WaitGroup
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// fmt.Println(r.Header)

	// Handle preflight
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		fmt.Println("--Method not allowed", r.Method)
		fmt.Print(r)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var newSettings Settings
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&newSettings); err != nil {
		fmt.Println("--Error decoding JSON--")
		http.Error(w, "Operation failed", http.StatusInternalServerError)
		return
	}

	if err := validateSingleTrue(newSettings); err != nil {
		fmt.Println("--Error validating single true fields--")
		http.Error(w, "Operation failed", http.StatusInternalServerError)
		return
	}

	data, err := json.MarshalIndent(newSettings, "", "  ")
	dir, _ := os.Executable()
	if err != nil || os.WriteFile(path.Join(dir + configPath), data, 0644) != nil {
		fmt.Println("--Error writing to file--")
		http.Error(w, "Operation failed", http.StatusInternalServerError)
		return
	}
	wg.Add(1)
	go manipulateContexts(newSettings)
	w.WriteHeader(http.StatusOK)
}
*/

func manipulateContexts(newSettings Settings, a *App) {
	var DummyContextManager []ContextPair
	for x := range ContextManager{
		(*(ContextManager[x].Cancel))()
		ctx, cancel := context.WithCancel(context.Background())
		DummyContextManager = append(DummyContextManager, ContextPair{&ctx, &cancel, ContextManager[x].NM})
		ContextManager[x].NM.SetMetricsAsSettings(newSettings)
		go ContextManager[x].NM.periodicReport(newSettings, ctx, a)
	}
	ContextManager = DummyContextManager
}

// Update (massive)
func (a *App) SubmitSettings(jsonString string) bool{
	var wg sync.WaitGroup
	var res Settings
	_ = json.Unmarshal([]byte(jsonString), &res)
	err := validateSingleTrue(res)
	if err != nil {
	    a.showJS(true, "Error changing settings")
	}
	configDir, _ := os.UserConfigDir()
	dataDir := filepath.Join(configDir, ".cloud-based-network-slicing-prototype")
    filePath := filepath.Join(dataDir, "settings.json")
	er := os.WriteFile(filePath, []byte(jsonString), 0644)
	if er != nil {
		a.showJS(true, "Error creating initial settings")
		return false
	}
	wg.Add(1)
	manipulateContexts(res, a)
	return true
}
