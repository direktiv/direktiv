import { beforeAll, describe, expect, it } from '@jest/globals'

import config from '../../common/config'
import helpers from '../../common/helpers'
import regex from '../../common/regex'
import request from '../../common/request'

const timestamps = {
	createdAt: expect.stringMatching(regex.timestampRegex),
	updatedAt: expect.stringMatching(regex.timestampRegex),
}

describe('Test namespace create calls', () => {
	beforeAll(helpers.deleteAllNamespaces)

	const testCases = [
		{
			input: {
				name: 'system',
			},
			want: {
				name: 'system',
				isSystemNamespace: true,
				mirror: null,
			},
		},
		{
			input: {
				name: 'api',
			},
			want: {
				name: 'api',
				isSystemNamespace: false,
				mirror: null,
			},
		},
		{
			input: {
				name: 'v1',
			},
			want: {
				name: 'v1',
				isSystemNamespace: false,
				mirror: null,
			},
		},
		{
			input: {
				name: 'v2',
			},
			want: {
				name: 'v2',
				isSystemNamespace: false,
				mirror: null,
			},
		},
		{
			input: {
				name: 'namespace',
			},
			want: {
				name: 'namespace',
				isSystemNamespace: false,
				mirror: null,
			},
		},
		{
			input: {
				name: 'namespaces',
			},
			want: {
				name: 'namespaces',
				isSystemNamespace: false,
				mirror: null,
			},
		},
		{
			input: {
				name: 'foo1',
			},
			want: {
				name: 'foo1',
				isSystemNamespace: false,
				mirror: null,
			},
		},
		{
			input: {
				name: 'foo2',
				mirror: {
					url: 'my_url',
					gitRef: 'main',
					authType: 'public',
				},
			},
			want: {
				name: 'foo2',
				isSystemNamespace: false,
				mirror: {
					url: 'my_url',
					gitRef: 'main',
					authType: 'public',
					insecure: false,
					...timestamps,
				},
			},
		},
		{
			input: {
				name: 'foo3',
				mirror: {
					url: 'my_url',
					insecure: true,
					gitRef: 'master',
					authType: 'public',
				},
			},
			want: {
				name: 'foo3',
				isSystemNamespace: false,
				mirror: {
					url: 'my_url',
					insecure: true,
					gitRef: 'master',
					authType: 'public',
					...timestamps,
				},
			},
		},
		{
			input: {
				name: 'foo4',
				mirror: {
					url: 'my_url',
					insecure: true,
					gitRef: 'master',
					authType: 'token',
					authToken: '12345',
				},
			},
			want: {
				name: 'foo4',
				isSystemNamespace: false,
				mirror: {
					url: 'my_url',
					insecure: true,
					gitRef: 'master',
					authType: 'token',
					...timestamps,
				},
			},
		},
		{
			input: {
				name: 'foo5',
				mirror: {
					url: 'my_url',
					insecure: true,
					gitRef: 'master',
					authType: 'ssh',
					publicKey: 'my-public-key',
					privateKey: 'my-private-key',
				},
			},
			want: {
				name: 'foo5',
				isSystemNamespace: false,
				mirror: {
					url: 'my_url',
					insecure: true,
					gitRef: 'master',
					publicKey: 'my-public-key',
					authType: 'ssh',
					...timestamps,
				},
			},
		},
	]

	for (let i = 0; i < testCases.length; i++) {
		const testCase = testCases[i]

		it(`should create a new namespace case ${ i }`, async () => {
			const res = await request(config.getDirektivHost())
				.post(`/api/v2/namespaces`)
				.send(testCase.input)
			expect(res.statusCode).toEqual(200)
			expect(res.body.data).toEqual({
				...testCase.want,
				...timestamps,
			})
		})
	}
})

describe('Test invalid namespace create calls', () => {
	beforeAll(helpers.deleteAllNamespaces)

	const testCases = [
		{
			input: {
				// invalid data
				name: '11',
			},
			wantError: {
				statusCode: 400,
				error: {
					code: 'request_data_invalid',
					message: 'invalid namespace name',
				},
			},
		},
	]

	for (let i = 0; i < testCases.length; i++) {
		const testCase = testCases[i]

		it(`should fail create a new namespace case ${ i }`, async () => {
			const res = await request(config.getDirektivHost())
				.post(`/api/v2/namespaces`)
				.send(testCase.input)
			expect(res.statusCode).toEqual(testCase.wantError.statusCode)
			expect(res.body.error).toEqual(
				testCase.wantError.error,
			)
		})
	}
})

describe('Test invalid namespace name', () => {
	beforeAll(helpers.deleteAllNamespaces)

	const testCases = [
		'test-flow-namespace-regex-A',
		'Test-flow-namespace-regex-a',
		'test-flow-namespace-reGex-a',
		'1test-flow-namespace-regex-a',
		'.test-flow-namespace-regex-a',
		'_test-flow-namespace-regex-a',
		'test-flow-namespace-regex-a_',
		'test-flow-namespace-regex-a.',
		'test-flow-namespace@regex-a',
		'test-flow-namespace+regex-a',
		'test-flow-namespace regex-a',
		'test-flow-namespace%20regex-a',
	]

	for (let i = 0; i < testCases.length; i++) {
		const testCase = testCases[i]

		it(`should fail create a new namespace case ${ i }`, async () => {
			const res = await request(config.getDirektivHost())
				.post(`/api/v2/namespaces`)
				.send({
					name: testCase,
				})
			expect(res.statusCode).toEqual(400)
			expect(res.body.error).toEqual({
				code: 'request_data_invalid',
				message: 'invalid namespace name',
			})
		})
	}
})

describe('Test valid namespace name', () => {
	beforeAll(helpers.deleteAllNamespaces)

	const testCases = [
		'a',
		'test-flow-namespace-regex-a',
		'test-flow-namespace-regex-1',
		'test-flow-namespace-regex-a_b.c',
	]

	for (let i = 0; i < testCases.length; i++) {
		const testCase = testCases[i]

		it(`should fail create a new namespace case ${ i }`, async () => {
			const res = await request(config.getDirektivHost())
				.post(`/api/v2/namespaces`)
				.send({
					name: testCase,
				})
			expect(res.statusCode).toEqual(200)
		})
	}
})

describe('Test error cases', () => {
	beforeAll(helpers.deleteAllNamespaces)

	it(`should create foo namespace`, async () => {
		const res = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces`)
			.send({
				name: 'foo',
			})
		expect(res.statusCode).toEqual(200)
	})

	it(`should fail create foo namespace`, async () => {
		const res = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces`)
			.send({
				name: 'foo',
			})
		expect(res.statusCode).toEqual(400)
		expect(res.body.error).toEqual({
			code: 'request_data_invalid',
			message: 'namespace name already used',
		})
	})
})

describe('Test missing fields create calls', () => {
	beforeAll(helpers.deleteAllNamespaces)

	const testCases = [
		{
			mirror: {
				url: 'my_url',
				gitRef: 'main',
			},
		},
		{
			name: 'foo4',
			mirror: {
				gitRef: 'main',
			},
		},
		{
			name: 'foo4',
			mirror: {
				url: 'my_url',
			},
		} ]

	for (let i = 0; i < testCases.length; i++) {
		const testCase = testCases[i]

		it(`should fail create a new namespace case ${ i }`, async () => {
			const res = await request(config.getDirektivHost())
				.post(`/api/v2/namespaces`)
				.send(testCase)
			// expect(res.statusCode).toEqual(400)
			expect(res.body.error.code).toEqual('request_data_invalid')
		})
	}
})
