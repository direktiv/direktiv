import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import request from '../common/request'
import { retry10 } from '../common/retry'

const testNamespace = 'js-inbound'

const endpointJSFile = `
direktiv_api: endpoint/v1
allow_anonymous: true
plugins:
  target:
    type: target-flow
    configuration:
        flow: /target.yaml
        content_type: application/json
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
methods: 
  - POST
path: /target`

const wf = `
direktiv_api: workflow/v1
states:
- id: helloworld
  type: noop
  transform:
    result: jq(.)
`

const consumer = `
direktiv_api: "consumer/v1"
username: "demo"
api_key: "apikey"
groups:
  - "group1"
`

const endpointConsumerFile = `
direktiv_api: endpoint/v1
allow_anonymous: false
plugins:
  target:
    type: target-flow
    configuration:
        flow: /target.yaml
        content_type: application/json
  auth:
    - type: key-auth
  inbound:
    - type: js-inbound
      configuration:
        script: |
          b = JSON.parse(input["Body"])
          b["user"] = input["Consumer"].Username
          input["Body"] = JSON.stringify(b) 
methods: 
  - POST
path: /target`

const endpointParamFile = `
direktiv_api: endpoint/v1
allow_anonymous: true
plugins:
  target:
    type: target-flow
    configuration:
        flow: /target.yaml
        content_type: application/json
  inbound:
    - type: js-inbound
      configuration:
        script: |
          b = JSON.parse(input["Body"])
          b["params"] = input["URLParams"].id
          input["Body"] = JSON.stringify(b) 
methods: 
  - POST
path: /target/{id}`

const endpointErrorFile = `
direktiv_api: endpoint/v1
allow_anonymous: true
plugins:
  target:
    type: target-flow
    configuration:
        flow: /target.yaml
        content_type: application/json
  inbound:
    - type: js-inbound
      configuration:
        script: |
          b = JSON.parse(input["Body"])
          b["error"] = "no access" 
          input["Body"] = JSON.stringify(b) 
          input["Headers"].Add("permission", "denied")
          input.Status = 403        
methods: 
  - POST
path: /target`

describe('Test js inbound plugin', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)
	// common.helpers.itShouldCreateNamespace(it, expect, testNamespace);

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'endpoint1.yaml', 'endpoint',
		endpointJSFile,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'target.yaml', 'workflow',
		wf,
	)

	retry10(`should have expected body after js`, async () => {
		const req = await request(common.config.getDirektivHost()).post(
			`/ns/` + testNamespace + `/target?Query1=value1&Query2=value2`,
		)
			.set('Header1', 'Value1')
			.send({ hello: 'world' })

		expect(req.statusCode).toEqual(200)
		expect(req.body.result.addheader).toEqual('value3')
		expect(req.body.result.addheaderdel).toEqual('')
		expect(req.body.result.addquery).toEqual('value1')
		expect(req.body.result.addquerydel).toEqual('')
	})
})

describe('Test js inbound plugin consumer', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'consumer.yaml', 'consumer',
		consumer,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'endpoint1.yaml', 'endpoint',
		endpointConsumerFile,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'target.yaml', 'workflow',
		wf,
	)

	retry10(`should have expected body after js`, async () => {
		const req = await request(common.config.getDirektivHost()).post(
			`/ns/` + testNamespace + `/target`,
		)
			.set('API-Token', 'apikey')
			.send({ hello: 'world' })

		expect(req.statusCode).toEqual(200)
		expect(req.body.result.user).toEqual('demo')
	})
})

describe('Test js inbound plugin url params', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'endpoint1.yaml', 'endpoint',
		endpointParamFile,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'target.yaml', 'workflow',
		wf,
	)

	retry10(`should have expected body after js`, async () => {
		const req = await request(common.config.getDirektivHost()).post(
			`/ns/` + testNamespace + `/target/myid`,
		)
			.send({ hello: 'world' })

		expect(req.statusCode).toEqual(200)
		expect(req.body.result.params).toEqual('myid')
	})
})

describe('Test js inbound plugin errors', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'endpoint1.yaml', 'endpoint',
		endpointErrorFile,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'target.yaml', 'workflow',
		wf,
	)

	retry10(`should have expected body after js`, async () => {
		const req = await request(common.config.getDirektivHost()).post(
			`/ns/` + testNamespace + `/target`,
		)
			.send({ hello: 'world' })

		expect(req.statusCode).toEqual(403)
	})
})
