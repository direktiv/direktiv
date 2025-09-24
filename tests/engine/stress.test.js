import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import helpers from '../common/helpers'
import request from "../common/request";
import common from "../common";
import {retry10} from "../common/retry";

const namespace = basename(__filename.replaceAll('.', '-'));
const fName = "file" + Math.random().toString(10).slice(2, 12) + ".wf.js"

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

	retry10(`should invoke /foo/${fName} workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespace }/instances?path=foo/${fName}`).send({ foo: "bar" })
		expect(req.statusCode).toEqual(200)
	})


	const total = 1000;    // total requests
	const batchSize = 100; // control concurrency in batches

	it(`fires ${total} requests`, async () => {
		const results = [];

		for (let start = 0; start < total; start += batchSize) {
			const batch = Array.from({ length: batchSize }, (_, i) => {
				const url = common.config.getDirektivHost() + `/api/v2/namespaces/${ namespace }/instances?path=foo/${fName}`;

				return fetch(url, {
					method: "POST",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify({ foo: "bar" }),
				}).then((res) => res.status);
			});

			// run batch concurrently
			const statuses = await Promise.all(batch);
			results.push(...statuses);

			console.log(`Batch done: ${start + batchSize}/${total}`);
		}

		console.log(JSON.stringify(results));

		// Assertions: all requests should be 200
		results.forEach((status, i) => {
			expect(status).toBe(200); // or 201 depending on your API
		});
	}, 60000); // extend timeout for big tests
})
