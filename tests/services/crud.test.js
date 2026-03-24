import { afterAll, beforeAll, describe, expect, it } from '@jest/globals'
import { bashServiceSrc, waitForServiceCondition } from './utils'

import config from '../common/config'
import helpers from '../common/helpers'
import regex from '../common/regex'
import request from '../common/request'

const namespace = helpers.randomNamespaceName()

const base = config.getDirektivBaseUrl()

describe('Service API', () => {
	beforeAll(async () => {
		const base = config.getDirektivBaseUrl()

		const nsRes = await request(base).post('/api/v2/namespaces').send({
			name: namespace,
		})
		expect(nsRes.statusCode).toEqual(200)
	})

	afterAll(async () => {
		return helpers.deleteNamespace(namespace)
	})

	it('creates a service file and spawns the service', async () => {
		const fileName = 'bash.svc.ts'

		const response = await request(base)
			.post(`/api/v2/namespaces/${namespace}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: fileName,
				type: 'service',
				mimeType: 'application/json',
				data: btoa(bashServiceSrc),
			})
		expect(response.statusCode).toEqual(200)

		expect(response.body).toMatchObject({
			data: {
				path: `/${fileName}`,
				type: 'service',
				createdAt: expect.stringMatching(regex.timestampRegex),
				updatedAt: expect.stringMatching(regex.timestampRegex),
				data: btoa(bashServiceSrc),
				errors: [],
			},
		})

		const expectedCondition = {
			type: 'Available',
			status: 'True',
			message: 'Deployment has minimum availability.',
		}

		const service = await waitForServiceCondition(
			namespace,
			`/${fileName}`,
			expectedCondition,
		)

		expect(service).toBeDefined()
	})
})
