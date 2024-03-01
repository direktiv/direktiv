import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import request from '../common/request'

const testNamespace = 'test-file-namespace'

describe('Test secret crud operations', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	it(`should create a new namespace`, async () => {
		const res = await request(common.config.getDirektivHost()).put(`/api/namespaces/${ testNamespace }`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			namespace: {
				name: testNamespace,
				// regex /^2.*Z$/ matches timestamps like 2023-03-01T14:19:52.383871512Z
				createdAt: expect.stringMatching(/^2.*Z$/),
				updatedAt: expect.stringMatching(/^2.*Z$/),
			},
		})
	})

	it(`should create a new secret`, async () => {
		const res = await request(common.config.getDirektivHost())
			.put(`/api/namespaces/${ testNamespace }/secrets/key1`)
			.set({
				'Content-Type': 'text/plain',
			})

			.send(`value1`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			namespace: testNamespace,
			key: 'key1',
		})
	})

	it(`should create another new secret`, async () => {
		const res = await request(common.config.getDirektivHost())
			.put(`/api/namespaces/${ testNamespace }/secrets/key2`)
			.set({
				'Content-Type': 'text/plain',
			})

			.send(`value2`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			namespace: testNamespace,
			key: 'key2',
		})
	})


	it(`should list all secrets`, async () => {
		const res = await request(common.config.getDirektivHost())
			.get(`/api/namespaces/${ testNamespace }/secrets`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			namespace: testNamespace,
			secrets: {
				pageInfo: null,
				results: [
					{ name: 'key1' },
					{ name: 'key2' },
				],
			},
		})
	})

	it(`should delete a secret`, async () => {
		const res = await request(common.config.getDirektivHost())
			.delete(`/api/namespaces/${ testNamespace }/secrets/key1`)
		expect(res.statusCode).toEqual(200)
	})

	it(`should list one secrets`, async () => {
		const res = await request(common.config.getDirektivHost())
			.get(`/api/namespaces/${ testNamespace }/secrets`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			namespace: testNamespace,
			secrets: {
				pageInfo: null,
				results: [ { name: 'key2' } ],
			},
		})
	})


	it(`should delete the second secret`, async () => {
		const res = await request(common.config.getDirektivHost())
			.delete(`/api/namespaces/${ testNamespace }/secrets/key2`)
		expect(res.statusCode).toEqual(200)
	})

	it(`should list empty`, async () => {
		const res = await request(common.config.getDirektivHost())
			.get(`/api/namespaces/${ testNamespace }/secrets`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			namespace: testNamespace,
			secrets: {
				pageInfo: null,
				results: [],
			},
		})
	})
})
