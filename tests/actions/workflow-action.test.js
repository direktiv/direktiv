import { afterEach, beforeEach, describe, expect, it } from '@jest/globals'
import {
	workflowEchoActionSrc,
	workflowFilesInActionSrc,
	workflowMultiCommandActionSrc,
	workflowTwoActionsSrc,
	workflowUsingActionSrc,
} from './fixtures'

import common from '../common'
import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'
import { waitForInstanceStatus } from '../instances/utils'
import { waitForResponseToMatch } from '../services/utils'

const baseUrl = config.getDirektivBaseUrl()
let namespace = ''

describe('Action usage from workflow', () => {

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
			15_000,
		)
		expect(instance).toBeDefined()
		expect(instance.body.data.status).toEqual('complete')

		const output = JSON.parse(instance.body.data.output)
		expect(output).toMatchObject({
			bash: [
				{
					success: true,
					result: 'myenvvalue',
				},
			],
		})
	})

	it('creates a workflow with two actions and lists them under services', async () => {
		// create workflow file
		const workflowResponse = await request(baseUrl)
			.post(`/api/v2/namespaces/${namespace}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: 'two-actions.wf.ts',
				type: 'workflow',
				mimeType: 'application/typescript',
				data: btoa(workflowTwoActionsSrc),
			})
		expect(workflowResponse.statusCode).toEqual(200)

		// execute workflow
		const executeResponse = await request(baseUrl).post(
			`/api/v2/namespaces/${namespace}/instances?path=two-actions.wf.ts`,
		)
		expect(executeResponse.statusCode).toEqual(200)

		const { id } = executeResponse.body.data
		const instance = await waitForInstanceStatus(
			baseUrl,
			namespace,
			id,
			'complete',
			15_000,
		)
		expect(instance).toBeDefined()
		expect(instance.body.data.status).toEqual('complete')

		// verify both actions are listed under services
		const serviceList = await waitForResponseToMatch(
			`/api/v2/namespaces/${namespace}/services`,
			{
				matchFn: (res) => {
					const images = (res.body?.data ?? []).map((s) => s.image)
					if (
						images.includes('direktiv/bash:dev') &&
						images.includes('direktiv/echo:dev')
					) {
						return res
					}
				},
			},
		)

		expect(serviceList.body).toMatchObject({
			data: expect.arrayContaining([
				expect.objectContaining({
					image: 'direktiv/bash:dev',
					filePath: '/two-actions.wf.ts',
					type: 'workflow',
				}),
				expect.objectContaining({
					image: 'direktiv/echo:dev',
					filePath: '/two-actions.wf.ts',
					type: 'workflow',
				}),
			]),
		})
	})

	it('executes the echo action and returns the input', async () => {
		const workflowResponse = await request(baseUrl)
			.post(`/api/v2/namespaces/${namespace}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: 'workflow-echo.wf.ts',
				type: 'workflow',
				mimeType: 'application/typescript',
				data: btoa(workflowEchoActionSrc),
			})
		expect(workflowResponse.statusCode).toEqual(200)

		const input = { hello: 'world' }

		const executeResponse = await request(baseUrl)
			.post(
				`/api/v2/namespaces/${namespace}/instances?path=workflow-echo.wf.ts`,
			)
			.send(input)
		expect(executeResponse.statusCode).toEqual(200)
		expect(executeResponse.body).toMatchObject({
			data: {
				id: expect.stringMatching(common.regex.uuidRegex),
				path: '/workflow-echo.wf.ts',
			},
		})

		const { id } = executeResponse.body.data

		const instance = await waitForInstanceStatus(
			baseUrl,
			namespace,
			id,
			'complete',
			15_000,
		)
		expect(instance).toBeDefined()
		expect(instance.body.data.status).toEqual('complete')

		const output = JSON.parse(instance.body.data.output)
		expect(output).toMatchObject(input)
	})

	it('executes multiple commands in order within one action run', async () => {
		const workflowResponse = await request(baseUrl)
			.post(`/api/v2/namespaces/${namespace}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: 'workflow-multi-command.wf.ts',
				type: 'workflow',
				mimeType: 'application/typescript',
				data: btoa(workflowMultiCommandActionSrc),
			})
		expect(workflowResponse.statusCode).toEqual(200)

		const executeResponse = await request(baseUrl).post(
			`/api/v2/namespaces/${namespace}/instances?path=workflow-multi-command.wf.ts`,
		)
		expect(executeResponse.statusCode).toEqual(200)
		expect(executeResponse.body).toMatchObject({
			data: {
				id: expect.stringMatching(common.regex.uuidRegex),
				path: '/workflow-multi-command.wf.ts',
			},
		})

		const { id } = executeResponse.body.data

		const instance = await waitForInstanceStatus(
			baseUrl,
			namespace,
			id,
			'complete',
			15_000,
		)
		expect(instance).toBeDefined()
		expect(instance.body.data.status).toEqual('complete')

		const output = JSON.parse(instance.body.data.output)
		expect(output).toMatchObject({
			bash: [
				{
					success: true,
					result: 'proof-one',
				},
				{
					success: true,
					result: 'proof-two',
				},
				{
					success: true,
					result: 'proof-three',
				},
			],
		})
	})

	it('mounts workflow variables and namespace files into an action run', async () => {
		// create workflow
		const workflowResponse = await request(baseUrl)
			.post(`/api/v2/namespaces/${namespace}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: 'files-in-action.wf.ts',
				type: 'workflow',
				mimeType: 'application/typescript',
				data: btoa(workflowFilesInActionSrc),
			})
		expect(workflowResponse.statusCode).toEqual(200)

		// set up namespace file mounted via scope "file"
		const fileResponse = await request(baseUrl)
			.post(`/api/v2/namespaces/${namespace}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: 'test.wf.ts',
				type: 'file',
				mimeType: 'application/typescript',
				data: btoa(workflowEchoActionSrc),
			})
		expect(fileResponse.statusCode).toEqual(200)

		// set up workflow-scoped variable mounted via scope "workflow"
		const varResponse = await request(baseUrl)
			.post(`/api/v2/namespaces/${namespace}/variables`)
			.send({
				name: 'wf-var-one',
				workflowPath: '/files-in-action.wf.ts',
				data: btoa('ht5erf'),
				mimeType: 'text/plain',
			})
		expect(varResponse.statusCode).toEqual(200)

		// execute workflow
		const executeResponse = await request(baseUrl).post(
			`/api/v2/namespaces/${namespace}/instances?path=files-in-action.wf.ts`,
		)
		expect(executeResponse.statusCode).toEqual(200)

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

		// 1) files are listed with ls command (including permissions for 0644)
		expect(output?.bash?.[0]?.success).toEqual(true)
		expect(output?.bash?.[0]?.result).toEqual(expect.stringContaining('test.wf.ts'))
		expect(output?.bash?.[0]?.result).toEqual(expect.stringMatching(/-rw-r--r--.*\stest\.wf\.ts$/m))

		// 2) mounted workflow variable content is readable
		expect(output?.bash?.[1]).toMatchObject({
			success: true,
			result: 'ht5erf',
		})

		// 3) mounted file content is readable
		expect(output?.bash?.[2]?.success).toEqual(true)
		expect(output?.bash?.[2]?.result).toEqual(expect.stringContaining('const flow: FlowDefinition'))
	})
})
