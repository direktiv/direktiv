import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const namespaceName = 'catchtest'

describe('Test catch state behaviour', () => {
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
		btoa(`states:
- id: a
  type: error
  error: testcode
  message: 'this is a test error'
  transform: 
    result: x
`))
	helpers.itShouldCreateFileV2(it, expect, namespaceName,
		'',
		'catch.yaml',
		'workflow',
		'text/plain',
		btoa(`
functions:
- id: child
  type: subflow
  workflow: '/error.yaml'
states:
- id: a
  type: action
  action:
    function: child
  catch:
  - error: 'testcode'
`))

	it(`should invoke the '/catch.yaml' workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/tree/catch.yaml?op=wait`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			error: {
				code: 'testcode',
				msg: 'this is a test error',
			},
		})
	})
})
