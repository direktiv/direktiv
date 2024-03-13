import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import common from '../common'
import config from '../common/config'
import helpers from '../common/helpers'
import regex from '../common/regex'
import request from '../common/request'
import { retry50 } from '../common/retry'

const namespace = basename(__filename)

describe('Test log api calls', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	beforeAll(common.helpers.deleteAllNamespaces)

	helpers.itShouldCreateFileV2(it, expect, namespace,
		'',
		'noop.yaml',
		'workflow',
		'text/plain',
		btoa(`
states:
- id: a
  type: noop
  transform:
    result: x`))

	retry50(`should contain instance log entries`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespace }/tree/noop.yaml?op=wait`)
		expect(req.statusCode).toEqual(200)
		const req1 = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespace }/instances`)
		expect(req.statusCode).toEqual(200)

		const req2 = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/logs?instance=${ req1.body.instances.results[0].id }`)
		expect(req2.statusCode).toEqual(200)
		expect(req2.body.data).not.toBeNull()
		expect(req2.body.data.length).toBeGreaterThanOrEqual(1)
	},
	)
	retry50(`should contain namespace log entries`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/logs`)
		expect(req.statusCode).toEqual(200)
		expect(req.body.data).not.toBeNull()
		expect(req.body.data).not.toBeNull()
		expect(req.body.data.length).toBeGreaterThanOrEqual(1)
	})
})
