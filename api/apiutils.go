package api

import (
	"aiolimas/logging"
	db_types "aiolimas/types"
	"net/http"
)
func writeSQLRowResults[T db_types.TableRepresentation](w http.ResponseWriter, results []T) {
	for _, row := range results {
		j, err := row.ToJson()
		if err != nil {
			logging.ELog(err)
			continue
		}
		w.Write(j)
		w.Write([]byte("\n"))
	}
}
