import { beforeAll, describe, expect, it } from '@jest/globals'
import { encode } from 'js-base64'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const namespaceName = 'canceltest'

describe('Test cancel state behaviour', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespaceName)

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
		const xreq = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespaceName }/instances?path=cancel.yaml`)
		expect(xreq.statusCode).toEqual(200)

		const instanceID = xreq.body.data.id

		await helpers.sleep(50)

		const creq = await request(common.config.getDirektivHost()).patch(`/api/v2/namespaces/${ namespaceName }/instances/${ instanceID }`)
			.set('Content-Type', 'application/json')
			.send({
				status: 'cancelled',
			})

		expect(creq.statusCode).toEqual(200)

		await helpers.sleep(50)

		const ireq = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespaceName }/instances/${ instanceID }`)
		expect(ireq.statusCode).toEqual(200)
		expect(ireq.body.data.status).toEqual('cancelled')
		expect(ireq.body.data.errorCode).toEqual('direktiv.cancels.api')
		expect(ireq.body.data.errorMessage).toEqual(encode('cancelled by api request'))
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
		const xreq = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespaceName }/instances?path=handle-cancel.yaml`)
		expect(xreq.statusCode).toEqual(200)

		const instanceID = xreq.body.data.id

		await helpers.sleep(50)

		const creq = await request(common.config.getDirektivHost()).patch(`/api/v2/namespaces/${ namespaceName }/instances/${ instanceID }`)
			.set('Content-Type', 'application/json')
			.send({
				status: 'cancelled',
			})

		expect(creq.statusCode).toEqual(200)

		await helpers.sleep(50)

		const ireq = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespaceName }/instances/${ instanceID }`)
		expect(ireq.statusCode).toEqual(200)
		expect(ireq.body.data.status).toEqual('cancelled')
		expect(ireq.body.data.errorCode).toEqual('direktiv.cancels.api')
		expect(ireq.body.data.errorMessage).toEqual(encode('cancelled by api request'))
	})

	// TODO: test that a parent can catch a child that was cancelled
	// TODO: test that cancelling a parent recurses down to all children
})
