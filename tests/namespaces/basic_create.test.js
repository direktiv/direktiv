import { beforeAll, describe, expect, it } from '@jest/globals'

import config from '../common/config'
import helpers from '../common/helpers'
import regex from '../common/regex'
import request from '../common/request'

const timestamps = {
	createdAt: expect.stringMatching(regex.timestampRegex),
	updatedAt: expect.stringMatching(regex.timestampRegex),
}

describe('Test namespace create calls', () => {
	beforeAll(helpers.deleteAllNamespaces)

	const testCases = [
		{
			input: {
				name: 'foo1',
			},
			want: {
				name: 'foo1',
				mirror: null,
			},
		},
		{
			input: {
				name: 'foo2',
				mirror: {
					url: 'my_url',
					gitRef: 'main',
				},
			},
			want: {
				name: 'foo2',
				mirror: {
					url: 'my_url',
					gitRef: 'main',
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
				},
			},
			want: {
				name: 'foo3',
				mirror: {
					url: 'my_url',
					insecure: true,
					gitRef: 'master',
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
