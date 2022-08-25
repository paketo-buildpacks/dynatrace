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

func testBaseURI(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	it("uses api-url", func() {
		Expect(dt.BaseURI(libcnb.Binding{Secret: map[string]string{"api-url": "test-url"}})).
			To(Equal("test-url"))
	})

	it("uses environment-id", func() {
		Expect(dt.BaseURI(libcnb.Binding{Secret: map[string]string{"environment-id": "test-id"}})).
			To(Equal("https://test-id.live.dynatrace.com/api"))
	})
}
