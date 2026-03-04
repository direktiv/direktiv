import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../../../utils/testutils";
import { describe, test } from "vitest";

describe("Record Expression schema", () => {
  test("accepts Record expression", () => {
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              Record: {
                foo: { Value: "spam" },
                somethingelse: { Value: false },
              },
            },
          },
        ],
      })
    );
  });

  test("accepts nested Record expression for user profile", () => {
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              Record: {
                user: {
                  Record: {
                    profile: {
                      Record: {
                        name: { Value: "alice" },
                        isActive: { Value: true },
                      },
                    },
                  },
                },
              },
            },
          },
        ],
      })
    );
  });

  test("rejects Record expression with invalid field expr", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          { kind: "when", body: { Record: { foo: { nope: true } } } },
        ],
      })
    );
  });

  test("rejects Record expression with additional top-level keys", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              Record: { foo: { Value: true } },
              Set: [],
            },
          },
        ],
      })
    );
  });
});
