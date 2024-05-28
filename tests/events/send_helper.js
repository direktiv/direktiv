import { expect } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

async function listInstancesAndFilter (ns, wf, status) {
	let append = ''

	if (wf)
		append = '&filter.field=AS&filter.type=CONTAINS&filter.val=' + wf

	let instancesResponse = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ ns }/instances?limit=100&offset=0` + append)
		.send()

	// if filter, then try to wait
	if (wf || status)
		for (let i = 0; i < 20; i++) {
			instancesResponse = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ ns }/instances?limit=100&offset=0` + append)
			if (status) {
				const idFind = instancesResponse.body.data.find(item => item.status === status)
				if (idFind)
					return idFind
			} else if (instancesResponse.body.meta.total === 1)
				return instancesResponse.body.data[0]

			await helpers.sleep(200)
			// eslint-disable-next-line
			instancesResponse = (function () {
			})()
		}

	if (instancesResponse)
		return instancesResponse.body
}

// send event and wait for it to appear in the event list baesd on id
async function sendEventAndList (ns, event) {
	const eventObject = JSON.parse(event)

	// requires cloudevent id
	expect(eventObject.id).not.toBeFalsy()

	// post event
	const workflowEventResponse = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ ns }/events/broadcast`)
		.set('Content-Type', 'application/json')
		.send(event)
	expect(workflowEventResponse.statusCode).toEqual(200)

	// wait for it to be registered
	const eventsResponse = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ ns }/events/history?limit=100&offset=0`)
		.send()

	for (let i = 0; i < eventsResponse.body.data.length; i++)
		if (eventsResponse.body.data[i].event.id === eventObject.id)
			return eventsResponse.body.data[i].event
}

export default {
	sendEventAndList,
	listInstancesAndFilter,
}
