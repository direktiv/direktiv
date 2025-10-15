import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'
import { retry10 } from '../common/retry'
import {fileURLToPath} from "url";

const namespace = basename(fileURLToPath(import.meta.url))

describe('Test gateway gitlab-webhook-auth plugin', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateYamlFile(it, expect, namespace,
		'/', 'ep1.yaml', 'endpoint', `
    x-direktiv-api: endpoint/v2
    x-direktiv-config:
        path: "/foo"
        allow_anonymous: false
        plugins:
          auth:
          - type: gitlab-webhook-auth
            configuration:
                secret: secret
          target:
            type: debug-target
    post:
      responses:
         "200":
           description: works
`,
	)

	retry10(`should access ep1.yaml endpoint`, async () => {
		const res = await request(common.config.getDirektivBaseUrl()).post(`/api/v2/namespaces/${ namespace }/gateway/foo`)
			.set('X-Gitlab-Token', 'secret')
			.send({ hello: 'world' })
		expect(res.statusCode).toEqual(200)
	})

	retry10(`should denied ep1.yaml endpoint`, async () => {
		const res = await request(common.config.getDirektivBaseUrl()).post(`/api/v2/namespaces/${ namespace }/gateway/foo`)
			.set('X-Gitlab-Token', 'wrongSecret')
			.send({ hello: 'world' })
		expect(res.statusCode).toEqual(403)
	})
})
