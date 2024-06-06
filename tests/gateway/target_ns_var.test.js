import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import request from '../common/request'
import { retry10 } from '../common/retry'

const testNamespace = 'system'

const limitedNamespace = 'limited_namespace'

const endpointNSVar = `
  direktiv_api: endpoint/v1
  allow_anonymous: true
  plugins:
    target:
      type: target-namespace-var
      configuration:
          namespace: ` + testNamespace + `
          variable: plain
  methods: 
    - GET
  path: endpoint1`

const endpointNSVarAllowed = `
  direktiv_api: endpoint/v1
  allow_anonymous: true
  plugins:
    target:
      type: target-namespace-var
      configuration:
          namespace: ` + limitedNamespace + `
          variable: plain
          content_type: text/test
  methods: 
    - GET
  path: /endpoint2`

const endpointNSVarBroken = `
  direktiv_api: endpoint/v1
  allow_anonymous: true
  plugins:
    target:
      type: target-namespace-var
  methods: 
    - GET
  path: ep3`

describe('Test target workflow var wrong config', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'ep3.yaml', 'endpoint',
		endpointNSVarBroken,
	)

	retry10(`should list all services`, async () => {
		const listRes = await request(common.config.getDirektivHost()).get(
			`/api/v2/namespaces/${ testNamespace }/gateway/routes`,
		)
		expect(listRes.statusCode).toEqual(200)
		expect(listRes.body.data.length).toEqual(1)
		expect(listRes.body.data[0]).toEqual({
			file_path: '/ep3.yaml',
			path: 'ep3',
			methods: [ 'GET' ],
			allow_anonymous: true,
			timeout: 0,
			server_path: '/ns/system/ep3',
			errors: [ "plugin 'target-namespace-var' err: variable required" ],
			warnings: [],
			plugins: { target: { type: 'target-namespace-var' } },
		})
	})
})

describe('Test target namespace variable plugin', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, limitedNamespace)
	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	it(`should set plain text variable`, async () => {
		const workflowVarResponse = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ testNamespace }/variables`)
			.send({
				name: 'plain',
				data: btoa('Hello World'),
				mimeType: 'text/plain',
			})
		expect(workflowVarResponse.statusCode).toEqual(200)
	})

	it(`should set plain text variable`, async () => {
		const workflowVarResponse = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ limitedNamespace }/variables`)
			.send({
				name: 'plain',
				data: btoa('Hello World 2'),
				mimeType: 'text/plain',
			})
		expect(workflowVarResponse.statusCode).toEqual(200)
	})

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'endpoint1.yaml', 'endpoint',
		endpointNSVar,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'endpoint2.yaml', 'endpoint',
		endpointNSVarAllowed,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		limitedNamespace,
		'/', 'endpoint1.yaml', 'endpoint',
		endpointNSVar,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		limitedNamespace,
		'/', 'endpoint2.yaml', 'endpoint',
		endpointNSVarAllowed,
	)

	retry10(`should return a ns var from magic namespace`, async () => {
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

	retry10(`should return a var from non-magic namespace`, async () => {
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
