import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const namespaceName = 'simpleeventstest'

describe('Test events states behaviour', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespaceName)

	helpers.itShouldCreateFileV2(it, expect, namespaceName,
		'',
		'generate-event.yaml',
		'workflow',
		'text/plain',
		btoa(`
states:
- id: wait 
  type: delay
  duration: PT1S
  transition: generate
- id: generate
  type: generateEvent
  event:
    type: test.simple
    source: "generate-event"
`))

	helpers.itShouldCreateFileV2(it, expect, namespaceName,
		'',
		'simple-listener.yaml',
		'workflow',
		'text/plain',
		btoa(`
functions:
- id: spinoff
  type: subflow
  workflow: 'generate-event.yaml'
states:
- id: spinoff
  type: action
  async: true
  action:
    function: spinoff
  transition: listen
- id: listen
  type: consumeEvent
  timeout: PT1M
  event:
    type: test.simple
  transform:
    result: x
`))

	it(`should walk through the execution of a workflow called /simple-listener.yaml`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/tree/simple-listener.yaml?op=wait`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			result: 'x',
		})
	})

	helpers.itShouldCreateFileV2(it, expect, namespaceName,
		'',
		'or-listener.yaml',
		'workflow',
		'text/plain',
		btoa(`
functions:
- id: spinoff
  type: subflow
  workflow: 'generate-event.yaml'
states:
- id: spinoff
  type: action
  async: true
  action:
    function: spinoff
  transition: listen
- id: listen
  type: consumeEvent
  event:
    type: test.simple
  timeout: PT1M
  transition: a
  transform: 'jq(.result = "x")'
- id: a
  type: noop
  transform: 'jq(.transitioned = "a")'
- id: b
  type: noop
  transform: 'jq(.transitioned = "b")'
`))

	it(`should walk through the execution of a workflow called /or-listener.yaml`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/tree/or-listener.yaml?op=wait`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			result: 'x',
			transitioned: 'a',
		})
		expect(req.body['test.simple']).toMatchObject({
			data: {},
			datacontenttype: 'application/json',
			source: 'generate-event',
			specversion: '1.0',
			type: 'test.simple',
		})
	})
})
