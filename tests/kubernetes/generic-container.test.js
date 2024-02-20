import common from "../common";
import request from "../common/request";
import retry from "jest-retries";


const testNamespace = "patches";


const genericContainerWorkflow = `
direktiv_api: workflow/v1
functions:
- id: test
  image: alpine
  type: knative-workflow
  cmd: /usr/share/direktiv/direktiv-cmd
states:
- id: test 
  type: action
  action:
    function: test
    input: 
      data:
        commands:
        - command: echo -n data
`


describe("Test generic container", () => {
  beforeAll(common.helpers.deleteAllNamespaces);

  common.helpers.itShouldCreateNamespace(it, expect, testNamespace);

  common.helpers.itShouldCreateFile(
    it,
    expect,
    testNamespace,
    "/wf1.yaml",
    genericContainerWorkflow
  );


  retry(`should invoke workflow`, 10, async () => {
    await sleep(500);
    const res = await request(common.config.getDirektivHost()).get(`/api/namespaces/${testNamespace}/tree/wf1.yaml?op=wait`)
    expect(res.statusCode).toEqual(200)
    expect(res.body.return[0].Output).toEqual("data")
  })


});


function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}
