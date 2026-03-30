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

export async function waitForResponseToMatch(endpoint, options) {
	const {
		timeoutMs = 5000,
		intervalMs = 200,
		matchFn,
		onTimeout,
		method = 'get',
		headers,
		body,
	} = options ?? {}

	if (typeof matchFn !== 'function') {
		throw new Error('waitForResponseToMatch requires an matchFn predicate')
	}

	const baseUrl = config.getDirektivBaseUrl()

	const deadline = Date.now() + timeoutMs

	while (Date.now() < deadline) {
		let req = request(baseUrl)[method](endpoint)

		if (headers) {
			for (const [key, value] of Object.entries(headers)) {
				req = req.set(key, value)
			}
		}

		if (body !== undefined) {
			req = req.send(body)
		}

		const res = await req
		const result = matchFn(res)

		if (result) {
			return result
		}

		await helpers.sleep(intervalMs)
	}

	if (typeof onTimeout === 'function') {
		throw onTimeout()
	}

	throw new Error(`request ${endpoint} did not satisfy predicate within ${timeoutMs}ms`)
}

export async function waitForServiceProperty(
	namespace,
	path,
	property,
	timeoutMs = 5000,
	intervalMs = 200,
) {
	return waitForResponseToMatch(`/api/v2/namespaces/${namespace}/services`, {
		timeoutMs,
		intervalMs,
		matchFn: (res) => {
			expect(res.statusCode).toEqual(200)
			const service = res.body?.data?.find((item) => item.filePath === path)
			if (service && isPartialMatch(service, property)) return service
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
	return waitForResponseToMatch(`/api/v2/namespaces/${namespace}/services`, {
		timeoutMs,
		intervalMs,
		matchFn: (res) => {
			expect(res.statusCode).toEqual(200)
			const service = res.body?.data?.find((item) => item.filePath === path)
			const match = service
				? findPartialMatch(service.conditions ?? [], condition)
				: undefined

			if (match) return service
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
	return waitForResponseToMatch(`/api/v2/namespaces/${namespace}/services`, {
		timeoutMs,
		intervalMs,
		matchFn: (res) => {
			expect(res.statusCode).toEqual(200)
			const service = res.body?.data?.find((item) => item.filePath === path)
			if (!service) return true
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
	return waitForResponseToMatch(`/api/v2/namespaces/${namespace}/services`, {
		timeoutMs,
		intervalMs: 200,
		matchFn: (res) => {
			expect(res.statusCode).toEqual(200)
			const count = res.body?.data?.length
			if (count === expectedCount) return res
		},
		onTimeout: () =>
			new Error(
				`services did not reach expected count ${expectedCount} within ${timeoutMs}ms`,
			),
	})
}

export async function waitForServicePodsCount(
	namespace,
	serviceID,
	expectedCount,
	timeoutMs = 5000,
	intervalMs = 200,
) {
	return waitForResponseToMatch(
		`/api/v2/namespaces/${namespace}/services/${serviceID}/pods`,
		{
		timeoutMs,
		intervalMs,
		matchFn: (res) => {
			expect(res.statusCode).toEqual(200)
			const count = res.body?.data?.length
			if (count === expectedCount) return res
		},
		onTimeout: () =>
			new Error(
				`service ${serviceID} pods did not reach expected count ${expectedCount} within ${timeoutMs}ms`,
			),
		},
	)
}
