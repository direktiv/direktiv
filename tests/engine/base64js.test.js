import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import request from '../common/request'

const testNamespace = 'js-base64'

const file = `
direktiv_api: workflow/v1
description: A simple 'no-op' state that returns 'Hello world!'
states:
- id: helloworld
  type: noop
  transform:
    result: |
      js(
        return btoa("demo")
      )
  transition: h
- id: h
  type: noop
  transform:
    out: |
      js(
        return atob(data["result"])
      )
`

describe('Test js base64 feature', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

    common.helpers.itShouldCreateYamlFileV2(
		it,
		expect,
		testNamespace,
		'/', 'wf.yaml', 'workflow',
		file,
	)

    it(`should invoke the workflow and encode and decode base64`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ testNamespace }/tree/wf.yaml?op=wait`)
			.send(`{"x": 5}`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			out: "demo",
		})
	})

})
