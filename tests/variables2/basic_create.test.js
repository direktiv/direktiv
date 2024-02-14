import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'
import request from "../common/request"

import common from '../common'
import config from '../common/config'
import helpers from '../common/helpers'
import regex from '../common/regex'

const namespace = basename(__filename)

describe('Test variable create calls', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateFileV2(it, expect, namespace, '/', 'wf1.yaml', 'workflow', 'text',
		btoa(helpers.dummyWorkflow('wf1.yaml')))

	const testCases = [
		{
			input: {
				name: 'foo1',
				data: btoa('bar1'),
				mimeType: 'mime1',
			},
			want: {
				name: 'foo1',
				data: btoa('bar1'),
				mimeType: 'mime1',
				size: 4,
				instanceId: '00000000-0000-0000-0000-000000000000',
				workflowPath: '',
			},
		},
		{
			input: {
				name: 'foo2',
				data: btoa('bar2'),
				mimeType: 'mime2',
				size: 4,
			},
			want: {
				name: 'foo2',
				data: btoa('bar2'),
				mimeType: 'mime2',
				size: 4,
				instanceId: '00000000-0000-0000-0000-000000000000',
				workflowPath: '',
			},
		},
		{
			input: {
				name: 'foo3',
				data: btoa('bar3'),
				mimeType: 'mime3',
			},
			want: {
				name: 'foo3',
				data: btoa('bar3'),
				mimeType: 'mime3',
				size: 4,
				instanceId: '00000000-0000-0000-0000-000000000000',
				workflowPath: '',
			},
		},
		{
			input: {
				name: 'foo4',
				data: btoa('bar4'),
				mimeType: 'mime4',
				workflowPath: '/wf1.yaml',
			},
			want: {
				name: 'foo4',
				data: btoa('bar4'),
				mimeType: 'mime4',
				size: 4,
				instanceId: '00000000-0000-0000-0000-000000000000',
				workflowPath: '/wf1.yaml',
			},
		},
	]

	for (let i = 0; i < testCases.length; i++) {
		const testCase = testCases[i]

		it(`should create a new variable case ${ i }`, async () => {
			const res = await request(config.getDirektivHost())
				.post(`/api/v2/namespaces/${ namespace }/variables`)
				.send(testCase.input)
			expect(res.statusCode).toEqual(200)
			expect(res.body.data).toMatchObject({
				id: expect.stringMatching(common.regex.uuidRegex),
				namespace,

				...testCase.want,

				createdAt: expect.stringMatching(regex.timestampRegex),
				updatedAt: expect.stringMatching(regex.timestampRegex),
			})
		})
	}
})

describe('Test invalid variable create calls', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	const testCases = [
		{
			input: {
				// invalid data
				name: 'foo1',
				data: 'invalid-base-64',
				mimeType: 'mime1',
			},
			wantError: {
				statusCode: 400,
				error: {
					code: 'request_body_not_json',
					message: "couldn't parse request payload in json format",
				},
			},
		},
		{
			input: {
				// invalid name
				name: '1',
				data: btoa('bar'),
				mimeType: 'mime1',
			},
			wantError: {
				statusCode: 400,
				error: {
					code: 'request_data_invalid',
					message: 'field name has invalid string',
				},
			},
		},
	]

	for (let i = 0; i < testCases.length; i++) {
		const testCase = testCases[i]

		it(`should fail create a new variable case ${ i }`, async () => {
			const res = await request(config.getDirektivHost())
				.post(`/api/v2/namespaces/${ namespace }/variables`)
				.send(testCase.input)
			expect(res.statusCode).toEqual(testCase.wantError.statusCode)
			expect(res.body.error).toMatchObject(
				testCase.wantError.error,
			)
		})
	}
})
