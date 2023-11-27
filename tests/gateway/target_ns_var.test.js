import common from "../common";
import request from "supertest";

const testNamespace = "gateway_namespace";

const limitedNamespace = "limited_namespace";

const endpointNSVar = `
  direktiv_api: endpoint/v1
  allow_anonymous: true
  plugins:
    target:
      type: target-namespace-var
      configuration:
          namespace: ` + testNamespace + `
          variable: plain
  methods: 
    - GET`

const endpointNSVarAllowed = `
  direktiv_api: endpoint/v1
  allow_anonymous: true
  plugins:
    target:
      type: target-namespace-var
      configuration:
          namespace: ` + limitedNamespace + `
          variable: plain
          content_type: text/test
  methods: 
    - GET`


describe("Test target namespace variable plugin", () => {
    beforeAll(common.helpers.deleteAllNamespaces);
  
    common.helpers.itShouldCreateNamespace(it, expect, limitedNamespace);
    common.helpers.itShouldCreateNamespace(it, expect, testNamespace);
  
    it(`should set plain text variable`, async () => {
      var workflowVarResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${testNamespace}/vars/plain`)
          .set('Content-Type', 'text/plain')
          .send("Hello World")
      expect(workflowVarResponse.statusCode).toEqual(200)
    })
  
    it(`should set plain text variable`, async () => {
      var workflowVarResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${limitedNamespace}/vars/plain`)
          .set('Content-Type', 'text/plain')
          .send("Hello World 2")
      expect(workflowVarResponse.statusCode).toEqual(200)
    })
  
  
    common.helpers.itShouldCreateFile(
      it,
      expect,
      testNamespace,
      "/endpoint1.yaml",
      endpointNSVar
    );
  
    common.helpers.itShouldCreateFile(
      it,
      expect,
      testNamespace,
      "/endpoint2.yaml",
      endpointNSVarAllowed
    );
  
    common.helpers.itShouldCreateFile(
      it,
      expect,
      limitedNamespace,
      "/endpoint1.yaml",
      endpointNSVar
    );
  
    common.helpers.itShouldCreateFile(
      it,
      expect,
      limitedNamespace,
      "/endpoint2.yaml",
      endpointNSVarAllowed
    );
  
  
    it(`should return a ns var from magic namespace`, async () => {
      const req = await request(common.config.getDirektivHost()).get(
        `/gw/endpoint1`
      );
      expect(req.statusCode).toEqual(200);
      expect(req.text).toEqual("Hello World")
      expect(req.header['content-type']).toEqual("text/plain")
    });
  
    it(`should return a var from magic namespace with namespace set`, async () => {
      const req = await request(common.config.getDirektivHost()).get(
        `/gw/endpoint2`
      );
      expect(req.statusCode).toEqual(200);
      expect(req.text).toEqual("Hello World 2")
      expect(req.header['content-type']).toEqual("text/test")
    });
  
    it(`should return a var from non-magic namespace`, async () => {
      const req = await request(common.config.getDirektivHost()).get(
        `/ns/` + limitedNamespace + `/endpoint2`
      );
      expect(req.statusCode).toEqual(200);
      expect(req.text).toEqual("Hello World 2")
    });
  
    it(`should not return a var`, async () => {
      const req = await request(common.config.getDirektivHost()).get(
        `/ns/` + limitedNamespace + `/endpoint1`
      );
      expect(req.statusCode).toEqual(403);
    });
    
  
  });
  