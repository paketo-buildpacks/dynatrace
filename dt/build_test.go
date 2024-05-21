/*
 * Copyright 2018-2024 the original author or authors.
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

const stackId = "test-stack-id"

func getExpectedDependency(serverUrl string) libpak.BuildpackDependency {
	return libpak.BuildpackDependency{
		ID:      "dynatrace-oneagent",
		Name:    "Dynatrace OneAgent",
		Version: "test-version",
		URI:     fmt.Sprintf("%s/v1/deployment/installer/agent/unix/paas/latest?bitness=64&skipMetadata=true&arch=x86&include=java&include=php", serverUrl),
		Stacks:  []string{stackId},
		PURL:    "pkg:generic/dynatrace-one-agent@test-version?arch=amd64",
		CPEs:    []string{"cpe:2.3:a:dynatrace:one-agent:test-version:*:*:*:*:*:*:*"},
	}
}

func getExpectedAllDependency(serverUrl string) libpak.BuildpackDependency {
	return libpak.BuildpackDependency{
		ID:      "dynatrace-oneagent",
		Name:    "Dynatrace OneAgent",
		Version: "test-version",
		URI:     fmt.Sprintf("%s/v1/deployment/installer/agent/unix/paas/latest?bitness=64&skipMetadata=true&arch=arm&include=all", serverUrl),
		Stacks:  []string{stackId},
		PURL:    "pkg:generic/dynatrace-one-agent@test-version?arch=arm64",
		CPEs:    []string{"cpe:2.3:a:dynatrace:one-agent:test-version:*:*:*:*:*:*:*"},
	}
}

func verifyBOM(bom *libcnb.BOM) {
	ExpectWithOffset(1, bom.Entries).To(HaveLen(2))
	ExpectWithOffset(1, bom.Entries[0].Name).To(Equal("dynatrace-oneagent"))
	ExpectWithOffset(1, bom.Entries[0].Launch).To(BeTrue())
	ExpectWithOffset(1, bom.Entries[0].Build).To(BeFalse())
	ExpectWithOffset(1, bom.Entries[1].Name).To(Equal("helper"))
	ExpectWithOffset(1, bom.Entries[1].Launch).To(BeTrue())
	ExpectWithOffset(1, bom.Entries[1].Build).To(BeFalse())

}

func verifyLayers(layers []libcnb.LayerContributor, serverUrl string, expectation func(string) libpak.BuildpackDependency) {
	ExpectWithOffset(1, layers).To(HaveLen(2))
	ExpectWithOffset(1, layers[0].Name()).To(Equal("dynatrace-oneagent"))
	ExpectWithOffset(1, layers[0].(dt.Agent).LayerContributor.Dependency).To(Equal(expectation(serverUrl)))
	ExpectWithOffset(1, layers[1].Name()).To(Equal("helper"))
	ExpectWithOffset(1, layers[1].(libpak.HelperLayerContributor).Names).To(Equal([]string{"properties"}))
}

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx    libcnb.BuildContext
		server *ghttp.Server
	)

	it.Before(func() {
		RegisterTestingT(t)
		server = ghttp.NewServer()

		t.Setenv("BP_ARCH", "amd64")

		ctx.Buildpack.Info.ID = "test-id"
		ctx.Buildpack.Info.Version = "test-version"
		ctx.StackID = stackId
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

		server.AppendHandlers(ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/v1/deployment/installer/agent/unix/paas/latest/metainfo"),
			ghttp.VerifyHeaderKV("Authorization", "Api-Token test-api-token"),
			ghttp.VerifyHeaderKV("User-Agent", "test-id/test-version"),
			ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]interface{}{"latestAgentVersion": "test-version"}),
		))

		ctx.Plan.Entries = append(ctx.Plan.Entries,
			libcnb.BuildpackPlanEntry{Name: "dynatrace-java"},
			libcnb.BuildpackPlanEntry{Name: "dynatrace-php"})
	})

	it.After(func() {
		server.Close()
	})

	it("contributes agent", func() {
		result, err := dt.Build{}.Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		verifyLayers(result.Layers, server.URL(), getExpectedDependency)
		verifyBOM(result.BOM)
	})

	it("also takes named binding into account", func() {
		ctx.Platform.Bindings = libcnb.Bindings{
			{
				Name: "DynatraceBinding",
				Type: "user-provided",
				Secret: map[string]string{
					"api-token": "custom-api-token",
					"api-url":   server.URL(),
				},
			},
		}

		server.SetHandler(0, ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/v1/deployment/installer/agent/unix/paas/latest/metainfo"),
			ghttp.VerifyHeaderKV("Authorization", "Api-Token custom-api-token"),
			ghttp.VerifyHeaderKV("User-Agent", "test-id/test-version"),
			ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]interface{}{"latestAgentVersion": "test-version"}),
		))

		result, err := dt.Build{}.Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		verifyLayers(result.Layers, server.URL(), getExpectedDependency)
		verifyBOM(result.BOM)
	})

	it("supports apitoken and apiurl", func() {
		ctx.Platform.Bindings = libcnb.Bindings{
			{
				Name: "DynatraceBinding",
				Type: "user-provided",
				Secret: map[string]string{
					"apitoken": "custom-apitoken",
					"apiurl":   server.URL(),
				},
			},
		}

		server.SetHandler(0, ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/v1/deployment/installer/agent/unix/paas/latest/metainfo"),
			ghttp.VerifyHeaderKV("Authorization", "Api-Token custom-apitoken"),
			ghttp.VerifyHeaderKV("User-Agent", "test-id/test-version"),
			ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]interface{}{"latestAgentVersion": "test-version"}),
		))

		result, err := dt.Build{}.Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		verifyLayers(result.Layers, server.URL(), getExpectedDependency)
		verifyBOM(result.BOM)
	})

	context("python", func() {
		it.Before(func() {
			t.Setenv("BP_ARCH", "arm64")

			ctx.Plan.Entries = append(ctx.Plan.Entries,
				libcnb.BuildpackPlanEntry{Name: "dynatrace-java"},
				libcnb.BuildpackPlanEntry{Name: "dynatrace-python"})
		})

		it("contributes all agent", func() {
			result, err := dt.Build{}.Build(ctx)
			Expect(err).NotTo(HaveOccurred())

			verifyLayers(result.Layers, server.URL(), getExpectedAllDependency)
			verifyBOM(result.BOM)
		})
	})
}
