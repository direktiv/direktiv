import { beforeAll, describe, expect, it } from '@jest/globals'
import { btoa } from 'js-base64'
import { basename } from 'path'
import { fileURLToPath } from 'url'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const namespace = basename(fileURLToPath(import.meta.url))

describe('List workflow success history', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateFile(
		it,
		expect,
		namespace,
		'/',
		'foo.wf.ts',
		'workflow',
		'application/x-typescript',
		btoa(`
function stateOne(payload) {
	print("RUN STATE FIRST");
	payload.one = 1;
	return transition(stateTwo, payload);
}

function stateTwo(payload) {
	print("RUN STATE SECOND");
	payload.two = 2;
    return transition(stateThree, payload);
}

function stateThree(payload) {
	print("RUN STATE THIRD");
	payload.three = 3;
    return finish(payload);
}`),
	)
	let instanceId = null

	it(`should invoke /foo.wf.ts workflow`, async () => {
		const res = await request(common.config.getDirektivBaseUrl())
			.post(
				`/api/v2/namespaces/${namespace}/instances?path=/foo.wf.ts&wait=true&fullOutput=true`,
			)
			.send({ foo: 'bar' })
		expect(res.statusCode).toEqual(200)
		instanceId = res.body.data.id
	})

	it(`should list /foo.wf.ts workflow history`, async () => {
		const res = await request(common.config.getDirektivBaseUrl()).get(
			`/api/v2/namespaces/${namespace}/instances/${instanceId}/history`,
		)
		expect(res.statusCode).toEqual(200)
		console.log(res.body.data)
		const history = res.body.data.map((item) => ({
			type: item.state,
			scope: item.metadata.WithScope,
			fn: item.fn,
			input: item.input,
			output: item.output,
			error: item.error,
			sequence: item.sequence,
		}))
		console.log(history)
		let firstSequence = history[0].sequence
		expect(history).toEqual([
			{
				scope: 'main',
				type: 'pending',
				fn: 'stateOne',
				input: { foo: 'bar' },
				output: undefined,
				sequence: firstSequence++,
			},
			{
				scope: 'main',
				type: 'running',
				fn: 'stateOne',
				input: { foo: 'bar' },
				output: undefined,
				sequence: firstSequence++,
			},
			{
				scope: 'main',
				type: 'running',
				fn: 'stateTwo',
				input: { foo: 'bar' },
				output: { foo: 'bar', one: 1 },
				sequence: firstSequence++,
			},
			{
				scope: 'main',
				type: 'running',
				fn: 'stateThree',
				input: { foo: 'bar' },
				output: { foo: 'bar', one: 1, two: 2 },
				sequence: firstSequence++,
			},
			{
				scope: 'main',
				type: 'complete',
				fn: undefined,
				input: { foo: 'bar' },
				output: { foo: 'bar', one: 1, three: 3, two: 2 },
				sequence: firstSequence++,
			},
		])
	})
})

