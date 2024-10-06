package backend

import (
	"bauklotze/pkg/api/types"
	"bauklotze/pkg/api/utils"
	"fmt"
	"github.com/containers/storage/pkg/regexp"
	"github.com/gorilla/schema"
	"net/http"
)

var NameRegex = regexp.Delayed("^[a-zA-Z0-9][a-zA-Z0-9_.-]*$")

func GetDiskUsage(w http.ResponseWriter, r *http.Request) {
	decoder := r.Context().Value(types.DecoderKey).(*schema.Decoder)
	name := utils.GetName(r)
	query := struct {
		Units *string `schema:"units"`
	}{}

	if err := decoder.Decode(&query, r.URL.Query()); err != nil {
		utils.Error(w, http.StatusBadRequest, fmt.Errorf("failed to parse parameters for %s: %w", r.URL.String(), err))
		return
	}

	if name == "" || !NameRegex.MatchString(name) {
		utils.Error(w, http.StatusBadRequest, fmt.Errorf("container name is required or malformed"))
		return
	}

	// SSH into Machine and run df -h / to get disk usage

}
