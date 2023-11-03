import common from "../common";
import request from "supertest";
import retry from "jest-retries";

const testNamespace = "gateway_namespace";

describe("Test gateway endpoints crud operations", () => {
  beforeAll(common.helpers.deleteAllNamespaces);

  common.helpers.itShouldCreateNamespace(it, expect, testNamespace);

  common.helpers.itShouldCreateFile(
    it,
    expect,
    testNamespace,
    "/g1.yaml",
    `
direktiv_api: endpoint/v1
method: POST
plugins:
  - type: example_plugin
    configuration:
      echo_value: test_value
  - type: target_workflow
    configuration:
      namespace: ${testNamespace}
      workflow: noop.yaml
    
`
  );

  common.helpers.itShouldCreateFile(
    it,
    expect,
    testNamespace,
    "/g2.yaml",
    `
direktiv_api: endpoint/v1
method: GET
plugins: 
  - type: example_plugin
    configuration:
        echo_value: test_value
  - type: target_workflow
    configuration:
      namespace: ${testNamespace}
      workflow: noop.yaml
`
  );

  it(`should list all endpoints`, async () => {
    const listRes = await request(common.config.getDirektivHost()).get(
      `/api/v2/namespaces/${testNamespace}/endpoints`
    );
    expect(listRes.statusCode).toEqual(200);
    expect(listRes.body.data.length).toEqual(2);
    expect(listRes.body.data).toEqual(
      expect.arrayContaining([
        {
          error: "",
          file_path: "/g1.yaml",
          method: "POST",
          plugins: [
            {
              configuration: { echo_value: "test_value" },
              type: "example_plugin",
            },
            {
              configuration: {
                namespace: "gateway_namespace",
                workflow: "noop.yaml",
              },
              type: "target_workflow",
            },
          ],
        },
        {
          error: "",
          file_path: "/g2.yaml",
          method: "GET",
          plugins: [
            {
              configuration: { echo_value: "test_value" },
              type: "example_plugin",
            },
            {
              configuration: {
                namespace: "gateway_namespace",
                workflow: "noop.yaml",
              },
              type: "target_workflow",
            },
          ],
        },
      ])
    );
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
  common.helpers.itShouldCreateFile(
    it,
    expect,
    testNamespace,
    "/g1.yaml",
    `
direktiv_api: endpoint/v1
method: GET
plugins: 
  - type: example_plugin
    configuration:
      some_echo_value: test_value
  - type: target_workflow
    configuration:
      namespace: ${testNamespace}
      workflow: noop.yaml
  `
  );
  it(`should execute endpoint plugins`, async () => {
    const req = await request(common.config.getDirektivHost()).get(
      `/api/v2/gw/g1.yaml`
    );

    expect(req.statusCode).toEqual(404);
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
