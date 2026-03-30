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

async function waitForService(namespace, path, options) {
	const {
		timeoutMs = 5000,
		intervalMs = 200,
		matchFn,
		onTimeout,
	} = options ?? {}

	if (typeof matchFn !== 'function') {
		throw new Error('waitForService requires an matchFn predicate')
	}

	const deadline = Date.now() + timeoutMs
	const baseUrl = config.getDirektivBaseUrl()

	while (Date.now() < deadline) {
		const res = await request(baseUrl).get(
			`/api/v2/namespaces/${namespace}/services`,
		)
		expect(res.statusCode).toEqual(200)

		const service = res.body?.data.find((item) => item.filePath === path)
		const result = matchFn(service, res)

		if (result?.done) {
			return result.value
		}

		await helpers.sleep(intervalMs)
	}

	if (typeof onTimeout === 'function') {
		throw onTimeout()
	}

	throw new Error(
		`service ${path} did not satisfy predicate within ${timeoutMs}ms`,
	)
}

export async function waitForServiceProperty(
	namespace,
	path,
	property,
	timeoutMs = 5000,
	intervalMs = 200,
) {
	return waitForService(namespace, path, {
		timeoutMs,
		intervalMs,
		matchFn: (service) => {
			if (service && isPartialMatch(service, property)) {
				return { done: true, value: service }
			}
			return { done: false }
		},
		onTimeout: () =>
			new Error(
				`service ${path} did not have expected property within ${timeoutMs}ms`,
			),
	})
}

export async function waitForServiceCondition(
	namespace,
	path,
	condition,
	timeoutMs = 5000,
	intervalMs = 200,
) {
	return waitForService(namespace, path, {
		timeoutMs,
		intervalMs,
		matchFn: (service) => {
			const match = service
				? findPartialMatch(service.conditions, condition)
				: undefined

			if (match) {
				return { done: true, value: service }
			}
			return { done: false }
		},
		onTimeout: () =>
			new Error(
				`service ${path} did not reach expected condition within ${timeoutMs}ms`,
			),
	})
}

export async function waitForServiceRemoved(
	namespace,
	path,
	timeoutMs = 5000,
	intervalMs = 200,
) {
	return waitForService(namespace, path, {
		timeoutMs,
		intervalMs,
		matchFn: (service) => {
			if (!service) {
				return { done: true, value: true }
			}
			return { done: false }
		},
		onTimeout: () =>
			new Error(`service ${path} still existed after ${timeoutMs}ms`),
	})
}

export async function waitForServiceCount(
	namespace,
	expectedCount,
	timeoutMs = 5000,
) {
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

export async function waitForServicePodsCount(
	namespace,
	serviceID,
	expectedCount,
	timeoutMs = 5000,
	intervalMs = 200,
) {
	const baseUrl = config.getDirektivBaseUrl()
	const deadline = Date.now() + timeoutMs

	while (Date.now() < deadline) {
		const res = await request(baseUrl).get(
			`/api/v2/namespaces/${namespace}/services/${serviceID}/pods`,
		)
		expect(res.statusCode).toEqual(200)

		const count = res.body?.data?.length
		if (count === expectedCount) {
			return res
		}

		await helpers.sleep(intervalMs)
	}

	throw new Error(
		`service ${serviceID} pods did not reach expected count ${expectedCount} within ${timeoutMs}ms`,
	)
}
