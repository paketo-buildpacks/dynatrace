/*
 * Copyright 2018-2022 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package dt

import (
	"fmt"
	"net/http"
	"os"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/crush"
)

type Agent struct {
	BuildpackID      string
	BuildpackVersion string
	LayerContributor libpak.DependencyLayerContributor
	Logger           bard.Logger
}

func NewAgent(
	dependency libpak.BuildpackDependency,
	cache libpak.DependencyCache,
	apiToken string,
	info libcnb.BuildpackInfo,
) (Agent, libcnb.BOMEntry) {

	contributor, entry := libpak.NewDependencyLayer(dependency, cache, libcnb.LayerTypes{
		Launch: true,
	})
	contributor.RequestModifierFuncs = append(contributor.RequestModifierFuncs,
		func(request *http.Request) (*http.Request, error) {
			request.Header.Set("Authorization", fmt.Sprintf("Api-Token %s", apiToken))
			return request, nil
		},
	)

	return Agent{
		BuildpackID:      info.ID,
		BuildpackVersion: info.Version,
		LayerContributor: contributor,
	}, entry
}

func (a Agent) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	a.LayerContributor.Logger = a.Logger

	return a.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		a.Logger.Bodyf("Expanding to %s", layer.Path)

		if err := crush.ExtractZip(artifact, layer.Path, 0); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to expand Dynatrace OneAgent\n%w", err)
		}

		layer.LaunchEnvironment.Default("BPI_DYNATRACE_BUILDPACK_ID", a.BuildpackID)
		layer.LaunchEnvironment.Default("BPI_DYNATRACE_BUILDPACK_VERSION", a.BuildpackVersion)
		layer.LaunchEnvironment.Default("DT_LOGSTREAM", "stdout")
		layer.LaunchEnvironment.Appendf("DT_CUSTOM_PROP", " ", "CloudNativeBuildpackVersion=%s", a.BuildpackVersion)
		layer.LaunchEnvironment.Prependf("LD_PRELOAD", string(os.PathListSeparator), "%s/agent/lib64/liboneagentproc.so", layer.Path)

		return layer, nil
	})
}

func (a Agent) Name() string {
	return a.LayerContributor.LayerName()
}
