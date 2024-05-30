import { beforeAll, describe, expect, it } from '@jest/globals'
import { btoa } from 'js-base64'
import { basename } from 'path'

import common from '../common'
import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'
import { retry10 } from '../common/retry'

const namespace = basename(__filename)

describe('Test target-flow-var plugin', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateYamlFile(it, expect, namespace, '/', 'wf1.yaml', 'workflow', `
direktiv_api: workflow/v1
description: A simple 'no-op' state that returns 'Hello world!'
states:
- id: helloworld
  type: noop
  transform:
    result: Hello world!
`)

	it(`should set plain text variable`, async () => {
		const workflowVarResponse = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespace }/variables`)
			.send({
				name: 'foo',
				data: btoa('Hello World 55'),
				mimeType: 'text/plain',
				workflowPath: '/wf1.yaml',
			})
		expect(workflowVarResponse.statusCode).toEqual(200)
	})

	helpers.itShouldCreateYamlFile(it, expect, namespace,
		'/', 'ep1.yaml', 'endpoint', `
direktiv_api: endpoint/v2
path: /ep1
methods: 
  - GET
allow_anonymous: true
plugins:
  target:
    type: target-flow-var
    configuration:
        namespace: ${ namespace }
        variable: foo
        flow: /wf1.yaml
`)
	retry10(`should execute wf1.yaml file`, async () => {
		const res = await request(config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/gateway2/ep1`)
		expect(res.statusCode).toEqual(200)
		expect(res.text).toEqual('Hello World 55')
		expect(res.headers['content-type']).toEqual('text/plain')
		expect(res.headers['content-length']).toEqual('14')
	})
})
