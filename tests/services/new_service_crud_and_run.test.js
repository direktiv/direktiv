import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import request from '../common/request'
import { retry10, retry50 } from '../common/retry'

const testNamespace = 'test-services'

describe('Test services crud operations', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFileV2(it, expect, testNamespace,
		'/','s1.yaml', 'service', `
direktiv_api: service/v1
image: direktiv/request
scale: 1
`)

	common.helpers.itShouldCreateYamlFileV2(it, expect, testNamespace,
		'/','s2.yaml', 'service', `
direktiv_api: service/v1
image: direktiv/request
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
					error: null,
					filePath: '/s1.yaml',
					id: 'test-services-s1-yaml-466337cb33',
					image: 'direktiv/request',
					namespace: 'test-services',
					scale: 1,
					size: 'medium',
					type: 'namespace-service',
				},
				{
					error: null,
					filePath: '/s2.yaml',
					id: 'test-services-s2-yaml-d396514862',
					image: 'direktiv/request',
					namespace: 'test-services',
					scale: 2,
					size: 'medium',
					type: 'namespace-service',
				},
			],
		})
	})

	retry50(`should list all service pods`, async () => {
		let sID = listRes.body.data[0].id
		let res = await request(common.config.getDirektivHost())
			.get(`/api/v2/namespaces/${ testNamespace }/services/${ sID }/pods`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			data: [
				{ id: expect.stringMatching(`^${ sID }(_|-)`) },
			],
		})

		sID = listRes.body.data[1].id
		res = await request(common.config.getDirektivHost())
			.get(`/api/v2/namespaces/${ testNamespace }/services/${ sID }/pods`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			data: [
				{ id: expect.stringMatching(`^${ sID }(_|-)`) },
				{ id: expect.stringMatching(`^${ sID }(_|-)`) },
			],
		})
	})

	retry10(`should list all services`, async () => {
		const res = await request(common.config.getDirektivHost())
			.get(`/api/v2/namespaces/${ testNamespace }/services`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			data: [
				{
					conditions: expect.arrayContaining([ expect.anything() ]),
					error: null,
					filePath: '/s1.yaml',
					id: 'test-services-s1-yaml-466337cb33',
					image: 'direktiv/request',
					namespace: 'test-services',
					scale: 1,
					size: 'medium',
					type: 'namespace-service',
				},
				{
					conditions: expect.arrayContaining([ expect.anything() ]),
					error: null,
					filePath: '/s2.yaml',
					id: 'test-services-s2-yaml-d396514862',
					image: 'direktiv/request',
					namespace: 'test-services',
					scale: 2,
					size: 'medium',
					type: 'namespace-service',
				},
			],
		})
	})

	it(`should rebuild all services`, async () => {
		let sID = listRes.body.data[0].id
		let res = await request(common.config.getDirektivHost())
			.post(`/api/v2/namespaces/${ testNamespace }/services/${ sID }/actions/rebuild`)
			.send()
		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual('')

		sID = listRes.body.data[1].id
		res = await request(common.config.getDirektivHost())
			.post(`/api/v2/namespaces/${ testNamespace }/services/${ sID }/actions/rebuild`)
			.send()
		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual('')
	})
})
