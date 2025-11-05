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

	helpers.itShouldCreateYamlFile(
		it,
		expect,
		namespace,
		'/',
		'p1.yaml',
		'page',
		`
direktiv_api: page/v1
type: page
blocks:
  - type: text
    content: Hello world from Pages!
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
    path: /ep1
    allow_anonymous: true
    plugins:
        target:
            type: target-page
            configuration:
                file: /p1.yaml
get:
    responses:
        "200":
        description: works
`,
	)

	retry10(`should execute foo.wf.ts file`, async () => {
		const res = await request(config.getDirektivBaseUrl()).get(
			`/api/v2/namespaces/${namespace}/gateway/ep1`,
		)
		expect(res.statusCode).toEqual(200)
		expect(res.text).toBe('page plugin works')
		expect(res.headers['content-type']).toEqual('text/plain; charset=utf-8')
	})
})
