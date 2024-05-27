import { beforeAll, describe, expect, it } from '@jest/globals'

import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'
import { retry10 } from '../common/retry'

const systemNamespace = 'system'
const normalNamespace = 'functionsfiles'

describe('Test system services behaviour', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, systemNamespace)
	helpers.itShouldCreateNamespace(it, expect, normalNamespace)

	helpers.itShouldCreateYamlFileV2(it, expect, systemNamespace,
		'/', 'bash.yaml', 'service', `
direktiv_api: service/v1
name: bash
image: direktiv/bash:dev
cmd: ""
scale: 1
`)

	helpers.itShouldCreateFileV2(it, expect, systemNamespace,
		'',
		`a.yaml`,
		'workflow',
		'text/plain',
		btoa(`
functions:
- id: bash
  type: knative-system
  service: /bash.yaml

states:
- id: set-value-fn
  type: action
  action:
    function: bash
    input: 
      commands:
      - command: bash -c 'echo a'
`))

	retry10(`should list services on the system namespace`, async () => {
		const req = await request(config.getDirektivHost()).get(`/api/v2/namespaces/${ systemNamespace }/services`)

		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			data: [
				{
					error: null,
					filePath: '/bash.yaml',
					id: 'system-bash-yaml-c57284f6aa',
					image: 'direktiv/bash:dev',
					namespace: 'system',
					scale: 1,
					size: 'medium',
					type: 'system-service',
				},
			],
		})
	})

	it(`should invoke the '/a.yaml' workflow on the system namespace`, async () => {
		const req = await request(config.getDirektivHost()).post(`/api/v2/namespaces/${ systemNamespace }/instances?path=a.yaml&wait=true`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			return: {
				bash: [
					{
						result: 'a',
						success: true,
					},
				],
			},
		})
	})

	helpers.itShouldCreateFileV2(it, expect, normalNamespace,
		'',
		`a.yaml`,
		'workflow',
		'text/plain',
		btoa(`
functions:
- id: bash
  type: knative-system
  service: /bash.yaml

states:
- id: set-value-fn
  type: action
  action:
    function: bash
    input: 
      commands:
      - command: bash -c 'echo a'
`))

	it(`should invoke the '/a.yaml' workflow on the non-system namespace`, async () => {
		const req = await request(config.getDirektivHost()).post(`/api/v2/namespaces/${ normalNamespace }/instances?path=a.yaml&wait=true`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			return: {
				bash: [
					{
						result: 'a',
						success: true,
					},
				],
			},
		})
	})

	helpers.itShouldCreateFileV2(it, expect, normalNamespace,
		'',
		`b.yaml`,
		'workflow',
		'text/plain',
		btoa(`
functions:
- id: bash
  type: knative-system
  service: /bash.yaml
    
states:
- id: set-c
  type: setter
  variables:
  - key: c
    scope: instance
    value: 11
  transition: set-value-fn
    
- id: set-value-fn
  type: action
  action:
    function: bash
    input: 
      commands:
      - command: bash -c 'cat a'
      - command: bash -c 'echo -n 5 > out/namespace/a'
      - command: bash -c 'cat b'
      - command: bash -c 'echo -n 7 > out/workflow/b'
      - command: bash -c 'cat c'
      - command: bash -c 'echo -n 11 > out/instance/c'
      - command: bash -c 'cat d'
      - command: bash -c 'echo -n 13 > out/instance/d'
      - command: bash -c 'cat e'
    files:
    - key: a
      scope: namespace
    - key: b
      scope: workflow
    - key: c
      scope: instance
    - key: d
      scope: instance
    - key: '/e.yaml'
      as: e
      scope: file
  transition: get-values
    
- id: get-values
  type: getter
  variables:
  - key: a
    scope: namespace
  - key: b
    scope: workflow
  - key: c
    scope: instance
  - key: d
    scope: instance
  - key: '/e.yaml'
    as: e
    scope: file
`))

	it(`should invoke the '/b.yaml' workflow on the non-system namespace, testing function files`, async () => {
		const req = await request(config.getDirektivHost()).post(`/api/v2/namespaces/${ normalNamespace }/instances?path=b.yaml&wait=true`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			var: {
				a: 5,
				b: 7,
				c: 11,
				d: 13,
				e: null,
			},
		})
		expect(req.body.return.bash[0]).toMatchObject({
			result: '',
			success: true,
		})
		expect(req.body.return.bash[2]).toMatchObject({
			result: '',
			success: true,
		})
		expect(req.body.return.bash[4]).toMatchObject({
			result: 11,
			success: true,
		})
		expect(req.body.return.bash[6]).toMatchObject({
			result: '',
			success: true,
		})
		expect(req.body.return.bash[8].result).toBe('')
	})
})
