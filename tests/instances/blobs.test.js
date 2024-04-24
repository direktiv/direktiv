import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import regex from '../common/regex'
import helpers from '../common/helpers'
import request from '../common/request'

const namespaceName = 'blobstest'

let id = ''

describe('Test wait success API behaviour', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	it(`should create a namespace`, async () => {
		const req = await request(common.config.getDirektivHost()).put(`/api/namespaces/${ namespaceName }`)

		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			namespace: {
				createdAt: expect.stringMatching(common.regex.timestampRegex),
				updatedAt: expect.stringMatching(common.regex.timestampRegex),
				name: namespaceName,
			},
		})
	})

	helpers.itShouldCreateFileV2(it, expect, namespaceName,
		'',
		'noop.yaml',
		'workflow',
		'text/plain',
		btoa(`
states:
- id: a
  type: noop
  transform:
    result: x`))

	it(`should invoke the 'noop.yaml' workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespaceName }/instances?path=noop.yaml`)
		.send({"a": 2})
		expect(req.statusCode).toEqual(200)

		id = req.body.data.id

		await sleep(200)
	})

	it(`should get the instance's input data`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespaceName }/instances/${ id }/input`)

		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			data: {
				input: "eyJhIjoyfQ==",
				createdAt: expect.stringMatching(regex.timestampRegex),
				endedAt: expect.stringMatching(regex.timestampRegex),
				definition: expect.stringMatching(regex.base64Regex),
				errorCode: "", 
				flow: [ "a" ],
				id: expect.stringMatching(regex.uuidRegex), 
				invoker: "api",
				lineage: [],
				path: "/noop.yaml",
				status: "complete", 
				traceId: expect.anything(),
			},
		})
	})

	it(`should get the instance's output data`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespaceName }/instances/${ id }/output`)

		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			data: {
				output: 'eyJyZXN1bHQiOiJ4In0=',
				createdAt: expect.stringMatching(regex.timestampRegex),
				endedAt: expect.stringMatching(regex.timestampRegex),
				definition: expect.stringMatching(regex.base64Regex),
				errorCode: "", 
				flow: [ "a" ],
				id: expect.stringMatching(regex.uuidRegex), 
				invoker: "api",
				lineage: [],
				path: "/noop.yaml",
				status: "complete", 
				traceId: expect.anything(),
			},
		})
	})
})

function sleep (ms) {
	return new Promise(resolve => setTimeout(resolve, ms))
}
