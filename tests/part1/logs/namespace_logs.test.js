import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'
import { fileURLToPath } from 'url'

import common from '../../common'
import helpers from '../../common/helpers'
import request from '../../common/request'
import { retry50 } from '../../common/retry'

const namespace = basename(fileURLToPath(import.meta.url))

describe('Test namespace log api calls', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)
	helpers.itShouldCreateFile(it, expect, namespace,
		'',
		'noop.yaml',
		'workflow',
		'text/plain',
		btoa(`
states:
- id: a
  type: noop
  transform:
    result: jq(.test)`))

	it(`generate namespace logs`, async () => {
		await request(common.config.getDirektivBaseUrl()).post(`/api/v2/namespaces/${ namespace }/instances?path=noop.yaml&wait=true`)
			.set('Content-Type', 'application/json')
			.send('{ "test":"me" }')
	})

	it(`generate namespace logs error`, async () => {
		await request(common.config.getDirektivBaseUrl()).post(`/api/v2/namespaces/${ namespace }/instances?path=noop.yaml&wait=true`)
	})

	retry50(`has one error message and at least on started message`, async () => {
		const logRes = await request(common.config.getDirektivBaseUrl()).get(`/api/v2/namespaces/${ namespace }/logs`)
		expect(logRes.statusCode).toEqual(200)

		expect(logRes.body.data).toEqual(
			expect.arrayContaining([
				expect.objectContaining({
					level: 'ERROR',
				}),
			]),
		)

		expect(logRes.body.data).toEqual(
			expect.arrayContaining([
				expect.objectContaining({
					msg: 'workflow has been started',
				}),
			]),
		)
	})
})
