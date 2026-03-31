import { afterEach, beforeEach, describe, expect, it } from '@jest/globals'
import { bashServiceSrc, workflowUsingSystemServiceSrc } from './fixtures'

import common from '../common'
import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'
import { waitForInstanceStatus } from '../instances/utils'
import { waitForServiceCondition } from './utils'

const baseUrl = config.getDirektivBaseUrl()
const systemNamespace = 'system'
let normalNamespace = ''

describe('System Service API', () => {
	beforeEach(async () => {
		normalNamespace = helpers.randomNamespaceName()
		const systemNsResponse = await request(baseUrl)
			.post('/api/v2/namespaces')
			.send({
				name: systemNamespace,
			})
		expect(systemNsResponse.statusCode).toEqual(200)
		const normalNsResponse = await request(baseUrl)
			.post('/api/v2/namespaces')
			.send({
				name: normalNamespace,
			})
		expect(normalNsResponse.statusCode).toEqual(200)
	})

	afterEach(async () => {
		return Promise.all([
			helpers.deleteNamespace(systemNamespace),
			helpers.deleteNamespace(normalNamespace),
		])
	})

	it('creates a system service, available to workflows in any namespace', async () => {
		const fileName = `system-service.svc.json`

		// create service file in the "system" namespace
		const createResponse = await request(baseUrl)
			.post(`/api/v2/namespaces/${systemNamespace}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: fileName,
				type: 'service',
				mimeType: 'application/json',
				data: btoa(bashServiceSrc),
			})
		expect(createResponse.statusCode).toEqual(200)

		// create workflow file using the system service
		const workflowResponse = await request(baseUrl)
			.post(`/api/v2/namespaces/${normalNamespace}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: 'workflow.wf.ts',
				type: 'workflow',
				mimeType: 'application/typescript',
				data: btoa(workflowUsingSystemServiceSrc),
			})
		expect(workflowResponse.statusCode).toEqual(200)

		// confirm service is running
		const expectedCondition = {
			type: 'Available',
			status: 'True',
		}

		const service = await waitForServiceCondition(
			systemNamespace,
			`/system-service.svc.json`,
			expectedCondition,
		)
		expect(service).toBeDefined()

		// execute workflow
		const executeResponse = await request(baseUrl).post(
			`/api/v2/namespaces/${normalNamespace}/instances?path=workflow.wf.ts`,
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
			normalNamespace,
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
		expect(output.bash[0].result).toMatch(/^total 8\n/m)
		expect(output.bash[0].result).toMatch(
			/^drwxrwxrwt\s+2\s+root\s+root\s+4096\s+.*\s\.$/m,
		)
		expect(output.bash[0].result).toMatch(
			/^drwxr-xr-x\s+1\s+root\s+root\s+4096\s+.*\s\.\.$/m,
		)
	})
})
