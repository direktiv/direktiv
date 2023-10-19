import request from 'supertest'
import common from "../common";

const testNamespace = "test-services"

describe('Test services crud operations', () => {
    beforeAll(common.helpers.deleteAllNamespaces)

    common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

    common.helpers.itShouldCreateServiceFile(it, expect, testNamespace,
        "/s1.yaml", `
direktiv_api: service/v1
name: s1
image: redis
cmd: redis-server
scale: 1
`)

    common.helpers.itShouldCreateServiceFile(it, expect, testNamespace,
        "/s2.yaml", `
direktiv_api: service/v1
name: s2
image: redis
cmd: redis-server
scale: 2
`)

});


function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}