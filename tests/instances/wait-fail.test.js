import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const namespaceName = 'waitfailtest'

describe('Test wait fail API behaviour', () => {
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
		'error.yaml',
		'workflow',
		'text/plain',
		btoa(`
states:
- id: a
  type: error
  error: errA
  message: "error A"
  transform:
    result: x`))

	it(`should invoke the 'error.yaml' workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespaceName }/instances?path=error.yaml&wait=true`)

		expect(req.statusCode).toEqual(500)
        expect(req.header['direktiv-instance-error-code']).toEqual("errA")
        expect(req.header['direktiv-instance-error-message']).toEqual("error A")
		expect(req.body).toMatchObject({
            error: {
                code: 'errA',
                message: 'error A'
            }
		})
	})

	it(`should invoke the '/error.yaml' workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespaceName }/instances?path=%2Ferror.yaml&wait=true`)

		expect(req.statusCode).toEqual(500)
        expect(req.header['direktiv-instance-error-code']).toEqual("errA")
        expect(req.header['direktiv-instance-error-message']).toEqual("error A")
		expect(req.body).toMatchObject({
            error: {
                code: 'errA',
                message: 'error A'
            }
		})
	})
})
