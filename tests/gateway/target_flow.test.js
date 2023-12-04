import common from "../common";
import request from "supertest";

const testNamespace = "gateway_namespace";

const limitedNamespace = "limited_namespace";


const workflow = `
  direktiv_api: workflow/v1
  description: A simple 'no-op' state that returns 'Hello world!'
  states:
  - id: helloworld
    type: noop
    transform:
      result: Hello world!
`

const endpointWorkflow = `
  direktiv_api: endpoint/v1
  allow_anonymous: true
  plugins:
    target:
      type: target-flow
      configuration:
          namespace: ` + testNamespace + `
          flow: /workflow.yaml
  methods: 
    - GET
  path: /endpoint1`

const endpointWorkflowAllowed = `
  direktiv_api: endpoint/v1
  allow_anonymous: true
  plugins:
    target:
      type: target-flow
      configuration:
          namespace: ` + limitedNamespace + `
          flow: /workflow.yaml
          content_type: text/json
  methods: 
    - GET
  path: /endpoint2`

describe("Test target workflow variable plugin", () => {
    beforeAll(common.helpers.deleteAllNamespaces);
  
    common.helpers.itShouldCreateNamespace(it, expect, limitedNamespace);
    common.helpers.itShouldCreateNamespace(it, expect, testNamespace);
  
    common.helpers.itShouldCreateFile(
      it,
      expect,
      testNamespace,
      "/workflow.yaml",
      workflow
    );

    common.helpers.itShouldCreateFile(
        it,
        expect,
        limitedNamespace,
        "/workflow.yaml",
        workflow
    );
  
    common.helpers.itShouldCreateFile(
        it,
        expect,
        limitedNamespace,
        "/endpoint1.yaml",
        endpointWorkflow
    );

    common.helpers.itShouldCreateFile(
        it,
        expect,
        limitedNamespace,
        "/endpoint2.yaml",
        endpointWorkflowAllowed
    );

    common.helpers.itShouldCreateFile(
        it,
        expect,
        testNamespace,
        "/endpoint1.yaml",
        endpointWorkflow
    );

    common.helpers.itShouldCreateFile(
        it,
        expect,
        testNamespace,
        "/endpoint2.yaml",
        endpointWorkflowAllowed
    );

    it(`should return a workflow run from magic namespace`, async () => {
    const req = await request(common.config.getDirektivHost()).get(
        `/gw/endpoint1`
    );
        expect(req.statusCode).toEqual(200);
        expect(req.text).toEqual("{\"result\":\"Hello world!\"}")
    });

    it(`should return a flow run from magic namespace with namespace set`, async () => {
        const req = await request(common.config.getDirektivHost()).get(
            `/gw/endpoint2`
        );
        expect(req.statusCode).toEqual(200);
        expect(req.text).toEqual("{\"result\":\"Hello world!\"}")
        expect(req.header['content-type']).toEqual("text/json")
    });

    it(`should return a workflow var from non-magic namespace`, async () => {
    const req = await request(common.config.getDirektivHost()).get(
        `/ns/` + limitedNamespace + `/endpoint2`
    );
        expect(req.statusCode).toEqual(200);
        expect(req.text).toEqual("{\"result\":\"Hello world!\"}")
    });

    it(`should not return a var`, async () => {
    const req = await request(common.config.getDirektivHost()).get(
        `/ns/` + limitedNamespace + `/endpoint1`
    );
        expect(req.statusCode).toEqual(403);
    });

  
  });