import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const namespaceName = 'canceltest'

var id = ''

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
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			data: {
                created_at: expect.stringMatching(common.regex.timestampRegex),
                definition: expect.stringMatching(common.regex.base64Regex),
                ended_at: null,
                error_code: "",
                id: expect.stringMatching(common.regex.uuidRegex),
                invoker: "api",
                path: "/delay.yaml",
                status: "pending",
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
			status: 'cancelled'
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
                created_at: expect.stringMatching(common.regex.timestampRegex),
                definition: expect.stringMatching(common.regex.base64Regex),
                ended_at: expect.stringMatching(common.regex.timestampRegex),
                error_code: "direktiv.cancels.api",
                id: expect.stringMatching(common.regex.uuidRegex),
                invoker: "api",
                path: "/delay.yaml",
                status: "cancelled",
            },
		})
	})
})

function sleep (ms) {
	return new Promise(resolve => setTimeout(resolve, ms))
}