describe('List workflow success with subflow history', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateFile(
		it,
		expect,
		namespace,
		'/',
		'subflow.wf.ts',
		'workflow',
		'application/x-typescript',
		btoa(`
function stateOne(payload) {
	payload.subflowOne = 1;
	return transition(stateTwo, payload);
}
function stateTwo(payload) {
	payload.subflowTwo = 2;
    return finish(payload);
}
`),
	)

	helpers.itShouldCreateFile(
		it,
		expect,
		namespace,
		'/',
		'main.wf.ts',
		'workflow',
		'application/x-typescript',
		btoa(`
function stateOne(payload) {
	payload.mainOne = 1;
	let newPayload = execSubflow("/subflow.wf.ts", payload);
	
	return transition(stateTwo, newPayload);
}
function stateTwo(payload) {
	payload.mainTwo = 2;
    return finish(payload);
}
`),
	)

	let instanceId = null

	it(`should invoke /foo.wf.ts workflow`, async () => {
		const res = await request(common.config.getDirektivBaseUrl())
			.post(
				`/api/v2/namespaces/${namespace}/instances?path=/main.wf.ts&wait=true&fullOutput=true`,
			)
			.send({ foo: 'bar' })
		expect(res.statusCode).toEqual(200)
		instanceId = res.body.data.id
	})

	it(`should list /main.wf.ts workflow history`, async () => {
		const res = await request(common.config.getDirektivBaseUrl()).get(
			`/api/v2/namespaces/${namespace}/instances/${instanceId}/history`,
		)
		expect(res.statusCode).toEqual(200)
		console.log(res.body.data)
		const history = res.body.data.map((item) => ({
			type: item.state,
			scope: item.metadata.WithScope,
			fn: item.fn,
			input: item.input,
			output: item.output,
			error: item.error,
			sequence: item.sequence,
		}))
		console.log(history)
		let firstSequence = history[0].sequence
		let subflowID = history[2].scope
		expect(history).toEqual([
			{
				scope: 'main',
				type: 'pending',
				fn: 'stateOne',
				input: { foo: 'bar' },
				output: undefined,
				sequence: firstSequence++,
			},
			{
				scope: 'main',
				type: 'running',
				fn: 'stateOne',
				input: { foo: 'bar' },
				output: undefined,
				sequence: firstSequence++,
			},
			{
				type: 'pending',
				scope: subflowID,
				fn: 'stateOne',
				input: { foo: 'bar', mainOne: 1 },
				output: undefined,
				sequence: firstSequence++,
			},
			{
				type: 'running',
				scope: subflowID,
				fn: 'stateOne',
				input: { foo: 'bar', mainOne: 1 },
				output: undefined,
				sequence: firstSequence++,
			},
			{
				type: 'running',
				scope: subflowID,
				fn: 'stateTwo',
				input: { foo: 'bar', mainOne: 1 },
				output: { foo: 'bar', mainOne: 1, subflowOne: 1 },
				sequence: firstSequence++,
			},
			{
				type: 'complete',
				scope: subflowID,
				fn: undefined,
				input: { foo: 'bar', mainOne: 1 },
				output: { foo: 'bar', mainOne: 1, subflowOne: 1, subflowTwo: 2 },
				sequence: firstSequence++,
			},
			{
				scope: 'main',
				type: 'running',
				fn: 'stateTwo',
				input: { foo: 'bar' },
				output: { foo: 'bar', mainOne: 1, subflowOne: 1, subflowTwo: 2 },
				sequence: firstSequence++,
			},
			{
				scope: 'main',
				type: 'complete',
				fn: undefined,
				input: { foo: 'bar' },
				output: {
					foo: 'bar',
					mainOne: 1,
					subflowOne: 1,
					subflowTwo: 2,
					mainTwo: 2,
				},
				sequence: firstSequence++,
			},
		])
	})
})

describe('List workflow failed at (two) history', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateFile(
		it,
		expect,
		namespace,
		'/',
		'foo.wf.ts',
		'workflow',
		'application/x-typescript',
		btoa(`
function stateOne(payload) {
	print("RUN STATE FIRST");
	payload.one = 1;
	return transition(stateTwo, payload);
}

function stateTwo(payload) {
	threw new Error("logic failed");
	print("RUN STATE SECOND");
    return transition(stateThree, payload);
}

function stateThree(payload) {
	print("RUN STATE THIRD");
	payload.three = 3;
    return finish(payload);
}`),
	)
	let instanceId = null

	it(`should invoke /foo.wf.ts workflow`, async () => {
		const res = await request(common.config.getDirektivBaseUrl())
			.post(
				`/api/v2/namespaces/${namespace}/instances?path=/foo.wf.ts&wait=true&fullOutput=true`,
			)
			.send({ foo: 'bar' })
		expect(res.statusCode).toEqual(200)
		instanceId = res.body.data.id
	})

	it(`should list /foo.wf.ts workflow history`, async () => {
		const res = await request(common.config.getDirektivBaseUrl()).get(
			`/api/v2/namespaces/${namespace}/instances/${instanceId}/history`,
		)
		expect(res.statusCode).toEqual(200)
		console.log(res.body.data)
		const history = res.body.data.map((item) => ({
			type: item.state,
			scope: item.metadata.WithScope,
			fn: item.fn,
			input: item.input,
			output: item.output,
			error: item.error,
			sequence: item.sequence,
		}))
		console.log(history)
		let firstSequence = history[0].sequence
		expect(history).toEqual([
			{
				scope: 'main',
				type: 'pending',
				fn: 'stateOne',
				input: { foo: 'bar' },
				output: undefined,
				error: undefined,
				sequence: firstSequence++,
			},
			{
				scope: 'main',
				type: 'running',
				fn: 'stateOne',
				input: { foo: 'bar' },
				output: undefined,
				error: undefined,
				sequence: firstSequence++,
			},
			{
				scope: 'main',
				type: 'running',
				fn: 'stateTwo',
				input: { foo: 'bar' },
				output: { foo: 'bar', one: 1 },
				error: undefined,
				sequence: firstSequence++,
			},
			{
				scope: 'main',
				type: 'failed',
				fn: undefined,
				input: { foo: 'bar' },
				output: undefined,
				error:
					'invoke start: ReferenceError: threw is not defined at stateTwo (foo.wf.ts:9:1(1))',
				sequence: firstSequence++,
			},
		])
	})
})

