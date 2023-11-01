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
    - type: example_plugin
      configuration:
          echo_value: test_value
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
    - type: example_plugin
      configuration:
          echo_value: test_value
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
          error: "",
          plugins: [
            {
              configuration: {},
            },
          ],
        },
        {
          method: "GET",
          workflow: "action.yaml",
          namespace: "ns",
          error: "",
          plugins: [
            {
              configuration: {},
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
workflow: my-workflow.yaml
namespace: ${testNamespace}
plugins: 
    - type: example_plugin
      configuration:
          some_echo_value: test_value
  `
  );
  it(`should execute endpoint plugins`, async () => {
    const req = await request(common.config.getDirektivHost()).get(
      `/api/v2/gw/g1.yaml`
    );

    expect(req.statusCode).toEqual(404);
  });
  it(`should create a new direktiv file`, async () => {
    const res = await request(common.config.getDirektivHost())
      .put(
        `/api/namespaces/${testNamespace}/tree/my-workflow.yaml?op=create-workflow`
      )
      .set({
        "Content-Type": "text/plain",
      }).send(`
      description: A simple 'no-op' state that returns 'Hello world!'
      states:
      - id: helloworld
        type: noop
        transform:
          result: Hello world!`);

    expect(res.statusCode).toEqual(200);

    const req = await request(common.config.getDirektivHost()).get(
      `/api/v2/gw/g1.yaml`
    );

    expect(req.statusCode).toEqual(200);
  });
});

describe("Test plugin schema endpoint", () => {
  beforeAll(common.helpers.deleteAllNamespaces);
  common.helpers.itShouldCreateNamespace(it, expect, testNamespace);

  it(`should return all plugin schemas`, async () => {
    const req = await request(common.config.getDirektivHost()).get(
      `/api/v2/namespaces/${testNamespace}/plugins`
    );

    expect(req.body).toMatchObject({
      data: {
        example_plugin: {
          $defs: {
            examplePluginConfig: {
              additionalProperties: false,
              properties: { echo_value: { type: "string" } },
              required: ["echo_value"],
              type: "object",
            },
          },
          $id: "https://github.com/direktiv/direktiv/pkg/refactor/gateway/example-plugin-config",
          $ref: "#/$defs/examplePluginConfig",
          $schema: "https://json-schema.org/draft/2020-12/schema",
        },
      },
    });

    expect(req.statusCode).toEqual(200);
  });
});
