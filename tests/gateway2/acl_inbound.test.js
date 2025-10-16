import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'
import { fileURLToPath } from 'url'

import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'
import { retry10 } from '../common/retry'

const namespace = basename(fileURLToPath(import.meta.url))

describe('Test gateway basic-auth plugin', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateYamlFile(
		it,
		expect,
		namespace,
		'/',
		'c1.yaml',
		'consumer',
		`
direktiv_api: "consumer/v1"
username: user1
password: pwd1
api_key: key1
tags:
- tag1
groups:
- group1
`,
	)

	helpers.itShouldCreateYamlFile(
		it,
		expect,
		namespace,
		'/',
		'ep1.yaml',
		'endpoint',
		`
x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/foo1"
    allow_anonymous: false
    plugins:
      auth:
      - type: basic-auth  
      target:
        type: debug-target
      inbound:
      - type: acl
        configuration:
          allow_groups: ["group2"]
post:
   responses:
      "200":
        description: works
`,
	)

	helpers.itShouldCreateYamlFile(
		it,
		expect,
		namespace,
		'/',
		'ep2.yaml',
		'endpoint',
		`
x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/foo2"
    allow_anonymous: false
    plugins:
      auth:
      - type: basic-auth 
      target:
        type: debug-target
      inbound:
      - type: acl
        configuration:
          allow_groups: ["group1"]
post:
   responses:
      "200":
        description: works
`,
	)

	retry10(`should denied ep1.yaml endpoint`, async () => {
		const res = await request(config.getDirektivBaseUrl())
			.post(`/api/v2/namespaces/${namespace}/gateway/foo1`)
			.send({})
			.auth('user1', 'pwd1')
		expect(res.statusCode).toEqual(403)
	})

	retry10(`should access ep2.yaml endpoint`, async () => {
		const res = await request(config.getDirektivBaseUrl())
			.post(`/api/v2/namespaces/${namespace}/gateway/foo2`)
			.send({ foo: 'bar' })
			.auth('user1', 'pwd1')
		expect(res.statusCode).toEqual(200)
	})
})
