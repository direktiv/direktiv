import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'

const namespace = basename(__filename)

describe('Test namespace delete calls', () => {
	beforeAll(helpers.deleteAllNamespaces)

	let createRes
	it(`should create a new namespace`, async () => {
		createRes = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces`)
			.send({
				name: 'foo',
				data: btoa('bar'),
			})
		expect(createRes.statusCode).toEqual(200)
	})

	it(`should delete a namespace`, async () => {
		const namespaceName = createRes.body.data.name
		const res = await request(config.getDirektivHost())
			.delete(`/api/v2/namespaces/${ namespaceName }`)
		expect(res.statusCode).toEqual(200)
	})
})

describe('Test invalid namespace delete calls', () => {
	beforeAll(helpers.deleteAllNamespaces)

	const testCases = [
		{
			// invalid id.
			name: 'something',
			wantError: {
				statusCode: 404,
				error: {
					code: 'resource_not_found',
					message: 'requested resource is not found',
				},
			},
		},
	]

	for (let i = 0; i < testCases.length; i++) {
		const testCase = testCases[i]

		it(`should fail delete a namespace case ${ i }`, async () => {
			const res = await request(config.getDirektivHost())
				.delete(`/api/v2/namespaces/${ testCase.name }`)
				.send(testCase.input)
			expect(res.statusCode).toEqual(testCase.wantError.statusCode)
			expect(res.body.error).toEqual(
				testCase.wantError.error,
			)
		})
	}
})
