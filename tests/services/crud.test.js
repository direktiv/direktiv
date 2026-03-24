import { afterAll, beforeAll, describe, expect, it } from '@jest/globals'
import {
	bashServiceSrc,
	echoServiceSrc,
	waitForServiceCondition,
	waitForServiceProperty,
	waitForServiceRemoved,
} from './utils'

import config from '../common/config'
import helpers from '../common/helpers'
import regex from '../common/regex'
import request from '../common/request'

const namespace = helpers.randomNamespaceName()

const baseUrl = config.getDirektivBaseUrl()

describe('Service API', () => {
	beforeAll(async () => {
		const nsRes = await request(baseUrl).post('/api/v2/namespaces').send({
			name: namespace,
		})
		expect(nsRes.statusCode).toEqual(200)
	})

	afterAll(async () => {
		return helpers.deleteNamespace(namespace)
	})

	it('creates and deletes a service', async () => {
		const fileName = 'bash.svc.ts'

		// create service
		const createResponse = await request(baseUrl)
			.post(`/api/v2/namespaces/${namespace}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: fileName,
				type: 'service',
				mimeType: 'application/json',
				data: btoa(bashServiceSrc),
			})
		expect(createResponse.statusCode).toEqual(200)

		// confirm expected response from POST request
		expect(createResponse.body).toMatchObject({
			data: {
				path: `/${fileName}`,
				type: 'service',
				createdAt: expect.stringMatching(regex.timestampRegex),
				updatedAt: expect.stringMatching(regex.timestampRegex),
				data: btoa(bashServiceSrc),
				errors: [],
			},
		})

		// confirm expected response from GET request
		const getResponse = await request(baseUrl).get(
			`/api/v2/namespaces/${namespace}/files/${fileName}`,
		)
		expect(getResponse.statusCode).toEqual(200)
		expect(getResponse.body).toMatchObject({
			data: {
				path: `/${fileName}`,
				type: 'service',
				createdAt: expect.stringMatching(regex.timestampRegex),
				updatedAt: expect.stringMatching(regex.timestampRegex),
				data: btoa(bashServiceSrc),
			},
		})

		// confirm service is running
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

		// delete service
		const deleteResponse = await request(baseUrl)
			.delete(`/api/v2/namespaces/${namespace}/files/${fileName}`)
			.set('Content-Type', 'application/json')
			.send({
				name: fileName,
				type: 'service',
				mimeType: 'application/json',
				data: btoa(bashServiceSrc),
			})
		expect(deleteResponse.statusCode).toEqual(200)

		const getAfterDeleteResponse = await request(baseUrl).get(
			`/api/v2/namespaces/${namespace}/files/${fileName}`,
		)

		// confirm get request does not find deleted service
		expect(getAfterDeleteResponse.statusCode).toEqual(404)

		// confirm removed service is no longer running
		expect(await waitForServiceRemoved(namespace, `/${fileName}`)).toBe(true)
	})

	it('updates a service', async () => {
		const fileName = 'bash-update.svc.ts'
		const filePath = `/${fileName}`

		// create service
		const createResponse = await request(baseUrl)
			.post(`/api/v2/namespaces/${namespace}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: fileName,
				type: 'service',
				mimeType: 'application/json',
				data: btoa(bashServiceSrc),
			})
		expect(createResponse.statusCode).toEqual(200)

		// confirm service is running
		const expectedCondition = {
			type: 'Available',
			status: 'True',
			message: 'Deployment has minimum availability.',
		}

		const serviceBeforePatch = await waitForServiceCondition(
			namespace,
			filePath,
			expectedCondition,
		)

		expect(serviceBeforePatch).toBeDefined()
		expect(serviceBeforePatch.image).toEqual('direktiv/bash:dev')

		// update service
		const patchResponse = await request(baseUrl)
			.patch(`/api/v2/namespaces/${namespace}/files/${fileName}`)
			.set('Content-Type', 'application/json')
			.send({
				data: btoa(echoServiceSrc),
			})
		expect(patchResponse.statusCode).toEqual(200)

		// confirm service has been updated and is running
		const serviceAfterPatch = await waitForServiceProperty(
			namespace,
			filePath,
			{ image: 'direktiv/echo' },
		)

		expect(serviceAfterPatch).toBeDefined()

		const serviceListResponse = await request(baseUrl).get(
			`/api/v2/namespaces/${namespace}/services`,
		)
		// confirm that only one service exists (old service terminated)
		expect(serviceListResponse.body?.data.length).toEqual(1)
	})
})
