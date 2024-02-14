import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'
import request from 'supertest'

import common from '../common'
import config from '../common/config'
import helpers from '../common/helpers'
import regex from '../common/regex'

const namespace = basename(__filename)

describe('Test variable delete calls', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	let createRes
	it(`should create a new variable`, async () => {
		createRes = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces/${ namespace }/variables`)
			.send({
				name: 'foo',
				data: btoa("bar"),
				mimeType: 'mime',
			})
		expect(createRes.statusCode).toEqual(200)
	})

	it(`should delete a variable`, async () => {
		const varId = createRes.body.data.id
		const res = await request(config.getDirektivHost())
			.delete(`/api/v2/namespaces/${ namespace }/variables/${varId}`)
		expect(res.statusCode).toEqual(200)
	})
})

describe('Test invalid variable delete calls', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	const testCases = [
		{
			// invalid id.
			id: "12345",
			wantError: {
				statusCode: 400,
				error: {
					code: 'request_data_invalid',
					message: "variable id is invalid uuid string",
				},
			},
		},
		{
			// none existent id.
			id: "cb673820-0d1d-43c9-9fa5-dce177ee42b1",
			wantError: {
				statusCode: 404,
				error: {
					code: 'resource_not_found',
					message: "requested resource is not found",
				},
			},
		},
	]

	for (let i = 0; i < testCases.length; i++) {
		const testCase = testCases[i]

		it(`should fail create a new variable case ${ i }`, async () => {
			const res = await request(config.getDirektivHost())
				.delete(`/api/v2/namespaces/${ namespace }/variables/${testCase.id}`)
				.send(testCase.input)
			expect(res.statusCode).toEqual(testCase.wantError.statusCode)
			expect(res.body.error).toMatchObject(
				testCase.wantError.error,
			)
		})
	}
})
