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

package helper_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/dynatrace/v4/helper"
)

func testProperties(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		p      helper.Properties
		server *ghttp.Server
	)

	it.Before(func() {
		RegisterTestingT(t)
		server = ghttp.NewServer()
	})

	it.After(func() {
		server.Close()
	})

	it("does not contribute properties if no binding exists", func() {
		Expect(p.Execute()).To(BeNil())
	})

	context("with binding", func() {
		it.Before(func() {
			p.Bindings = libcnb.Bindings{
				{
					Name: "test-binding",
					Type: "Dynatrace",
					Secret: map[string]string{
						"api-token": "test-api-token",
						"api-url":   server.URL(),
						"test-key":  "test-value",
					},
				},
			}
		})

		it("returns error if $BPI_DYNATRACE_BUILDPACK_ID is not set", func() {
			_, err := p.Execute()
			Expect(err).To(MatchError("$BPI_DYNATRACE_BUILDPACK_ID must be set"))
		})

		context("$BPI_DYNATRACE_BUILDPACK_ID", func() {
			it.Before(func() {
				Expect(os.Setenv("BPI_DYNATRACE_BUILDPACK_ID", "test-id")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("BPI_DYNATRACE_BUILDPACK_ID")).To(Succeed())
			})

			it("returns error if $BPI_DYNATRACE_BUILDPACK_VERSION is not set", func() {
				_, err := p.Execute()
				Expect(err).To(MatchError("$BPI_DYNATRACE_BUILDPACK_VERSION must be set"))
			})

			context("$BPI_DYNATRACE_BUILDPACK_VERSION", func() {
				it.Before(func() {
					Expect(os.Setenv("BPI_DYNATRACE_BUILDPACK_VERSION", "test-version")).To(Succeed())
				})

				it.After(func() {
					Expect(os.Unsetenv("BPI_DYNATRACE_BUILDPACK_VERSION")).To(Succeed())
				})

				it("contributes properties if binding exists", func() {
					server.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v1/deployment/installer/agent/connectioninfo"),
						ghttp.VerifyHeaderKV("Authorization", "Api-Token test-api-token"),
						ghttp.VerifyHeaderKV("User-Agent", "test-id/test-version"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]interface{}{
							"tenantUUID":             "test-tenant-uuid",
							"tenantToken":            "test-tenant-token",
							"communicationEndpoints": []string{"test-communication-endpoint-1", "test-communication-endpoint-2"},
						}),
					))

					Expect(p.Execute()).To(Equal(map[string]string{
						"DT_CONNECTION_POINT": "test-communication-endpoint-1;test-communication-endpoint-2",
						"DT_TENANT":           "test-tenant-uuid",
						"DT_TENANTTOKEN":      "test-tenant-token",
						"DT_TEST_KEY":         "test-value",
					}))
				})
			})
		})
	})

}
