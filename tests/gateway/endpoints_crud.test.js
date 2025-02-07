import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import request from '../common/request'
import { retry10 } from '../common/retry'

const testNamespace = 'system'

const endpoint1 = `x-direktiv-api: endpoint/v2
x-direktiv-config:
  path: "/endpoint1"
  allow_anonymous: false
  plugins:
    auth:
    - type: key-auth
      configuration:
         key_name: secret
    target:
      type: instant-response
      configuration:
        status_code: 201
        status_message: "TEST1"
get:
  responses:
    "200":
      description: works`

const endpoint2 = `x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/endpoint2"
    allow_anonymous: true
    plugins:
      auth:
      - type: key-auth
        configuration:
           key_name: secret
      target:
        type: instant-response
        configuration:
          status_code: 202
          status_message: "TEST2"
get:
   responses:
      "200":
          description: works`

const endpoint3 = `x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/endpoint3/longer/path"
    allow_anonymous: true
    plugins:
      auth:
      - type: key-auth
        configuration:
           key_name: secret
      target:
        type: instant-response
        configuration:
          status_code: 201
          status_message: "TEST1"
get:
   responses:
      "200":
        description: works`

const endpoint4 = `x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/endpoint4/longer/path/{id}"
    allow_anonymous: true
    plugins:
      auth:
      - type: key-auth
        configuration:
           key_name: secret
      target:
        type: instant-response
        configuration:
          status_code: 201
          status_message: "TEST1"
get:
   responses:
      "200":
        description: works`

const consumer1 = `
direktiv_api: "consumer/v1"
username: consumer1
password: pwd
api_key: key1
tags:
- tag1
groups:
- group1`

const consumer2 = `
direktiv_api: "consumer/v1"
username: consumer2
password: pwd
api_key: key2
tags:
- tag2
groups:
- group2`

const endpointBroken = `x-direktiv-api: endpoint/v2
x-direktiv-config:
  path: "ep4"
  allow_anonymous: true
  plugins:
    outbound: 
      type: js-outbound
get:
  responses:
    "200":
      description: works`

describe('Test wrong endpoint config', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'endpointbroken.yaml', 'endpoint',
		endpointBroken,
	)

	retry10(`should list all endpoints`, async () => {
		const listRes = await request(common.config.getDirektivHost()).get(
			`/api/v2/namespaces/${ testNamespace }/gateway/routes`,
		)
		expect(listRes.statusCode).toEqual(200)
		expect(listRes.body.data.length).toEqual(1)
		expect(listRes.body.data[0]).toEqual(
			{
				spec: expect.anything(),
				file_path: '/endpointbroken.yaml',
				errors: [ 'yaml: unmarshal errors:\n  line 5: cannot unmarshal !!map into []core.PluginConfig' ],
				warnings: [],
				server_path: '',
			},
		)
	})
})

describe('Test gateway endpoints on create', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	retry10(`should list all endpoints`, async () => {
		const listRes = await request(common.config.getDirektivHost()).get(
			`/api/v2/namespaces/${ testNamespace }/gateway/routes`,
		)
		expect(listRes.statusCode).toEqual(200)
		expect(listRes.body.data.length).toEqual(0)
		expect(listRes.body.data).toEqual(
			expect.arrayContaining(
				[],
			),
		)
	})
})

describe('Test gateway get single endpoint', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'endpoint1.yaml', 'endpoint',
		endpoint1,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'endpoint2.yaml', 'endpoint',
		endpoint2,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'endpoint3.yaml', 'endpoint',
		endpoint3,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'endpoint4.yaml', 'endpoint',
		endpoint4,
	)

	retry10(`should list simple endpoint`, async () => {
		const listRes = await request(common.config.getDirektivHost()).get(
			`/api/v2/namespaces/${ testNamespace }/gateway/routes?path=/endpoint1`,
		)
		expect(listRes.statusCode).toEqual(200)
		expect(listRes.body.data.length).toEqual(1)
		expect(listRes.body.data[0]).toEqual(expect.objectContaining({
			spec: expect.objectContaining({
				get: expect.anything(),
				"x-direktiv-config": {
					allow_anonymous: false,
					path: "/endpoint1",
					plugins: {
						auth: [
						{
							configuration: {
								key_name: "secret"
							},
							type: "key-auth"
							}
						],
						target: {
							configuration: {
								status_code: 201,
								status_message: "TEST1"
							},
							type: "instant-response"
						}
					}
				}
			}),
			errors: [],
			warnings: [],
			server_path: '/ns/system/endpoint1',
			file_path: '/endpoint1.yaml',
		}))
	})

	retry10(`should list long path endpoint`, async () => {
		const listRes = await request(common.config.getDirektivHost()).get(
			`/api/v2/namespaces/${ testNamespace }/gateway/routes?path=/endpoint3/longer/path`,
		)
		expect(listRes.statusCode).toEqual(200)
		expect(listRes.body.data.length).toEqual(1)
		expect(listRes.body.data[0].spec["x-direktiv-config"].path).toEqual("/endpoint3/longer/path")
	})

	retry10(`should list long path endpoint with var`, async () => {
		const listRes = await request(common.config.getDirektivHost()).get(
			`/api/v2/namespaces/${ testNamespace }/gateway/routes?path=/endpoint4/longer/path/{id}`,
		)
		expect(listRes.statusCode).toEqual(200)
		expect(listRes.body.data.length).toEqual(1)
		expect(listRes.body.data[0].spec["x-direktiv-config"].path).toEqual("/endpoint4/longer/path/{id}")
	})
})

