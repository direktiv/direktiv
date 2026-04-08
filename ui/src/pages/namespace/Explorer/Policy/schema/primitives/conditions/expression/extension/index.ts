import { isValidDecimalLiteral, isValidDurationLiteral } from "./utils";
import type { ExpressionSchemaType } from "../types";
import { strictSingleKeyObject } from "../utils";
import { z } from "zod";

const CedarExtensionConstructorNames = [
  "datetime",
  "decimal",
  "duration",
  "ip",
] as const;

const CedarExtensionMethodNames = [
  "isIpv4",
  "isIpv6",
  "isLoopback",
  "isMulticast",
  "isInRange",
  "offset",
  "durationSince",
  "toDate",
  "toTime",
  "toMilliseconds",
  "toSeconds",
  "toMinutes",
  "toHours",
  "toDays",
  "lessThan",
  "lessThanOrEqual",
  "greaterThan",
  "greaterThanOrEqual",
] as const;

const _CedarExtensionNames = [
  ...CedarExtensionConstructorNames,
  ...CedarExtensionMethodNames,
] as const;

export type ExtensionIdentifier = (typeof _CedarExtensionNames)[number];

// Cedar constructor validation only accepts string-literal Value expressions,
// not arbitrary child expressions that happen to evaluate to strings.
const literalStringValueSchema = z.object({ Value: z.string() }).strict();

const cedarDatetimeLiteralSchema = z.union([
  z.string().date(),
  z.string().datetime({ offset: true }),
]);

const cedarIpLiteralSchema = z.union([
  z.string().ip({ version: "v4" }),
  z.string().ip({ version: "v6" }),
  z.string().cidr({ version: "v4" }),
  z.string().cidr({ version: "v6" }),
]);

const decimalLiteralArgumentSchema = literalStringValueSchema.refine(
  ({ Value }) => isValidDecimalLiteral(Value),
  "decimal() requires a valid Cedar decimal literal"
);

// Slightly looser than Cedar because this does not enforce every fractional-
// second edge case from Cedar's datetime grammar, but it is good enough here.
const datetimeLiteralArgumentSchema = literalStringValueSchema.refine(
  ({ Value }) => cedarDatetimeLiteralSchema.safeParse(Value).success,
  "datetime() requires a valid Cedar datetime literal"
);

const durationLiteralArgumentSchema = literalStringValueSchema.refine(
  ({ Value }) => isValidDurationLiteral(Value),
  "duration() requires a valid Cedar duration literal"
);

const ipLiteralArgumentSchema = literalStringValueSchema.refine(
  ({ Value }) => cedarIpLiteralSchema.safeParse(Value).success,
  "ip() requires a valid Cedar IP literal"
);

const createExtensionCallSchema = <Name extends ExtensionIdentifier>(
  name: Name,
  argsSchema: z.ZodTypeAny
) => strictSingleKeyObject(name, argsSchema);

export const ExtensionExpressionSchema = (
  expressionSchema: ExpressionSchemaType
) => {
  const constructorSchemas = [
    // Cedar: decimal("100.00")
    createExtensionCallSchema(
      "decimal",
      z.tuple([decimalLiteralArgumentSchema])
    ),
    // Cedar: datetime("2024-10-15T11:35:00Z")
    createExtensionCallSchema(
      "datetime",
      z.tuple([datetimeLiteralArgumentSchema])
    ),
    // Cedar: duration("2h30m")
    createExtensionCallSchema(
      "duration",
      z.tuple([durationLiteralArgumentSchema])
    ),
    // Cedar: ip("10.0.0.0/8")
    createExtensionCallSchema("ip", z.tuple([ipLiteralArgumentSchema])),
  ] as const;

  const receiverOnlyMethodSchema = [
    // Cedar: ip("127.0.0.1").isIpv4()
    createExtensionCallSchema("isIpv4", z.tuple([expressionSchema])),
    // Cedar: ip("::1").isIpv6()
    createExtensionCallSchema("isIpv6", z.tuple([expressionSchema])),
    // Cedar: ip("127.0.0.1").isLoopback()
    createExtensionCallSchema("isLoopback", z.tuple([expressionSchema])),
    // Cedar: ip("ff00::2").isMulticast()
    createExtensionCallSchema("isMulticast", z.tuple([expressionSchema])),
    // Cedar: datetime("2024-10-15T11:35:00Z").toDate()
    createExtensionCallSchema("toDate", z.tuple([expressionSchema])),
    // Cedar: datetime("2024-10-15T11:35:00Z").toTime()
    createExtensionCallSchema("toTime", z.tuple([expressionSchema])),
    // Cedar: duration("2h30m").toMilliseconds()
    createExtensionCallSchema("toMilliseconds", z.tuple([expressionSchema])),
    // Cedar: duration("2h30m").toSeconds()
    createExtensionCallSchema("toSeconds", z.tuple([expressionSchema])),
    // Cedar: duration("2h30m").toMinutes()
    createExtensionCallSchema("toMinutes", z.tuple([expressionSchema])),
    // Cedar: duration("2h30m").toHours()
    createExtensionCallSchema("toHours", z.tuple([expressionSchema])),
    // Cedar: duration("48h").toDays()
    createExtensionCallSchema("toDays", z.tuple([expressionSchema])),
  ] as const;

  const receiverAndArgumentMethodSchema = [
    // Cedar: context.source_ip.isInRange(ip("10.0.0.0/8"))
    createExtensionCallSchema(
      "isInRange",
      z.tuple([expressionSchema, expressionSchema])
    ),
    // Cedar: datetime("2024-10-15T11:35:00Z").offset(duration("1h"))
    createExtensionCallSchema(
      "offset",
      z.tuple([expressionSchema, expressionSchema])
    ),
    // Cedar: datetime("2024-10-16T11:35:00Z").durationSince(datetime("2024-10-15T11:35:00Z"))
    createExtensionCallSchema(
      "durationSince",
      z.tuple([expressionSchema, expressionSchema])
    ),
    // Cedar: decimal("1.23").lessThan(decimal("1.24"))
    createExtensionCallSchema(
      "lessThan",
      z.tuple([expressionSchema, expressionSchema])
    ),
    // Cedar: decimal("1.23").lessThanOrEqual(decimal("1.24"))
    createExtensionCallSchema(
      "lessThanOrEqual",
      z.tuple([expressionSchema, expressionSchema])
    ),
    // Cedar: decimal("1.25").greaterThan(decimal("1.24"))
    createExtensionCallSchema(
      "greaterThan",
      z.tuple([expressionSchema, expressionSchema])
    ),
    // Cedar: decimal("1.24").greaterThanOrEqual(decimal("1.24"))
    createExtensionCallSchema(
      "greaterThanOrEqual",
      z.tuple([expressionSchema, expressionSchema])
    ),
  ] as const;

  return z.union([
    ...constructorSchemas,
    ...receiverOnlyMethodSchema,
    ...receiverAndArgumentMethodSchema,
  ]);
};
