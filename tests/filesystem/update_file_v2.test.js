import request from 'supertest'
import common from "../common";
import regex from "../common/regex";

const testNamespace = "test-file-namespace"

describe('Test filesystem tree update paths', () => {
    beforeAll(common.helpers.deleteAllNamespaces)

    common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

    common.helpers.itShouldCheckPathExistsV2(it, expect, testNamespace, "/dir1", false)
    common.helpers.itShouldCheckPathExistsV2(it, expect, testNamespace, "/foo1", false)

    common.helpers.itShouldCreateDirV2(it, expect, testNamespace, "/", "dir1")
    common.helpers.itShouldCreateFileV2(it, expect, testNamespace, "/", "foo1", "workflow", "text", common.helpers.dummyWorkflow("foo1"))

    common.helpers.itShouldCheckPathExistsV2(it, expect, testNamespace, "/dir1", true)
    common.helpers.itShouldCheckPathExistsV2(it, expect, testNamespace, "/foo1", true)


    common.helpers.itShouldUpdatePathV2(it, expect, testNamespace, "/foo1", "/foo2", )

    common.helpers.itShouldCheckPathExistsV2(it, expect, testNamespace, "/dir1", true)
    common.helpers.itShouldCheckPathExistsV2(it, expect, testNamespace, "/foo1", false)
    common.helpers.itShouldCheckPathExistsV2(it, expect, testNamespace, "/foo2", true)
})


describe('Test filesystem tree change dir', () => {
    beforeAll(common.helpers.deleteAllNamespaces)

    common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

    common.helpers.itShouldCreateDirV2(it, expect, testNamespace, "/", "dir1")
    common.helpers.itShouldCreateDirV2(it, expect, testNamespace, "/dir1", "dir2")

    common.helpers.itShouldCheckPathExistsV2(it, expect, testNamespace, "/dir1", true)
    common.helpers.itShouldCheckPathExistsV2(it, expect, testNamespace, "/dir2", false)
    common.helpers.itShouldCheckPathExistsV2(it, expect, testNamespace, "/dir1/dir2", true)


    common.helpers.itShouldUpdatePathV2(it, expect, testNamespace, "/dir1/dir2", "/dir2", )

    common.helpers.itShouldCheckPathExistsV2(it, expect, testNamespace, "/dir1", true)
    common.helpers.itShouldCheckPathExistsV2(it, expect, testNamespace, "/dir2", true)
    common.helpers.itShouldCheckPathExistsV2(it, expect, testNamespace, "/dir1/dir2", false)
})
