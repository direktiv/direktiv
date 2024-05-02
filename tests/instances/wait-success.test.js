import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import regex from '../common/regex'
import request from '../common/request'

const namespaceName = 'waitsuccesstest'

describe('Test wait success API behaviour', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespaceName)

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
		const req = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespaceName }/instances?path=noop.yaml&wait=true`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			result: 'x',
		})
	})

	it(`should invoke the '/noop.yaml' workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespaceName }/instances?path=%2Fnoop.yaml&wait=true`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			result: 'x',
		})
	})

	it(`should perform a list request`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespaceName }/instances`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			data: [
				{
					createdAt: expect.stringMatching(regex.timestampRegex),
					endedAt: expect.stringMatching(regex.timestampRegex),
					definition: expect.stringMatching(regex.base64Regex),
					errorCode: null,
					flow: [ 'a' ],
					id: expect.stringMatching(regex.uuidRegex),
					invoker: 'api',
					lineage: [],
					path: '/noop.yaml',
					status: 'complete',
					traceId: expect.anything(),
				},
				{
					createdAt: expect.stringMatching(regex.timestampRegex),
					endedAt: expect.stringMatching(regex.timestampRegex),
					definition: expect.stringMatching(regex.base64Regex),
					errorCode: null,
					flow: [ 'a' ],
					id: expect.stringMatching(regex.uuidRegex),
					invoker: 'api',
					lineage: [],
					path: '/noop.yaml',
					status: 'complete',
					traceId: expect.anything(),
				},
			],
		})
	})
})
