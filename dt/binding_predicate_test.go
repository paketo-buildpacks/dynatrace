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
	"github.com/paketo-buildpacks/dynatrace/v4/dt"
	"github.com/sclevine/spec"
)

func testIsDynatraceBinding(t *testing.T, context spec.G, it spec.S) {
	it("returns false for initial binding", func() {
		result := dt.IsDynatraceBinding(libcnb.Binding{})
		Expect(result).To(BeFalse())
	})

	it("returns false for non dynatrace binding", func() {
		result := dt.IsDynatraceBinding(libcnb.Binding{Name: "foo", Type: "bar"})
		Expect(result).To(BeFalse())
	})

	it("returns true for the type Dynatrace", func() {
		result := dt.IsDynatraceBinding(libcnb.Binding{Name: "foo", Type: "Dynatrace"})
		Expect(result).To(BeTrue())
	})

	it("returns true for the type dynatrace", func() {
		result := dt.IsDynatraceBinding(libcnb.Binding{Name: "foo", Type: "dynatrace"})
		Expect(result).To(BeTrue())
	})

	it("returns true for the name Dynatrace", func() {
		result := dt.IsDynatraceBinding(libcnb.Binding{Name: "Dynatrace", Type: "user-provided"})
		Expect(result).To(BeTrue())
	})

	it("returns true if the name contains dynatrace", func() {
		result := dt.IsDynatraceBinding(libcnb.Binding{Name: "my-dynatrace-binding", Type: "user-provided"})
		Expect(result).To(BeFalse())
	})

}
