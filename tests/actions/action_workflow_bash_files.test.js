import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'
import { basename } from 'path'

const namespace = basename(__filename)
const testWorkflow = 'test-workflow-bash.yaml'

describe('Test workflow bash commands via action', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateFile(it, expect, namespace,
		'',
		testWorkflow,
		'workflow',
		'text/plain',
		btoa(`
functions:
- id: bash
  image: gcr.io/direktiv/functions/bash:1.0
  type: knative-workflow
states:
- id: bash 
  type: action
  action:
    function: bash
`))

	it(`should echo input via bash action from ${testWorkflow} workflow`, async () => {
		await helpers.sleep(500)
		const res = await request(common.config.getDirektivHost())
			.post(`/api/v2/namespaces/${namespace}/instances?path=${testWorkflow}&wait=true`)
			.send({
				commands: [
					{
						command: "echo '{\"hello\":\"world\"}'"
					}
				]
			});

		expect(res.statusCode).toEqual(200)
		expect(res.body.return.bash).toMatchObject(
			[{ "result": { "hello": "world" }, "success": true }])
	})
	it(`should upload not exec file via bash action from ${testWorkflow} workflow due to bad permissions`, async () => {
		const res = await request(common.config.getDirektivHost())
			.post(`/api/v2/namespaces/${namespace}/instances?path=${testWorkflow}&wait=true`)
			.send({
				files: [
					{ name: "hello.sh", data: "#!/bin/bash\necho 'Hello World'" }
				],
				commands: [
					{
						command: "./hello.sh"
					}
				]
			});
		expect(res.statusCode).toEqual(500)
		expect(res.body.error).toMatchObject({
			code: 'io.direktiv.command.error',
			message: 'fork/exec ./hello.sh: permission denied'
		})
	})
	it(`should upload and exec file via bash action from ${testWorkflow} workflow`, async () => {
		const res = await request(common.config.getDirektivHost())
			.post(`/api/v2/namespaces/${namespace}/instances?path=${testWorkflow}&wait=true`)
			.send({
				files: [
					{ name: "hello.sh", data: "#!/bin/bash\necho 'Hello World'", mode: '0755' }
				],
				commands: [
					{
						command: "./hello.sh"
					}
				]
			});
		expect(res.statusCode).toEqual(200)
		expect(res.body.return.bash).toMatchObject(
			[{ "result": "Hello World", "success": true }])
	})
	it(`should return exported env via bash action from ${testWorkflow} workflow`, async () => {
		const res = await request(common.config.getDirektivHost())
			.post(`/api/v2/namespaces/${namespace}/instances?path=${testWorkflow}&wait=true`)
			.send({
				commands: [
					{
						command: 'touch executed && cat executed',
					}
				]
			});
		expect(res.statusCode).toEqual(200)
		expect(res.body.return.bash).toMatchObject(
			[{ "result": "", "success": true }])
	})
	it(`files from prior action should not exists`, async () => {
		const res = await request(common.config.getDirektivHost())
			.post(`/api/v2/namespaces/${namespace}/instances?path=${testWorkflow}&wait=true`)
			.send({
				commands: [
					{
						command: 'cat executed',
					}
				]
			});

		expect(res.statusCode).toEqual(500)
		expect(res.body.error).toMatchObject(
			{ "code": "io.direktiv.command.error", "message": "cat: executed: No such file or directory" })
	})
	it(`should execute both commands it via bash action from ${testWorkflow} workflow`, async () => {
		const res = await request(common.config.getDirektivHost())
			.post(`/api/v2/namespaces/${namespace}/instances?path=${testWorkflow}&wait=true`)
			.send({
				commands: [
					{
						command: 'touch executed',
						command: 'sleep 1 && cat executed'
					}
				]
			});
		expect(res.statusCode).toEqual(200)
		expect(res.body.return.bash).toMatchObject(
			[{ "result": "", "success": true }])
	})
})
