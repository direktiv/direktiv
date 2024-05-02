import { describe, expect, it } from '@jest/globals'

import config from '../../common/config'
import request from '../../common/request'

describe('Test the status information API', () => {
	it(`should request status information`, async () => {
		const r = await request(config.getDirektivHost()).get(`/api/v2/status`)
		expect(r.statusCode).toEqual(200)

		expect(r.body.data).toEqual({
			version: expect.anything(),
			isEnterprise: false,
			requiresAuth: false,
		})
	})
})
