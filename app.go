package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// App struct
type App struct {
	ctx      context.Context
	configMu sync.Mutex
}

// Device represents an input device
type Device struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	Type string `json:"type"`
}

// Button represents a device button configuration
type Button struct {
	Number         int    `json:"number"`
	CurrentAction  string `json:"currentAction"`
	IsConfigurable bool   `json:"isConfigurable"`
}

// WheelAction represents one direction of a touch ring (or scroll strip)
type WheelAction struct {
	Property      string `json:"property"`
	Label         string `json:"label"`
	CurrentAction string `json:"currentAction"`
}

// DeviceMappings holds the user-entered mappings persisted for one device
type DeviceMappings struct {
	Buttons map[int]string    `json:"buttons"`
	Wheels  map[string]string `json:"wheels"`
}

// wheelProperties lists the xsetwacom properties that make up touch rings
// and touch strips. Availability is probed per device.
var wheelProperties = []struct {
	Property string
	Label    string
}{
	{"AbsWheelUp", "Touch Ring (counter-clockwise)"},
	{"AbsWheelDown", "Touch Ring (clockwise)"},
	{"AbsWheel2Up", "Touch Ring 2 (counter-clockwise)"},
	{"AbsWheel2Down", "Touch Ring 2 (clockwise)"},
	{"RelWheelUp", "Scroll Wheel (up)"},
	{"RelWheelDown", "Scroll Wheel (down)"},
	{"StripLeftUp", "Left Strip (up)"},
	{"StripLeftDown", "Left Strip (down)"},
	{"StripRightUp", "Right Strip (up)"},
	{"StripRightDown", "Right Strip (down)"},
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. Saved mappings are re-applied in the
// background because xsetwacom settings do not survive reboots or replugs.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	go a.applySavedMappings()
}

// --- Config persistence -----------------------------------------------------

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "tabmap", "config.json"), nil
}

func (a *App) loadConfig() map[string]*DeviceMappings {
	cfg := map[string]*DeviceMappings{}
	path, err := configPath()
	if err != nil {
		return cfg
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}
	_ = json.Unmarshal(data, &cfg)
	return cfg
}

func (a *App) saveConfig(cfg map[string]*DeviceMappings) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (a *App) deviceMappings(cfg map[string]*DeviceMappings, deviceName string) *DeviceMappings {
	m := cfg[deviceName]
	if m == nil {
		m = &DeviceMappings{}
		cfg[deviceName] = m
	}
	if m.Buttons == nil {
		m.Buttons = map[int]string{}
	}
	if m.Wheels == nil {
		m.Wheels = map[string]string{}
	}
	return m
}

// GetSavedMappings returns the persisted mappings for a device (empty if none)
func (a *App) GetSavedMappings(deviceName string) DeviceMappings {
	a.configMu.Lock()
	defer a.configMu.Unlock()
	cfg := a.loadConfig()
	return *a.deviceMappings(cfg, deviceName)
}

func (a *App) rememberButton(deviceName string, buttonNumber int, action string) {
	a.configMu.Lock()
	defer a.configMu.Unlock()
	cfg := a.loadConfig()
	m := a.deviceMappings(cfg, deviceName)
	if strings.TrimSpace(action) == "" {
		delete(m.Buttons, buttonNumber)
	} else {
		m.Buttons[buttonNumber] = action
	}
	_ = a.saveConfig(cfg)
}

func (a *App) rememberWheel(deviceName string, property string, action string) {
	a.configMu.Lock()
	defer a.configMu.Unlock()
	cfg := a.loadConfig()
	m := a.deviceMappings(cfg, deviceName)
	if strings.TrimSpace(action) == "" {
		delete(m.Wheels, property)
	} else {
		m.Wheels[property] = action
	}
	_ = a.saveConfig(cfg)
}

// applySavedMappings re-applies every persisted mapping. Devices that are not
// currently connected simply fail their xsetwacom calls and are skipped.
func (a *App) applySavedMappings() {
	a.configMu.Lock()
	cfg := a.loadConfig()
	a.configMu.Unlock()

	for deviceName, m := range cfg {
		if m == nil {
			continue
		}
		for num, action := range m.Buttons {
			_ = applyButton(deviceName, num, action)
		}
		for prop, action := range m.Wheels {
			_ = applyWheel(deviceName, prop, action)
		}
	}
}

// --- Device queries ----------------------------------------------------------

// GetDevices returns all detected input devices from xsetwacom
func (a *App) GetDevices() ([]Device, error) {
	cmd := exec.Command("xsetwacom", "--list", "devices")
	output, err := cmd.Output()
	if err != nil {
		return []Device{}, fmt.Errorf("xsetwacom not found or no devices detected: %v", err)
	}

	devices := []Device{}
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse xsetwacom output format: "Device Name \t id: X \t type: TYPE"
		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		id := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(parts[1]), "id:"))
		deviceType := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(parts[2]), "type:"))

		devices = append(devices, Device{
			Name: name,
			ID:   id,
			Type: deviceType,
		})
	}

	return devices, nil
}

