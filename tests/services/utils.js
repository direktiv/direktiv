import { expect, request } from '@jest/globals'

import helpers from '../common/helpers'

export const bashServiceSrc = `{
  "image": "direktiv/bash:dev",
  "scale": 1,
  "size": "small"
}`

export async function waitForServiceStatus(
	baseUrl,
	namespace,
	path,
	status,
	timeoutMs = 5000,
) {
	const deadline = Date.now() + timeoutMs

	while (Date.now() < deadline) {
		const res = await request(baseUrl).get(
			`/api/v2/namespaces/${namespace}/services`,
		)
		expect(res.statusCode).toEqual(200)

		const instance = res.body?.data.find((item) => item.path === path)

		if (instance.status === status) {
			return instance
		}

		await helpers.sleep(200)
	}

	throw new Error(
		`service ${path} did not reach status ${status} within ${timeoutMs}ms`,
	)
}
