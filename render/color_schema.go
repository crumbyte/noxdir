package render

import (
	"encoding/json"
	"fmt"
	"os"
)

type ChartColors struct {
	Border  string `json:"border"`
	Sector1 string `json:"sector1"`
	Sector2 string `json:"sector2"`
	Sector3 string `json:"sector3"`
	Sector4 string `json:"sector4"`
	Sector5 string `json:"sector5"`
	Sector6 string `json:"sector6"`
	Sector7 string `json:"sector7"`
	Sector8 string `json:"sector8"`
	Sector9 string `json:"sector9"`
}

type DrivesStatusBarColors struct {
	ModeBG     string `json:"modeBackground"`
	CapacityBG string `json:"capacityBackground"`
	FreeBG     string `json:"freeBackground"`
	UsedBG     string `json:"usedBackground"`
}

type DirsStatusBarColors struct {
	PathBG  string `json:"pathBackground"`
	ModeBG  string `json:"modeBackground"`
	SizeBG  string `json:"sizeBackground"`
	DirsBG  string `json:"dirsBackground"`
	FilesBG string `json:"filesBackground"`
	ErrorBG string `json:"errorBackground"`
}

type StatusBarColors struct {
	Text      string                `json:"text"`
	BlockText string                `json:"blockText"`
	BG        string                `json:"background"`
	VersionBG string                `json:"versionBackground"`
	Drives    DrivesStatusBarColors `json:"drives"`
	Dirs      DirsStatusBarColors   `json:"dirs"`
}

// The ColorSchema schema contains color values for most UI elements, such as
// text and border colors, element backgrounds, etc. The Style component uses
// the ColorSchema instance during rendering elements. Each color value must be
// represented as a hex ("#FFBF69") or ANSI ("240") string color code.
//
// DefaultColorSchema always used as a base schema and all customizations are
// applied over it.
type ColorSchema struct {
	StatusBar         StatusBarColors `json:"statusBar"`
	ChartColors       ChartColors     `json:"chart"`
	TableHeaderBorder string          `json:"tableHeaderBorder"`
	SelectedRowText   string          `json:"selectedRowText"`
	SelectedRowBG     string          `json:"selectedRowBackground"`
	TopFilesText      string          `json:"topFilesText"`
	HelpText          string          `json:"helpText"`
	BindingText       string          `json:"bindingText"`
	DialogBoxBorder   string          `json:"dialogBoxBorder"`
	ConfirmButtonText string          `json:"confirmButtonText"`
	ConfirmButtonBG   string          `json:"confirmButtonBackground"`
	ActiveButtonText  string          `json:"activeButtonText"`
	ActiveButtonBG    string          `json:"activeButtonBackground"`
	FilterText        string          `json:"filterText"`
	ScanProgressBar   PG              `json:"scanProgressBar"`
	UsageProgressBar  PG              `json:"usageProgressBar"`
}

// DecodeFileColorSchema reads the color schema from the file by the provided
// path and applies it to the *ColorSchema instance. An error will be returned
// if the path is invalid or the JSON color schema content cannot be decoded.
func DecodeFileColorSchema(path string, cs *ColorSchema) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open color schema file %s: %w", path, err)
	}

	defer func(f *os.File) {
		_ = f.Close()
	}(file)

	if err = json.NewDecoder(file).Decode(cs); err != nil {
		return fmt.Errorf("decode color schema file %s: %w", path, err)
	}

	return nil
}

func DefaultColorSchema() ColorSchema {
	return ColorSchema{
		ScanProgressBar: PG{
			ColorProfile: 0,
			StartColor:   "#833AB4",
			EndColor:     "#FCB045",
		},
		UsageProgressBar: PG{
			ColorProfile: 3,
			FullChar:     "🟥",
			EmptyChar:    "🟩",
			HidePercent:  true,
		},
		StatusBar: StatusBarColors{
			VersionBG: "#8338EC",
			Text:      "#C1C6B2",
			BG:        "#353533",
			BlockText: "#FFFDF5",
			Drives: DrivesStatusBarColors{
				ModeBG:     "#FF5F87",
				CapacityBG: "#FF5F87",
				FreeBG:     "#FF5F87",
				UsedBG:     "#FF5F87",
			},
			Dirs: DirsStatusBarColors{
				PathBG:  "#FF5F87",
				ModeBG:  "#FF8531",
				SizeBG:  "#FF5F87",
				DirsBG:  "#FF5F87",
				FilesBG: "#FF5F87",
				ErrorBG: "#FF303E",
			},
		},
		ChartColors: ChartColors{
			Border:  "240",
			Sector1: "#ffbe0b",
			Sector2: "#fb5607",
			Sector3: "#ff006e",
			Sector4: "#8338ec",
			Sector5: "#3a86ff",
			Sector6: "#00f5d4",
			Sector7: "#fef9ef",
			Sector8: "#ff85a1",
			Sector9: "#b5838d",
		},
		TableHeaderBorder: "240",
		SelectedRowText:   "#262626",
		SelectedRowBG:     "#EBBD34",
		TopFilesText:      "#EBBD34",
		HelpText:          "#696868",
		BindingText:       "#FFBF69",
		DialogBoxBorder:   "#FF303E",
		ConfirmButtonText: "#FFFDF5",
		ConfirmButtonBG:   "#353533",
		ActiveButtonText:  "#FFFDF5",
		ActiveButtonBG:    "#FF8531",
		FilterText:        "#EBBD34",
	}
}
