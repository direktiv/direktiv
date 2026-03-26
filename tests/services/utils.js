import config from '../common/config'
import { expect } from '@jest/globals'
import helpers from '../common/helpers'
import request from '../common/request'

function isPartialMatch(object, match) {
	const result = Object.entries(match).every(
		([key, value]) => object[key] === value,
	)
	return result
}

function findPartialMatch(arr, match) {
	return arr.find((item) => isPartialMatch(item, match))
}

export async function waitForServiceProperty(
	namespace,
	path,
	property,
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

		if (service && isPartialMatch(service, property)) {
			return service
		}

		await helpers.sleep(200)
	}

	throw new Error(
		`service ${path} did not have expected property within ${timeoutMs}ms`,
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

export async function waitForServiceCount(namespace, expectedCount, timeoutMs = 5000) {
	const baseUrl = config.getDirektivBaseUrl()
	const deadline = Date.now() + timeoutMs

	while (Date.now() < deadline) {
		const res = await request(baseUrl).get(
			`/api/v2/namespaces/${namespace}/services`,
		)
		expect(res.statusCode).toEqual(200)

		const count = res.body?.data?.length
		if (count === expectedCount) {
			return res
		}

		await helpers.sleep(200)
	}

	throw new Error(
		`services did not reach expected count ${expectedCount} within ${timeoutMs}ms`,
	)
}