// GetDeviceButtons returns available buttons for a specific device
func (a *App) GetDeviceButtons(deviceName string) ([]Button, error) {
	buttons := []Button{}

	// Probe buttons 1-32 (extended range for tablets with many buttons).
	// xsetwacom exits non-zero / prints to stderr for unavailable buttons.
	for i := 1; i <= 32; i++ {
		output, err := exec.Command("xsetwacom", "--get", deviceName, "Button", strconv.Itoa(i)).Output()
		if err != nil {
			continue
		}

		currentAction := strings.TrimSpace(string(output))
		if currentAction == "" || strings.Contains(currentAction, "does not exist") {
			continue
		}

		buttons = append(buttons, Button{
			Number:         i,
			CurrentAction:  currentAction,
			IsConfigurable: true,
		})
	}

	return buttons, nil
}

// GetWheelActions returns the touch ring / strip actions available on a device.
// Properties the device does not expose are omitted.
func (a *App) GetWheelActions(deviceName string) ([]WheelAction, error) {
	actions := []WheelAction{}

	for _, wp := range wheelProperties {
		output, err := exec.Command("xsetwacom", "--get", deviceName, wp.Property).Output()
		if err != nil {
			continue
		}

		current := strings.TrimSpace(string(output))
		if current == "" || strings.Contains(current, "does not exist") {
			continue
		}

		actions = append(actions, WheelAction{
			Property:      wp.Property,
			Label:         wp.Label,
			CurrentAction: current,
		})
	}

	return actions, nil
}

// --- Setting mappings ----------------------------------------------------------

// formatAction converts a user-friendly action string into xsetwacom's
// action mapping syntax, passed as a single argument.
//
//	"ctrl+z"        -> "key ctrl z"
//	"space"         -> "key space"
//	"key +ctrl +z"  -> unchanged (already explicit)
//	"button 3"      -> "button 3"
//	"pan"           -> "pan"
func formatAction(action string) string {
	action = strings.TrimSpace(action)
	lower := strings.ToLower(action)

	// Already in explicit xsetwacom syntax
	for _, prefix := range []string{"key ", "button ", "pan", "modetoggle", "displaytoggle"} {
		if strings.HasPrefix(lower, prefix) || lower == strings.TrimSpace(prefix) {
			return action
		}
	}

	// Convert "ctrl+shift+z" style combos to "key ctrl shift z"
	keys := strings.FieldsFunc(action, func(r rune) bool {
		return r == '+' || r == ' '
	})
	return "key " + strings.Join(keys, " ")
}

func applyButton(deviceName string, buttonNumber int, action string) error {
	var mapping string
	if strings.TrimSpace(action) == "" {
		// Restore default: the button emits its own button number
		mapping = "button +" + strconv.Itoa(buttonNumber)
	} else {
		mapping = formatAction(action)
	}

	out, err := exec.Command("xsetwacom", "--set", deviceName, "Button", strconv.Itoa(buttonNumber), mapping).CombinedOutput()
	if err != nil || strings.Contains(string(out), "Unable to parse") {
		return fmt.Errorf("failed to set button %d to '%s': %s", buttonNumber, mapping, strings.TrimSpace(string(out)))
	}
	return nil
}

func applyWheel(deviceName string, property string, action string) error {
	var mapping string
	if strings.TrimSpace(action) == "" {
		// Defaults per the wacom driver: rings/strips scroll (buttons 4/5)
		if strings.HasSuffix(property, "Up") {
			mapping = "button +4"
		} else {
			mapping = "button +5"
		}
	} else {
		mapping = formatAction(action)
	}

	out, err := exec.Command("xsetwacom", "--set", deviceName, property, mapping).CombinedOutput()
	if err != nil || strings.Contains(string(out), "Unable to parse") {
		return fmt.Errorf("failed to set %s to '%s': %s", property, mapping, strings.TrimSpace(string(out)))
	}
	return nil
}

// SetButtonAction sets the action for a specific button and persists it.
// An empty action restores the button to its default (plain button press).
func (a *App) SetButtonAction(deviceName string, buttonNumber int, action string) error {
	if err := applyButton(deviceName, buttonNumber, action); err != nil {
		return err
	}
	a.rememberButton(deviceName, buttonNumber, action)
	return nil
}

// SetWheelAction sets the action for a touch ring / strip direction and persists it.
// An empty action restores the default scroll behaviour for that property.
func (a *App) SetWheelAction(deviceName string, property string, action string) error {
	valid := false
	for _, wp := range wheelProperties {
		if wp.Property == property {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("unknown wheel property: %s", property)
	}

	if err := applyWheel(deviceName, property, action); err != nil {
		return err
	}
	a.rememberWheel(deviceName, property, action)
	return nil
}
