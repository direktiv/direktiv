import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'
import {retry, retry10} from '../common/retry'

function quantile (arr, q) {
	if (!arr.length) return NaN
	const a = [ ...arr ].sort((x, y) => x - y)
	const pos = (a.length - 1) * q
	const base = Math.floor(pos)
	const rest = pos - base
	return a[base] + (a[base + 1] - a[base]) * (rest || 0)
}

const randomStr = Math.random().toString(10).slice(2, 12)
const namespace = basename(__filename.replaceAll('.', '-')) + randomStr
const fName = 'file' + randomStr + '.wf.js'

describe('Stress test js engine', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)
	helpers.itShouldCreateDir(it, expect, namespace, '/', 'foo')

	helpers.itShouldCreateFile(it, expect, namespace, '/foo', fName, 'file', 'text/plain',
		btoa(`
function stateOne(payload) {
	print("RUN STATE FIRST");
    return transition(stateTwo, payload);
}
function stateTwo(payload) {
	print("RUN STATE SECOND");
    return finish(payload);
}
`))

	retry10(`should invoke /foo/${ fName } workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespace }/instances?path=foo/${ fName }`)
			.send({ foo: 'bar' })
		expect(req.statusCode).toEqual(200)
	})

	const cases = [
		{
			total: 10,
			batchSize: 2,
		},
		{
			total: 50,
			batchSize: 5,
		},
		{
			total: 100,
			batchSize: 10,
		},
		{
			total: 1000,
			batchSize: 100,
		},
	]
	for (let i = 0; i < cases.length; i++) {
		const total = cases[i].total
		const batchSize = cases[i].batchSize

		it(`fires ${ total } requests in ${ batchSize } batches`, async () => {
			const results = []

			const durations = [] // ms
			let ok = 0,
				fail = 0

			for (let start = 0; start < total; start += batchSize) {
				const batch = Array.from({ length: batchSize }, (_, i) => {
					const url = common.config.getDirektivHost() + `/api/v2/namespaces/${ namespace }/instances?path=foo/${ fName }`

					const t0 = performance.now()
					return fetch(url, {
						method: 'POST',
						headers: { 'Content-Type': 'application/json' },
						body: JSON.stringify({ foo: 'bar' }),
					}).then(res => {
						const t1 = performance.now()
						durations.push(t1 - t0)
						if (res.ok) ok++; else fail++
						return res.status
					})
						.catch(err => {
							const t1 = performance.now()
							durations.push(t1 - t0)
							fail++
							return 0 // mark as failed
						})
				})

				// run batch concurrently
				const statuses = await Promise.all(batch)
				results.push(...statuses)

				console.log(`Batch done: ${ start + batchSize }/${ total }`)
			}

			const sum = durations.reduce((a, b) => a + b, 0)
			const avg = sum / durations.length // average response time (ms)
			const min = Math.min(...durations)
			const max = Math.max(...durations)
			const p50 = quantile(durations, 0.50)
			const p90 = quantile(durations, 0.90)
			const p95 = quantile(durations, 0.95)
			const p99 = quantile(durations, 0.99)

			console.log(JSON.stringify(results))

			console.log('\nLatency (ms):',
				{ count: durations.length, avg: +avg.toFixed(2), min: +min.toFixed(2),
					p50: +p50.toFixed(2), p90: +p90.toFixed(2), p95: +p95.toFixed(2),
					p99: +p99.toFixed(2), max: +max.toFixed(2), ok, fail })

			// Assertions: all requests should be 200
			results.forEach((status, i) => {
				expect(status).toBe(200) // or 201 depending on your API
			})

			expect(fail).toBe(0)
			expect(avg).toBeLessThan(300) // e.g., average < 300ms
			expect(p95).toBeLessThan(600) // e.g., p95 < 600ms
		}, 60000) // extend timeout for big tests
	}

	retry(`should have all success instances`, 100,async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/instances/stats`)
		console.log(req.body.data)
		expect(req.statusCode).toEqual(200)
		expect(req.body.data).toEqual({succeeded: 1161})
	}, 1000)
})
