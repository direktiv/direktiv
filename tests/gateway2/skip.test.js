import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'
import { retry10 } from '../common/retry'

const namespace = basename(__filename)

describe('Test gateway basic-auth plugin', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateYamlFile(it, expect, namespace,
		'/', 'wf2.yaml', 'workflow', `
direktiv_api: workflow/v1
states:
- id: a
  type: noop
`)

	helpers.itShouldCreateYamlFile(it, expect, namespace,
		'/', 'ep1.yaml', 'endpoint', `
x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: /ep1
    allow_anonymous: true
    timeout: 1
    plugins:
        target:
            type: target-flow
            configuration:
                namespace: ${ namespace }
                flow: /wf2.yaml
get:
    responses:
        "200":
            description: works
`,
	)

	helpers.itShouldCreateYamlFile(it, expect, namespace,
		'/', 'ep2.yaml', 'endpoint', `
x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: /ep3
    allow_anonymous: true
    timeout: 10
    skip_openapi: true
    plugins:
        target:
            type: target-flow
            configuration:
                namespace: ${ namespace }
                flow: /wf2.yaml
get:
    responses:
        "200":
            description: works
`,
	)

	retry10(`should show two routes`, async () => {
		const res = await request(config.getDirektivBaseUrl()).get(`/api/v2/namespaces/${ namespace }/gateway/routes`)
			.send({})
		expect(res.statusCode).toEqual(200)
		expect(res.body.data).toHaveLength(2)
	})

	retry10(`should show one route`, async () => {
		const res = await request(config.getDirektivBaseUrl()).get(`/api/v2/namespaces/${ namespace }/gateway/info`)
			.send({})
		expect(res.statusCode).toEqual(200)
		expect(res.body.data.spec.paths['/ep1']).not.toBeUndefined()
		expect(res.body.data.spec.paths['/ep2']).toBeUndefined()
	})
})
