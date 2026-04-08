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

describe('Test action timeout feature', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	const testCases = [
		{
			name: 'action-timeout-fail.wf.ts',
			input: {},
			wantStatus: 'failed',
			wantErrorMessage: expect.stringContaining('timeout'),
			file: `
		var echo = generateAction({
			image: "ubuntu:24.04",
			cmd: "/usr/share/direktiv/direktiv-cmd",
			size: "small",
			timeout: "PT2S",
			envs: [
				{
					name: "MY_VAR",
					value: "test",
				},
			],
			retries: 0,
		});

		function stateFirst(input) {
			var payload = {
				commands: [
					{
						command: "sleep 4",
					},
				],
			};

			let result = echo(payload);

			return transition(stateSecond, result);
		}

		function stateSecond(input) {
			return finish(input);
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
		it(`action with timeout PT2S should timeout when command takes 4s`, async () => {
			const res = await request(common.config.getDirektivBaseUrl())
				.post(
					`/api/v2/namespaces/${namespace}/instances?path=/${testCase.name}&wait=true&fullOutput=true`,
				)
				.send(testCase.input)
			expect(res.statusCode).toEqual(200)
			expect(res.body.data.status).toEqual(testCase.wantStatus)
			expect(res.body.data.errorMessage).toEqual(testCase.wantErrorMessage)
		})
	}
})