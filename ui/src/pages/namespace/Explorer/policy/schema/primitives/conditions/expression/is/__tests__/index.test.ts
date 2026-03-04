import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../../../utils/testutils";
import { describe, test } from "vitest";

describe("Is Expression schema", () => {
  test("accepts is expression with optional in clause", () => {
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              is: {
                left: { Var: "principal" },
                entity_type: "User",
                in: { Value: { __entity: { type: "Group", id: "friends" } } },
              },
            },
          },
        ],
      })
    );
  });

  test("accepts is expression without in clause", () => {
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              is: {
                left: { Var: "principal" },
                entity_type: "User",
              },
            },
          },
        ],
      })
    );
  });

  test("rejects is expression without entity_type", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              is: {
                left: { Var: "principal" },
                in: { Value: { __entity: { type: "Group", id: "friends" } } },
              },
            },
          },
        ],
      })
    );
  });

  test("rejects non-string entity_type", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              is: {
                left: { Var: "principal" },
                entity_type: 1,
              },
            },
          },
        ],
      })
    );
  });
});
