import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import config from '../../common/config'
import helpers from '../../common/helpers'
import request from '../../common/request'
import {fileURLToPath} from "url";

const namespace = basename(fileURLToPath(import.meta.url))

describe('Test variable delete calls', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	it(`should create and delete one var`, async () => {
		const createRes = await request(config.getDirektivBaseUrl())
			.post(`/api/v2/namespaces/${ namespace }/variables`)
			.send({
				name: 'foo',
				data: btoa('bar'),
				mimeType: 'mime',
			})
		expect(createRes.statusCode).toEqual(200)

		const varId = createRes.body.data.id
		const res = await request(config.getDirektivBaseUrl())
			.delete(`/api/v2/namespaces/${ namespace }/variables/${ varId }`)
		expect(res.statusCode).toEqual(200)
	})

	it(`should create two vars and delete them with multiple delete endpoint`, async () => {
		const createRes1 = await request(config.getDirektivBaseUrl())
			.post(`/api/v2/namespaces/${ namespace }/variables`)
			.send({
				name: 'foo',
				data: btoa('bar'),
				mimeType: 'mime',
			})
		expect(createRes1.statusCode).toEqual(200)

		const createRes2 = await request(config.getDirektivBaseUrl())
			.post(`/api/v2/namespaces/${ namespace }/variables`)
			.send({
				name: 'foo1',
				data: btoa('bar'),
				mimeType: 'mime',
			})
		expect(createRes2.statusCode).toEqual(200)

		const varId1 = createRes1.body.data.id
		const varId2 = createRes2.body.data.id

		const res = await request(config.getDirektivBaseUrl())
			.delete(`/api/v2/namespaces/${ namespace }/variables?ids=${ varId1 },${ varId2 }`)
		expect(res.statusCode).toEqual(200)
	})
})

describe('Test invalid variable delete calls', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	const testCases = [
		{
			// invalid id.
			id: '12345',
			wantError: {
				statusCode: 400,
				error: {
					code: 'request_data_invalid',
					message: 'variable id is invalid uuid string',
				},
			},
		},
		{
			// none existent id.
			id: 'cb673820-0d1d-43c9-9fa5-dce177ee42b1',
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

		it(`should fail delete a variable case ${ i }`, async () => {
			const res = await request(config.getDirektivBaseUrl())
				.delete(`/api/v2/namespaces/${ namespace }/variables/${ testCase.id }`)
				.send(testCase.input)
			expect(res.statusCode).toEqual(testCase.wantError.statusCode)
			expect(res.body.error).toEqual(
				testCase.wantError.error,
			)
		})
	}
})
