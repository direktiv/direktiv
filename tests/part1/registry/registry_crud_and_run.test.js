import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../../common'
import request from '../../common/request'
import regex from '../../common/regex'

const testNamespace = 'test-services'

describe('Test services crud operations', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	it(`should create a registry`, async () => {
		const res = await request(common.config.getDirektivHost())
			.post(`/api/v2/namespaces/${testNamespace}/registries`)
			.send({
				url: 'docker.io',
				user: 'me',
				password: 'secret',
			})
		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual({
			data: {
				createdAt: expect.stringMatching(regex.timestampRegex),
				namespace: 'test-services',
				id: 'secret-c163796084d652e67cb0',
				url: 'docker.io',
				user: 'me',
			},
		})
	})

	it(`should create a registry`, async () => {
		const res = await request(common.config.getDirektivHost())
			.post(`/api/v2/namespaces/${testNamespace}/registries`)
			.send({
				url: 'docker2.io',
				user: 'me2',
				password: 'secret2',
			})
		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual({
			data: {
				createdAt: expect.stringMatching(regex.timestampRegex),
				namespace: 'test-services',
				id: 'secret-7a95ae8578ed80f27403',
				url: 'docker2.io',
				user: 'me2',
			},
		})
	})

	it(`should list all registries`, async () => {
		const listRes = await request(common.config.getDirektivHost())
			.get(`/api/v2/namespaces/${testNamespace}/registries`)
		expect(listRes.statusCode).toEqual(200)
		expect(listRes.body.data.length).toEqual(2)
		const sortedResponseData = listRes.body.data.sort((a, b) => a.id.localeCompare(b.id));

		const expectedData = [
			{
				createdAt: expect.stringMatching(regex.timestampRegex),
				namespace: 'test-services',
				id: 'secret-c163796084d652e67cb0',
				url: 'docker.io',
				user: 'me',
			},
			{
				createdAt: expect.stringMatching(regex.timestampRegex),
				namespace: 'test-services',
				id: 'secret-7a95ae8578ed80f27403',
				url: 'docker2.io',
				user: 'me2',
			}
		].sort((a, b) => a.id.localeCompare(b.id));

		expect(sortedResponseData).toEqual(expectedData);
	})

	it(`should delete a registry`, async () => {
		const res = await request(common.config.getDirektivHost())
			.delete(`/api/v2/namespaces/${testNamespace}/registries/secret-c163796084d652e67cb0`)
		expect(res.statusCode).toEqual(200)
	})

	it(`should list all registries after delete`, async () => {
		const listRes = await request(common.config.getDirektivHost())
			.get(`/api/v2/namespaces/${testNamespace}/registries`)
		expect(listRes.statusCode).toEqual(200)
		expect(listRes.body.data.length).toEqual(1)
	})
})
