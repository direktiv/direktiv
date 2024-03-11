import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const namespaceName = 'errorstatetest'

describe('Test error state behaviour', () => {
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
  error: testcode
  message: 'this is a test error'
  transform: 
    result: x
`))

	it(`should invoke the '/error.yaml' workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/tree/error.yaml?op=wait`)

		expect(req.statusCode).toEqual(500)
		expect(req.headers['direktiv-instance-error-code']).toEqual('testcode')
		expect(req.headers['direktiv-instance-error-message']).toEqual('this is a test error')
		expect(req.body).toMatchObject({})
	})

	helpers.itShouldCreateFileV2(it, expect, namespaceName,
		'',
		'caller.yaml',
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
`))

	it(`should invoke the '/caller.yaml' workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/tree/caller.yaml?op=wait`)

		expect(req.statusCode).toEqual(500)
		expect(req.headers['direktiv-instance-error-code']).toEqual('testcode')
		expect(req.headers['direktiv-instance-error-message']).toEqual('this is a test error')
		expect(req.body).toMatchObject({})
	})

	helpers.itShouldCreateFileV2(it, expect, namespaceName,
		'',
		'error-and-continue.yaml',
		'workflow',
		'text/plain',
		btoa(`
states:
- id: a
  type: error
  error: testcode
  message: 'this is a test error'
  transition: b
- id: b
  type: noop
`))

	it(`should invoke the '/error-and-continue.yaml' workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/tree/error-and-continue.yaml?op=wait`)

		expect(req.statusCode).toEqual(500)
		expect(req.headers['direktiv-instance-error-code']).toEqual('testcode')
		expect(req.headers['direktiv-instance-error-message']).toEqual('this is a test error')
		expect(req.body).toMatchObject({})
	})

	helpers.itShouldCreateFileV2(it, expect, namespaceName,
		'',
		'double-error.yaml',
		'workflow',
		'text/plain',
		btoa(`
states:
- id: a
  type: error
  error: testcode
  message: 'this is a test error'
  transition: b
- id: b
  type: error
  error: testcode2
  message: 'this is a test error 2'
`))
})
