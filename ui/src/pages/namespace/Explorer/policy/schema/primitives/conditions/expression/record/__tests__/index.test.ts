import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../../../utils/testutils";
import { describe, test } from "vitest";

describe("Record Expression schema", () => {
  test("accepts Record expression", () => {
    // Cedar: when { {"foo": "spam", "somethingelse": false}.foo == "spam" };
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              "==": {
                left: {
                  ".": {
                    left: {
                      Record: {
                        foo: { Value: "spam" },
                        somethingelse: { Value: false },
                      },
                    },
                    attr: "foo",
                  },
                },
                right: { Value: "spam" },
              },
            },
          },
        ],
      })
    );
  });

  test("accepts nested Record expression for user profile", () => {
    // Cedar: when { {"user": {"profile": {"name": "alice", "isActive": true}}}.user.profile.isActive };
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              ".": {
                left: {
                  ".": {
                    left: {
                      ".": {
                        left: {
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
                        attr: "user",
                      },
                    },
                    attr: "profile",
                  },
                },
                attr: "isActive",
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
