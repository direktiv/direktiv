import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import request from '../common/request'
import { retry10 } from '../common/retry'

const testNamespace = 'system'

const limitedNamespace = 'limited_namespace'

const endpointNSFile = `
x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/endpoint1"
    allow_anonymous: true
    plugins:
      target:
        type: target-namespace-file
        configuration:
          namespace: ` + testNamespace + `
          file: /endpoint1.yaml
get:
   responses:
      "200":
        description: works
`

const endpointNSFileAllowed = `
x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/endpoint2"
    allow_anonymous: true
    plugins:
      target:
        type: target-namespace-file
        configuration:
          file: /endpoint1.yaml
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
        type: something-wrong
get:
   responses:
      "200":
        description: works`

const mimetypeSet = `
x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/endpoint-mimetype"
    allow_anonymous: true
    plugins:
      target:
        type: target-namespace-file
        configuration:
          file: /mimetype.yaml
          content_type: application/whatever
get:
   responses:
      "200":
        description: works`


const mimetypeNotSet = `
x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/endpoint-no-mimetype"
    allow_anonymous: true
    plugins:
      target:
        type: target-namespace-file
        configuration:
          file: /mimetype.yaml
get:
   responses:
      "200":
        description: works`

describe('Test target file wrong config', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'ep3.yaml', 'endpoint',
		endpointBroken,
	)

	retry10(`should fail with wrong config`, async () => {
		const listRes = await request(common.config.getDirektivHost()).get(
			`/api/v2/namespaces/${ testNamespace }/gateway/routes`,
		)
		expect(listRes.statusCode).toEqual(200)
		expect(listRes.body.data.length).toEqual(1)
		expect(listRes.body.data[0]).toEqual({
			spec: expect.anything(),
			file_path: '/ep3.yaml',
			server_path: '/ns/system/endpoint3',
			errors: [ "plugin 'something-wrong' err: doesn't exist" ],
			warnings: [],
		})
	})
})

describe('Test mimetype for file target', () => {
	beforeAll(common.helpers.deleteAllNamespaces)
	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'mimetype.yaml', 'endpoint',
		mimetypeSet,
	)

	retry10(`should return a configured mimetype`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/ns/system/endpoint-mimetype`,
		)
		expect(req.headers['content-type']).toEqual('application/whatever')
	})

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'no-mimetype.yaml', 'endpoint',
		mimetypeNotSet,
	)

	retry10(`should return a guess mimetype (yaml)`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/ns/system/endpoint-no-mimetype`,
		)
		expect(req.headers['content-type']).toEqual('application/yaml')
	})
})

describe('Test target namespace file plugin', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, limitedNamespace)
	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'endpoint1.yaml', 'endpoint',
		endpointNSFile,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'endpoint2.yaml', 'endpoint',
		endpointNSFileAllowed,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		limitedNamespace,
		'/', 'endpoint1.yaml', 'endpoint',
		endpointNSFile,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		limitedNamespace,
		'/', 'endpoint2.yaml', 'endpoint',
		endpointNSFileAllowed,
	)

	retry10(`should return a file from magic namespace`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/ns/system/endpoint1`,
		)
		expect(req.statusCode).toEqual(200)
		expect(req.text).toEqual(endpointNSFile)
	})

	retry10(`should return a file from magic namespace without namespace set`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/ns/system/endpoint2`,
		)
		expect(req.statusCode).toEqual(200)
		expect(req.text).toEqual(endpointNSFile)
	})

	retry10(`should return a file from non-magic namespace`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/ns/` + limitedNamespace + `/endpoint2`,
		)
		expect(req.statusCode).toEqual(200)
		expect(req.text).toEqual(endpointNSFile)
	})

	retry10(`should not return a file across namespaces`, async () => {
		const req = await request(common.config.getDirektivHost()).get(
			`/ns/` + limitedNamespace + `/endpoint1`,
		)
		expect(req.statusCode).toEqual(403)
	})
})
