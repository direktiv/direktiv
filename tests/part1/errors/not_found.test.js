import { describe, expect, it } from '@jest/globals'

import common from '../../common'
import request from '../../common/request'

describe('Test path not found', () => {
	const methods = [
		'get',
		'post',
		'put',
		'delete',
	]

	const paths = [
		'/api/something',
		'/api/something/',
		'/api/v2/something',
		'/api/v1/something/',
		'/api/v1/something/not/found',
		'/api/v2/something/not/found/',
	]

	paths.forEach(path => {
		methods.forEach(method => {
			it(`should return not_found for path:${ path } with method:${ method }`, async () => {
				const res = await request(common.config.getDirektivHost())[method](path)
				expect(res.statusCode).toEqual(404)
				expect(res.body).toMatchObject({
					error: {
						code: 'request_path_not_found',
						message: 'request http path is not found',
					},
				},

				)
			})
		})
	})
})