describe('List workflow failed at (three) history', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateFile(
		it,
		expect,
		namespace,
		'/',
		'foo.wf.ts',
		'workflow',
		'application/x-typescript',
		btoa(`
function stateOne(payload) {
	print("RUN STATE FIRST");
	payload.one = 1;
	return transition(stateTwo, payload);
}

function stateTwo(payload) {
	print("RUN STATE SECOND");
	payload.two = 2;
    return transition(stateThree, payload);
}

function stateThree(payload) {
	threw new Error("logic failed");
	print("RUN STATE THIRD");
	payload.three = 3;
    return finish(payload);
}`),
	)
	let instanceId = null

	it(`should invoke /foo.wf.ts workflow`, async () => {
		const res = await request(common.config.getDirektivBaseUrl())
			.post(
				`/api/v2/namespaces/${namespace}/instances?path=/foo.wf.ts&wait=true&fullOutput=true`,
			)
			.send({ foo: 'bar' })
		expect(res.statusCode).toEqual(200)
		instanceId = res.body.data.id
	})

	it(`should list /foo.wf.ts workflow history`, async () => {
		const res = await request(common.config.getDirektivBaseUrl()).get(
			`/api/v2/namespaces/${namespace}/instances/${instanceId}/history`,
		)
		expect(res.statusCode).toEqual(200)
		console.log(res.body.data)
		const history = res.body.data.map((item) => ({
			type: item.state,
			scope: item.metadata.WithScope,
			fn: item.fn,
			input: item.input,
			output: item.output,
			error: item.error,
			sequence: item.sequence,
		}))
		console.log(history)
		let firstSequence = history[0].sequence
		expect(history).toEqual([
			{
				scope: 'main',
				type: 'pending',
				fn: 'stateOne',
				input: { foo: 'bar' },
				output: undefined,
				error: undefined,
				sequence: firstSequence++,
			},
			{
				scope: 'main',
				type: 'running',
				fn: 'stateOne',
				input: { foo: 'bar' },
				output: undefined,
				error: undefined,
				sequence: firstSequence++,
			},
			{
				scope: 'main',
				type: 'running',
				fn: 'stateTwo',
				input: { foo: 'bar' },
				output: { foo: 'bar', one: 1 },
				error: undefined,
				sequence: firstSequence++,
			},
			{
				scope: 'main',
				type: 'running',
				fn: 'stateThree',
				input: { foo: 'bar' },
				output: { foo: 'bar', one: 1, two: 2 },
				error: undefined,
				sequence: firstSequence++,
			},
			{
				scope: 'main',
				type: 'failed',
				fn: undefined,
				input: { foo: 'bar' },
				output: undefined,
				error:
					'invoke start: ReferenceError: threw is not defined at stateThree (foo.wf.ts:15:1(1))',
				sequence: firstSequence++,
			},
		])
	})
})
