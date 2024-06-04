import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'
import { retry10 } from '../common/retry'

const namespace = basename(__filename)

describe('Test gateway2 basic-auth plugin', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateYamlFile(it, expect, namespace,
		'/', 'c1.yaml', 'consumer', `
direktiv_api: "consumer/v1"
username: user1
password: pwd1
api_key: key1
tags:
- tag1
groups:
- group1
`)

	helpers.itShouldCreateYamlFile(it, expect, namespace,
		'/', 'ep1.yaml', 'endpoint', `
direktiv_api: endpoint/v1
path: /foo1
allow_anonymous: false
methods:
  - POST
plugins:
  auth:
    - type: basic-auth   
  target:
    type: debug-target
  inbound:
    - type: acl
      configuration:
        allow_groups: ["group2"]
`)

	helpers.itShouldCreateYamlFile(it, expect, namespace,
		'/', 'ep2.yaml', 'endpoint', `
direktiv_api: endpoint/v1
path: /foo2
allow_anonymous: false
methods:
  - POST
plugins:
  auth:
    - type: basic-auth   
  target:
    type: debug-target
  inbound:
    - type: acl
      configuration:
        allow_groups: ["group1"]
`)
	retry10(`should denied ep1.yaml endpoint`, async () => {
		const res = await request(config.getDirektivHost()).post(`/api/v2/namespaces/${ namespace }/gateway2/foo1`)
			.send({})
			.auth('user1', 'pwd1')
		expect(res.statusCode).toEqual(403)
	})

	retry10(`should access ep2.yaml endpoint`, async () => {
		const res = await request(config.getDirektivHost()).post(`/api/v2/namespaces/${ namespace }/gateway2/foo2`)
			.send({ foo: 'bar' })
			.auth('user1', 'pwd1')
		expect(res.statusCode).toEqual(200)
	})
})
