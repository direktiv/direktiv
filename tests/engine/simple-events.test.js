import request from "../common/request"

import common from "../common"

const namespaceName = "simpleeventstest"


describe('Test events states behaviour', () => {
    beforeAll(common.helpers.deleteAllNamespaces)


    it(`should create a namespace`, async () => {
        var req = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}`)

        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
                name: namespaceName,
            },
        })
    })

    it(`should create a workflow called /generate-event.yaml`, async () => {

        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${namespaceName}/tree/generate-event.yaml?op=create-workflow`)
            .set({
                'Content-Type': 'text/plain',
            })
            .send(`
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
`)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: namespaceName,
        })
    })

    it(`should create a workflow called /simple-listener.yaml`, async () => {

        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${namespaceName}/tree/simple-listener.yaml?op=create-workflow`)
            .set({
                'Content-Type': 'text/plain',
            })
            .send(`
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
`)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: namespaceName,
        })
    })

    it(`should walk through the execution of a workflow called /simple-listener.yaml`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/simple-listener.yaml?op=wait`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            result: 'x',
        })
    })

    it(`should create a workflow called /or-listener.yaml`, async () => {

      const res = await request(common.config.getDirektivHost())
          .put(`/api/namespaces/${namespaceName}/tree/or-listener.yaml?op=create-workflow`)
          .set({
              'Content-Type': 'text/plain',
          })
          .send(`
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
`)

      expect(res.statusCode).toEqual(200)
      expect(res.body).toMatchObject({
          namespace: namespaceName,
      })
  })

  it(`should walk through the execution of a workflow called /or-listener.yaml`, async () => {
      const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/or-listener.yaml?op=wait`)
      expect(req.statusCode).toEqual(200)
      expect(req.body).toMatchObject({
          result: 'x',
          transitioned: 'a',
      })
      expect(req.body["test.simple"]).toMatchObject({
        data: {},
        datacontenttype: 'application/json',
        source: 'generate-event',
        specversion: '1.0',
        type: 'test.simple'
    })
  })

})