import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import request from '../common/request'
import { retry10, retry50 } from '../common/retry'

const testNamespace = 'system'
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
x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/endpoint1"
    allow_anonymous: true
    plugins:
      target:
        type: target-flow
        configuration:
          namespace: ` + testNamespace + `
          flow: /workflow.yaml
get:
   responses:
      "200":
        description: works`

const endpointTargetLimitedNamespaceWorkflow = `
x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/endpoint1"
    allow_anonymous: true
    plugins:
      target:
        type: target-flow
        configuration:
          namespace: ` + limitedNamespace + `
          flow: /workflow.yaml
get:
   responses:
      "200":
        description: works`

const endpointPOSTWorkflow = `
x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/endpoint1"
    allow_anonymous: true
    plugins:
      target:
        type: target-flow
        configuration:
          namespace: ` + testNamespace + `
          flow: /workflow.yaml
post:
   responses:
      "200":
        description: works`

const endpointComplexPOSTWorkflow = `
x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/endpoint1"
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
post:
   responses:
      "200":
        description: works`

const endpointWorkflowAllowed = `
x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/endpoint2"
    allow_anonymous: true
    plugins:
      target:
        type: target-flow
        configuration:
          namespace: ` + limitedNamespace + `
          flow: /workflow.yaml
          content_type: text/json
get:
   responses:
      "200":
        description: works`

const endpointBroken = `
x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/endpoint3"
    allow_anonymous: true
    plugins:
      target:
        type: target-flow
get:
   responses:
      "200":
        description: works`

const endpointErrorWorkflow = `
x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/endpoint3"
    allow_anonymous: true
    plugins:
      target:
        type: target-flow
        configuration:
          flow: /ep3.yaml
get:
   responses:
      "200":
        description: works`

const errorWorkflow = `direktiv_api: workflow/v1
states:
- id: a
  type: error
  error: badinput
  message: 'Missing or invalid value for required input.'
`

const endpointNoContentType = `
x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/endpointct"
    allow_anonymous: true
    plugins:
      target:
        type: target-flow
        configuration:
          flow: /contentType.yaml
get:
   responses:
      "200":
        description: works`

const endpointContentType = `
x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/endpointcttest"
    allow_anonymous: true
    plugins:
      target:
        type: target-flow
        configuration:
          flow: /contentType.yaml
          content_type: test/me
get:
   responses:
      "200":
        description: works`

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

	common.helpers.itShouldCreateYamlFile(
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
		expect(listRes.body.data[0]).toEqual({
			spec: expect.anything(),
			file_path: '/ep3.yaml',
			server_path: '/ns/system/endpoint3',
			errors: [ "plugin 'target-flow' err: flow required" ],
			warnings: [],
		})
	})
})

describe('Test target workflow with errors', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'ep3.yaml', 'workflow',
		errorWorkflow,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'eperr3.yaml', 'endpoint',
		endpointErrorWorkflow,
	)

	retry50(`should return a workflow run from magic namespace`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/ns/system/endpoint3`,
		)
		expect(req.statusCode).toEqual(500)
		expect(req.body.error).toEqual({
			endpointFile: '/eperr3.yaml',
			message: 'errCode: badinput, errMessage: Missing or invalid value for required input., instanceId: ',
		})
	})
})

describe('Test target workflow plugin', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, limitedNamespace)
	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'workflow.yaml', 'workflow',
		workflow,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		limitedNamespace,
		'/', 'workflow.yaml', 'workflow',
		workflow,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		limitedNamespace,
		'/', 'endpoint1.yaml', 'endpoint',
		endpointWorkflow,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		limitedNamespace,
		'/', 'endpoint2.yaml', 'endpoint',
		endpointWorkflowAllowed,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'endpoint1.yaml', 'endpoint',
		endpointWorkflow,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'endpoint2.yaml', 'endpoint',
		endpointWorkflowAllowed,
	)

	retry50(`should return a workflow run from magic namespace`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/ns/system/endpoint1`,
		)
		expect(req.statusCode).toEqual(200)
		expect(req.text).toEqual('{"result":"Hello world!"}')
	})

	retry10(`should return a flow run from magic namespace with namespace set`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/ns/system/endpoint2`,
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

	retry10(`should not return a workflow in non-magic namespace`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/ns/` + limitedNamespace + `/endpoint1`,
		)
		expect(req.statusCode).toEqual(403)
	})
})

describe('Test POST method for target workflow plugin', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, limitedNamespace)
	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'workflow.yaml', 'workflow',
		workflowEcho,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'endpoint1.yaml', 'endpoint',
		endpointPOSTWorkflow,
	)

	retry50(`should return a workflow run from magic namespace`, async () => {
		const req = await request(common.config.getDirektivHost()).post(
			`/ns/system/endpoint1`,
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

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'workflow.yaml', 'workflow',
		workflowEcho,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'endpoint1.yaml', 'endpoint',
		endpointComplexPOSTWorkflow,
	)

	retry50(`should return a workflow run from magic namespace`, async () => {
		const req = await request(common.config.getDirektivHost()).post(
			`/ns/system/endpoint1`,
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

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		limitedNamespace,
		'/', 'workflow.yaml', 'workflow',
		workflow,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'workflow.yaml', 'workflow',
		workflowNotToBetriggered,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'endpoint7.yaml', 'endpoint',
		endpointTargetLimitedNamespaceWorkflow,
	)

	retry10(`should return a workflow run from limited namespace`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/ns/system/endpoint1`,
		)
		expect(req.statusCode).toEqual(200)
		expect(req.text).toEqual('{"result":"Hello world!"}')
	})
})

describe('Test target workflow default contenttype', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'epnoct.yaml', 'endpoint',
		endpointNoContentType,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'epct.yaml', 'endpoint',
		endpointContentType,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'contentType.yaml', 'workflow',
		contentType,
	)

	retry10(`should return a json content type`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/ns/system/endpointct`,
		)

		expect(req.headers['content-type']).toEqual('application/json')
		expect(req.statusCode).toEqual(200)
	})

	retry10(`should return a configured content type`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/ns/system/endpointcttest`,
		)

		expect(req.headers['content-type']).toEqual('test/me')
		expect(req.statusCode).toEqual(200)
	})
})
