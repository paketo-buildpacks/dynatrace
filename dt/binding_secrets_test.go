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

var createBinding = func(keysAndValues ...string) libcnb.Binding {
	secrets := make(map[string]string, len(keysAndValues)/2)

	for i := 0; i < len(keysAndValues); i = i + 2 {
		secrets[keysAndValues[i]] = keysAndValues[i+1]
	}

	return libcnb.Binding{Secret: secrets}
}

func testBaseURI(t *testing.T, _ spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	it("uses api-url", func() {
		Expect(dt.BaseURI(createBinding("api-url", "test-url"))).
			To(Equal("test-url"))
	})

	it("uses api-url when both api-url and apiurl are set", func() {
		Expect(dt.BaseURI(createBinding("api-url", "test-url", "apiurl", "other-url"))).
			To(Equal("test-url"))
	})

	it("uses apiurl", func() {
		Expect(dt.BaseURI(createBinding("apiurl", "other-url"))).
			To(Equal("other-url"))
	})

	it("uses environment-id", func() {
		Expect(dt.BaseURI(createBinding("environment-id", "test-id"))).
			To(Equal("https://test-id.live.dynatrace.com/api"))
	})
}

func testAPIToken(t *testing.T, _ spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	it("uses api-token", func() {
		Expect(dt.APIToken(createBinding("api-token", "test-token"))).
			To(Equal("test-token"))
	})

	it("uses api-token when both api-token and apitoken are set", func() {
		Expect(dt.APIToken(createBinding("api-token", "test-token", "apitoken", "other-token"))).
			To(Equal("test-token"))
	})

	it("uses apitoken", func() {
		Expect(dt.APIToken(createBinding("apitoken", "other-token"))).
			To(Equal("other-token"))
	})
}
