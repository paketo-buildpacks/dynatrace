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
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/libpak"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/dynatrace/v4/dt"
)

func testAgent(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx libcnb.BuildContext
	)

	it.Before(func() {
		var err error

		ctx.Buildpack.Info.ID = "test-id"
		ctx.Buildpack.Info.Version = "test-version"

		ctx.Layers.Path, err = ioutil.TempDir("", "java-agent-layers")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(ctx.Layers.Path)).To(Succeed())
	})

	it("contributes agent", func() {
		dep := libpak.BuildpackDependency{
			URI:    "https://localhost/stub-dynatrace-agent.zip",
			SHA256: "1da90986465057b9b455363124a52fac78f025e5295c989807e090018cc37dc1",
		}
		dc := libpak.DependencyCache{CachePath: "testdata"}

		j, _ := dt.NewAgent(dep, dc, "test-api-token", ctx.Buildpack.Info)
		layer, err := ctx.Layers.Layer("test-layer")
		Expect(err).NotTo(HaveOccurred())

		layer, err = j.Contribute(layer)
		Expect(err).NotTo(HaveOccurred())

		Expect(layer.Launch).To(BeTrue())
		Expect(filepath.Join(layer.Path, "fixture-marker")).To(BeARegularFile())
		Expect(layer.LaunchEnvironment["BPI_DYNATRACE_BUILDPACK_ID.default"]).To(Equal("test-id"))
		Expect(layer.LaunchEnvironment["BPI_DYNATRACE_BUILDPACK_VERSION.default"]).To(Equal("test-version"))
		Expect(layer.LaunchEnvironment["DT_LOGSTREAM.default"]).To(Equal("stdout"))
		Expect(layer.LaunchEnvironment["DT_CUSTOM_PROP.delim"]).To(Equal(" "))
		Expect(layer.LaunchEnvironment["DT_CUSTOM_PROP.append"]).To(Equal("CloudNativeBuildpackVersion=test-version"))
		Expect(layer.LaunchEnvironment["LD_PRELOAD.delim"]).To(Equal(string(os.PathListSeparator)))
		Expect(layer.LaunchEnvironment["LD_PRELOAD.prepend"]).To(Equal(fmt.Sprintf("%s/agent/lib64/liboneagentproc.so", layer.Path)))
	})

	it("modifies dependency request with Authorization header", func() {
		dep := libpak.BuildpackDependency{
			URI:    "https://localhost/stub-dynatrace-agent.zip",
			SHA256: "1da90986465057b9b455363124a52fac78f025e5295c989807e090018cc37dc1",
		}
		dc := libpak.DependencyCache{CachePath: "testdata"}

		j, _ := dt.NewAgent(dep, dc, "test-api-token", ctx.Buildpack.Info)

		req, err := j.LayerContributor.RequestModifierFuncs[0](&http.Request{Header: http.Header{}})
		Expect(err).NotTo(HaveOccurred())

		Expect(req.Header.Get("Authorization")).To(Equal("Api-Token test-api-token"))
	})
}
