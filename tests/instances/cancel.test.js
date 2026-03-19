import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import request from '../common/request'

const namespaceName = 'canceltest'

const delayWorkflowSource = `
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateFirst",
};

function stateFirst(): StateFunction<unknown> {
	sleep(10)
  return finish({ data: "finished after waiting for 10s" })  
}
`

function sleep(ms) {
	return new Promise((resolve) => setTimeout(resolve, ms))
}

describe('instance cancel API', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	beforeAll(async () => {
		const base = common.config.getDirektivBaseUrl()

		const nsRes = await request(base)
			.post('/api/v2/namespaces')
			.send({ name: namespaceName })
		expect(nsRes.statusCode).toEqual(200)

		const fileRes = await request(base)
			.post(`/api/v2/namespaces/${namespaceName}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: 'delay.wf.ts',
				type: 'workflow',
				mimeType: 'application/typescript',
				data: btoa(delayWorkflowSource),
			})
		expect(fileRes.statusCode).toEqual(200)
	})

	it('cancels a workflow instance via PATCH and GET returns status cancelled with expected metadata', async () => {
		const base = common.config.getDirektivBaseUrl()

		const createRes = await request(base).post(
			`/api/v2/namespaces/${namespaceName}/instances?path=delay.wf.ts`,
		)
		expect(createRes.statusCode).toEqual(200)
		expect(createRes.body).toMatchObject({
			data: {
				createdAt: expect.stringMatching(common.regex.timestampRegex),
				definition: expect.stringMatching(common.regex.base64Regex),
				id: expect.stringMatching(common.regex.uuidRegex),
				invoker: 'api',
				path: '/delay.wf.ts',
				status: 'pending',
			},
		})

		const { id } = createRes.body.data

		await sleep(200)

		const patchRes = await request(base)
			.patch(`/api/v2/namespaces/${namespaceName}/instances/${id}`)
			.set('Content-Type', 'application/json')
			.send({ status: 'cancelled' })
		expect(patchRes.statusCode).toEqual(200)

		await sleep(500)

		const getRes = await request(base).get(
			`/api/v2/namespaces/${namespaceName}/instances/${id}`,
		)
		expect(getRes.statusCode).toEqual(200)
		expect(getRes.body).toMatchObject({
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
