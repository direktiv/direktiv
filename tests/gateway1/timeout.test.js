import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'
import { retry10 } from '../common/retry'
import {fileURLToPath} from "url";

const namespace = basename(fileURLToPath(import.meta.url))

describe('Test gateway basic-auth plugin', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldTSWorkflow(it, expect, namespace, '/', 'foo.wf.ts', `
function stateFirst(input) {
	sleep(5)
	return finish("Hello world!")
}
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
                flow: /foo.wf.ts
get:
    responses:
        "200":
            description: works
`,
	)

	helpers.itShouldCreateYamlFile(it, expect, namespace,
		'/', 'ep3.yaml', 'endpoint', `
x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: /ep3
    allow_anonymous: true
    timeout: 10
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

	helpers.itShouldCreateYamlFile(it, expect, namespace,
		'/', 'ep2.yaml', 'endpoint', `
x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: /ep2
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

	retry10(`should execute gateway ep1.yaml endpoint`, async () => {
		const res = await request(config.getDirektivBaseUrl()).get(`/ns/${ namespace }/ep1`)
			.send({})
		expect(res.statusCode).toEqual(503)
	})

	retry10(`should execute gateway ep2.yaml endpoint`, async () => {
		const res = await request(config.getDirektivBaseUrl()).get(`/ns/${ namespace }/ep2`)
			.send({})
		expect(res.statusCode).toEqual(200)
	})

	retry10(`should execute gateway ep3.yaml endpoint`, async () => {
		const res = await request(config.getDirektivBaseUrl()).get(`/ns/${ namespace }/ep3`)
			.send({})
		expect(res.statusCode).toEqual(200)
	})
})
