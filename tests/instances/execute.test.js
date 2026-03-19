import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const namespaceName = 'executetest'

const executeWorkflowSource = `
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT5S",
  state: "stateFirst",
};

function stateFirst(): StateFunction<unknown> {
  return finish({ result: "ok" });
}
`

async function waitForInstanceCompletion(
	baseUrl,
	namespace,
	id,
	timeoutMs = 5000,
) {
	const deadline = Date.now() + timeoutMs

	while (Date.now() < deadline) {
		const res = await request(baseUrl).get(
			`/api/v2/namespaces/${namespace}/instances/${id}`,
		)
		expect(res.statusCode).toEqual(200)

		const status = res.body?.data?.status
		if (
			status === 'complete' ||
			status === 'failed' ||
			status === 'cancelled'
		) {
			return res
		}

		await helpers.sleep(200)
	}

	throw new Error(
		`instance ${id} did not reach terminal state within ${timeoutMs}ms`,
	)
}

describe('instance execute API', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	beforeAll(async () => {
		const base = common.config.getDirektivBaseUrl()

		const nsRes = await request(base).post('/api/v2/namespaces').send({
			name: namespaceName,
		})
		expect(nsRes.statusCode).toEqual(200)

		const fileRes = await request(base)
			.post(`/api/v2/namespaces/${namespaceName}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: 'flow.wf.ts',
				type: 'workflow',
				mimeType: 'application/typescript',
				data: btoa(executeWorkflowSource),
			})
		expect(fileRes.statusCode).toEqual(200)
	})

	it('executes a workflow instance and returns completed instance data', async () => {
		const base = common.config.getDirektivBaseUrl()

		const createRes = await request(base).post(
			`/api/v2/namespaces/${namespaceName}/instances?path=flow.wf.ts`,
		)
		expect(createRes.statusCode).toEqual(200)
		expect(createRes.body).toMatchObject({
			data: {
				createdAt: expect.stringMatching(common.regex.timestampRegex),
				definition: expect.stringMatching(common.regex.base64Regex),
				id: expect.stringMatching(common.regex.uuidRegex),
				invoker: 'api',
				path: '/flow.wf.ts',
			},
		})

		const { id } = createRes.body.data

		const getRes = await waitForInstanceCompletion(base, namespaceName, id)
		expect(getRes.body).toMatchObject({
			data: {
				id,
				status: 'complete',
				path: '/flow.wf.ts',
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
