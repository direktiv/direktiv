import request from 'supertest'

import common from "../common"

const namespaceName = "eventfilters"


const eventFilter1 = "filter1"
const filter1 = `
if (event["source"] == "mysource") {
  nslog("rename source")
  event["source"] = "newsource"
}

if (event["source"] == "hello") {
  nslog("drop me")
  return null
}

return event
`

const filter1patch = `
if (event["source"] == "mysource") {
  nslog("rename source")
  event["source"] = "newsource"
}

if (event["source"] == "hello") {
  nslog("drop me patch")
  return null
}

return event
`

const eventFilter2 = "filter2"
const filter2 = `
nslog("event in")
return event
`

const eventFilter3 = "filter3"
const filter3 = `
THIS IS BROKEN!!!
`

const basevent = (type, source) => `{
  "specversion" : "1.0",
  "type" : "${type}",
  "source" : "${source}",
  "datacontenttype" : "application/json",
  "data" : {
      "hello": "world",
      "123": 456
  }
}`

describe('Test events filter', () => {
    beforeAll(common.helpers.deleteAllNamespaces)
    // afterAll(common.helpers.deleteAllNamespaces)

    it(`should create namespace`, async () => {
        var createNamespaceResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}`)
        expect(createNamespaceResponse.statusCode).toEqual(200)
    })

    it(`basic filter handling`, async () => {

        // workflow with start
        var createFilterResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${namespaceName}/eventfilter/${eventFilter1}`)
            .send(filter1)
        expect(createFilterResponse.statusCode).toEqual(200)

        var createFilterResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${namespaceName}/eventfilter/${eventFilter2}`)
            .send(filter2)
        expect(createFilterResponse.statusCode).toEqual(200)

        var createFilterResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${namespaceName}/eventfilter/${eventFilter3}`)
            .send(filter3)
        expect(createFilterResponse.statusCode).toEqual(400)


        var getFilterResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/eventfilter/${eventFilter1}`)
            .send()
        expect(getFilterResponse.statusCode).toEqual(200)
        expect(getFilterResponse.body.filtername).toEqual(eventFilter1)

        var getFilterResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/eventfilter/${eventFilter2}`)
            .send()
        expect(getFilterResponse.statusCode).toEqual(200)
        expect(getFilterResponse.body.filtername).toEqual(eventFilter2)


        var getFiltersResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/eventfilter`)
            .send()
        expect(getFiltersResponse.statusCode).toEqual(200)
        expect(getFiltersResponse.body.eventFilter.length).toEqual(2)

        var getFilterResponse = await request(common.config.getDirektivHost()).delete(`/api/namespaces/${namespaceName}/eventfilter/${eventFilter2}`)
            .send()
        expect(getFilterResponse.statusCode).toEqual(200)

        var getFiltersResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/eventfilter`)
            .send()
        expect(getFiltersResponse.statusCode).toEqual(200)
        expect(getFiltersResponse.body.eventFilter.length).toEqual(1)


        var createFilterResponse = await request(common.config.getDirektivHost()).patch(`/api/namespaces/${namespaceName}/eventfilter/${eventFilter1}`)
            .send(filter1patch)
        expect(createFilterResponse.statusCode).toEqual(200)


        var getFilterResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/eventfilter/${eventFilter1}`)
            .send()
        expect(getFilterResponse.statusCode).toEqual(200)
        expect(getFilterResponse.body.filtername).toEqual(eventFilter1)
        expect(getFilterResponse.body.jsCode).toMatch(/(drop me patch)/i)


    })

    it(`run filter`, async () => {

        // send event through, no change
        var workflowEventResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${namespaceName}/broadcast/${eventFilter1}`)
            .set('Content-Type', 'application/json')
            .send(basevent("nochange", "nochange"))
        expect(workflowEventResponse.statusCode).toEqual(200)


        // send event through, change
        var workflowEventResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${namespaceName}/broadcast/${eventFilter1}`)
            .set('Content-Type', 'application/json')
            .send(basevent("mysource", "mysource"))
        expect(workflowEventResponse.statusCode).toEqual(200)


        // send event through, block
        var workflowEventResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${namespaceName}/broadcast/${eventFilter1}`)
            .set('Content-Type', 'application/json')
            .send(basevent("hello", "hello"))
        expect(workflowEventResponse.statusCode).toEqual(200)

        await new Promise((r) => setTimeout(r, 500));

        // there should be two events, one blocked
        var eventsResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/events?limit=100&offset=0`)
            .send()

        expect(eventsResponse.body.events.pageInfo.total).toEqual(2)

        var idFind = eventsResponse.body.events.results.find(item => item.source === "newsource");
        expect(idFind).not.toBeFalsy();

        var idFind = eventsResponse.body.events.results.find(item => item.source === "nochange");
        expect(idFind).not.toBeFalsy();

    });

    it(`run filter load test`, async () => {

        var event = basevent("mytype", "value1")

        for (let i = 0; i < 2000; i++) {

            var workflowEventResponse = await request(common.config.getDirektivHost()).post(`/api/namespaces/${namespaceName}/broadcast/${eventFilter1}`)
                .set('Content-Type', 'application/json')
                .send(event)

        }

    });

})
