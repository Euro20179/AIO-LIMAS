package api

import (
	db_types "aiolimas/types"
	"net/http"
)
func writeSQLRowResults[T db_types.TableRepresentation](w http.ResponseWriter, results []T) {
	for _, row := range results {
		j, err := row.ToJson()
		if err != nil {
			println(err.Error())
			continue
		}
		w.Write(j)
		w.Write([]byte("\n"))
	}
}
