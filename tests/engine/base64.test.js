import { beforeAll, describe, expect, it } from '@jest/globals'

import { basename } from 'path'
import { btoa } from 'js-base64'
import common from '../common'
import { fileURLToPath } from 'url'
import helpers from '../common/helpers'
import request from '../common/request'

const namespace =
	helpers.randomLowercaseString(3) +
	'-' +
	basename(fileURLToPath(import.meta.url))

describe('Test base64 functions', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	const testCases = [
		{
			name: 'base64_encode.wf.ts',
			file: `
		function stateOne(payload) {
			let encoded = base64Encode("hello world");
			return finish(encoded);
		}
`,
			expectedOutput: 'aGVsbG8gd29ybGQ=',
		},
		{
			name: 'base64_decode.wf.ts',
			file: `
		function stateOne(payload) {
			let decoded = base64Decode("aGVsbG8gd29ybGQ=");
			return finish(decoded);
		}
`,
			expectedOutput: 'hello world',
		},
		{
			name: 'base64_roundtrip.wf.ts',
			file: `
		function stateOne(payload) {
			let original = "test input data";
			let encoded = base64Encode(original);
			let decoded = base64Decode(encoded);
			return finish(decoded);
		}
`,
			expectedOutput: 'test input data',
		},
		{
			name: 'base64_invalid.wf.ts',
			file: `
		function stateOne(payload) {
			let decoded = base64Decode("not-valid!!!");
			return finish(decoded);
		}
`,
			wantError: true,
			errorContains: 'invalid base64',
		},
	]

	for (const testCase of testCases) {
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

		if (testCase.wantError) {
			it(`should error on invalid base64 in /${testCase.name}`, async () => {
				const res = await request(common.config.getDirektivBaseUrl())
					.post(
						`/api/v2/namespaces/${namespace}/instances?path=/${testCase.name}&wait=true`,
					)
					.send({})
				expect([200, 500]).toContain(res.statusCode)
				if (res.statusCode === 200) {
					expect(res.body.data.status).toEqual('failed')
					expect(res.body.data.errorMessage).toContain(testCase.errorContains)
				}
			})
		} else {
			it(`should execute /${testCase.name} and return "${testCase.expectedOutput}"`, async () => {
				const res = await request(common.config.getDirektivBaseUrl())
					.post(
						`/api/v2/namespaces/${namespace}/instances?path=/${testCase.name}&wait=true&fullOutput=true`,
					)
					.send({})
				expect(res.statusCode).toEqual(200)
				expect(res.body.data.status).toEqual('complete')
				expect(res.body.data.errorMessage).toBeNull()
				expect(res.body.data.output).toEqual(
					JSON.stringify(testCase.expectedOutput),
				)
			})
		}
	}
})
