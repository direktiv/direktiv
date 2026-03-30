import { afterEach, beforeEach, describe, expect, it } from '@jest/globals'
import { waitForResponseToMatch, waitForServiceCondition } from './utils'

import { bashServiceScale3Src } from './fixtures'
import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'

const namespace = helpers.randomNamespaceName()
const baseUrl = config.getDirektivBaseUrl()

describe('Service rebuild', () => {
	beforeEach(async () => {
		const nsRes = await request(baseUrl).post('/api/v2/namespaces').send({
			name: namespace,
		})
		expect(nsRes.statusCode).toEqual(200)
	})

	afterEach(async () => {
		return helpers.deleteNamespace(namespace)
	})

	it('rebuild tears down pods and respawns them', async () => {
		const fileName = 'bash-rebuild.svc.json'
		const filePath = `/${fileName}`

		// create service
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
			10_000,
		)
		expect(service).toBeDefined()

		// confirm pods spawned and store their ids
		const podsBeforeRes = await waitForResponseToMatch(
			`/api/v2/namespaces/${namespace}/services/${service.id}/pods`,

			{
				matchFn: (res) => {
					expect(res.statusCode).toEqual(200)
					const pods = res.body?.data
					if (Array.isArray(pods) && pods.length === 3) return res
				},
				onTimeout: () =>
					new Error(
						`service ${service.id} did not report 3 pods before rebuild`,
					),
			},
		)

		const podIDsBefore = (podsBeforeRes.body?.data ?? [])
			.map((p) => p?.id)
			.filter(Boolean)
		expect(podIDsBefore.length).toEqual(3)

		// trigger rebuild
		const rebuildRes = await request(baseUrl)
			.post(
				`/api/v2/namespaces/${namespace}/services/${service.id}/actions/rebuild`,
			)
			.send()
		expect(rebuildRes.statusCode).toEqual(200)
		expect(rebuildRes.body).toEqual('')

		// confirm new pods spawned
		const podsAfterRes = await waitForResponseToMatch(
			`/api/v2/namespaces/${namespace}/services/${service.id}/pods`,
			{
				matchFn: (res) => {
					expect(res.statusCode).toEqual(200)
					const pods = res.body?.data
					if (!Array.isArray(pods) || pods.length !== 3) return

					const newIDs = pods.map((p) => p?.id).filter(Boolean)
					if (newIDs.length !== 3) return

					const allNew = newIDs.every((id) => !podIDsBefore.includes(id))
					if (allNew) return res
				},
				timeoutMs: 10_000,
				onTimeout: () =>
					new Error(
						`service ${service.id} did not respawn 3 pods with all-new IDs`,
					),
			},
		)

		const podIDsAfter = (podsAfterRes.body?.data ?? [])
			.map((p) => p?.id)
			.filter(Boolean)
		expect(podIDsAfter.length).toEqual(3)
		expect(podIDsAfter.every((id) => !podIDsBefore.includes(id))).toBe(true)

		// confirm old pods removed
		await waitForResponseToMatch(
			`/api/v2/namespaces/${namespace}/services/${service.id}/pods`,
			{
				matchFn: (res) => {
					expect(res.statusCode).toEqual(200)
					const pods = res.body?.data
					if (!Array.isArray(pods)) return

					const ids = pods.map((p) => p?.id).filter(Boolean)
					if (ids.some((id) => podIDsBefore.includes(id))) return

					return true
				},
				onTimeout: () =>
					new Error(
						`service ${service.id} still listed at least one pre-rebuild pod ID`,
					),
			},
		)
	})
})
