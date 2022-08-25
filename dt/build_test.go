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

package dt_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/paketo-buildpacks/libpak"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/dynatrace/v4/dt"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx    libcnb.BuildContext
		server *ghttp.Server
	)

	it.Before(func() {
		RegisterTestingT(t)
		server = ghttp.NewServer()

		ctx.Buildpack.Info.ID = "test-id"
		ctx.Buildpack.Info.Version = "test-version"
		ctx.Buildpack.API = "0.7"

		ctx.Platform.Bindings = libcnb.Bindings{
			{
				Name: "test-binding",
				Type: "Dynatrace",
				Secret: map[string]string{
					"api-token": "test-api-token",
					"api-url":   server.URL(),
				},
			},
		}
	})

	it.After(func() {
		server.Close()
	})

	it("contributes agent for API 0.7+", func() {
		server.AppendHandlers(ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/v1/deployment/installer/agent/unix/paas/latest/metainfo"),
			ghttp.VerifyHeaderKV("Authorization", "Api-Token test-api-token"),
			ghttp.VerifyHeaderKV("User-Agent", "test-id/test-version"),
			ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]interface{}{"latestAgentVersion": "test-version"}),
		))

		ctx.Plan.Entries = append(ctx.Plan.Entries,
			libcnb.BuildpackPlanEntry{Name: "dynatrace-java"},
			libcnb.BuildpackPlanEntry{Name: "dynatrace-php"})
		ctx.StackID = "test-stack-id"

		result, err := dt.Build{}.Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(2))
		Expect(result.Layers[0].Name()).To(Equal("dynatrace-oneagent"))
		Expect(result.Layers[0].(dt.Agent).LayerContributor.Dependency).To(Equal(libpak.BuildpackDependency{
			ID:      "dynatrace-oneagent",
			Name:    "Dynatrace OneAgent",
			Version: "test-version",
			URI:     fmt.Sprintf("%s/v1/deployment/installer/agent/unix/paas/latest?bitness=64&skipMetadata=true&include=java&include=php", server.URL()),
			Stacks:  []string{ctx.StackID},
			PURL:    "pkg:generic/dynatrace-one-agent@test-version?arch=amd64",
			CPEs:    []string{"cpe:2.3:a:dynatrace:one-agent:test-version:*:*:*:*:*:*:*"},
		}))
		Expect(result.Layers[1].Name()).To(Equal("helper"))
		Expect(result.Layers[1].(libpak.HelperLayerContributor).Names).To(Equal([]string{"properties"}))

		Expect(result.BOM.Entries).To(HaveLen(2))
		Expect(result.BOM.Entries[0].Name).To(Equal("dynatrace-oneagent"))
		Expect(result.BOM.Entries[0].Launch).To(BeTrue())
		Expect(result.BOM.Entries[0].Build).To(BeFalse())
		Expect(result.BOM.Entries[1].Name).To(Equal("helper"))
		Expect(result.BOM.Entries[1].Launch).To(BeTrue())
		Expect(result.BOM.Entries[1].Build).To(BeFalse())
	})
	it("contributes agent for API <= 0.6", func() {
		server.AppendHandlers(ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/v1/deployment/installer/agent/unix/paas/latest/metainfo"),
			ghttp.VerifyHeaderKV("Authorization", "Api-Token test-api-token"),
			ghttp.VerifyHeaderKV("User-Agent", "test-id/test-version"),
			ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]interface{}{"latestAgentVersion": "test-version"}),
		))

		ctx.Plan.Entries = append(ctx.Plan.Entries,
			libcnb.BuildpackPlanEntry{Name: "dynatrace-java"},
			libcnb.BuildpackPlanEntry{Name: "dynatrace-php"})
		ctx.StackID = "test-stack-id"
		ctx.Buildpack.API = "0.6"

		result, err := dt.Build{}.Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(2))
		Expect(result.Layers[0].Name()).To(Equal("dynatrace-oneagent"))
		Expect(result.Layers[0].(dt.Agent).LayerContributor.Dependency).To(Equal(libpak.BuildpackDependency{
			ID:      "dynatrace-oneagent",
			Name:    "Dynatrace OneAgent",
			Version: "test-version",
			URI:     fmt.Sprintf("%s/v1/deployment/installer/agent/unix/paas/latest?bitness=64&skipMetadata=true&include=java&include=php", server.URL()),
			Stacks:  []string{ctx.StackID},
			PURL:    "pkg:generic/dynatrace-one-agent@test-version?arch=amd64",
			CPEs:    []string{"cpe:2.3:a:dynatrace:one-agent:test-version:*:*:*:*:*:*:*"},
		}))
		Expect(result.Layers[1].Name()).To(Equal("helper"))
		Expect(result.Layers[1].(libpak.HelperLayerContributor).Names).To(Equal([]string{"properties"}))

		Expect(result.BOM.Entries).To(HaveLen(2))
		Expect(result.BOM.Entries[0].Name).To(Equal("dynatrace-oneagent"))
		Expect(result.BOM.Entries[0].Launch).To(BeTrue())
		Expect(result.BOM.Entries[0].Build).To(BeFalse())
		Expect(result.BOM.Entries[1].Name).To(Equal("helper"))
		Expect(result.BOM.Entries[1].Launch).To(BeTrue())
		Expect(result.BOM.Entries[1].Build).To(BeFalse())
	})

}
