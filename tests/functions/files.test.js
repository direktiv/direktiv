import request from 'supertest'

import common from "../common"

const namespaceName = "functionsfiles"

describe('Test function files behaviour', () => {
    beforeAll(common.helpers.deleteAllNamespaces)
    afterAll(common.helpers.deleteAllNamespaces)

    it(`should create a namespace`, async () => {
        var req = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            namespace: {
                createdAt: expect.stringMatching(common.regex.timestampRegex),
                updatedAt: expect.stringMatching(common.regex.timestampRegex),
                name: namespaceName,
                oid: expect.stringMatching(common.regex.uuidRegex),
            },
        })
    })

    it(`should create a bash service`, async () => {
        var req = await request(common.config.getDirektivHost())
            .post(`/api/functions/namespaces/${namespaceName}`)
            .send({
                cmd: "",
                image: "direktiv/bash:dev",
                minScale: 1,
                name: "bash",
                size: 1
            })
        expect(req.statusCode).toEqual(200)
        expect(req.body).toEqual({})
    })

    it(`should create a workflow called /a.yaml`, async () => {

        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${namespaceName}/tree/a.yaml?op=create-workflow`)
            .set({
                'Content-Type': 'text/plain',
            })
            .send(`
functions:
- id: bash
  type: knative-namespace
  service: bash

states:
- id: set-c
  type: setter
  variables:
  - key: c
    scope: instance
    value: 11
  transition: set-value-fn

- id: set-value-fn
  type: action
  action:
    function: bash
    input: 
      commands:
      - command: bash -c 'cat a'
      - command: bash -c 'echo -n 5 > out/namespace/a'
      - command: bash -c 'cat b'
      - command: bash -c 'echo -n 7 > out/workflow/b'
      - command: bash -c 'cat c'
      - command: bash -c 'echo -n 11 > out/instance/c'
      - command: bash -c 'cat d'
      - command: bash -c 'echo -n 13 > out/instance/d'
      - command: bash -c 'cat e'
    files:
    - key: a
      scope: namespace
    - key: b
      scope: workflow
    - key: c
      scope: instance
    - key: d
      scope: instance
    - key: '/e.yaml'
      as: e
      scope: file
  transition: get-values

- id: get-values
  type: getter
  variables:
  - key: a
    scope: namespace
  - key: b
    scope: workflow
  - key: c
    scope: instance
  - key: d
    scope: instance
  - key: '/e.yaml'
    as: e
    scope: file
`)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: namespaceName,
        })
    })

    it(`should invoke the '/a.yaml' workflow on a fresh namespace`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/a.yaml?op=wait`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            var: {
                a: 5,
                b: 7,
                c: 11,
                d: 13,
                e: null
            }
        })
        expect(req.body.return.bash[0]).toMatchObject({
            result: "",
            success: true
        })
        expect(req.body.return.bash[2]).toMatchObject({
            result: "",
            success: true
        })
        expect(req.body.return.bash[4]).toMatchObject({
            result: 11,
            success: true
        })
        expect(req.body.return.bash[6]).toMatchObject({
            result: "",
            success: true
        })
        expect(req.body.return.bash[8].result).toBe("")
    }, 30000)

    it(`should create a workflow called /e.yaml`, async () => {

        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${namespaceName}/tree/e.yaml?op=create-workflow`)
            .set({
                'Content-Type': 'text/plain',
            })
            .send(`
states:
- id: a
  type: noop
  transform:
    result: x`)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: namespaceName,
        })
    })

    it(`should invoke the '/a.yaml' workflow on a non-fresh namespace`, async () => {
        const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/tree/a.yaml?op=wait`)
        expect(req.statusCode).toEqual(200)
        expect(req.body).toMatchObject({
            var: {
                a: 5,
                b: 7,
                c: 11,
                d: 13,
                e: "CnN0YXRlczoKLSBpZDogYQogIHR5cGU6IG5vb3AKICB0cmFuc2Zvcm06CiAgICByZXN1bHQ6IHg="
            }
        })
        expect(req.body.return.bash[0]).toMatchObject({
            result: 5,
            success: true
        })
        expect(req.body.return.bash[2]).toMatchObject({
            result: 7,
            success: true
        })
        expect(req.body.return.bash[4]).toMatchObject({
            result: 11,
            success: true
        })
        expect(req.body.return.bash[6]).toMatchObject({
            result: "",
            success: true
        })
        expect(req.body.return.bash[8].result).not.toBe("")
    })
})