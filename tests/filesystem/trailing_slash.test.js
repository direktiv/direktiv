import common from "../common";

const testNamespace = "test-file-namespace"

describe('Test filesystem tree read operations', () => {
    beforeAll(common.helpers.deleteAllNamespaces)

    common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

    common.helpers.itShouldCreateFileV2(it, expect, testNamespace,
        "",
        "foo1",
        "workflow",
        "text",
        common.helpers.dummyWorkflow("foo1"))

    common.helpers.itShouldCreateFileV2(it, expect, testNamespace,
        "/",
        "foo2",
        "workflow",
        "text",
        common.helpers.dummyWorkflow("foo2"))


    common.helpers.itShouldCreateDirV2(it, expect, testNamespace, "", "dir1")
    common.helpers.itShouldCreateDirV2(it, expect, testNamespace, "/", "dir2")

})
