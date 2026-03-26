import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../../common'
import helpers from '../../common/helpers'
import request from '../../common/request'

const namespaceName = 'callpathtest'

describe('Test subflow behaviour', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespaceName)

	helpers.itShouldCreateDir(it, expect, namespaceName, '/', 'a')

	helpers.itShouldCreateFile(
		it,
		expect,
		namespaceName,
		'',
		`child.wf.ts`,
		'workflow',
		'application/typescript',
		btoa(`
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateFirst",
};

function stateFirst(data): StateFunction<string> {
  return finish({
    meta: "the workflow input is returned in the input field",
    input: data,
  });
}
`),
	)

	helpers.itShouldCreateFile(
		it,
		expect,
		namespaceName,
		'',
		`parent.wf.ts`,
		'workflow',
		'application/typescript',
		btoa(`
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateFirst",
};

function stateFirst(): StateFunction<unknown> {
  let resp = execSubflow(
    "/child.wf.ts",
    "This is the data passed to the subflow",
  );
  return finish(resp);
}
`),
	)

	it(`the parent workflow should invoke the child workflow and return its response`, async () => {
		const req = await request(common.config.getDirektivBaseUrl()).post(
			`/api/v2/namespaces/${namespaceName}/instances?path=%2Fparent.wf.ts&wait=true`,
		)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toEqual({
			data: {
				input: 'This is the data passed to the subflow',
				meta: 'the workflow input is returned in the input field',
			},
		})
	})
})
