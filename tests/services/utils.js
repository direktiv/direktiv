import config from '../common/config'
import { expect } from '@jest/globals'
import helpers from '../common/helpers'
import request from '../common/request'

export const bashServiceSrc = `{
  "image": "direktiv/bash:dev",
  "scale": 1,
  "size": "small"
}`

function findPartialMatch(arr, match) {
	return arr.find((item) =>
		Object.entries(match).every(([key, value]) => item[key] === value),
	)
}

export async function waitForServiceCondition(
	namespace,
	path,
	condition,
	timeoutMs = 5000,
) {
	const deadline = Date.now() + timeoutMs
	const baseUrl = config.getDirektivBaseUrl()

	while (Date.now() < deadline) {
		const res = await request(baseUrl).get(
			`/api/v2/namespaces/${namespace}/services`,
		)
		expect(res.statusCode).toEqual(200)

		const service = res.body?.data.find((item) => item.filePath === path)

		let match

		if (service) {
			match = findPartialMatch(service.conditions, condition)
		}

		if (match) {
			return service
		}

		await helpers.sleep(200)
	}

	throw new Error(
		`service ${path} did not reach expected condition within ${timeoutMs}ms`,
	)
}

export async function waitForServiceRemoved(namespace, path, timeoutMs = 5000) {
	const deadline = Date.now() + timeoutMs
	const baseUrl = config.getDirektivBaseUrl()

	while (Date.now() < deadline) {
		const res = await request(baseUrl).get(
			`/api/v2/namespaces/${namespace}/services`,
		)
		expect(res.statusCode).toEqual(200)

		const service = res.body?.data.find((item) => item.filePath === path)

		if (!service) {
			return true
		}

		await helpers.sleep(200)
	}

	throw new Error(`service ${path} still existed after ${timeoutMs}ms`)
}
