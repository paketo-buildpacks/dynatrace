package dt

import (
	"strings"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak/bindings"
)

func IsDynatraceBinding(bind libcnb.Binding) bool {
	if bindings.OfType("Dynatrace")(bind) {
		return true
	}

	if strings.Contains(strings.ToLower(bind.Name), "dynatrace") {
		return true
	}

	return false
}
