import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import request from '../common/request'
import { retry10 } from '../common/retry'

const testNamespace = 'test-services'

describe('Test services crud operations', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFileV2(it, expect, testNamespace,
		'/', 's1.yaml', 'service', `
direktiv_api: service/v1
image: redis
cmd: redis-server
scale: 1
envs:
- name: foo1
  value: bar1
- name: foo2
  value: bar2
`)

	common.helpers.itShouldCreateYamlFileV2(it, expect, testNamespace,
		'/', 's2.yaml', 'service', `
direktiv_api: service/v1
image: redis
cmd: redis-server
scale: 2
`)

	let listRes
	retry10(`should list all services`, async () => {
		listRes = await request(common.config.getDirektivHost())
			.get(`/api/v2/namespaces/${ testNamespace }/services`)
		expect(listRes.statusCode).toEqual(200)
		expect(listRes.body).toMatchObject({
			data: [
				{
					cmd: 'redis-server',
					error: null,
					filePath: '/s1.yaml',
					id: 'test-services-s1-yaml-466337cb33',
					image: 'redis',
					namespace: 'test-services',
					scale: 1,
					size: 'medium',
					type: 'namespace-service',
					envs: [
						{ name: 'foo1',
							value: 'bar1' },
						{ name: 'foo2',
							value: 'bar2' },
					],
				},
				{
					cmd: 'redis-server',
					error: null,
					filePath: '/s2.yaml',
					id: 'test-services-s2-yaml-d396514862',
					image: 'redis',
					namespace: 'test-services',
					scale: 2,
					size: 'medium',
					type: 'namespace-service',
					envs: [],
				},
			],
		})
	})
})
