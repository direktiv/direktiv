import { describe, expect, it } from '@jest/globals'

import config from '../common/config'
import request from '../common/request'

describe('Test the version information API', () => {
	it(`should request version information`, async () => {
		const r = await request(config.getDirektivHost()).get(`/api/v2/version`)
		expect(r.statusCode).toEqual(200)

		expect(r.body).toMatchObject({
			data: expect.anything(),
		})
	})
})
