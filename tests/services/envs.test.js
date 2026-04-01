import { afterEach, beforeEach, describe, expect, it } from '@jest/globals'
import { waitForResponseToMatch, waitForServiceCondition } from './utils'

import { bashServiceWithEnvsSrc } from './fixtures'
import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'

const baseUrl = config.getDirektivBaseUrl()
let namespace = ''

describe('Service environment variables', () => {
	beforeEach(async () => {
		namespace = helpers.randomNamespaceName()
		const nsRes = await request(baseUrl).post('/api/v2/namespaces').send({
			name: namespace,
		})
		expect(nsRes.statusCode).toEqual(200)
	})

	afterEach(async () => {
		return helpers.deleteNamespace(namespace)
	})

	it('lists envs on services endpoint', async () => {
		const fileName = 'bash-envs.svc.json'
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

		await waitForResponseToMatch(`/api/v2/namespaces/${namespace}/services`, {
			matchFn: (res) => {
				expect(res.statusCode).toEqual(200)

				const match = res.body?.data?.find((item) => item.filePath === filePath)

				if (!match) {
					return false
				}

				expect(match).toMatchObject({
					filePath,
					envs: expect.arrayContaining(expectedEnvs),
				})

				return true
			},
			onTimeout: () =>
				new Error(`service ${filePath} was not listed within timeout`),
		})
	})
})
