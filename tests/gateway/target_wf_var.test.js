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

const endpointWorkflowVar = `
x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/endpoint1"
    allow_anonymous: true
    plugins:
      target:
        type: target-flow-var
        configuration:
          namespace: ` + testNamespace + `
          flow: /workflow.yaml
          variable: test
get:
   responses:
      "200":
        description: works
`

// const endpointWorkflowVar = `
//   direktiv_api: endpoint/v1
//   allow_anonymous: true
//   plugins:
//     target:
//       type: target-flow-var
//       configuration:
//           namespace: ` + testNamespace + `
//           flow: /workflow.yaml
//           variable: test
//   methods:
//     - GET
//   path: /endpoint1`

const endpointWorkflowVarAllowed = `
x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/endpoint2"
    allow_anonymous: true
    plugins:
      target:
        type: target-flow-var
        configuration:
          namespace: ` + limitedNamespace + `
          flow: /workflow.yaml
          variable: test
          content_type: text/test
get:
   responses:
      "200":
        description: works
`

// const endpointWorkflowVarAllowed = `
//   direktiv_api: endpoint/v1
//   allow_anonymous: true
//   plugins:
//     target:
//       type: target-flow-var
//       configuration:
//           namespace: ` + limitedNamespace + `
//           flow: /workflow.yaml
//           variable: test
//           content_type: text/test
//   methods:
//     - GET
//   path: endpoint2`

const endpointWorkflkowVarBroken = `
x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/ep3"
    allow_anonymous: true
    plugins:
      target:
        type: target-flow-var
get:
   responses:
      "200":
        description: works
`

// const endpointWorkflkowVarBroken = `
//   direktiv_api: endpoint/v1
//   allow_anonymous: true
//   plugins:
//     target:
//       type: target-flow-var
//   methods:
//     - GET
//   path: ep3`

describe('Test target workflow var wrong config', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'ep3.yaml', 'endpoint',
		endpointWorkflkowVarBroken,
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
			server_path: '/ns/system/ep3',
			errors: [ "plugin 'target-flow-var' err: variable required" ],
			warnings: [],
		})
	})
})

describe('Test target workflow variable plugin', () => {
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
		endpointWorkflowVar,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		limitedNamespace,
		'/', 'endpoint2.yaml', 'endpoint',
		endpointWorkflowVarAllowed,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'endpoint1.yaml', 'endpoint',
		endpointWorkflowVar,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'endpoint2.yaml', 'endpoint',
		endpointWorkflowVarAllowed,
	)

	it(`should set plain text variable`, async () => {
		const workflowVarResponse = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ testNamespace }/variables`)
			.send({
				name: 'test',
				workflowPath: '/workflow.yaml',
				data: btoa('Hello World'),
				mimeType: 'text/plain',
			})
		expect(workflowVarResponse.statusCode).toEqual(200)
	})

	it(`should set plain text variable for worklfow in limited namespace`, async () => {
		const workflowVarResponse = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ limitedNamespace }/variables`)
			.send({
				name: 'test',
				workflowPath: '/workflow.yaml',
				data: btoa('Hello World 2'),
				mimeType: 'text/plain',
			})
		expect(workflowVarResponse.statusCode).toEqual(200)
	})

	retry50(`should return a workflow var from magic namespace`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/ns/system/endpoint1`,
		)
		expect(req.statusCode).toEqual(200)
		expect(req.text).toEqual('Hello World')
		expect(req.header['content-type']).toEqual('text/plain')
	})

	retry10(`should return a var from magic namespace with namespace set`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/ns/system/endpoint2`,
		)
		expect(req.statusCode).toEqual(200)
		expect(req.text).toEqual('Hello World 2')
		expect(req.header['content-type']).toEqual('text/test')
	})

	retry10(`should return a workflow var from non-magic namespace`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/ns/` + limitedNamespace + `/endpoint2`,
		)
		expect(req.statusCode).toEqual(200)
		expect(req.text).toEqual('Hello World 2')
	})

	retry10(`should not return a var`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/ns/` + limitedNamespace + `/endpoint1`,
		)
		expect(req.statusCode).toEqual(403)
	})
})
