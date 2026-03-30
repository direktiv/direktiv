import { afterEach, beforeEach, describe, expect, it } from '@jest/globals'

import { bashServiceWithEnvsSrc } from './fixtures'
import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'
import { waitForServiceCondition } from './utils'

const namespace = helpers.randomNamespaceName()
const baseUrl = config.getDirektivBaseUrl()

describe('Service environment variables', () => {
	beforeEach(async () => {
		const nsRes = await request(baseUrl).post('/api/v2/namespaces').send({
			name: namespace,
		})
		expect(nsRes.statusCode).toEqual(200)
	})

	afterEach(async () => {
		return helpers.deleteNamespace(namespace)
	})

	it('lists envs on services endpoint', async () => {
		const fileName = 'bash-envs.svc.ts'
		const filePath = `/${fileName}`

		const createResponse = await request(baseUrl)
			.post(`/api/v2/namespaces/${namespace}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: fileName,
				type: 'service',
				mimeType: 'application/json',
				data: btoa(bashServiceWithEnvsSrc),
			})
		expect(createResponse.statusCode).toEqual(200)

		const expectedCondition = {
			type: 'Available',
			status: 'True',
			message: 'Deployment has minimum availability.',
		}
		const service = await waitForServiceCondition(
			namespace,
			filePath,
			expectedCondition,
		)
		expect(service).toBeDefined()

		const expectedEnvs = [
			{ name: 'FOO1', value: 'bar1' },
			{ name: 'FOO2', value: 'bar2' },
		]

		const deadline = Date.now() + 5000
		while (Date.now() < deadline) {
			const listRes = await request(baseUrl).get(
				`/api/v2/namespaces/${namespace}/services`,
			)
			expect(listRes.statusCode).toEqual(200)

			const listed = listRes.body?.data?.find((item) => item.filePath === filePath)
			if (listed) {
				expect(listed).toMatchObject({
					filePath,
					envs: expect.arrayContaining(expectedEnvs),
				})
				return
			}

			await helpers.sleep(200)
		}

		throw new Error(`service ${filePath} was not listed within 5000ms`)
	})
})

