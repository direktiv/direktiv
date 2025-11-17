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

describe('Test js engine', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	const testCases = [
		{
			name: 'basicAction.wf.ts',
			input: { foo: 'bar' },
			wantOutput: JSON.stringify({ foo: 'bar', input: { foo: 'bar' } }),
			wantErrorMessage: null,
			wantStatus: 'complete',
			file: `
		var echo = generateAction({
			type: "local",
			size: "medium",
			image: "mendhak/http-https-echo:latest"
		});
		function stateOne(payload) {
			let result = echo({ 
				body: {foo: "bar", input: payload}
				headers: {
				"content-type": "application/json"
				}
			});
			return finish(result.json);
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
			expect(res.statusCode).toEqual(200)
			expect(res.body.data.status).toEqual(testCase.wantStatus)
			expect(res.body.data.errorMessage).toEqual(testCase.wantErrorMessage)
			expect(res.body.data.output).toEqual(testCase.wantOutput)
		})
	}
})
