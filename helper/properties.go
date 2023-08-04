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

package helper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/bindings"

	"github.com/paketo-buildpacks/dynatrace/v4/dt"
)

type Properties struct {
	Bindings libcnb.Bindings
	Logger   bard.Logger
}

func (p Properties) Execute() (map[string]string, error) {
	b, ok, err := bindings.ResolveOne(p.Bindings, dt.IsDynatraceBinding)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve binding Dynatrace\n%w", err)
	} else if !ok {
		return nil, nil
	}

	p.Logger.Info("Configuring Dynatrace properties")

	id, ok := os.LookupEnv("BPI_DYNATRACE_BUILDPACK_ID")
	if !ok {
		return nil, fmt.Errorf("$BPI_DYNATRACE_BUILDPACK_ID must be set")
	}

	version, ok := os.LookupEnv("BPI_DYNATRACE_BUILDPACK_VERSION")
	if !ok {
		return nil, fmt.Errorf("$BPI_DYNATRACE_BUILDPACK_VERSION must be set")
	}

	e := make(map[string]string)

	uri := fmt.Sprintf("%s/v1/deployment/installer/agent/connectioninfo", dt.BaseURI(b))

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create new GET request for %s\n%w", uri, err)
	}
	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", id, version))
	req.Header.Set("Authorization", fmt.Sprintf("Api-Token %s", dt.APIToken(b)))

	client := http.Client{Transport: &http.Transport{Proxy: http.ProxyFromEnvironment}}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to request %s\n%w", uri, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("could not download %s: %d", uri, resp.StatusCode)
	}

	raw := struct {
		Tenant                  string   `json:"tenantUUID"`
		TenantToken             string   `json:"tenantToken"`
		CommunicationsEndpoints []string `json:"communicationEndpoints"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("unable to decode payload\n%w", err)
	}

	e["DT_TENANT"] = raw.Tenant
	e["DT_TENANTTOKEN"] = raw.TenantToken
	e["DT_CONNECTION_POINT"] = strings.Join(raw.CommunicationsEndpoints, ";")

	delete(b.Secret, "api-token")
	delete(b.Secret, "apitoken")
	delete(b.Secret, "api-url")
	delete(b.Secret, "apiurl")
	delete(b.Secret, "environment-id")

	for k, v := range b.Secret {
		s := strings.ToUpper(k)
		s = strings.ReplaceAll(s, "-", "_")
		s = strings.ReplaceAll(s, ".", "_")

		e[fmt.Sprintf("DT_%s", s)] = v
	}

	return e, nil
}
