import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'
import { fileURLToPath } from 'url'

import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'
import { retry10 } from '../common/retry'

const namespace = basename(fileURLToPath(import.meta.url))

describe('Test gateway reconciling', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	it(`should create a new foo secret`, async () => {
		const res = await request(config.getDirektivBaseUrl())
			.post(`/api/v2/namespaces/${namespace}/secrets`)
			.send({
				name: 'foo',
				data: btoa('bar'),
			})
		expect(res.statusCode).toEqual(200)
	})

	helpers.itShouldCreateYamlFile(
		it,
		expect,
		namespace,
		'/',
		'wf1.yml',
		'workflow',
		`
direktiv_api: workflow/v1
description: A simple 'no-op' state that returns 'Hello world!'
states:
- id: step1
  type: noop
  transform:
    result: Hello world!
`,
	)

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
password: fetchSecret(foo)
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
        path: "/foo"
        allow_anonymous: false
        plugins:
          auth:
          - type: basic-auth  
          target:
            type: debug-target
    post:
      responses:
         "200":
           description: works
`,
	)

	retry10(`should get access denied ep1.yaml endpoint`, async () => {
		const res = await request(config.getDirektivBaseUrl())
			.post(`/api/v2/namespaces/${namespace}/gateway/foo`)
			.send({})
			.auth('user1', 'falsePassword')
		expect(res.statusCode).toEqual(403)
		expect(res.body).toEqual({
			error: {
				endpointFile: '/ep1.yaml',
				message: 'authentication failed',
			},
		})
	})

	retry10(`should execute protected ep1.yaml endpoint`, async () => {
		const res = await request(config.getDirektivBaseUrl())
			.post(`/api/v2/namespaces/${namespace}/gateway/foo`)
			.send({})
			.auth('user1', 'bar')
		expect(res.statusCode).toEqual(200)
		expect(res.body.data.text).toEqual('from debug plugin')
	})
})
