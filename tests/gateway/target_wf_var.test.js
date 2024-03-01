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

const endpointWorkflowVar = `
  direktiv_api: endpoint/v1
  allow_anonymous: true
  plugins:
    target:
      type: target-flow-var
      configuration:
          namespace: ` + testNamespace + `
          flow: /workflow.yaml
          variable: test
  methods: 
    - GET
  path: /endpoint1`

const endpointWorkflowVarAllowed = `
  direktiv_api: endpoint/v1
  allow_anonymous: true
  plugins:
    target:
      type: target-flow-var
      configuration:
          namespace: ` + limitedNamespace + `
          flow: /workflow.yaml
          variable: test
          content_type: text/test
  methods: 
    - GET
  path: endpoint2`

const endpointWorkflkowVarBroken = `
  direktiv_api: endpoint/v1
  allow_anonymous: true
  plugins:
    target:
      type: target-flow-var
  methods: 
    - GET
  path: ep3`

describe('Test target workflow var wrong config', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateFile(
		it,
		expect,
		testNamespace,
		'/ep3.yaml',
		endpointWorkflkowVarBroken,
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
						path: '/ep3',
						methods: [ 'GET' ],
						allow_anonymous: true,
						timeout: 0,
						server_path: '/gw/ep3',
						errors: [ 'flow and variable required' ],
						warnings: [],
						plugins: { target: { type: 'target-flow-var' } },
					},
				],
			),
		)
	})
})

describe('Test target workflow variable plugin', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, limitedNamespace)
	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateFile(
		it,
		expect,
		testNamespace,
		'/workflow.yaml',
		workflow,
	)

	common.helpers.itShouldCreateFile(
		it,
		expect,
		limitedNamespace,
		'/workflow.yaml',
		workflow,
	)

	common.helpers.itShouldCreateFile(
		it,
		expect,
		limitedNamespace,
		'/endpoint1.yaml',
		endpointWorkflowVar,
	)

	common.helpers.itShouldCreateFile(
		it,
		expect,
		limitedNamespace,
		'/endpoint2.yaml',
		endpointWorkflowVarAllowed,
	)

	common.helpers.itShouldCreateFile(
		it,
		expect,
		testNamespace,
		'/endpoint1.yaml',
		endpointWorkflowVar,
	)

	common.helpers.itShouldCreateFile(
		it,
		expect,
		testNamespace,
		'/endpoint2.yaml',
		endpointWorkflowVarAllowed,
	)

	it(`should set plain text variable for worklfow`, async () => {
		const workflowVarResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${ testNamespace }/tree/workflow.yaml?op=set-var&var=test`)
			.set('Content-Type', 'text/plain')
			.send('Hello World')
		expect(workflowVarResponse.statusCode).toEqual(200)
	})


	it(`should set plain text variable for worklfow in limited namespace`, async () => {
		const workflowVarResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${ limitedNamespace }/tree/workflow.yaml?op=set-var&var=test`)
			.set('Content-Type', 'text/plain')
			.send('Hello World 2')
		expect(workflowVarResponse.statusCode).toEqual(200)
	})


	retry10(`should return a workflow var from magic namespace`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/gw/endpoint1`,
		)
		expect(req.statusCode).toEqual(200)
		expect(req.text).toEqual('Hello World')
		expect(req.header['content-type']).toEqual('text/plain')
	})

	retry10(`should return a var from magic namespace with namespace set`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/gw/endpoint2`,
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
		expect(req.statusCode).toEqual(500)
	})
})
