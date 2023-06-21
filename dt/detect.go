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

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/bindings"
)

type Detect struct {
	Logger bard.Logger
}

func (d Detect) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	_, ok, err := bindings.ResolveOne(context.Platform.Bindings, IsDynatraceBinding)
	if err != nil {
		return libcnb.DetectResult{}, fmt.Errorf("unable to resolve binding Dynatrace\n%w", err)
	} else if !ok {
		d.Logger.Info("SKIPPED: No binding for 'Dynatrace' found (type or name)")
		return libcnb.DetectResult{Pass: false}, nil
	}

	return libcnb.DetectResult{
		Pass: true,
		Plans: []libcnb.BuildPlan{
			{
				Provides: []libcnb.BuildPlanProvide{
					{Name: "dynatrace-apache"},
				},
				Requires: []libcnb.BuildPlanRequire{
					{Name: "dynatrace-apache"},
					{Name: "httpd"},
				},
			},
			{
				Provides: []libcnb.BuildPlanProvide{
					{Name: "dynatrace-dotnet"},
				},
				Requires: []libcnb.BuildPlanRequire{
					{Name: "dynatrace-dotnet"},
					{Name: "dotnet-runtime"},
					{Name: "node"},
				},
			},
			{
				Provides: []libcnb.BuildPlanProvide{
					{Name: "dynatrace-dotnet"},
				},
				Requires: []libcnb.BuildPlanRequire{
					{Name: "dynatrace-dotnet"},
					{Name: "dotnet-core-aspnet-runtime"},
					{Name: "node"},
				},
			},
			{
				Provides: []libcnb.BuildPlanProvide{
					{Name: "dynatrace-go"},
				},
				Requires: []libcnb.BuildPlanRequire{
					{Name: "dynatrace-go"},
					{Name: "go"},
				},
			},
			{
				Provides: []libcnb.BuildPlanProvide{
					{Name: "dynatrace-java"},
				},
				Requires: []libcnb.BuildPlanRequire{
					{Name: "dynatrace-java"},
					{Name: "jvm-application"},
				},
			},
			{
				Provides: []libcnb.BuildPlanProvide{
					{Name: "dynatrace-nginx"},
				},
				Requires: []libcnb.BuildPlanRequire{
					{Name: "dynatrace-nginx"},
					{Name: "nginx"},
				},
			},
			{
				Provides: []libcnb.BuildPlanProvide{
					{Name: "dynatrace-nodejs"},
				},
				Requires: []libcnb.BuildPlanRequire{
					{Name: "dynatrace-nodejs"},
					{Name: "node"},
					{Name: "node_modules"},
				},
			},
			{
				Provides: []libcnb.BuildPlanProvide{
					{Name: "dynatrace-php"},
				},
				Requires: []libcnb.BuildPlanRequire{
					{Name: "dynatrace-php"},
					{Name: "php"},
				},
			},
		},
	}, nil
}
