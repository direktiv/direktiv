import request from 'supertest'

import common from "../common"

const namespaceName = "root"

const cliExecutable = "/direktivctl"

const { exec } = require('child_process');
const fs = require('fs');
const prefix = `-a https://${common.config.getDirektivHost()} -t password `
const flagNamespace = `-n ${namespaceName}`

describe('Test the direktiv-cli-tool', () => {
    beforeAll(common.helpers.deleteAllNamespaces)
    afterAll(common.helpers.deleteAllNamespaces)

    it(`test namespace doesn not exists`, async() => { 
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
        assertStdErrContainsString("workflows push /tests/jest/simplewf.yaml", "pushing workflow")
        var readRevsResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/root/tree/simplewf`)
        expect(readRevsResponse.statusCode).toEqual(200)
    })
})

const assertStdErrContainsString = ((cmd,want)=>{
    exec(`${cliExecutable} ${prefix} ${flagNamespace} ${cmd}`, (err, stdout, stderr) => {
        //expect(`${cliExecutable} ${prefix} ${flagNamespace} ${cmd}`).toStrictEqual("expect.stringContaining(want)")
        expect(stderr).toStrictEqual(expect.stringContaining(want))
    })
})
const assertStdOutContainsString = ((cmd,want)=>{
    exec(`${cliExecutable} ${prefix} ${flagNamespace} ${cmd}`, (err, stdout, stderr) => {
        //expect(`${cliExecutable} ${prefix} ${flagNamespace} ${cmd}`).toStrictEqual("expect.stringContaining(want)")
        expect(stdout).toStrictEqual(expect.stringContaining(want))
    })
})