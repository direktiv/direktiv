import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../../../utils/testutils";
import { describe, test } from "vitest";

describe("Var Expression schema", () => {
  test("accepts all valid Var values", () => {
    /*
      Cedar:
      when { principal == User::"alice" };
      when { action == Action::"readFile" };
      when { resource in Folder::"Public" };
      when { context.tls_version == "1.3" };
    */
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              "==": {
                left: { Var: "principal" },
                right: { Value: { __entity: { type: "User", id: "alice" } } },
              },
            },
          },
          {
            kind: "when",
            body: {
              "==": {
                left: { Var: "action" },
                right: {
                  Value: { __entity: { type: "Action", id: "readFile" } },
                },
              },
            },
          },
          {
            kind: "when",
            body: {
              in: {
                left: { Var: "resource" },
                right: {
                  Value: { __entity: { type: "Folder", id: "Public" } },
                },
              },
            },
          },
          {
            kind: "when",
            body: {
              "==": {
                left: {
                  ".": { left: { Var: "context" }, attr: "tls_version" },
                },
                right: { Value: "1.3" },
              },
            },
          },
        ],
      })
    );
  });

  test("rejects invalid Var value", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [{ kind: "when", body: { Var: "actor" } }],
      })
    );
  });
});
