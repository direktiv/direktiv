import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import request from '../common/request'
import { retry10 } from '../common/retry'

const testNamespace = 'gateway'
const limitedNamespace = 'limited_namespace'

const workflow = `
  direktiv_api: workflow/v1
  description: A simple 'no-op' state that returns 'Hello world!'
  states:
  - id: helloworld
    type: noop
    transform:
      result: Hello world!
`

const workflowNotToBetriggered = `
  direktiv_api: workflow/v1
  description: A simple 'no-op' state that returns 'Hello world!'
  states:
  - id: helloworld
    type: noop
    transform:
      result: This wf should not be triggered!
`

const workflowEcho = `
  direktiv_api: workflow/v1
  description: A simple 'no-op' state that returns 'Hello world!'
  states:
  - id: helloworld
    type: noop
    transform:
      result: 'jq(.)'
`

const endpointWorkflow = `
  direktiv_api: endpoint/v1
  allow_anonymous: true
  plugins:
    target:
      type: target-flow
      configuration:
          namespace: ` + testNamespace + `
          flow: /workflow.yaml
  methods: 
    - GET
  path: /endpoint1`

const endpointTargetLimitedNamespaceWorkflow = `
  direktiv_api: endpoint/v1
  allow_anonymous: true
  plugins:
    target:
      type: target-flow
      configuration:
          namespace: ` + limitedNamespace + `
          flow: /workflow.yaml
  methods: 
    - GET
  path: /endpoint1`

const endpointPOSTWorkflow = `
  direktiv_api: endpoint/v1
  allow_anonymous: true
  plugins:
    target:
      type: target-flow
      configuration:
          namespace: ` + testNamespace + `
          flow: /workflow.yaml
  methods: 
    - POST
  path: /endpoint1`

const endpointComplexPOSTWorkflow = `
  direktiv_api: endpoint/v1
  allow_anonymous: true
  plugins:
    inbound:
      - type: js-inbound
        configuration:
          script: b = JSON.parse(input["Body"]); b["message"] = "Changed"; input["Body"] = JSON.stringify(b);
    target:
      type: target-flow
      configuration:
          namespace: ` + testNamespace + `
          flow: /workflow.yaml
  methods: 
    - POST
  path: /endpoint1`

const endpointWorkflowAllowed = `
  direktiv_api: endpoint/v1
  allow_anonymous: true
  plugins:
    target:
      type: target-flow
      configuration:
          namespace: ` + limitedNamespace + `
          flow: /workflow.yaml
          content_type: text/json
  methods: 
    - GET
  path: /endpoint2`

const endpointBroken = `
  direktiv_api: endpoint/v1
  allow_anonymous: true
  plugins:
    target:
      type: target-flow
  methods: 
    - GET
  path: /endpoint3`

const endpointErrorWorkflow = `direktiv_api: endpoint/v1
allow_anonymous: true
plugins:
  target:
    type: target-flow
    configuration:
      flow: /ep3.yaml
methods: 
  - GET
path: /endpoint3`

const errorWorkflow = `direktiv_api: workflow/v1
states:
- id: a
  type: error
  error: badinput
  message: 'Missing or invalid value for required input.'
`

const endpointNoContentType = `direktiv_api: endpoint/v1
allow_anonymous: true
plugins:
  target:
    type: target-flow
    configuration:
      flow: /contentType.yaml
methods: 
  - GET
path: /endpointct`

const endpointContentType = `direktiv_api: endpoint/v1
allow_anonymous: true
plugins:
  target:
    type: target-flow
    configuration:
      flow: /contentType.yaml
      content_type: test/me
methods: 
  - GET
path: /endpointcttest`

const contentType = `
direktiv_api: workflow/v1
states:
- id: helloworld
  type: noop
  transform:
    result: Hello world!
`

describe('Test target workflow wrong config', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFileV2(
		it,
		expect,
		testNamespace,
		'/', 'ep3.yaml', 'endpoint',
		endpointBroken,
	)

	retry10(`should list all services`, async () => {
		const listRes = await request(common.config.getDirektivHost()).get(
			`/api/v2/namespaces/${ testNamespace }/gateway/routes`,
		)
		expect(listRes.statusCode).toEqual(200)
		expect(listRes.body.data.length).toEqual(1)
		expect(listRes.body.data).toEqual(
			expect.arrayContaining(
				[
					{
						file_path: '/ep3.yaml',
						path: '/endpoint3',
						methods: [ 'GET' ],
						allow_anonymous: true,
						server_path: '/gw/endpoint3',
						timeout: 0,
						errors: [ 'flow required' ],
						warnings: [],
						plugins: { target: { type: 'target-flow' } },
					},
				],
			),
		)
	})
})

