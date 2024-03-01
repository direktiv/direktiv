import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import request from '../common/request'

const testNamespace = 'test-namespace'


describe('Test namespaces crud operations', () => {
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

	it(`should get the new namespace`, async () => {
		const res = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ testNamespace }/tree`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			namespace: testNamespace,
		})
	})

	it(`should delete the new namespace`, async () => {
		const res = await request(common.config.getDirektivHost()).delete(`/api/namespaces/${ testNamespace }?recursive=true`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({})
	})

	it(`should get 404 after the new namespace deletion`, async () => {
		const res = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ testNamespace }/tree`)
		expect(res.statusCode).toEqual(404)
		expect(res.body).toMatchObject({
			code: 404,
			message: 'ErrNotFound',
		})
	})
})
