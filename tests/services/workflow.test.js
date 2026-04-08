import { afterEach, beforeEach, describe, expect, it } from '@jest/globals'
import { bashServiceSrc, workflowUsingServiceSrc } from './fixtures'

import common from '../common'
import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'
import { waitForInstanceStatus } from '../instances/utils'
import { waitForServiceCondition } from './utils'

const baseUrl = config.getDirektivBaseUrl()
let namespace = ''

describe('Service usage from workflow', () => {
	beforeEach(async () => {
		namespace = helpers.randomNamespaceName()
		const nsRes = await request(baseUrl).post('/api/v2/namespaces').send({
			name: namespace,
		})
		expect(nsRes.statusCode).toEqual(200)
	})

	afterEach(async () => {
		return helpers.deleteNamespace(namespace)
	})

	it('creates a service and uses it from a workflow in the same namespace', async () => {
		const serviceFileName = 'service.svc.json'

		// create service file in the namespace
		const createResponse = await request(baseUrl)
			.post(`/api/v2/namespaces/${namespace}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: serviceFileName,
				type: 'service',
				mimeType: 'application/json',
				data: btoa(bashServiceSrc),
			})
		expect(createResponse.statusCode).toEqual(200)

		// create workflow file using the namespace service
		const workflowResponse = await request(baseUrl)
			.post(`/api/v2/namespaces/${namespace}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: 'workflow.wf.ts',
				type: 'workflow',
				mimeType: 'application/typescript',
				data: btoa(workflowUsingServiceSrc),
			})
		expect(workflowResponse.statusCode).toEqual(200)

		// confirm service is running
		const expectedCondition = {
			type: 'Available',
			status: 'True',
		}

		const service = await waitForServiceCondition(
			namespace,
			`/${serviceFileName}`,
			expectedCondition,
		)
		expect(service).toBeDefined()

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
				},
			],
		})

		expect(output.bash[0].result).toMatch(
			"direktiv-bash-service-ok",
		)
	})
})

