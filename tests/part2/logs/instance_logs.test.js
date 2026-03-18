import { afterAll, describe, expect, it } from '@jest/globals'

import common from '../../common'
import request from '../../common/request'

function sleep(ms) {
	return new Promise((resolve) => setTimeout(resolve, ms))
}

function uniqueNamespace(prefix) {
	const rand = Math.random().toString(16).slice(2, 10)
	return `${prefix}-${Date.now()}-${rand}`
}

async function createNamespace(ns) {
	const res = await request(common.config.getDirektivBaseUrl())
		.post(`/api/v2/namespaces`)
		.send({ name: ns })
	expect(res.statusCode).toBe(200)
	return res
}

async function deleteNamespace(ns) {
	try {
		await request(common.config.getDirektivBaseUrl()).delete(
			`/api/v2/namespaces/${ns}`,
		)
	} catch {
		// best-effort cleanup
	}
}

async function upsertTypescriptWorkflow({ namespace, name, source }) {
	const res = await request(common.config.getDirektivBaseUrl())
		.post(`/api/v2/namespaces/${namespace}/files`)
		.set('Content-Type', 'application/json')
		.send({
			name,
			type: 'workflow',
			mimeType: 'application/typescript',
			data: btoa(source),
		})

	expect(res.statusCode).toBe(200)
	return res
}

async function invokeWorkflow({ namespace, path, expectStatus }) {
	const res = await request(common.config.getDirektivBaseUrl()).post(
		`/api/v2/namespaces/${namespace}/instances?path=${encodeURIComponent(path)}&wait=true`,
	)
	expect(res.statusCode).toBe(expectStatus)
	return res
}

async function getLatestInstanceId({ namespace, filterVal }) {
	const res = await request(common.config.getDirektivBaseUrl()).get(
		`/api/v2/namespaces/${namespace}/instances?filter.field=AS&filter.type=CONTAINS&filter.val=${encodeURIComponent(filterVal)}`,
	)
	expect(res.statusCode).toBe(200)
	expect(Array.isArray(res.body?.data)).toBe(true)
	expect(res.body.data.length).toBeGreaterThan(0)
	expect(typeof res.body.data[0]?.id).toBe('string')
	return res.body.data[0].id
}

async function fetchInstanceLogs({ namespace, instanceId }) {
	return await request(common.config.getDirektivBaseUrl()).get(
		`/api/v2/namespaces/${namespace}/logs?instance=${encodeURIComponent(instanceId)}`,
	)
}

describe('instance logs (new)', () => {
	const namespacesToCleanUp = []

	afterAll(async () => {
		await Promise.all(namespacesToCleanUp.map((ns) => deleteNamespace(ns)))
	})

	it('logs response for a successful workflow execution after 1 second', async () => {
		const namespace = uniqueNamespace('instance-logs-success')
		namespacesToCleanUp.push(namespace)

		const workflowPath = 'successful.wf.ts'
		const expectedWorkflow = `/${workflowPath}`
		const workflowSource = `
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateFirst",
};

function stateFirst(): StateFunction<unknown> {
  return finish({ data: "hello world" })  
}
`

		await createNamespace(namespace)
		await upsertTypescriptWorkflow({
			namespace,
			name: workflowPath,
			source: workflowSource,
		})

		await invokeWorkflow({ namespace, path: workflowPath, expectStatus: 200 })
		await sleep(1000)

		const instanceId = await getLatestInstanceId({
			namespace,
			filterVal: 'successful',
		})

		const logRes = await fetchInstanceLogs({ namespace, instanceId })
		expect(logRes.statusCode).toBe(200)
		expect(Array.isArray(logRes.body?.data)).toBe(true)

		const entries = logRes.body.data
		expect(entries.length).toBeGreaterThan(0)

		expect(entries).toEqual(
			expect.arrayContaining([
				expect.objectContaining({
					level: 'INFO',
					msg: expect.stringContaining('flow starting'),
					namespace,
					workflow: expect.objectContaining({
						workflow: expectedWorkflow,
						instance: instanceId,
					}),
				}),
				expect.objectContaining({
					level: 'INFO',
					msg: `transitioning to 'stateFirst'`,
					namespace,
					workflow: expect.objectContaining({
						state: 'stateFirst',
						workflow: expectedWorkflow,
						instance: instanceId,
					}),
				}),
				expect.objectContaining({
					level: 'INFO',
					msg: 'instance terminated',
					namespace,
					workflow: expect.objectContaining({
						status: 'completed',
						state: 'stateFirst',
						workflow: expectedWorkflow,
						instance: instanceId,
					}),
				}),
			]),
		)
	})

	it('logs response for a workflow that fails with error after 1 second', async () => {
		const namespace = uniqueNamespace('instance-logs-error')
		namespacesToCleanUp.push(namespace)

		const workflowPath = 'error.wf.ts'
		const expectedWorkflow = `/${workflowPath}`
		const workflowSource = `
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateFirst",
};

function stateFirst(): StateFunction<unknown> {
  throw Error("This was set up to fail");

  return finish("unreachable");
}
`

		await createNamespace(namespace)
		await upsertTypescriptWorkflow({
			namespace,
			name: workflowPath,
			source: workflowSource,
		})

		await invokeWorkflow({ namespace, path: workflowPath, expectStatus: 500 })
		await sleep(1000)

		const instanceId = await getLatestInstanceId({
			namespace,
			filterVal: 'error',
		})

		const logRes = await fetchInstanceLogs({ namespace, instanceId })
		expect(logRes.statusCode).toBe(200)
		expect(Array.isArray(logRes.body?.data)).toBe(true)

		const entries = logRes.body.data
		expect(entries.length).toBeGreaterThan(0)

		expect(entries).toEqual(
			expect.arrayContaining([
				expect.objectContaining({
					level: 'INFO',
					msg: expect.stringContaining('flow starting'),
					namespace,
					workflow: expect.objectContaining({
						workflow: expectedWorkflow,
						instance: instanceId,
					}),
				}),
				expect.objectContaining({
					level: 'INFO',
					msg: `transitioning to 'stateFirst'`,
					namespace,
					workflow: expect.objectContaining({
						state: 'stateFirst',
						workflow: expectedWorkflow,
						instance: instanceId,
					}),
				}),
				expect.objectContaining({
					level: 'ERROR',
					msg: expect.stringContaining(
						'error during flow: Error: This was set up to fail',
					),
					namespace,
					workflow: expect.objectContaining({
						status: 'error',
						state: 'stateFirst',
						workflow: expectedWorkflow,
						instance: instanceId,
					}),
				}),
				expect.objectContaining({
					level: 'INFO',
					msg: 'instance terminated',
					namespace,
					workflow: expect.objectContaining({
						status: 'error',
						state: 'stateFirst',
						workflow: expectedWorkflow,
						instance: instanceId,
					}),
				}),
			]),
		)
	})
})

