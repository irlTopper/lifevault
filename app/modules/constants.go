package modules

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/revel/revel"
)

const APIVersion = "0.1"

var DeskTempFolder string // Our own temp folder in os temp folder eg /tmp/teamworkdesk

// Create a unique folder "teamworkdesk" within the os temp folder
// We destroy this folder at startup
func SetupTempFolder() {
	tempDir := os.TempDir()
	DeskTempFolder = filepath.FromSlash(tempDir + "/teamworkdesk/")

	// Delete the old folder from last time
	err := os.RemoveAll(DeskTempFolder)
	if err != nil {
		revel.ERROR.Print(fmt.Sprintf("Error removing old Temp folder at %v: %v", DeskTempFolder, err))
	}

	// Create it again
	if err = os.MkdirAll(DeskTempFolder, 0777); err != nil {
		panic(fmt.Sprintf("Error creating temp folder %v", DeskTempFolder))
	}
	revel.INFO.Println("Temp folder made at", DeskTempFolder)
}
