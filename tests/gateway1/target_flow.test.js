import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'
import { fileURLToPath } from 'url'

import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'
import { retry10 } from '../common/retry'

const namespace = basename(fileURLToPath(import.meta.url))

describe('Test target-flow plugin', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldTSWorkflow(it, expect, namespace, '/', 'foo.wf.ts', `
function stateFirst(input) {
	return finish("Hello world!")
}
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
                flow: /foo.wf.ts
get:
    responses:
        "200":
        description: works
`,
	)

	retry10(`should execute foo.wf.ts file`, async () => {
		const res = await request(config.getDirektivBaseUrl()).get(`/api/v2/namespaces/${ namespace }/gateway/ep1`)

		expect(res.statusCode).toEqual(200)
		const got = JSON.parse(res.body.data.output)
		const want = 'Hello world!'
		expect(got).toBe(want)
	})
})
