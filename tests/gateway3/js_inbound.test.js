import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import request from '../common/request'
import { retry10 } from '../common/retry'

const testNamespace = 'js-inbound'

const endpointJSFile = `x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/target"
    allow_anonymous: true
    plugins:
      inbound:
      - type: js-inbound
        configuration:
          script: |
              input["Headers"].Del("Header1")
              input["Headers"].Add("Header3", "value3")
              input["Queries"].Del("Query2")
              b = JSON.parse(input["Body"])
              b["addquery"] = input["Queries"].Get("Query1")
              b["addquerydel"] = input["Queries"].Get("Query2")

              b["addheader"] = input["Headers"].Get("Header3")
              b["addheaderdel"] = input["Headers"].Get("Header1")
              input["Body"] = JSON.stringify(b) 
      target:
        type: target-flow
        configuration:
          flow: /foo.wf.ts
          content_type: application/json
post:
   responses:
      "200":
        description: works`

const wf = `
function stateFirst(input) {
	return finish(input)
}
`

const consumer = `
direktiv_api: "consumer/v1"
username: "demo"
api_key: "apikey"
groups:
  - "group1"
`

const endpointConsumerFile = `x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/target"
    allow_anonymous: false
    plugins:
      auth:
      - type: key-auth
      inbound:
      - type: js-inbound
        configuration:
          script: |
            b = JSON.parse(input["Body"])
            b["user"] = input["Consumer"].Username
            input["Body"] = JSON.stringify(b) 
      target:
        type: target-flow
        configuration:
          flow: /foo.wf.ts
          content_type: application/json
post:
   responses:
      "200":
        description: works`

const endpointParamFile = `x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/target/{id}"
    allow_anonymous: true
    plugins:
      inbound:
      - type: js-inbound
        configuration:
          script: |
            b = JSON.parse(input["Body"])
            b["params"] = input["URLParams"].id
            input["Body"] = JSON.stringify(b) 
      target:
        type: target-flow
        configuration:
          flow: /foo.wf.ts
          content_type: application/json
post:
   responses:
      "200":
        description: works`

const endpointErrorFile = `x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/target"
    allow_anonymous: true
    plugins:
      inbound:
      - type: js-inbound
        configuration:
          script: |
            b = JSON.parse(input["Body"])
            b["error"] = "no access" 
            input["Body"] = JSON.stringify(b) 
            input["Headers"].Add("permission", "denied")
            input.Status = 403 
      target:
        type: target-flow
        configuration:
          flow: /foo.wf.ts
          content_type: application/json
post:
   responses:
      "200":
        description: works`

describe('Test js inbound plugin', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)
	// common.helpers.itShouldCreateNamespace(it, expect, testNamespace);

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/',
		'endpoint1.yaml',
		'endpoint',
		endpointJSFile,
	)

	common.helpers.itShouldTSWorkflow(
		it,
		expect,
		testNamespace,
		'/',
		'foo.wf.ts',
		wf,
	)

	retry10(`should have expected body after js`, async () => {
		const req = await request(common.config.getDirektivBaseUrl())
			.post(`/ns/` + testNamespace + `/target?Query1=value1&Query2=value2`)
			.set('Header1', 'Value1')
			.send({ hello: 'world' })

		const got = req.body.data

		expect(req.statusCode).toEqual(200)
		expect(got.addheader).toEqual('value3')
		expect(got.addheaderdel).toEqual('')
		expect(got.addquery).toEqual('value1')
		expect(got.addquerydel).toEqual('')
	})
})

describe('Test js inbound plugin consumer', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/',
		'consumer.yaml',
		'consumer',
		consumer,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/',
		'endpoint1.yaml',
		'endpoint',
		endpointConsumerFile,
	)

	common.helpers.itShouldTSWorkflow(
		it,
		expect,
		testNamespace,
		'/',
		'foo.wf.ts',
		wf,
	)

	retry10(`should have expected body after js`, async () => {
		const req = await request(common.config.getDirektivBaseUrl())
			.post(`/ns/` + testNamespace + `/target`)
			.set('API-Token', 'apikey')
			.send({ hello: 'world' })

		const got = req.body.data

		expect(req.statusCode).toEqual(200)
		expect(got.user).toEqual('demo')
	})
})

describe('Test js inbound plugin url params', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/',
		'endpoint1.yaml',
		'endpoint',
		endpointParamFile,
	)

	common.helpers.itShouldTSWorkflow(
		it,
		expect,
		testNamespace,
		'/',
		'foo.wf.ts',
		wf,
	)

	retry10(`should have expected body after js`, async () => {
		const req = await request(common.config.getDirektivBaseUrl())
			.post(`/ns/` + testNamespace + `/target/myid`)
			.send({ hello: 'world' })

		const got = req.body.data

		expect(req.statusCode).toEqual(200)
		expect(got.params).toEqual('myid')
	})
})

describe('Test js inbound plugin errors', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/',
		'endpoint1.yaml',
		'endpoint',
		endpointErrorFile,
	)

	common.helpers.itShouldTSWorkflow(
		it,
		expect,
		testNamespace,
		'/',
		'foo.wf.ts',
		wf,
	)

	retry10(`should have expected body after js`, async () => {
		const req = await request(common.config.getDirektivBaseUrl())
			.post(`/ns/` + testNamespace + `/target`)
			.send({ hello: 'world' })

		expect(req.statusCode).toEqual(403)
	})
})
