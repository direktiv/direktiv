import { afterEach, beforeEach, describe, expect, it } from '@jest/globals'
import { waitForServiceCondition, waitForServicePodsCount } from './utils'

import { bashServiceScale3Src } from './fixtures'
import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'

const namespace = helpers.randomNamespaceName()
const baseUrl = config.getDirektivBaseUrl()

describe('Service pods', () => {
	beforeEach(async () => {
		const nsRes = await request(baseUrl).post('/api/v2/namespaces').send({
			name: namespace,
		})
		expect(nsRes.statusCode).toEqual(200)
	})

	afterEach(async () => {
		return helpers.deleteNamespace(namespace)
	})

	it('spawns expected number of pods', async () => {
		const fileName = 'bash-pods.svc.ts'
		const filePath = `/${fileName}`

		const createResponse = await request(baseUrl)
			.post(`/api/v2/namespaces/${namespace}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: fileName,
				type: 'service',
				mimeType: 'application/json',
				data: btoa(bashServiceScale3Src),
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
		expect(service.scale).toEqual(2)

		const podsRes = await waitForServicePodsCount(
			namespace,
			service.id,
			service.scale,
		)
		expect(podsRes.body?.data?.length).toEqual(service.scale)
	})
})
