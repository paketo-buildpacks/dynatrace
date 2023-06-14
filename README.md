# `docker.io/paketobuildpacks/dynatrace`
The Paketo Buildpack for Dynatrace is a Cloud Native Buildpack that contributes the Dynatrace OneAgent and configures it to connect to the service.

## Behavior
This buildpack will participate if one the following conditions are met

* A binding exists with `name` of `Dynatrace`
* A binding exists with `type` of `Dynatrace`

**Note**: The binding must include the following required Secret values to successfully contribute Dynatrace

**Note**: If both bindings (`name` **and** `type`) exist, the binding with `name` will have precedence.

| Key | Value Description
| -------------------- | -----------
| `api-url`<br/>  **or** <br/> `environment-id` | The base URL of the Dynatrace API. If not set, `environment-id` must be set. <br/> --- <br/> If `api-url` is not set, a URL is configured in the form: https://<`environment-id`>.live.dynatrace.com/api
| `api-token` | (Required) The token for communicating with the Dynatrace service.


The buildpack will do the following for .NET, Go, Apache HTTPd, Java, Nginx, NodeJS, and PHP applications:

* Contributes a OneAgent including the appropriate libraries to a layer and configures `$LD_PRELOAD` to use it
* Sets `$DT_TENANT`, `$DT_TENANTTOKEN`, and `$DT_CONNECTION_POINT` at launch time.
* Transforms the contents of the binding secret to environment variables with the pattern `DT_<KEY>=<VALUE>`
  * Excluding `api-token`, `api-url`, and `environment-id`

## Bindings
The buildpack optionally accepts the following bindings:

### Type: `dependency-mapping`
|Key                   | Value   | Description
|----------------------|---------|------------
|`<dependency-digest>` | `<uri>` | If needed, the buildpack will fetch the dependency with digest `<dependency-digest>` from `<uri>`

## License

This buildpack is released under version 2.0 of the [Apache License][a].

[a]: http://www.apache.org/licenses/LICENSE-2.0
