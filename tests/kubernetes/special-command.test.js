import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'
import { retry10 } from '../common/retry'

const testNamespace = 'command'

const genericContainerWorkflow = `
direktiv_api: workflow/v1
functions:
- id: get
  image: ubuntu:24.04
  type: knative-workflow
  cmd: /usr/share/direktiv/direktiv-cmd
states:
- id: getter 
  type: action
  action:
    function: get
    input: 
      data:
        commands:
        - command: bash -c "env | grep HELLO"
          envs:
          - name: HELLO
            value: WORLD
`

const stopWorkflow = `
direktiv_api: workflow/v1
functions:
- id: get
  image: ubuntu:24.04
  type: knative-workflow
  cmd: /usr/share/direktiv/direktiv-cmd
states:
- id: getter 
  type: action
  action:
    function: get
    input: 
      data:
        commands:
        - command: lsaasdasd
          stop: true
        - command: ls
`

const stopWorkflow2 = `
direktiv_api: workflow/v1
functions:
- id: get
  image: ubuntu:24.04
  type: knative-workflow
  cmd: /usr/share/direktiv/direktiv-cmd
states:
- id: getter 
  type: action
  action:
    function: get
    input: 
      data:
        commands:
        - command: lsaasdasd
        - command: ls
`

const supressWorkflow = `
direktiv_api: workflow/v1
functions:
- id: get
  image: ubuntu:24.04
  type: knative-workflow
  cmd: /usr/share/direktiv/direktiv-cmd
states:
- id: getter 
  type: action
  action:
    function: get
    input: 
      data:
        commands:
        - command: echo hello1
        - command: echo hello2
          suppress_output: true
`

const filesWorkflow = `
direktiv_api: workflow/v1
functions:
- id: get
  image: ubuntu:24.04
  type: knative-workflow
  cmd: /usr/share/direktiv/direktiv-cmd
states:
- id: getter 
  type: action
  action:
    function: get
    input: 
      files:
      - name: script.sh
        content: |
          #!/bin/bash

          cat hello.txt
        permission: 0755
      - name: hello.txt
        content: HELLO
        permission: 0444

      data:
        commands:
        - command: ./script.sh
        - command: stat -c "%a" hello.txt 
`

describe('Test special command with files and permission', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/',
		'wf1.yaml',
		'workflow',
		filesWorkflow,
	)

	retry10(`should invoke workflow`, async () => {
		const res = await request(common.config.getDirektivBaseUrl()).post(
			`/api/v2/namespaces/${testNamespace}/instances?path=wf1.yaml&wait=true`,
		)
		expect(res.statusCode).toEqual(200)
		expect(res.body.return[0].Output).toEqual('HELLO')
		expect(res.body.return[1].Output).toEqual(444)
	})
})

describe('Test special command with env', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/',
		'wf2.yaml',
		'workflow',
		genericContainerWorkflow,
	)

	retry10(`should invoke workflow`, async () => {
		const res = await request(common.config.getDirektivBaseUrl()).post(
			`/api/v2/namespaces/${testNamespace}/instances?path=wf2.yaml&wait=true`,
		)
		expect(res.statusCode).toEqual(200)
		expect(res.body.return[0].Output).toEqual('HELLO=WORLD\n')
	})
})

describe('Test special command with supress', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/',
		'wf3.yaml',
		'workflow',
		supressWorkflow,
	)

	// this prints both but doesn't show on logs
	retry10(`should invoke workflow`, async () => {
		const res = await request(common.config.getDirektivBaseUrl()).post(
			`/api/v2/namespaces/${testNamespace}/instances?path=wf3.yaml&wait=true`,
		)
		expect(res.statusCode).toEqual(200)
		expect(res.body.return[0].Output).toEqual('hello1\n')
		expect(res.body.return[1].Output).toEqual('hello2\n')
	})

	retry10(`should not contain instance log entries`, async () => {
		const instRes = await request(common.config.getDirektivBaseUrl()).get(
			`/api/v2/namespaces/${testNamespace}/instances?filter.field=AS&filter.type=CONTAINS&filter.val=wf3`,
		)
		expect(instRes.statusCode).toEqual(200)

		const logRes = await request(common.config.getDirektivBaseUrl()).get(
			`/api/v2/namespaces/${testNamespace}/logs?instance=${instRes.body.data[0].id}`,
		)
		expect(logRes.statusCode).toEqual(200)

		expect(logRes.body.data).toEqual(
			expect.arrayContaining([
				expect.objectContaining({
					msg: 'hello1\n',
				}),
			]),
		)

		// suppress does not log but adds to result
		expect(logRes.body.data).toEqual(
			expect.arrayContaining([
				expect.not.objectContaining({
					msg: 'hello2\n',
				}),
			]),
		)
	})
})

describe('Test special command with stop', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/',
		'wf5.yaml',
		'workflow',
		stopWorkflow,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/',
		'wf6.yaml',
		'workflow',
		stopWorkflow2,
	)

	it(`should wait a second for the services to sync`, async () => {
		await helpers.sleep(1000)
	})

	retry10(`should invoke workflow`, async () => {
		const res = await request(common.config.getDirektivBaseUrl()).post(
			`/api/v2/namespaces/${testNamespace}/instances?path=wf5.yaml&wait=true`,
		)
		expect(res.statusCode).toEqual(500)
	})

	retry10(`should invoke workflow`, async () => {
		const res = await request(common.config.getDirektivBaseUrl()).post(
			`/api/v2/namespaces/${testNamespace}/instances?path=wf6.yaml&wait=true`,
		)
		expect(res.statusCode).toEqual(200)
		expect(res.body.return.length).toBe(2)
	})
})
