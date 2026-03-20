import { CedarJsonSchema, type CedarJsonSchemaInputType } from "..";
import { describe, expect, test } from "vitest";

const expectValidCedarJsonSchema = (input: CedarJsonSchemaInputType) => {
  expect(CedarJsonSchema.safeParse(input).success).toBe(true);
  expect(CedarJsonSchema.parse(input)).toEqual(input);
};

const expectInvalidCedarJsonSchema = (input: CedarJsonSchemaInputType) => {
  expect(CedarJsonSchema.safeParse(input).success).toBe(false);
};

const createBaseSchema = (
  overrides: Partial<CedarJsonSchemaInputType[string]> = {}
): CedarJsonSchemaInputType => ({
  Demo: {
    entityTypes: {
      User: {},
    },
    actions: {
      view: {
        appliesTo: {
          principalTypes: ["User"],
          resourceTypes: ["User"],
        },
      },
    },
    ...overrides,
  },
});

describe("Cedar JSON schema", () => {
  test("accepts namespaces with entity types, actions, annotations, and common types", () => {
    const input: CedarJsonSchemaInputType = {
      PhotoFlash: {
        annotations: {
          doc: "Photo sharing app",
        },
        entityTypes: {
          User: {
            annotations: {
              doc: "Application user",
            },
            memberOfTypes: ["UserGroup"],
            shape: {
              type: "Record",
              attributes: {
                department: {
                  type: "String",
                },
                profile: {
                  type: "Record",
                  attributes: {
                    age: {
                      type: "Long",
                      required: false,
                    },
                    location: {
                      type: "EntityOrCommon",
                      name: "GeoContext",
                    },
                  },
                },
              },
            },
            tags: {
              type: "String",
            },
          },
          UserGroup: {
            enum: ["admins", "reviewers"],
            annotations: {
              doc: "Known user groups",
            },
          },
        },
        commonTypes: {
          GeoContext: {
            type: "Record",
            attributes: {
              ip: {
                type: "Extension",
                name: "ipaddr",
              },
              region: {
                type: "String",
              },
            },
          },
        },
        actions: {
          read: {
            appliesTo: {
              principalTypes: [],
              resourceTypes: [],
            },
          },
          viewPhoto: {
            annotations: {
              doc: "Views a photo",
            },
            memberOf: [
              {
                id: "read",
              },
              {
                id: "viewAnything",
                type: "Shared::Action",
              },
            ],
            appliesTo: {
              principalTypes: ["User"],
              resourceTypes: ["UserGroup"],
              context: {
                type: "GeoContext",
              },
            },
          },
        },
      },
    };

    expect(CedarJsonSchema.parse(input)).toEqual(input);
  });

  test("accepts the empty namespace", () => {
    const input: CedarJsonSchemaInputType = {
      "": {
        entityTypes: {
          User: {},
        },
        actions: {
          view: {
            appliesTo: {
              principalTypes: ["User"],
              resourceTypes: ["User"],
            },
          },
        },
      },
    };

    expect(CedarJsonSchema.safeParse(input).success).toBe(true);
  });

  test("rejects reserved namespace segments and invalid action group references", () => {
    expect(
      CedarJsonSchema.safeParse({
        __cedar: {
          entityTypes: {
            User: {},
          },
          actions: {
            view: {
              appliesTo: {
                principalTypes: ["User"],
                resourceTypes: ["User"],
              },
            },
          },
        },
      }).success
    ).toBe(false);

    expect(
      CedarJsonSchema.safeParse({
        Demo: {
          entityTypes: {
            User: {},
          },
          actions: {
            view: {
              memberOf: [
                {
                  id: "read",
                  type: "Shared::NotAction",
                },
              ],
              appliesTo: {
                principalTypes: ["User"],
                resourceTypes: ["User"],
              },
            },
          },
        },
      }).success
    ).toBe(false);
  });

  test("accepts valid nested input", () => {
    const input: CedarJsonSchemaInputType = createBaseSchema({
      commonTypes: {
        GeoContext: {
          type: "Record",
          attributes: {
            ip: {
              type: "Extension",
              name: "ipaddr",
            },
          },
        },
      },
      entityTypes: {
        User: {
          shape: {
            type: "Record",
            attributes: {
              profile: {
                type: "Record",
                attributes: {
                  location: {
                    type: "EntityOrCommon",
                    name: "GeoContext",
                    required: false,
                  },
                },
              },
            },
          },
        },
      },
    });

    expectValidCedarJsonSchema(input);
  });

  test("rejects extra namespace keys at type level", () => {
    expectInvalidCedarJsonSchema(
      createBaseSchema({
        // @ts-expect-error - namespace declarations are strict
        invalid: {},
      })
    );
  });

  test("rejects non-string annotation values at type level", () => {
    expectInvalidCedarJsonSchema(
      createBaseSchema({
        entityTypes: {
          User: {
            annotations: {
              // @ts-expect-error - annotation values must be strings
              doc: 123,
            },
          },
        },
      })
    );
  });

  test("rejects invalid action memberOf entry shapes at type level", () => {
    expectInvalidCedarJsonSchema(
      createBaseSchema({
        actions: {
          view: {
            memberOf: [
              {
                // @ts-expect-error - action group ids must be strings
                id: 123,
              },
            ],
            appliesTo: {
              principalTypes: ["User"],
              resourceTypes: ["User"],
            },
          },
        },
      })
    );

    expectInvalidCedarJsonSchema(
      createBaseSchema({
        actions: {
          view: {
            memberOf: [
              {
                id: "read",
                // @ts-expect-error - action group references are strict
                foo: "bar",
              },
            ],
            appliesTo: {
              principalTypes: ["User"],
              resourceTypes: ["User"],
            },
          },
        },
      })
    );
  });

  test("rejects appliesTo values with wrong types at type level", () => {
    expectInvalidCedarJsonSchema(
      createBaseSchema({
        actions: {
          view: {
            appliesTo: {
              // @ts-expect-error - principalTypes must be an array
              principalTypes: "User",
              resourceTypes: ["User"],
            },
          },
        },
      })
    );

    expectInvalidCedarJsonSchema(
      createBaseSchema({
        actions: {
          view: {
            appliesTo: {
              principalTypes: ["User"],
              resourceTypes: ["User"],
              // @ts-expect-error - required must be boolean
              context: {
                type: "Record",
                attributes: {
                  isAuthenticated: {
                    type: "Boolean",
                    required: "sometimes",
                  },
                },
              },
            },
          },
        },
      })
    );
  });

  test("rejects enum and entity references with wrong value types at type level", () => {
    expectInvalidCedarJsonSchema(
      createBaseSchema({
        entityTypes: {
          User: {
            // @ts-expect-error - enum entries must be strings
            enum: [1, 2, 3],
          },
        },
      })
    );

    expectInvalidCedarJsonSchema(
      createBaseSchema({
        entityTypes: {
          User: {
            // @ts-expect-error - entity references require a string name
            shape: {
              type: "Record",
              attributes: {
                manager: {
                  type: "Entity",
                  name: 99,
                },
              },
            },
          },
        },
      })
    );
  });
});
