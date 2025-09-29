import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'
import { retry10 } from '../common/retry'

const namespace = basename(__filename)

describe('Test target-flow plugin', () => {
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

	helpers.itShouldCreateYamlFile(it, expect, namespace,
		'/', 'ep1.yaml', 'endpoint', `
x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: /ep1
    allow_anonymous: true
    plugins:
        target:
            type: target-flow
            configuration:
                namespace: ${ namespace }
                flow: /wf1.yaml
get:
    responses:
        "200":
        description: works
`,
	)

	retry10(`should execute wf1.yaml file`, async () => {
		const res = await request(config.getDirektivBaseUrl()).get(`/api/v2/namespaces/${ namespace }/gateway/ep1`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			result: 'Hello world!',
		})
	})
})
