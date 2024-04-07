import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const namespaceName = 'waitsuccesstest'

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
})
