import { beforeAll, describe, expect, it } from '@jest/globals'

import { basename } from 'path'
import common from '../../common'
import { fileURLToPath } from 'url'
import helpers from '../../common/helpers'
import request from '../../common/request'

const namespace = basename(fileURLToPath(import.meta.url))

// Todo: Update and unskip in TDI-238 
// namespace logs currently do not log the expected messages
// update assertions to match log output
describe.skip('Test namespace log api calls', () => {
	beforeAll(helpers.deleteAllNamespaces)

	async function retryInIt(retries, delayMs, handler) {
		let latestError
		for (let tries = 0; tries < retries; tries++)
			try {
				await helpers.sleep(delayMs)
				await handler()
				return
			} catch (error) {
				latestError = error
			}
		throw latestError
	}

	const workflowSource = `
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateFirst",
};

const error = 'input must contain { "data": "string" or number }';

function formatMessage(data: string | number, type: string) {
  return { message: \`$\{data} is a $\{type}\` };
}

function stateFirst(input): StateFunction<unknown> {
  const { data } = input;
  if (!data) {
    throw Error(error);
  }
  return transition(stateSecond, data);
}

function stateSecond(data): StateFunction<unknown> {
  const type = typeof data;
  if (type === "string" || type === "number") {
    const message = formatMessage(data, type);
    return finish(message);
  }
  return finish({ error });
}
`

	it('logs when a workflow has started', async () => {
		const ns = `${namespace}-started`
		const createNSRes = await request(common.config.getDirektivBaseUrl())
			.post(`/api/v2/namespaces`)
			.send({ name: ns })
		expect(createNSRes.statusCode).toEqual(200)

		const createFileRes = await request(common.config.getDirektivBaseUrl())
			.post(`/api/v2/namespaces/${ns}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: 'hello.wf.ts',
				type: 'workflow',
				mimeType: 'application/typescript',
				data: btoa(workflowSource),
			})
		expect(createFileRes.statusCode).toEqual(200)

		await request(common.config.getDirektivBaseUrl())
			.post(`/api/v2/namespaces/${ns}/instances?path=hello.wf.ts&wait=true`)
			.set('Content-Type', 'application/json')
			.send('{ "test" : "me" }')

		await retryInIt(50, 500, async () => {
			const logRes = await request(common.config.getDirektivBaseUrl()).get(
				`/api/v2/namespaces/${ns}/logs`,
			)
			expect(logRes.statusCode).toEqual(200)
			expect(logRes.body.data).toEqual(
				expect.arrayContaining([
					expect.objectContaining({
						msg: 'workflow has been started',
					}),
				]),
			)
		})
	})

	it('logs when a workflow has failed', async () => {
		const ns = `${namespace}-error`
		const createNSRes = await request(common.config.getDirektivBaseUrl())
			.post(`/api/v2/namespaces`)
			.send({ name: ns })
		expect(createNSRes.statusCode).toEqual(200)

		const createFileRes = await request(common.config.getDirektivBaseUrl())
			.post(`/api/v2/namespaces/${ns}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: 'hello.wf.ts',
				type: 'workflow',
				mimeType: 'application/typescript',
				data: btoa(workflowSource),
			})

		expect(createFileRes.statusCode).toEqual(200)

		await request(common.config.getDirektivBaseUrl()).post(
			`/api/v2/namespaces/${ns}/instances?path=hello.wf.ts&wait=true`,
		)

		await retryInIt(50, 500, async () => {
			const logRes = await request(common.config.getDirektivBaseUrl()).get(
				`/api/v2/namespaces/${ns}/logs`,
			)
			expect(logRes.statusCode).toEqual(200)
			expect(logRes.body.data).toEqual(
				expect.arrayContaining([
					expect.objectContaining({
						msg: 'workflow has failed',
					}),
				]),
			)
		})
	})
})
