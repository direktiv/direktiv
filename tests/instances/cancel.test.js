import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const namespaceName = 'canceltest'

let id = ''

describe('Test wait success API behaviour', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespaceName)

	helpers.itShouldCreateFileV2(it, expect, namespaceName,
		'',
		'delay.yaml',
		'workflow',
		'text/plain',
		btoa(`
states:
- id: a
  type: delay
  duration: 'PT10S'
  transform:
    result: x`))

	it(`should invoke the 'delay.yaml' workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespaceName }/instances?path=delay.yaml`)
		.send({
			name: 'foo',
			data: btoa('bar'),
		})
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			data: {
				createdAt: expect.stringMatching(common.regex.timestampRegex),
				definition: expect.stringMatching(common.regex.base64Regex),
				id: expect.stringMatching(common.regex.uuidRegex),
				invoker: 'api',
				path: '/delay.yaml',
				status: 'pending',
			},
		})

		id = req.body.data.id

		await sleep(200)
	})

	it(`should fail to cancel the instance`, async () => {
		const req = await request(common.config.getDirektivHost()).patch(`/api/v2/namespaces/${ namespaceName }/instances/${ id }`)
		expect(req.statusCode).toEqual(400)
		expect(req.body).toMatchObject({})

		await sleep(500)
	})

	it(`should cancel the instance`, async () => {
		const req = await request(common.config.getDirektivHost()).patch(`/api/v2/namespaces/${ namespaceName }/instances/${ id }`)
			.set('Content-Type', 'application/json')
			.send({
				status: 'cancelled',
			})
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({})

		await sleep(500)
	})

	it(`should verify that the instance has been cancelled`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespaceName }/instances/${ id }`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			data: {
				createdAt: expect.stringMatching(common.regex.timestampRegex),
				definition: expect.stringMatching(common.regex.base64Regex),
				endedAt: expect.stringMatching(common.regex.timestampRegex),
				errorCode: 'direktiv.cancels.api',
				id: expect.stringMatching(common.regex.uuidRegex),
				invoker: 'api',
				path: '/delay.yaml',
				status: 'cancelled',
				inputLength: 28,
				outputLength: 0,
				metadataLength: 0,
			},
		})
	})
})

function sleep (ms) {
	return new Promise(resolve => setTimeout(resolve, ms))
}
