import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'
import { fileURLToPath } from 'url'

import config from '../../common/config'
import helpers from '../../common/helpers'
import request from '../../common/request'

const namespace = basename(fileURLToPath(import.meta.url))

describe('Test filesystem tree read operations', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespace)

	it(`should fail creating file with invalid base64 data`, async () => {
		const res = await request(config.getDirektivBaseUrl())
			.post(`/api/v2/namespaces/${namespace}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: 'foo',
				type: 'workflow',
				mimeType: 'text/plain',
				data: 'some_invalid_base64',
			})
		expect(res.statusCode).toEqual(400)
		expect(res.body).toMatchObject({
			error: {
				code: 'request_data_invalid',
				message: 'file data has invalid base64 string',
			},
		})
	})
})
