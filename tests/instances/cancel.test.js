import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const namespaceName = 'canceltest'

let id = ''

describe('Test wait success API behaviour', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespaceName)

	helpers.itShouldCreateFile(
		it,
		expect,
		namespaceName,
		'',
		'delay.wf.ts',
		'workflow',
		'application/typescript',
		btoa(`
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateFirst",
};

function stateFirst(): StateFunction<unknown> {
	sleep(10)
  return finish({ data: "finished after waiting for 10s" })  
}
`),
	)

	it(`should invoke the 'delay.wf.ts' workflow`, async () => {
		const req = await request(common.config.getDirektivBaseUrl()).post(
			`/api/v2/namespaces/${namespaceName}/instances?path=delay.wf.ts`,
		)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			data: {
				createdAt: expect.stringMatching(common.regex.timestampRegex),
				definition: expect.stringMatching(common.regex.base64Regex),
				id: expect.stringMatching(common.regex.uuidRegex),
				invoker: 'api',
				path: '/delay.wf.ts',
				status: 'pending',
			},
		})

		id = req.body.data.id

		await sleep(200)
	})

	it(`should cancel the instance`, async () => {
		await sleep(1000)

		const req = await request(common.config.getDirektivBaseUrl())
			.patch(`/api/v2/namespaces/${namespaceName}/instances/${id}`)
			.set('Content-Type', 'application/json')
			.send({
				status: 'cancelled',
			})

		expect(req.statusCode).toEqual(200)

		await sleep(500)
	})

	it(`should verify that the instance has been cancelled`, async () => {
		const req = await request(common.config.getDirektivBaseUrl()).get(
			`/api/v2/namespaces/${namespaceName}/instances/${id}`,
		)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			data: {
				id,
				status: 'cancelled',
				path: '/delay.wf.ts',
				invoker: 'api',
				namespace: namespaceName,
				createdAt: expect.stringMatching(common.regex.timestampRegex),
				startedAt: expect.stringMatching(common.regex.timestampRegex),
				endedAt: expect.stringMatching(common.regex.timestampRegex),
				definition: expect.stringMatching(common.regex.base64Regex),
				errorCode: null,
				errorMessage: null,
			},
		})
	})
})

function sleep(ms) {
	return new Promise((resolve) => setTimeout(resolve, ms))
}
