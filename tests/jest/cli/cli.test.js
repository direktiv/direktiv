import request from 'supertest'

import common from "../common"
import { setMaxIdleHTTPParsers } from 'http';

const namespaceName = "root"

const cliExecutable = "/direktivctl"

const { exec } = require('child_process');
const fs = require('fs');
const prefix = `-a https://${common.config.getDirektivHost()} -t password `
const flagNamespace = `-n ${namespaceName}`

describe('Test the direktiv-cli-tool', () => {
    beforeAll(common.helpers.deleteAllNamespaces)
    afterAll(common.helpers.deleteAllNamespaces)

    it(`test namespace does not exists`, async() => { 
        if (!fs.existsSync(cliExecutable)){ return }
        assertStdErrContainsString("info", "404 Not Found")
    })
    it(`test namespace exists`, async() => {
        if (!fs.existsSync(cliExecutable)){ return }
        const createResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${namespaceName}`)
        expect(createResponse.statusCode).toEqual(200)
        assertStdErrContainsString("info", `namespace: ${namespaceName}`)
    })
    it(`test push wf`, async() => {
        if (!fs.existsSync(cliExecutable)){ return }
        const filepath =  "/tests/jest/"
        const filename = "simplewf"
        const fileextension = "yaml"
        await assertStdErrShouldNotContainsString(`workflows push ${filepath}`, `pushing workflow ${filename}`)
        const res = await request(common.config.getDirektivHost()).get(`/api/namespaces/root/tree/${filename}`)
        expect(res.statusCode).toEqual(200)
    })
    it(`test push wf w relative path`, async() => {
        if (!fs.existsSync(cliExecutable)){ return }
        const filepath =  "../../tests/jest/"
        const filename = "simplewf"
        const fileextension = "yaml"
        await assertStdErrShouldNotContainsString(`workflows push ${filepath}`, `pushing workflow ${filename}`)
        const res = await request(common.config.getDirektivHost()).get(`/api/namespaces/root/tree/${filename}`)
        expect(res.statusCode).toEqual(200)
    })
})

const assertStdErrContainsString = ((cmd,want)=>{
    return new Promise(function (resolve, reject) {
        exec(`${cliExecutable} ${prefix} ${flagNamespace} ${cmd}`, (err, stdout, stderr) => {
            expect(stderr).toStrictEqual(expect.stringContaining(want))
            resolve()
        })
    })
})
const assertStdErrShouldNotContainsString = ((cmd,want)=>{
    return new Promise(function (resolve, reject) {
        exec(`${cliExecutable} ${prefix} ${flagNamespace} ${cmd}`, (err, stdout, stderr) => {
            expect(stderr).not.toStrictEqual(expect.stringContaining(want))
            resolve()
        })
    })
})