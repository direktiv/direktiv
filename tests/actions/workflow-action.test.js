import { afterEach, beforeEach, describe, expect, it } from '@jest/globals'

import common from '../common'
import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'
import { waitForInstanceStatus } from '../instances/utils'
import { workflowUsingActionSrc } from './fixtures'

const baseUrl = config.getDirektivBaseUrl()
const namespace = helpers.randomNamespaceName()

describe('Action usage from workflow', () => {
	beforeEach(async () => {
		const nsRes = await request(baseUrl).post('/api/v2/namespaces').send({
			name: namespace,
		})
		expect(nsRes.statusCode).toEqual(200)
	})

	afterEach(async () => {
		return helpers.deleteNamespace(namespace)
	})

	it('creates a workflow with an inline action and executes it', async () => {
		// create workflow file with inline action
		const workflowResponse = await request(baseUrl)
			.post(`/api/v2/namespaces/${namespace}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: 'workflow.wf.ts',
				type: 'workflow',
				mimeType: 'application/typescript',
				data: btoa(workflowUsingActionSrc),
			})
		expect(workflowResponse.statusCode).toEqual(200)

		// execute workflow
		const executeResponse = await request(baseUrl).post(
			`/api/v2/namespaces/${namespace}/instances?path=workflow.wf.ts`,
		)
		expect(executeResponse.statusCode).toEqual(200)
		expect(executeResponse.body).toMatchObject({
			data: {
				id: expect.stringMatching(common.regex.uuidRegex),
				path: '/workflow.wf.ts',
			},
		})

		const { id } = executeResponse.body.data

		const instance = await waitForInstanceStatus(
			baseUrl,
			namespace,
			id,
			'complete',
		)
		expect(instance).toBeDefined()
		expect(instance.body.data.status).toEqual('complete')

		const output = JSON.parse(instance.body.data.output)
		expect(output).toMatchObject({
			bash: [
				{
					success: true,
					result: "myenvvalue",
				},
			],
		})
	})
})

