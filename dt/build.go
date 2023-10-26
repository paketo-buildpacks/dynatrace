/*
 * Copyright 2018-2023 the original author or authors.
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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/bindings"
)

type Build struct {
	Logger bard.Logger
}

func (b Build) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	b.Logger.Title(context.Buildpack)
	result := libcnb.NewBuildResult()

	pr := libpak.PlanEntryResolver{Plan: context.Plan}

	dc, err := libpak.NewDependencyCache(context)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create dependency cache\n%w", err)
	}
	dc.Logger = b.Logger

	s, _, err := bindings.ResolveOne(context.Platform.Bindings, IsDynatraceBinding)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to resolve binding Dynatrace\n%w", err)
	}

	v, err := b.AgentVersion(s, context.Buildpack.Info)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to determine agent version\n%w", err)
	}

	uri := fmt.Sprintf("%s/v1/deployment/installer/agent/unix/paas/latest?bitness=64&skipMetadata=true", BaseURI(s))

	for _, t := range []string{"apache", "dotnet", "go", "java", "nginx", "nodejs", "php", "python"} {
		if _, ok, err := pr.Resolve(fmt.Sprintf("dynatrace-%s", t)); err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to resolve dynatrace-%s plan entry\n%w", t, err)
		} else if ok {
			uri = fmt.Sprintf("%s&include=%s", uri, t)
		}
	}

	dep := libpak.BuildpackDependency{
		ID:      "dynatrace-oneagent",
		Name:    "Dynatrace OneAgent",
		Version: v,
		URI:     uri,
		SHA256:  "",
		Stacks:  []string{context.StackID},
		PURL:    fmt.Sprintf("pkg:generic/dynatrace-one-agent@%s?arch=amd64", v),
		CPEs:    []string{fmt.Sprintf("cpe:2.3:a:dynatrace:one-agent:%s:*:*:*:*:*:*:*", v)},
	}

	a, be := NewAgent(dep, dc, APIToken(s), context.Buildpack.Info)
	a.Logger = b.Logger
	result.Layers = append(result.Layers, a)
	result.BOM.Entries = append(result.BOM.Entries, be)

	h, be := libpak.NewHelperLayer(context.Buildpack, "properties")
	h.Logger = b.Logger
	result.Layers = append(result.Layers, h)
	result.BOM.Entries = append(result.BOM.Entries, be)

	return result, nil
}

func (Build) AgentVersion(binding libcnb.Binding, info libcnb.BuildpackInfo) (string, error) {
	uri := fmt.Sprintf("%s/v1/deployment/installer/agent/unix/paas/latest/metainfo", BaseURI(binding))

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return "", fmt.Errorf("unable to create new GET request for %s\n%w", uri, err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Api-Token %s", APIToken(binding)))
	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", info.ID, info.Version))

	client := http.Client{Transport: &http.Transport{Proxy: http.ProxyFromEnvironment}}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("unable to request %s\n%w", uri, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("could not download %s: %d", uri, resp.StatusCode)
	}

	raw := struct {
		LatestAgentVersion string `json:"latestAgentVersion"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return "", fmt.Errorf("unable to decode payload\n%w", err)
	}

	return raw.LatestAgentVersion, nil
}
