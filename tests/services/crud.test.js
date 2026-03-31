import { afterEach, beforeEach, describe, expect, it } from '@jest/globals'
import { bashServiceSrc, echoServiceSrc } from './fixtures'
import {
	waitForServiceCondition,
	waitForServiceCount,
	waitForServiceProperty,
	waitForServiceRemoved,
} from './utils'

import config from '../common/config'
import helpers from '../common/helpers'
import regex from '../common/regex'
import request from '../common/request'

const baseUrl = config.getDirektivBaseUrl()
let namespace = ''

describe('Service API', () => {
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

	it('creates and deletes a service', async () => {
		const fileName = 'bash.svc.json'

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
			10_000,
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
		const fileName = 'bash-update.svc.json'
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
			10_000,
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

	it('renames a service', async () => {
		const fileName = 'asd.svc.json'
		const newFileName = 'def.wf.ts'

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

		const serviceBeforeRename = await waitForServiceCondition(
			namespace,
			`/${fileName}`,
			expectedCondition,
			10_000,
		)
		expect(serviceBeforeRename).toBeDefined()

		// rename service
		const renameResponse = await request(baseUrl)
			.patch(`/api/v2/namespaces/${namespace}/files/${fileName}`)
			.set('Content-Type', 'application/json')
			.send({ path: `/${newFileName}` })
		expect(renameResponse.statusCode).toEqual(200)
		expect(renameResponse.body?.data?.path).toEqual(`/${newFileName}`)

		// confirm service file is updated
		const getOldPathResponse = await request(baseUrl).get(
			`/api/v2/namespaces/${namespace}/files/${fileName}`,
		)
		expect(getOldPathResponse.statusCode).toEqual(404)

		const getNewPathResponse = await request(baseUrl).get(
			`/api/v2/namespaces/${namespace}/files/${newFileName}`,
		)
		expect(getNewPathResponse.statusCode).toEqual(200)
		expect(getNewPathResponse.body?.data?.path).toEqual(`/${newFileName}`)

		// confirm /services lists the renamed filePath
		await waitForServiceRemoved(namespace, `/${fileName}`)
		await waitForServiceProperty(namespace, `/${newFileName}`, {})

		const serviceListResponse = await request(baseUrl).get(
			`/api/v2/namespaces/${namespace}/services`,
		)
		expect(serviceListResponse.statusCode).toEqual(200)
		expect(serviceListResponse.body).toMatchObject({
			data: expect.arrayContaining([
				expect.objectContaining({ filePath: `/${newFileName}` }),
			]),
		})

		// confirm service is running after rename
		const serviceAfterRename = await waitForServiceCondition(
			namespace,
			`/${newFileName}`,
			expectedCondition,
			10_000,
		)
		expect(serviceAfterRename).toBeDefined()

		// confirm service is no longer running after old file is deleted
		await waitForServiceRemoved(namespace, `/${fileName}`)
	})

	it('lists services', async () => {
		const fileName1 = 'list-1.svc.json'
		const fileName2 = 'list-2.svc.json'
		const fileName3 = 'list-3.svc.json'

		// create 3 services (different filenames + different images)
		const src1 = bashServiceSrc
		const src2 = bashServiceSrc.replace(
			'direktiv/bash:dev',
			'direktiv/http-request:dev',
		)
		const src3 = bashServiceSrc.replace(
			'direktiv/bash:dev',
			'direktiv/echo:dev',
		)

		const createResponse1 = await request(baseUrl)
			.post(`/api/v2/namespaces/${namespace}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: fileName1,
				type: 'service',
				mimeType: 'application/json',
				data: btoa(src1),
			})
		expect(createResponse1.statusCode).toEqual(200)

		const createResponse2 = await request(baseUrl)
			.post(`/api/v2/namespaces/${namespace}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: fileName2,
				type: 'service',
				mimeType: 'application/json',
				data: btoa(src2),
			})
		expect(createResponse2.statusCode).toEqual(200)

		const createResponse3 = await request(baseUrl)
			.post(`/api/v2/namespaces/${namespace}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: fileName3,
				type: 'service',
				mimeType: 'application/json',
				data: btoa(src3),
			})
		expect(createResponse3.statusCode).toEqual(200)

		// list services + assert all 3 are present
		await waitForServiceCount(namespace, 3)

		const serviceListResponse = await request(baseUrl).get(
			`/api/v2/namespaces/${namespace}/services`,
		)
		expect(serviceListResponse.statusCode).toEqual(200)

		expect(serviceListResponse.body).toMatchObject({
			data: expect.arrayContaining([
				expect.objectContaining({
					filePath: `/${fileName1}`,
					image: 'direktiv/bash:dev',
				}),
				expect.objectContaining({
					filePath: `/${fileName2}`,
					image: 'direktiv/http-request:dev',
				}),
				expect.objectContaining({
					filePath: `/${fileName3}`,
					image: 'direktiv/echo:dev',
				}),
			]),
		})
	})
})
