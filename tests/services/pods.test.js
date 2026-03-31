import { afterEach, beforeEach, describe, expect, it } from '@jest/globals'
import { bashServiceScale0Src, bashServiceScale3Src } from './fixtures'
import { waitForServiceCondition, waitForServicePodsCount } from './utils'

import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'

const baseUrl = config.getDirektivBaseUrl()
let namespace = ''

describe('Service pods', () => {
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

	it('does not spawn pods when scale is 0', async () => {
		const fileName = 'bash-pods-scale-0.svc.json'
		const filePath = `/${fileName}`

		const createResponse = await request(baseUrl)
			.post(`/api/v2/namespaces/${namespace}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: fileName,
				type: 'service',
				mimeType: 'application/json',
				data: btoa(bashServiceScale0Src),
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

		const listRes = await waitForServicePodsCount(namespace, service.id, 0)
		expect(listRes.body?.data?.length).toEqual(0)
	})

	it('spawns expected number of pods', async () => {
		const fileName = 'bash-pods.svc.json'
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
		expect(service.scale).toEqual(3)

		const podsRes = await waitForServicePodsCount(
			namespace,
			service.id,
			service.scale,
		)
		expect(podsRes.body?.data?.length).toEqual(service.scale)
	})
})
