import { beforeAll, describe, expect, it } from '@jest/globals'

import config from '../common/config'
import helpers from '../common/helpers'
import regex from '../common/regex'
import request from '../common/request'

const namespaceName = 'functionsfiles'

describe('Test function files behaviour', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespaceName)

	helpers.itShouldCreateYamlFileV2(it, expect, namespaceName,
		'/', 'bash.yaml', 'service', `
direktiv_api: service/v1
name: bash
image: direktiv/bash:dev
cmd: ""
scale: 1
`)

	helpers.itShouldCreateFileV2(it, expect, namespaceName,
		'',
		`a.yaml`,
		'workflow',
		'text/plain',
		btoa(`
functions:
- id: bash
  type: knative-namespace
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

	it(`should invoke the '/a.yaml' workflow on a fresh namespace`, async () => {
		const req = await request(config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/tree/a.yaml?op=wait`)
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

	helpers.itShouldCreateFileV2(it, expect, namespaceName,
		'',
		`e.yaml`,
		'workflow',
		'text/plain',
		btoa(`
states:
- id: a
  type: noop
  transform:
    result: x`))

	it(`should invoke the '/a.yaml' workflow on a non-fresh namespace`, async () => {
		const req = await request(config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/tree/a.yaml?op=wait`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			var: {
				a: 5,
				b: 7,
				c: 11,
				d: 13,
				e: 'CnN0YXRlczoKLSBpZDogYQogIHR5cGU6IG5vb3AKICB0cmFuc2Zvcm06CiAgICByZXN1bHQ6IHg=',
			},
		})
		expect(req.body.return.bash[0]).toMatchObject({
			result: 5,
			success: true,
		})
		expect(req.body.return.bash[2]).toMatchObject({
			result: 7,
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
		expect(req.body.return.bash[8].result).not.toBe('')
	})
})