describe('Test target workflow with errors', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFileV2(
		it,
		expect,
		testNamespace,
		'/', 'ep3.yaml', 'workflow',
		errorWorkflow,
	)

	common.helpers.itShouldCreateYamlFileV2(
		it,
		expect,
		testNamespace,
		'/', 'eperr3.yaml', 'endpoint',
		endpointErrorWorkflow,
	)

	retry10(`should return a workflow run from magic namespace`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/gw/endpoint3`,
		)
		expect(req.statusCode).toEqual(500)
		expect(req.text).toContain('error executing workflow: badinput: Missing or invalid value for required input.')
	})
})

describe('Test target workflow plugin', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, limitedNamespace)
	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFileV2(
		it,
		expect,
		testNamespace,
		'/', 'workflow.yaml', 'workflow',
		workflow,
	)

	common.helpers.itShouldCreateYamlFileV2(
		it,
		expect,
		limitedNamespace,
		'/', 'workflow.yaml', 'workflow',
		workflow,
	)

	common.helpers.itShouldCreateYamlFileV2(
		it,
		expect,
		limitedNamespace,
		'/', 'endpoint1.yaml', 'endpoint',
		endpointWorkflow,
	)

	common.helpers.itShouldCreateYamlFileV2(
		it,
		expect,
		limitedNamespace,
		'/', 'endpoint2.yaml', 'endpoint',
		endpointWorkflowAllowed,
	)

	common.helpers.itShouldCreateYamlFileV2(
		it,
		expect,
		testNamespace,
		'/', 'endpoint1.yaml', 'endpoint',
		endpointWorkflow,
	)

	common.helpers.itShouldCreateYamlFileV2(
		it,
		expect,
		testNamespace,
		'/', 'endpoint2.yaml', 'endpoint',
		endpointWorkflowAllowed,
	)

	retry10(`should return a workflow run from magic namespace`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/gw/endpoint1`,
		)
		expect(req.statusCode).toEqual(200)
		expect(req.text).toEqual('{"result":"Hello world!"}')
	})

	retry10(`should return a flow run from magic namespace with namespace set`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/gw/endpoint2`,
		)
		expect(req.statusCode).toEqual(200)
		expect(req.text).toEqual('{"result":"Hello world!"}')
		expect(req.header['content-type']).toEqual('text/json')
	})

	retry10(`should return a workflow var from non-magic namespace`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/ns/` + limitedNamespace + `/endpoint2`,
		)
		expect(req.statusCode).toEqual(200)
		expect(req.text).toEqual('{"result":"Hello world!"}')
	})

	retry10(`should not return a workflow in onn-magic namespace`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/ns/` + limitedNamespace + `/endpoint1`,
		)
		expect(req.statusCode).toEqual(500)
	})
})

describe('Test POST method for target workflow plugin', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, limitedNamespace)
	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFileV2(
		it,
		expect,
		testNamespace,
		'/', 'workflow.yaml', 'workflow',
		workflowEcho,
	)

	common.helpers.itShouldCreateYamlFileV2(
		it,
		expect,
		testNamespace,
		'/', 'endpoint1.yaml', 'endpoint',
		endpointPOSTWorkflow,
	)

	retry10(`should return a workflow run from magic namespace`, async () => {
		const req = await request(common.config.getDirektivHost()).post(
			`/gw/endpoint1`,
		)
			.send({ message: 'Hi' })
		expect(req.statusCode).toEqual(200)
		expect(req.text).toEqual('{"result":{"message":"Hi"}}')
	})
})

describe('Test Complex POST method for target workflow plugin', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, limitedNamespace)
	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFileV2(
		it,
		expect,
		testNamespace,
		'/', 'workflow.yaml', 'workflow',
		workflowEcho,
	)

	common.helpers.itShouldCreateYamlFileV2(
		it,
		expect,
		testNamespace,
		'/', 'endpoint1.yaml', 'endpoint',
		endpointComplexPOSTWorkflow,
	)

	retry10(`should return a workflow run from magic namespace`, async () => {
		const req = await request(common.config.getDirektivHost()).post(
			`/gw/endpoint1`,
		)
			.send({ message: 'Hi' })
		expect(req.statusCode).toEqual(200)
		expect(req.text).toEqual('{"result":{"message":"Changed"}}')
	})
})

describe('Test scope for target workflow plugin', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, limitedNamespace)
	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFileV2(
		it,
		expect,
		limitedNamespace,
		'/', 'workflow.yaml', 'workflow',
		workflow,
	)

	common.helpers.itShouldCreateYamlFileV2(
		it,
		expect,
		testNamespace,
		'/', 'workflow.yaml', 'workflow',
		workflowNotToBetriggered,
	)

	common.helpers.itShouldCreateYamlFileV2(
		it,
		expect,
		testNamespace,
		'/', 'endpoint7.yaml', 'endpoint',
		endpointTargetLimitedNamespaceWorkflow,
	)

	retry10(`should return a workflow run from limited namespace`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/gw/endpoint1`,
		)
		expect(req.statusCode).toEqual(200)
		expect(req.text).toEqual('{"result":"Hello world!"}')
	})
})

describe('Test target workflow default contenttype', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFileV2(
		it,
		expect,
		testNamespace,
		'/', 'epnoct.yaml', 'endpoint',
		endpointNoContentType,
	)

	common.helpers.itShouldCreateYamlFileV2(
		it,
		expect,
		testNamespace,
		'/', 'epct.yaml', 'endpoint',
		endpointContentType,
	)

	common.helpers.itShouldCreateYamlFileV2(
		it,
		expect,
		testNamespace,
		'/', 'contentType.yaml', 'workflow',
		contentType,
	)

	retry10(`should return a json content type`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/gw/endpointct`,
		)

		expect(req.headers['content-type']).toEqual('application/json')
		expect(req.statusCode).toEqual(200)
	})

	retry10(`should return a configured content type`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/gw/endpointcttest`,
		)

		expect(req.headers['content-type']).toEqual('test/me')
		expect(req.statusCode).toEqual(200)
	})
})
