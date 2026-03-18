import { afterAll, describe, expect, it } from '@jest/globals'

import common from '../../common'
import request from '../../common/request'

function sleep(ms) {
	return new Promise((resolve) => setTimeout(resolve, ms))
}

async function pollUntil(label, fn, { timeoutMs = 3000, intervalMs = 200 } = {}) {
	const deadline = Date.now() + timeoutMs
	let lastError

	while (true) {
		try {
			const result = await fn()
			if (result !== undefined && result !== null) return result
		} catch (err) {
			lastError = err
		}

		if (Date.now() >= deadline) {
			if (lastError) throw lastError
			throw new Error(`Timed out waiting for ${label} after ${timeoutMs}ms`)
		}

		await sleep(intervalMs)
	}
}

async function eventually(label, fn, { timeoutMs = 3000, intervalMs = 200 } = {}) {
	const deadline = Date.now() + timeoutMs
	let lastError

	while (true) {
		try {
			return await fn()
		} catch (err) {
			lastError = err
		}

		if (Date.now() >= deadline) {
			if (lastError) throw lastError
			throw new Error(`Timed out waiting for ${label} after ${timeoutMs}ms`)
		}

		await sleep(intervalMs)
	}
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
		// best-effort cleanup, fail silently.
	}
}

async function createTypescriptWorkflow({ namespace, name, source }) {
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

async function tryGetLatestInstanceId({ namespace, filterVal }) {
	const res = await request(common.config.getDirektivBaseUrl()).get(
		`/api/v2/namespaces/${namespace}/instances?filter.field=AS&filter.type=CONTAINS&filter.val=${encodeURIComponent(filterVal)}`,
	)

	if (res.statusCode !== 200) return null
	if (!Array.isArray(res.body?.data) || res.body.data.length < 1) return null
	if (typeof res.body.data[0]?.id !== 'string') return null

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

	it('logs response for a successful workflow execution', async () => {
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
		await createTypescriptWorkflow({
			namespace,
			name: workflowPath,
			source: workflowSource,
		})

		await invokeWorkflow({ namespace, path: workflowPath, expectStatus: 200 })

		const instanceId = await pollUntil(
			'instance to appear',
			async () =>
				await tryGetLatestInstanceId({
					namespace,
					filterVal: 'successful',
				}),
			{ timeoutMs: 3000, intervalMs: 200 },
		)

		await eventually(
			'logs to contain expected success entries',
			async () => {
				const res = await fetchInstanceLogs({ namespace, instanceId })
				expect(res.statusCode).toBe(200)
				expect(Array.isArray(res.body?.data)).toBe(true)

				expect(res.body.data).toEqual(
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

				return res
			},
			{ timeoutMs: 3000, intervalMs: 200 },
		)
	})

	it('logs response for a workflow that fails with error', async () => {
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
		await createTypescriptWorkflow({
			namespace,
			name: workflowPath,
			source: workflowSource,
		})

		await invokeWorkflow({ namespace, path: workflowPath, expectStatus: 500 })

		const instanceId = await pollUntil(
			'instance to appear',
			async () =>
				await tryGetLatestInstanceId({
					namespace,
					filterVal: 'error',
				}),
			{ timeoutMs: 3000, intervalMs: 200 },
		)

		await eventually(
			'logs to contain expected error entries',
			async () => {
				const res = await fetchInstanceLogs({ namespace, instanceId })
				expect(res.statusCode).toBe(200)
				expect(Array.isArray(res.body?.data)).toBe(true)

				expect(res.body.data).toEqual(
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

				return res
			},
			{ timeoutMs: 3000, intervalMs: 200 },
		)
	})
})

