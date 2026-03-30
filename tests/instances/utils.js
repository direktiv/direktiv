import { expect } from '@jest/globals'
import helpers from '../common/helpers'
import request from '../common/request'

export const okWorkflowSource = `
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT5S",
  state: "stateFirst",
};

function stateFirst(): StateFunction<unknown> {
  return finish({ result: "ok" });
}
`

export const delayWorkflowSource = `
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateFirst",
};

function stateFirst(): StateFunction<unknown> {
	sleep(5)
  return finish({ data: "finished after waiting for 5s" })  
}
`

export const delayErrorWorkflowSource = `
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateFirst",
};

function stateFirst(): StateFunction<unknown> {
	sleep(5)
  throw Error("This was set up to fail after 5s")

  return finish({ data: "unreachable" })  
}
`

export async function waitForInstanceStatus(
	baseUrl,
	namespace,
	id,
	status,
	timeoutMs = 5000,
) {
	const deadline = Date.now() + timeoutMs

	while (Date.now() < deadline) {
		const res = await request(baseUrl).get(
			`/api/v2/namespaces/${namespace}/instances/${id}`,
		)
		expect(res.statusCode).toEqual(200)

		console.log(res.body?.data)

		const instanceStatus = res.body?.data?.status
		if (instanceStatus === status) {
			return res
		}

		await helpers.sleep(200)
	}

	throw new Error(
		`instance ${id} did not reach status ${status} within ${timeoutMs}ms`,
	)
}
