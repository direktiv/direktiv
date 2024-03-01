import { expect } from '@jest/globals'

import common from '../common'
import request from '../common/request'
import helpers from "../common/helpers";


async function listInstancesAndFilter (ns, wf, status) {

	let append = ''

	if (wf)
		append = '&filter.field=AS&filter.type=CONTAINS&filter.val=' + wf


	let instancesResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ ns }/instances?limit=10&offset=0` + append)
		.send()

	// if filter, then try to wait
	if (wf || status)
		for (let i = 0; i < 20; i++) {
			instancesResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ ns }/instances?limit=10&offset=0` + append)
			if (status) {
				const idFind = instancesResponse.body.instances.results.find(item => item.status === status)
				if (idFind)
					return idFind

			} else if (instancesResponse.body.instances.pageInfo.total == 1)
				return instancesResponse.body.instances.results[0]

			await helpers.sleep(100)
			instancesResponse = (function () {

			})()
		}


	if (instancesResponse)
		return instancesResponse.body


}

// send event and wait for it to appear in the event list baesd on id
async function sendEventAndList (ns, event) {

	const eventObject = JSON.parse(event)
	let idFind

	// requires cloudevent id
	expect(eventObject.id).not.toBeFalsy()

	// post event
	const workflowEventResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${ ns }/broadcast`)
		.set('Content-Type', 'application/json')
		.send(event)
	expect(workflowEventResponse.statusCode).toEqual(200)

	// wait for it to be registered
	for (let i = 0; i < 50; i++) {
		const eventsResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ ns }/events?limit=100&offset=0`)
			.send()
		idFind = eventsResponse.body.events.results.find(item => item.id === eventObject.id)
		if (idFind)
			break

		await helpers.sleep(100)
	}
	return idFind
}

export default {
	sendEventAndList,
	listInstancesAndFilter,
}
