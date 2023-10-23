import common from "../common";
import request from "supertest";
import retry from "jest-retries";

const testNamespace = "gateway_namespace";

describe("Test gateway endpoints crud operations", () => {
  beforeAll(common.helpers.deleteAllNamespaces);

  common.helpers.itShouldCreateNamespace(it, expect, testNamespace);

  common.helpers.itShouldCreateEndpointFile(
    it,
    expect,
    testNamespace,
    "/g1.yaml",
    `
direktiv_api: endpoint/v1
method: POST
workflow: action.yaml
namespace: ns
plugins: 
    - id: example_plugin
      configuration: ""
`
  );

  common.helpers.itShouldCreateEndpointFile(
    it,
    expect,
    testNamespace,
    "/g2.yaml",
    `
direktiv_api: endpoint/v1
method: GET
workflow: action.yaml
namespace: ns
plugins: 
    - id: example_plugin
      configuration: ""
`
  );

  it(`should list all endpoints`, async () => {
    const listRes = await request(common.config.getDirektivHost()).get(
      `/api/v2/namespaces/${testNamespace}/endpoints`
    );
    expect(listRes.statusCode).toEqual(200);
    expect(listRes.body.data.length).toEqual(2);
    expect(listRes.body).toMatchObject({
      data: [
        {
          method: "POST",
          workflow: "action.yaml",
          namespace: "ns",
          plugins: [
            {
              configuration: "",
              id: "example_plugin",
            },
          ],
        },
        {
          method: "GET",
          workflow: "action.yaml",
          namespace: "ns",
          plugins: [
            {
              configuration: "",
              id: "example_plugin",
            },
          ],
        },
      ],
    });
  });

  common.helpers.itShouldDeleteFile(it, expect, testNamespace, "/g1.yaml");

  it(`should list all endpoints`, async () => {
    const listRes = await request(common.config.getDirektivHost()).get(
      `/api/v2/namespaces/${testNamespace}/endpoints`
    );
    expect(listRes.statusCode).toEqual(200);
    expect(listRes.body.data.length).toEqual(1);
    expect(listRes.body).toMatchObject({
      data: [
        {
          method: "GET",
        },
      ],
    });
  });
});

describe("Test availability of gateway endpoints", () => {
  beforeAll(common.helpers.deleteAllNamespaces);

  common.helpers.itShouldCreateNamespace(it, expect, testNamespace);

  common.helpers.itShouldCreateEndpointFile(
    it,
    expect,
    testNamespace,
    "/g1.yaml",
    `
direktiv_api: endpoint/v1
method: GET
workflow: action.yaml
namespace: ns
plugins: 
    - id: example_plugin
      configuration: ""
  `
  );
  it(`should execute endpoint plugins`, async () => {
    const req = await request(common.config.getDirektivHost()).get(
      `/api/v2/gw/g1.yaml`
    );

    expect(req.statusCode).toEqual(200);
  });
});
