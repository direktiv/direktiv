import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const namespaceName = 'wfeventsv2'

const basicEvent = `{
    "specversion" : "1.0",
    "type" : "testerDuplicate",
    "source" : "https://direktiv.io/test",
    "id": "123"
}`

describe('Test send events v2 api', () => {
	beforeAll(common.helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespaceName)
	it(`should send event to namespace`, async () => {
		const sendEventResponse = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespaceName }/events/broadcast`)
			.set('Content-Type', 'application/json')
			.send(basicEvent)
		expect(sendEventResponse.statusCode).toEqual(200)
	})

})