import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../../../utils/testutils";
import { describe, test } from "vitest";

type ExtensionBodyInput = ReturnType<
  typeof createBasePolicy
>["conditions"][number]["body"];

const expectValidExtensionBody = (body: ExtensionBodyInput) => {
  expectValidPolicy(
    createBasePolicy({
      conditions: [{ kind: "when", body }],
    })
  );
};

describe("Extension Expression schema", () => {
  test("accepts decimal constructor expression", () => {
    expectValidExtensionBody({ decimal: [{ Value: "10.0" }] });
  });

  test("accepts datetime constructor expression", () => {
    expectValidExtensionBody({ datetime: [{ Value: "2024-10-15T11:35:00Z" }] });
  });

  test("accepts duration constructor expression", () => {
    expectValidExtensionBody({ duration: [{ Value: "2h30m" }] });
  });

  test("accepts ip constructor expression", () => {
    expectValidExtensionBody({ ip: [{ Value: "127.0.0.1/24" }] });
  });

  test("accepts isIpv4 method expression", () => {
    expectValidExtensionBody({ isIpv4: [{ ip: [{ Value: "127.0.0.1" }] }] });
  });

  test("accepts isIpv6 method expression", () => {
    expectValidExtensionBody({ isIpv6: [{ ip: [{ Value: "::1" }] }] });
  });

  test("accepts isLoopback method expression", () => {
    expectValidExtensionBody({
      isLoopback: [{ ip: [{ Value: "127.0.0.1" }] }],
    });
  });

  test("accepts isMulticast method expression", () => {
    expectValidExtensionBody({
      isMulticast: [{ ip: [{ Value: "ff00::2" }] }],
    });
  });

  test("accepts isInRange method expression", () => {
    expectValidExtensionBody({
      isInRange: [
        { ip: [{ Value: "10.0.0.1" }] },
        { ip: [{ Value: "10.0.0.0/8" }] },
      ],
    });
  });

  test("accepts offset method expression", () => {
    expectValidExtensionBody({
      offset: [
        { datetime: [{ Value: "2024-10-15T11:35:00Z" }] },
        { duration: [{ Value: "1h" }] },
      ],
    });
  });

  test("accepts durationSince method expression", () => {
    expectValidExtensionBody({
      durationSince: [
        { datetime: [{ Value: "2024-10-16T11:35:00Z" }] },
        { datetime: [{ Value: "2024-10-15T11:35:00Z" }] },
      ],
    });
  });

  test("accepts toDate method expression", () => {
    expectValidExtensionBody({
      toDate: [{ datetime: [{ Value: "2024-10-15T11:35:00Z" }] }],
    });
  });

  test("accepts toTime method expression", () => {
    expectValidExtensionBody({
      toTime: [{ datetime: [{ Value: "2024-10-15T11:35:00Z" }] }],
    });
  });

  test("accepts toMilliseconds method expression", () => {
    expectValidExtensionBody({
      toMilliseconds: [{ duration: [{ Value: "1h30m45s" }] }],
    });
  });

  test("accepts toSeconds method expression", () => {
    expectValidExtensionBody({
      toSeconds: [{ duration: [{ Value: "1h30m45s" }] }],
    });
  });

  test("accepts toMinutes method expression", () => {
    expectValidExtensionBody({
      toMinutes: [{ duration: [{ Value: "1h30m45s" }] }],
    });
  });

  test("accepts toHours method expression", () => {
    expectValidExtensionBody({
      toHours: [{ duration: [{ Value: "1h30m45s" }] }],
    });
  });

  test("accepts toDays method expression", () => {
    expectValidExtensionBody({
      toDays: [{ duration: [{ Value: "48h" }] }],
    });
  });

  test("accepts lessThan method expression", () => {
    expectValidExtensionBody({
      lessThan: [
        { decimal: [{ Value: "1.23" }] },
        { decimal: [{ Value: "1.24" }] },
      ],
    });
  });

  test("accepts lessThanOrEqual method expression", () => {
    expectValidExtensionBody({
      lessThanOrEqual: [
        { decimal: [{ Value: "1.23" }] },
        { decimal: [{ Value: "1.23" }] },
      ],
    });
  });

  test("accepts greaterThan method expression", () => {
    expectValidExtensionBody({
      greaterThan: [
        { decimal: [{ Value: "2.00" }] },
        { decimal: [{ Value: "1.24" }] },
      ],
    });
  });

  test("accepts greaterThanOrEqual method expression", () => {
    expectValidExtensionBody({
      greaterThanOrEqual: [
        { decimal: [{ Value: "2.00" }] },
        { decimal: [{ Value: "2.00" }] },
      ],
    });
  });

  test("rejects unknown extension names", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            // @ts-expect-error - only documented Cedar extension names are allowed
            body: { customExtension: [{ Value: "10.0" }] },
          },
        ],
      })
    );
  });

  test("rejects reserved expression keys as extensions", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            // @ts-expect-error - reserved expression keys cannot be extensions
            body: { is: [{ Value: "10.0" }] },
          },
        ],
      })
    );
  });

  test("rejects constructor expressions with non-array args", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            // @ts-expect-error - constructor args must be tuples
            body: { decimal: { Value: "10.0" } },
          },
        ],
      })
    );
  });

  test("rejects malformed decimal literals", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: { decimal: [{ Value: "1234" }] },
          },
        ],
      })
    );
  });

  test("rejects malformed datetime literals", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: { datetime: [{ Value: "2024-10-15T11:38:02ZZ" }] },
          },
        ],
      })
    );
  });

  test("rejects malformed duration literals", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: { duration: [{ Value: "1s1d" }] },
          },
        ],
      })
    );
  });

  test("rejects malformed ip literals", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: { ip: [{ Value: "380.0.0.1" }] },
          },
        ],
      })
    );
  });

  test("rejects methods with the wrong arity", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            // @ts-expect-error - isInRange requires receiver and range
            body: { isInRange: [{ ip: [{ Value: "10.0.0.1" }] }] },
          },
        ],
      })
    );
  });

  test("rejects receiver-only methods with extra args", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              // @ts-expect-error - toDate only accepts the receiver
              toDate: [
                { datetime: [{ Value: "2024-10-15T11:35:00Z" }] },
                { duration: [{ Value: "1h" }] },
              ],
            },
          },
        ],
      })
    );
  });
});
