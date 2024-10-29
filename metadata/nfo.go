package metadata

import (
	"fmt"
	"os"
	"path"

	"aiolimas/types"
)

//returns the path to nfo file, or "" if no path exists
func NFOExists(entry *db_types.InfoEntry) string {
	location := entry.Location

	stat, err := os.Stat(entry.Location)
	if  err != nil {
		return ""
	}

	expectedNFO := fmt.Sprintf("%s.nfo", stat.Name())

	if stat.IsDir() {
		entries, err := os.ReadDir(location)
		if err != nil{
			return ""
		}
		for _, entry := range entries {
			if entry.Name() == expectedNFO {
				return path.Join(location, entry.Name())
			}
		}
		return ""
	}

	fullExpectedNFO := path.Join(location, expectedNFO)
	_, err = os.Stat(fullExpectedNFO)
	if err != nil{
		return ""
	}
	return fullExpectedNFO
}
