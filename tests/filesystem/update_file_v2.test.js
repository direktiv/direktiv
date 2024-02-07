import request from 'supertest'
import common from "../common";
import regex from "../common/regex";

const testNamespace = "test-file-namespace"

describe('Test filesystem tree update operations', () => {
    beforeAll(common.helpers.deleteAllNamespaces)

    common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

    common.helpers.itShouldCreateDirV2(it, expect, testNamespace, "/", "dir1")
    common.helpers.itShouldCreateFileV2(it, expect, testNamespace,
        "/dir1",
        "foo1",
        "workflow",
        "text/plain",
        btoa(common.helpers.dummyWorkflow("foo1")))

    common.helpers.itShouldUpdateFileV2(it, expect, testNamespace,
        "/dir1/foo1",
        {
            absolutePath: "/dir1/foo2",
            data: btoa(common.helpers.dummyWorkflow("foo2"))},
        )
})
