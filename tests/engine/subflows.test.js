import { beforeAll, describe, expect, it } from '@jest/globals'
import { btoa } from 'js-base64'
import { basename } from 'path'
import { fileURLToPath } from 'url'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

function randomLowercaseString(length) {
	const letters = 'abcdefghijklmnopqrstuvwxyz'
	let result = ''
	for (let i = 0; i < length; i++) {
		result += letters[Math.floor(Math.random() * 26)]
	}
	return result
}

const namespace =
	randomLowercaseString(3) + '-' + basename(fileURLToPath(import.meta.url))

describe('Test Subflows', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	const subflowFiles = [
		{
			name: 'okSubflow.wf.ts',
			file: `
		function stateOne(payload) {
			let events = payload.events;
			if (events == undefined) {
				events = [];
			}
			events.push("processed on okSubflow.wf.ts -> stateOne");
			payload.events = events;
			
			return finish(payload);
		}
`},
		{
			name: 'errorSubflow.wf.ts',
			file: `
		function stateOne(payload) {
			throw "error in errorSubflow.wf.ts -> stateOne";
			return finish(payload);
		}
`,
		},
	]

	for (let i = 0; i < subflowFiles.length; i++) {
		helpers.itShouldCreateFile(
			it,
			expect,
			namespace,
			'/',
			subflowFiles[i].name,
			'workflow',
			'application/x-typescript',
			btoa(subflowFiles[i].file),
		)
	}


	const testCases = [
		{
			name: 'basic.wf.ts',
			input: { foo: 'bar' },
			wantOutput: { foo: 'bar', events:[
				"processed on okSubflow.wf.ts -> stateOne",
			]},
			wantErrorMessage: null,
			wantStatus: 'complete',
			file: `
		function stateOne(payload) {
			let result = execSubflow("/okSubflow.wf.ts", payload);
			return finish(result);
		}
`,
		},
	]

	for (let i = 0; i < testCases.length; i++) {
		const testCase = testCases[i]
		helpers.itShouldCreateFile(
			it,
			expect,
			namespace,
			'/',
			testCase.name,
			'workflow',
			'application/x-typescript',
			btoa(testCase.file),
		)
		it(`should invoke /${testCase.name} workflow with &wait=true&fullOutput=true`, async () => {
			const res = await request(common.config.getDirektivBaseUrl())
				.post(
					`/api/v2/namespaces/${namespace}/instances?path=/${testCase.name}&wait=true&fullOutput=true`,
				)
				.send(testCase.input)
			console.log(res.body)
			expect(res.statusCode).toEqual(200)
			expect(res.body.data.status).toEqual(testCase.wantStatus)
			expect(res.body.data.errorMessage).toEqual(testCase.wantErrorMessage)
			expect(JSON.parse(res.body.data.output)).toEqual(testCase.wantOutput)
		})
	}


})
