import request from 'supertest'

import config from "./config";
import common from "./index";

async function deleteAllNamespaces() {
    let listResponse = await request(config.getDirektivHost()).get(`/api/namespaces`)
    if(listResponse.statusCode !== 200) {
        throw Error(`none ok namespaces list statusCode(${listResponse.statusCode})`)
    }

    for (const namespace of listResponse.body.results) {
        let response = await request(config.getDirektivHost()).delete(`/api/namespaces/${namespace.name}?recursive=true`)

        if(response.statusCode !== 200) {
            throw Error(`none ok namespace(${namespace.name}) delete statusCode(${response.statusCode})`)
        }
    }
}

async function itShouldCreateNamespace(it, expect, ns) {
    it(`should create a new namespace ${ns}`, async () => {
        const res = await request(common.config.getDirektivHost()).put(`/api/namespaces/${ns}`)
        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: {
                name: ns,
                // regex /^2.*Z$/ matches timestamps like 2023-03-01T14:19:52.383871512Z
                createdAt: expect.stringMatching(/^2.*Z$/),
                updatedAt: expect.stringMatching(/^2.*Z$/),
            }
        })
    })
}

async function itShouldCreateFile(it, expect, ns, path, content) {
    it(`should create a new file ${path}`, async () => {
        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${ns}/tree${path}?op=create-workflow`)
            .set({
                'Content-Type': 'text/plain',
            })

            .send(content)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: ns,
        })
    })
}

function dummyWorkflow(someText) {
    return `
direktiv_api: workflow/v1
description: A simple 'no-op' state that returns 'Hello world!'
states:
- id: helloworld
  type: noop
`
}

async function itShouldCreateDirectory(it, expect, ns, path) {
    it(`should create a directory ${path}`, async () => {
        const res = await request(common.config.getDirektivHost())
            .put(`/api/namespaces/${ns}/tree${path}?op=create-directory`)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: ns,
        })
    })
}

async function itShouldUpdateFile(it, expect, ns, path, content) {
    it(`should update existing file ${path}`, async () => {
        const res = await request(common.config.getDirektivHost())
            .post(`/api/namespaces/${ns}/tree${path}?op=update-workflow`)
            .set({
                'Content-Type': 'text/plain',
            })

            .send(content)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({
            namespace: ns,
        })
    })
}

async function itShouldDeleteFile(it, expect, ns, path) {
    it(`should delete a file ${path}`, async () => {
        const res = await request(common.config.getDirektivHost())
            .delete(`/api/namespaces/${ns}/tree${path}?op=delete-node`)

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({})
    })
}

async function itShouldRenameFile(it, expect, ns, path, newPath) {
    it(`should delete a file ${path}`, async () => {
        const res = await request(common.config.getDirektivHost())
            .post(`/api/namespaces/${ns}/tree${path}?op=rename-node`)
            .send({new: newPath})

        expect(res.statusCode).toEqual(200)
        expect(res.body).toMatchObject({})
    })
}

export default {
    deleteAllNamespaces,
    itShouldCreateNamespace,
    itShouldCreateFile,
    itShouldDeleteFile,
    itShouldRenameFile,
    itShouldUpdateFile,
    itShouldCreateDirectory,
    dummyWorkflow,
}