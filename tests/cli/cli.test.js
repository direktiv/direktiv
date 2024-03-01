import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import request from '../common/request'

const namespaceName = 'root'

const cliExecutable = 'direktivctl'

const { exec } = require('child_process')
const fs = require('fs')

const filepath = '/tests/cli/mockdata/direktiv-project'

const prefix = common.config.getDirektivHost().includes('http')
	? `-a ${ common.config.getDirektivHost() } -t password`
	: `-a http://${ common.config.getDirektivHost() } -t password`
const flagNamespace = `-n ${ namespaceName }`

describe('Test the direktiv-cli-tool', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	it(`TODO: enable this e2e tests.`, async () => {})
	return

	it(`create namespace`, async () => {
		const createResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${ namespaceName }`)
		expect(createResponse.statusCode).toEqual(200)
	})
	it(`test info show ns`, async () => {
		assertStdErrContainsString(`${ filepath }`, 'workflows info', `namespace: ${ namespaceName }`)
	})
	it(`test info project exists`, async () => {
		assertStdErrContainsString(`${ filepath }`, 'workflows info', `namespace: ${ namespaceName }`)
		assertStdErrContainsString(`${ filepath }`, 'workflows info', 'direktiv-project')
	})
	it(`test push project abs`, async () => {
		const filename = 'simplewf'
		const fileextension = 'yaml'
		await assertStdErrContainsString(`${ filepath }`, `workflows push ${ filepath }`, `pushing workflow ${ filename }`)
		const res = await request(common.config.getDirektivHost()).get(`/api/namespaces/root/tree/${ filename }.${ fileextension }`)
		expect(res.statusCode).toEqual(200)
	})
	it(`test push project .`, async () => {
		const filename = 'simplewf'
		const fileextension = 'yaml'
		await assertStdErrContainsString(`${ filepath }`, `workflows push .`, `pushing workflow ${ filename }`)
		const res = await request(common.config.getDirektivHost()).get(`/api/namespaces/root/tree/${ filename }.${ fileextension }`)
		expect(res.statusCode).toEqual(200)
	})
	it(`test push subpath relative pwd`, async () => {
		const filename = 'simplewfInSubfolder'
		const fileextension = 'yaml'
		const path = `${ filepath }/subfolder`
		await assertStdErrContainsString(`${ path }`, `workflows push .`, `pushing workflow subfolder/${ filename }`)
		const res = await request(common.config.getDirektivHost()).get(`/api/namespaces/root/tree/subfolder/${ filename }.${ fileextension }`)
		expect(res.statusCode).toEqual(200)
	})
	it(`test push subpath abs pwd`, async () => {
		const filename = 'simplewfInSubfolder'
		const fileextension = 'yaml'
		const path = `${ filepath }/subfolder`
		await assertStdErrContainsString(`${ path }`, `workflows push ${ path }/`, `pushing workflow subfolder/${ filename }`)
		const res = await request(common.config.getDirektivHost()).get(`/api/namespaces/root/tree/subfolder/${ filename }.${ fileextension }`)
		expect(res.statusCode).toEqual(200)
	})
})

const assertStdErrContainsString = (path, cmd, want) => new Promise((resolve, reject) => {
	exec(`cd ${ path } && ${ cliExecutable } ${ prefix } ${ flagNamespace } ${ cmd }`, (err, stdout, stderr) => {
		expect(stderr).toStrictEqual(expect.stringContaining(want))
		resolve()
	})
})
const assertStdOutContainsString = (path, cmd, want) => new Promise((resolve, reject) => {
	exec(`cd ${ path } && ${ cliExecutable } ${ prefix } ${ flagNamespace } ${ cmd }`, (err, stdout, stderr) => {
		expect(stdout).toStrictEqual(expect.stringContaining(want))
		resolve()
	})
})
const assertStdErrShouldNotContainsString = (path, cmd, want) => new Promise((resolve, reject) => {
	exec(`cd ${ path } && ${ cliExecutable } ${ prefix } ${ flagNamespace } ${ cmd }`, (err, stdout, stderr) => {
		expect(stderr).not.toStrictEqual(expect.stringContaining(want))
		resolve()
	})
})
