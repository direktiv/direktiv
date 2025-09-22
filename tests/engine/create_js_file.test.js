import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import helpers from '../common/helpers'
import request from "../common/request";
import common from "../common";
import {retry10} from "../common/retry";

const namespace = basename(__filename.replaceAll('.', '-'));
const fName = "file" + Math.random().toString(10).slice(2, 12) + ".wf.js"

describe('Test js engine', () => {
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
    return payload;
}
`))
	retry10(`should invoke /foo/${fName} workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespace }/instances?path=foo/${fName}`).send({ foo: "bar" })
		console.log(req.statusCode, req.text)
		expect(req.statusCode).toEqual(200)
	})
})
