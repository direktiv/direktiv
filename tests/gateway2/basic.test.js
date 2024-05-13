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
		'/', 'ep1.yaml', 'endpoint', `
direktiv_api: endpoint/v2
path: /foo
allow_anonymous: true
methods:
  - POST
plugins:
  target:
    type: debug-target
`)

	retry10(`should execute gateway ep1.yaml endpoint`, async () => {
		const res = await request(config.getDirektivHost()).post(`/api/v2/namespaces/${ namespace }/gateway2/foo`)
			.send({})
		expect(res.statusCode).toEqual(200)
		expect(res.body.data.text).toEqual('from debug plugin')
	})
})
