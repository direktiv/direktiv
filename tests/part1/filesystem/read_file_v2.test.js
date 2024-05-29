import { beforeAll, describe, expect, it } from '@jest/globals'
import { btoa, btou } from 'js-base64'
import { basename } from 'path'

import config from '../../common/config'
import helpers from '../../common/helpers'
import regex from '../../common/regex'
import request from '../../common/request'

const namespace = basename(__filename)

describe('Test filesystem read single file', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespace)
	helpers.itShouldCreateFileV2(it, expect, namespace, '/', 'foo.yaml', 'file', 'text/plain', btoa('some foo data'))

	it(`should read file`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/files/foo.yaml`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			data: {
				path: '/foo.yaml',
				type: 'file',
				data: btoa('some foo data'),
				createdAt: expect.stringMatching(regex.timestampRegex),
				updatedAt: expect.stringMatching(regex.timestampRegex),
			},
		})
	})

	it(`should read raw file`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/files/foo.yaml?withRaw=true`)
		expect(res.statusCode).toEqual(200)
		expect(res.headers['content-type']).toEqual('text/plain')
		expect(res.headers['content-length']).toEqual('13')
		expect(res.text).toEqual('some foo data')
	})

	it(`should read raw file not found`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/files/something.yaml?withRaw=true`)
		expect(res.statusCode).toEqual(404)
	})

	it(`should read raw file not found`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/files/something.yaml`)
		expect(res.statusCode).toEqual(404)
		expect(res.body).toMatchObject({
			error: {
				code: 'resource_not_found',
				 message: 'filesystem path is not found',

			},
		})
	})
})
