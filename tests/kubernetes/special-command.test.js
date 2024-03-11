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
        - command: echo hello
        - command: echo hello
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

	common.helpers.itShouldCreateYamlFileV2(
		it,
		expect,
		testNamespace,
		'/', 'wf1.yaml', 'workflow',
		filesWorkflow,
	)

	retry10(`should invoke workflow`, async () => {
		const res = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ testNamespace }/tree/wf1.yaml?op=wait`)
		expect(res.statusCode).toEqual(200)
		expect(res.body.return[0].Output).toEqual('HELLO')
		expect(res.body.return[1].Output).toEqual('444\n')
	})
})

describe('Test special command with env', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFileV2(
		it,
		expect,
		testNamespace,
		'/', 'wf1.yaml', 'workflow',
		genericContainerWorkflow,
	)

	retry10(`should invoke workflow`, async () => {
		const res = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ testNamespace }/tree/wf1.yaml?op=wait`)
		expect(res.statusCode).toEqual(200)
		expect(res.body.return[0].Output).toEqual('HELLO=WORLD\n')
	})
})

describe('Test special command with supress', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFileV2(
		it,
		expect,
		testNamespace,
		'/', 'wf1.yaml', 'workflow',
		supressWorkflow,
	)

	retry10(`should invoke workflow`, async () => {
		const res = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ testNamespace }/tree/wf1.yaml?op=wait`)
		expect(res.statusCode).toEqual(200)
		expect(res.body.return[0].Output).toEqual('hello\n')
		expect(res.body.return[1].Output).toEqual('')
	})
})

describe('Test special command with stop', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFileV2(
		it,
		expect,
		testNamespace,
		'/', 'wf1.yaml', 'workflow',
		stopWorkflow,
	)

	common.helpers.itShouldCreateYamlFileV2(
		it,
		expect,
		testNamespace,
		'/', 'wf2.yaml', 'workflow',
		stopWorkflow2,
	)

	it(`should wait a second for the services to sync`, async () => {
		await helpers.sleep(1000)
	})

	retry10(`should invoke workflow`, async () => {
		const res = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ testNamespace }/tree/wf1.yaml?op=wait`)
		expect(res.statusCode).toEqual(500)
	})

	retry10(`should invoke workflow`, async () => {
		const res = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ testNamespace }/tree/wf2.yaml?op=wait`)
		expect(res.statusCode).toEqual(200)
		expect(res.body.return.length).toBe(2)
	})
})
