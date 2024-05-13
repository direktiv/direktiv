import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'
import { retry10 } from '../common/retry'

const namespace = basename(__filename)

describe('Test gateway2 reconciling', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	it(`should create a new foo secret`, async () => {
		const res = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces/${ namespace }/secrets`)
			.send({
				name: 'foo',
				data: btoa('bar'),
			})
		expect(res.statusCode).toEqual(200)
	})

	helpers.itShouldCreateYamlFileV2(it, expect, namespace,
		'/', 'wf1.yml', 'workflow', `
direktiv_api: workflow/v1
description: A simple 'no-op' state that returns 'Hello world!'
states:
- id: helloworld
  type: noop
  transform:
    result: Hello world!
`)

	helpers.itShouldCreateYamlFileV2(it, expect, namespace,
		'/', 'c1.yaml', 'consumer', `
direktiv_api: "consumer/v2"
username: user1
password: fetchSecret(foo)
api_key: key1
tags:
- tag1
groups:
- group1
`)

	helpers.itShouldCreateYamlFileV2(it, expect, namespace,
		'/', 'ep1.yaml', 'endpoint', `
direktiv_api: endpoint/v2
path: /foo
allow_anonymous: false
methods:
  - POST
plugins:
  target:
    type: debug-target
  auth:
    - type: basic-auth   
`)

	retry10(`should get access denied ep1.yaml endpoint`, async () => {
		const res = await request(config.getDirektivHost()).post(`/api/v2/namespaces/${ namespace }/gateway2/foo`)
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
		const res = await request(config.getDirektivHost()).post(`/api/v2/namespaces/${ namespace }/gateway2/foo`)
			.send({})
			.auth('user1', 'bar')
		expect(res.statusCode).toEqual(200)
		expect(res.body.data.text).toEqual('from debug plugin')
	})
})
