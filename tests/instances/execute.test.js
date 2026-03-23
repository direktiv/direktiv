import { beforeAll, describe, expect, it } from '@jest/globals'
import {
	delayErrorWorkflowSource,
	delayWorkflowSource,
	okWorkflowSource,
} from './utils'

import common from '../common'
import request from '../common/request'
import { waitForInstanceStatus } from './utils'

const namespaceName = 'executetest'

describe('instance API - execute and report state', () => {
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
				data: btoa(okWorkflowSource),
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

		const getRes = await waitForInstanceStatus(
			base,
			namespaceName,
			id,
			'complete',
		)

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

	it('executes a workflow instance and reports pending and complete states', async () => {
		const base = common.config.getDirektivBaseUrl()
		const workflowPath = '/delay-success.wf.ts'

		const fileRes = await request(base)
			.post(`/api/v2/namespaces/${namespaceName}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: workflowPath,
				type: 'workflow',
				mimeType: 'application/typescript',
				data: btoa(delayWorkflowSource),
			})
		expect(fileRes.statusCode).toEqual(200)

		const createRes = await request(base).post(
			`/api/v2/namespaces/${namespaceName}/instances?path=${workflowPath}`,
		)
		expect(createRes.statusCode).toEqual(200)
		expect(createRes.body).toMatchObject({
			data: {
				createdAt: expect.stringMatching(common.regex.timestampRegex),
				definition: expect.stringMatching(common.regex.base64Regex),
				id: expect.stringMatching(common.regex.uuidRegex),
				invoker: 'api',
				path: workflowPath,
			},
		})

		const { id } = createRes.body.data

		const runningRes = await waitForInstanceStatus(
			base,
			namespaceName,
			id,
			'running',
		)

		expect(runningRes.body).toMatchObject({
			data: {
				id,
				status: 'running',
				path: workflowPath,
				invoker: 'api',
				namespace: namespaceName,
			},
		})

		const completeRes = await waitForInstanceStatus(
			base,
			namespaceName,
			id,
			'complete',
			8000,
		)

		expect(completeRes.body).toMatchObject({
			data: {
				id,
				status: 'complete',
				path: workflowPath,
				invoker: 'api',
				namespace: namespaceName,
			},
		})
	})

	it('executes a workflow instance and reports pending and failed states', async () => {
		const base = common.config.getDirektivBaseUrl()
		const workflowPath = '/delay-fail.wf.ts'

		const fileRes = await request(base)
			.post(`/api/v2/namespaces/${namespaceName}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: workflowPath,
				type: 'workflow',
				mimeType: 'application/typescript',
				data: btoa(delayErrorWorkflowSource),
			})
		expect(fileRes.statusCode).toEqual(200)

		const createRes = await request(base).post(
			`/api/v2/namespaces/${namespaceName}/instances?path=${workflowPath}`,
		)
		expect(createRes.statusCode).toEqual(200)
		expect(createRes.body).toMatchObject({
			data: {
				createdAt: expect.stringMatching(common.regex.timestampRegex),
				definition: expect.stringMatching(common.regex.base64Regex),
				id: expect.stringMatching(common.regex.uuidRegex),
				invoker: 'api',
				path: workflowPath,
			},
		})

		const { id } = createRes.body.data

		const runningRes = await waitForInstanceStatus(
			base,
			namespaceName,
			id,
			'running',
		)

		expect(runningRes.body).toMatchObject({
			data: {
				id,
				status: 'running',
				path: workflowPath,
				invoker: 'api',
				namespace: namespaceName,
			},
		})

		const completeRes = await waitForInstanceStatus(
			base,
			namespaceName,
			id,
			'failed',
			8000,
		)

		expect(completeRes.body).toMatchObject({
			data: {
				id,
				status: 'failed',
				path: workflowPath,
				invoker: 'api',
				namespace: namespaceName,
			},
		})
	})
})
