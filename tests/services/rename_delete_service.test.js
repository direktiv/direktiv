import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import request from '../common/request'
import { retry10 } from '../common/retry'

const testNamespace = 'test-services'

describe('Test renaming services operations', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateFile(it, expect, testNamespace,
		'/s1.yaml', `
direktiv_api: service/v1
image: redis
cmd: redis-server
scale: 1
`)

	common.helpers.itShouldCreateFile(it, expect, testNamespace,
		'/s2.yaml', `
direktiv_api: service/v1
image: redis
cmd: redis-server
scale: 2
`)

	common.helpers.itShouldCreateFile(it, expect, testNamespace,
		'/w1.yaml', `
description: something
functions:
- id: get
  image: direktiv/request:v4
  type: knative-workflow
states:
- id: foo
  type: noop
`)

	common.helpers.itShouldCreateFile(it, expect, testNamespace,
		'/w2.yaml', `
description: something
functions:
- id: get
  image: direktiv/request:v4
  type: knative-workflow
states:
- id: foo
  type: noop
`)

	let listRes
	retry10(`should list all services`, async () => {
		listRes = await request(common.config.getDirektivHost())
			.get(`/api/v2/namespaces/${ testNamespace }/services`)
		expect(listRes.statusCode).toEqual(200)
		expect(listRes.body).toMatchObject({
			data: [
				{
					type: 'namespace-service',
					namespace: 'test-services',
					filePath: '/s1.yaml',
					name: '',
					image: 'redis',
					error: null,
					id: 'test-services-s1-yaml-466337cb33',
				},
				{
					type: 'namespace-service',
					namespace: 'test-services',
					filePath: '/s2.yaml',
					name: '',
					image: 'redis',
					error: null,
					id: 'test-services-s2-yaml-d396514862',
				},
				{
					type: 'workflow-service',
					namespace: 'test-services',
					filePath: '/w1.yaml',
					name: 'get',
					image: 'direktiv/request:v4',
					error: null,
					id: 'test-services-get-w1-yaml-e39f2311c0',
				},
				{
					type: 'workflow-service',
					namespace: 'test-services',
					filePath: '/w2.yaml',
					name: 'get',
					image: 'direktiv/request:v4',
					error: null,
					id: 'test-services-get-w2-yaml-9cca18d982',
				},

			],
		})
	})

	common.helpers.itShouldRenameFile(it, expect, testNamespace, '/s2.yaml', '/s3.yaml')
	common.helpers.itShouldRenameFile(it, expect, testNamespace, '/w2.yaml', '/w3.yaml')

	retry10(`should list all services`, async () => {
		listRes = await request(common.config.getDirektivHost())
			.get(`/api/v2/namespaces/${ testNamespace }/services`)
		expect(listRes.statusCode).toEqual(200)
		expect(listRes.body).toMatchObject({
			data: [
				{
					type: 'namespace-service',
					namespace: 'test-services',
					filePath: '/s1.yaml',
					name: '',
					image: 'redis',
					error: null,
					id: 'test-services-s1-yaml-466337cb33',
				},
				{
					type: 'namespace-service',
					namespace: 'test-services',
					filePath: '/s3.yaml',
					name: '',
					image: 'redis',
					error: null,
					id: 'test-services-s3-yaml-a8af2622f0',
				},
				{
					type: 'workflow-service',
					namespace: 'test-services',
					filePath: '/w1.yaml',
					name: 'get',
					image: 'direktiv/request:v4',
					error: null,
					id: 'test-services-get-w1-yaml-e39f2311c0',
				},
				{
					type: 'workflow-service',
					namespace: 'test-services',
					filePath: '/w3.yaml',
					name: 'get',
					image: 'direktiv/request:v4',
					error: null,
					id: 'test-services-get-w3-yaml-aa1f397b0a',
				},

			],
		})
	})

	common.helpers.itShouldDeleteFile(it, expect, testNamespace, '/s1.yaml')
	common.helpers.itShouldDeleteFile(it, expect, testNamespace, '/w1.yaml')

	retry10(`should list all services`, async () => {
		listRes = await request(common.config.getDirektivHost())
			.get(`/api/v2/namespaces/${ testNamespace }/services`)
		expect(listRes.statusCode).toEqual(200)
		expect(listRes.body).toMatchObject({
			data: [
				{
					type: 'namespace-service',
					namespace: 'test-services',
					filePath: '/s3.yaml',
					name: '',
					image: 'redis',
					error: null,
					id: 'test-services-s3-yaml-a8af2622f0',
				},
				{
					type: 'workflow-service',
					namespace: 'test-services',
					filePath: '/w3.yaml',
					name: 'get',
					image: 'direktiv/request:v4',
					error: null,
					id: 'test-services-get-w3-yaml-aa1f397b0a',
				},

			],
		})
	})
})
