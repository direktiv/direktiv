import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import common from '../common'
import config from '../common/config'
import helpers from '../common/helpers'
import regex from '../common/regex'
import request from '../common/request'

const namespace = basename(__filename)

describe('Test variable update calls', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	let createRes
	it(`should create a variable case`, async () => {
		createRes = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces/${ namespace }/variables`)
			.send({
				name: 'foo',
				data: btoa('bar'),
				mimeType: 'mime',
			})

		expect(createRes.statusCode).toEqual(200)
	})

	const testCases = [
		{
			input: {
				name: 'foo1',
			},
			want: {
				name: 'foo1',
				data: btoa('bar'),
				mimeType: 'mime',
				size: 3,
				type: 'namespace_variable',
				reference: namespace,
			},
		},
		{
			input: {
				data: btoa('bar2--'),
			},
			want: {
				name: 'foo1',
				data: btoa('bar2--'),
				mimeType: 'mime',
				size: 6,
				type: 'namespace_variable',
				reference: namespace,
			},
		},
		{
			input: {
				mimeType: 'mime2',
			},
			want: {
				name: 'foo1',
				data: btoa('bar2--'),
				mimeType: 'mime2',
				size: 6,
				type: 'namespace_variable',
				reference: namespace,
			},
		},
		{
			input: {
				name: 'foo3',
				mimeType: 'mime3',
			},
			want: {
				name: 'foo3',
				data: btoa('bar2--'),
				mimeType: 'mime3',
				size: 6,
				type: 'namespace_variable',
				reference: namespace,
			},
		},
		{
			input: {
				name: 'foo4',
				data: btoa('bar4--'),
				mimeType: 'mime4',
			},
			want: {
				name: 'foo4',
				data: btoa('bar4--'),
				mimeType: 'mime4',
				size: 6,
				type: 'namespace_variable',
				reference: namespace,
			},
		},
	]

	for (let i = 0; i < testCases.length; i++) {
		const testCase = testCases[i]

		// eslint-disable-next-line no-loop-func
		it(`should update variable case ${ i }`, async () => {
			const varId = createRes.body.data.id
			const res = await request(config.getDirektivHost())
				.patch(`/api/v2/namespaces/${ namespace }/variables/${ varId }`)
				.send(testCase.input)
			expect(res.statusCode).toEqual(200)
			expect(res.body.data).toEqual({
				id: expect.stringMatching(common.regex.uuidRegex),

				...testCase.want,

				createdAt: expect.stringMatching(regex.timestampRegex),
				updatedAt: expect.stringMatching(regex.timestampRegex),
			})
		})
	}
})

describe('Test invalid variable update calls', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	let createRes
	it(`should create a variable case`, async () => {
		createRes = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces/${ namespace }/variables`)
			.send({
				name: 'foo',
				data: btoa('bar'),
				mimeType: 'mime',
			})

		expect(createRes.statusCode).toEqual(200)
	})

	const testCases = [
		{
			input: {
				// invalid name
				name: '1',
			},
			wantError: {
				statusCode: 400,
				error: {
					code: 'request_data_invalid',
					message: 'field name has invalid string',
				},
			},
		},
		{
			input: {
				// invalid data
				data: 'some string',
			},
			wantError: {
				statusCode: 400,
				error: {
					code: 'request_body_not_json',
					message: "couldn't parse request payload in json format",
				},
			},
		},
	]

	for (let i = 0; i < testCases.length; i++) {
		const testCase = testCases[i]

		// eslint-disable-next-line no-loop-func
		it(`should fail updating a variable case ${ i }`, async () => {
			const varId = createRes.body.data.id
			const res = await request(config.getDirektivHost())
				.patch(`/api/v2/namespaces/${ namespace }/variables/${ varId }`)
				.send(testCase.input)
			expect(res.statusCode).toEqual(testCase.wantError.statusCode)
			expect(res.body.error).toEqual(
				testCase.wantError.error,
			)
		})
	}
})
