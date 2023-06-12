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
	"testing"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/dynatrace/v4/dt"
)

var expectedResult = libcnb.DetectResult{
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
}

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx    libcnb.DetectContext
		detect dt.Detect
	)

	it("fails detection without service", func() {
		actualResult, err := detect.Detect(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(actualResult.Pass).To(BeFalse())
		Expect(actualResult).To(Equal(libcnb.DetectResult{}))
	})

	it("passes with service of type dynatrace", func() {
		ctx.Platform.Bindings = libcnb.Bindings{
			{Name: "test-service", Type: "Dynatrace"},
		}

		actualResult, err := detect.Detect(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(actualResult.Pass).To(BeTrue())
		Expect(actualResult).To(Equal(expectedResult))
	})

	it("passes with service with name dynatrace", func() {
		ctx.Platform.Bindings = libcnb.Bindings{
			{Name: "Dynatrace", Type: "user-provided"},
		}

		actualResult, err := detect.Detect(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(actualResult.Pass).To(BeTrue())
		Expect(actualResult).To(Equal(expectedResult))
	})

	it("passes with services with name and of type dynatrace", func() {
		ctx.Platform.Bindings = libcnb.Bindings{
			{Name: "Dynatrace", Type: "user-provided"},
			{Name: "provided", Type: "Dynatrace"},
		}

		actualResult, err := detect.Detect(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(actualResult.Pass).To(BeTrue())
		Expect(actualResult).To(Equal(expectedResult))
	})

	it("errors with multiple services with name dynatrace", func() {
		ctx.Platform.Bindings = libcnb.Bindings{
			{Name: "Dynatrace", Type: "user-provided"},
			{Name: "Dynatrace", Type: "user-provided"},
		}

		_, err := detect.Detect(ctx)
		Expect(err).To(MatchError(ContainSubstring("unable to resolve")))
	})

}
