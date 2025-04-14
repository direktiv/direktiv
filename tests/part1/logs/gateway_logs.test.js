import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import common from '../../common'
import helpers from '../../common/helpers'
import request from '../../common/request'
import { retry50 } from '../../common/retry'

const namespace = basename(__filename)

describe('Test namespace log api calls', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)
	helpers.itShouldCreateFile(it, expect, namespace,
		'',
		'gw.yaml',
		'endpoint',
		'text/plain',
		btoa(`
x-direktiv-api: endpoint/v2
x-direktiv-config:
  path: /demo
  allow_anonymous: true
  plugins:
    auth: []
    inbound:
      - configuration:
          script: log("four")
        type: js-inbound
    outbound: []
    target:
      configuration:
        status_code: 200
        status_message: Hello
      type: instant-response
get:
  responses:
    "200":
      description: ""
`))

retry50(`call gateway`, async () => {
		await request(common.config.getDirektivHost()).get(`/ns/${ namespace }/demo`)

		const logRes = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/logs?route=%2Fdemo`)
		expect(logRes.statusCode).toEqual(200)
		expect(logRes.body.data).toEqual(
			expect.arrayContaining([
				expect.objectContaining({
					msg: 'four',
				}),
			]),
		)	
	})
})
