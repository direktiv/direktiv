import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'
import { retry10 } from '../common/retry'

const namespace = basename(__filename)

describe('Test gateway2 gitlab-webhook-auth plugin', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateYamlFile(it, expect, namespace,
		'/', 'ep1.yaml', 'endpoint', `
direktiv_api: endpoint/v1
path: /foo
methods: 
  - POST
allow_anonymous: false
plugins:
  target:
    type: debug-target
  auth:
    - type: gitlab-webhook-auth
      configuration:
        secret: secret
`)

	retry10(`should access ep1.yaml endpoint`, async () => {
		const res = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespace }/gateway/foo`)
			.set('X-Gitlab-Token', 'secret')
			.send({ hello: 'world' })
		expect(res.statusCode).toEqual(200)
	})

	retry10(`should denied ep1.yaml endpoint`, async () => {
		const res = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespace }/gateway/foo`)
			.set('X-Gitlab-Token', 'wrongSecret')
			.send({ hello: 'world' })
		expect(res.statusCode).toEqual(403)
	})
})
