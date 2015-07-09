package utility

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/irlTopper/lifevault/app/modules/logger/app"
)

const APIVersion = "0.1"

var TempFolder string // Our own temp folder in os temp folder eg /tmp/lifevault

// Create a unique folder "lifevault" within the os temp folder
// We destroy this folder at startup
func SetupTempFolder() {
	tempDir := os.TempDir()
	TempFolder = filepath.FromSlash(tempDir + "/lifevault/")

	// Delete the old folder from last time
	err := os.RemoveAll(TempFolder)
	if err != nil {
		logger.Log.Panic(fmt.Sprintf("Error removing old Temp folder at %v: %v", TempFolder, err))
	}

	// Create it again
	if err = os.MkdirAll(TempFolder, 0777); err != nil {
		panic(fmt.Sprintf("Error creating temp folder %v", TempFolder))
	}
}
