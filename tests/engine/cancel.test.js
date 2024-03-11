import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const namespaceName = 'canceltest'

describe('Test cancel state behaviour', () => {
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
		'cancel.yaml',
		'workflow',
		'text/plain',
		btoa(`
states:
- id: a
  type: delay
  duration: PT5S
  transform:
    result: x`))

	it(`should invoke the '/cancel.yaml' workflow`, async () => {
		const xreq = await request(common.config.getDirektivHost()).post(`/api/namespaces/${ namespaceName }/tree/cancel.yaml?op=execute`)
		expect(xreq.statusCode).toEqual(200)

		const instanceID = xreq.body.instance

		await helpers.sleep(50)

		const creq = await request(common.config.getDirektivHost()).post(`/api/namespaces/${ namespaceName }/instances/${ instanceID }/cancel`)
		expect(creq.statusCode).toEqual(200)

		await helpers.sleep(50)

		const ireq = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/instances/${ instanceID }`)
		expect(ireq.statusCode).toEqual(200)
		expect(ireq.body.instance.status).toEqual('failed')
		expect(ireq.body.instance.errorCode).toEqual('direktiv.cancels.api')
		expect(ireq.body.instance.errorMessage).toEqual('cancelled by api request')
	})

	helpers.itShouldCreateFileV2(it, expect, namespaceName,
		'',
		'handle-cancel.yaml',
		'workflow',
		'text/plain',
		btoa(`states:
- id: a
  type: delay
  duration: PT5S
  catch:
  - error: 'direktiv.cancels.api'
  transform:
    result: x`))

	it(`should invoke the '/handle-cancel.yaml' workflow`, async () => {
		const xreq = await request(common.config.getDirektivHost()).post(`/api/namespaces/${ namespaceName }/tree/handle-cancel.yaml?op=execute`)
		expect(xreq.statusCode).toEqual(200)

		const instanceID = xreq.body.instance

		await helpers.sleep(50)

		const creq = await request(common.config.getDirektivHost()).post(`/api/namespaces/${ namespaceName }/instances/${ instanceID }/cancel`)
		expect(creq.statusCode).toEqual(200)

		await helpers.sleep(50)

		const ireq = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/instances/${ instanceID }`)
		expect(ireq.statusCode).toEqual(200)
		expect(ireq.body.instance.status).toEqual('failed')
		expect(ireq.body.instance.errorCode).toEqual('direktiv.cancels.api')
		expect(ireq.body.instance.errorMessage).toEqual('cancelled by api request')
	})

	// TODO: test that a parent can catch a child that was cancelled
	// TODO: test that cancelling a parent recurses down to all children
})
