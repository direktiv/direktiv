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
    content: Anything you want to display
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
  allow_anonymous: true
  path: /ep1
  plugins:
    auth: []
    inbound: []
    outbound: []
    target:
      type: target-page
      configuration:
        file: /p1.yaml
get:
  responses:
    "200":
      description: ""
`,
	)

	retry10(`should execute foo.wf.ts file`, async () => {
		const res = await request(config.getDirektivBaseUrl()).get(
			`/api/v2/namespaces/${namespace}/gateway/ep1`,
		)
		expect(res.statusCode).toEqual(200)
		expect(res.text.slice(0, 15)).toBe('<!doctype html>')
		expect(res.headers['content-type']).toEqual('text/html; charset=utf-8')
	})
})