describe('Test gateway endpoints crud operations', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'endpoint1.yaml', 'endpoint',
		endpoint1,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'endpoint2.yaml', 'endpoint',
		endpoint2,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'consumer1.yaml', 'consumer',
		consumer1,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'consumer2.yaml', 'consumer',
		consumer2,
	)

	retry10(`should list all endpoints`, async () => {
		const listRes = await request(common.config.getDirektivHost()).get(
			`/api/v2/namespaces/${ testNamespace }/gateway/routes`,
		)
		expect(listRes.statusCode).toEqual(200)
		expect(listRes.body.data.length).toEqual(2)
		expect(listRes.body.data).toEqual(
			[
				{
					spec: expect.anything(),
					errors: [],
					warnings: [],
					server_path: '/ns/system/endpoint1',
					file_path: '/endpoint1.yaml',
				}, {
					spec: expect.anything(),
					errors: [],
					warnings: [],
					server_path: '/ns/system/endpoint2',
					file_path: '/endpoint2.yaml',
				},
			],
		)
	})

	retry10(`should list all consumers`, async () => {
		const listRes = await request(common.config.getDirektivHost()).get(
			`/api/v2/namespaces/${ testNamespace }/gateway/consumers`,
		)
		expect(listRes.statusCode).toEqual(200)
		expect(listRes.body.data.length).toEqual(2)

		const comp = (a, b) => a.file_path < b.file_path ? -1 : 1
		expect(listRes.body.data.sort(comp)).toEqual([
			{
				file_path: '/consumer2.yaml',
				errors: [],
				api_key: 'key2',
				groups: [ 'group2' ],
				password: 'pwd',
				tags: [ 'tag2' ],
				username: 'consumer2',
			},

			{
				file_path: '/consumer1.yaml',
				errors: [],
				api_key: 'key1',
				groups: [ 'group1' ],
				password: 'pwd',
				tags: [ 'tag1' ],
				username: 'consumer1',
			},
		].sort(comp))
	})

	common.helpers.itShouldDeleteFile(it, expect, testNamespace, '/endpoint1.yaml')
	common.helpers.itShouldDeleteFile(it, expect, testNamespace, '/consumer1.yaml')

	retry10(`should list one route after delete`, async () => {
		const listRes = await request(common.config.getDirektivHost()).get(
			`/api/v2/namespaces/${ testNamespace }/gateway/routes`,
		)
		expect(listRes.statusCode).toEqual(200)
		expect(listRes.body.data.length).toEqual(1)
	})

	retry10(`should list one consumer after delete`, async () => {
		const listRes = await request(common.config.getDirektivHost()).get(
			`/api/v2/namespaces/${ testNamespace }/gateway/consumers`,
		)
		expect(listRes.statusCode).toEqual(200)
		expect(listRes.body.data.length).toEqual(1)
	})
})

describe('Test availability of gateway endpoints', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'endpoint1.yaml', 'endpoint',
		endpoint1,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'endpoint2.yaml', 'endpoint',
		endpoint2,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'consumer1.yaml', 'consumer',
		consumer1,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'consumer2.yaml', 'consumer',
		consumer2,
	)

	retry10(`should not run endpoint without authentication`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/ns/system/endpoint1`,
		)
		expect(req.statusCode).toEqual(403)
	})

	retry10(`should run endpoint without authentication but allow anonymous`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/ns/system/endpoint2`,
		)
		expect(req.statusCode).toEqual(202)
	})

	retry10(`should run endpoint with key authentication`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/ns/system/endpoint1`,
		)
			.set('secret', 'key2')
		expect(req.statusCode).toEqual(201)
	})

	retry10(`should run endpoint with basic authentication`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/ns/system/endpoint2`,
		)
			.auth('consumer1', 'pwd')
		expect(req.statusCode).toEqual(202)
	})
})
