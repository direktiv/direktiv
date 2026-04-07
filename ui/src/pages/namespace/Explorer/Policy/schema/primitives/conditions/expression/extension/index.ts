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

const cedarIpLiteralSchema = z.union([
  z.string().ip({ version: "v4" }),
  z.string().ip({ version: "v6" }),
  z.string().cidr({ version: "v4" }),
  z.string().cidr({ version: "v6" }),
]);

const cedarDatetimeLiteralSchema = z
  .union([z.string().date(), z.string().datetime({ offset: true })])
  .refine(
    (value) => {
      if (!value.includes("T")) {
        return true;
      }

      if (value.endsWith("Z")) {
        return !value.includes(".") || /\.\d{3}Z$/.test(value);
      }

      return !value.includes(".") || /\.\d{3}[+-]\d{2}:\d{2}$/.test(value);
    },
    {
      message:
        "datetime() requires a valid Cedar datetime literal with optional millisecond precision",
    }
  );

const decimalLowerBound = BigInt("-9223372036854775808");
const decimalUpperBound = BigInt("9223372036854775807");
const longLowerBound = BigInt("-9223372036854775808");
const longUpperBound = BigInt("9223372036854775807");

// Cedar decimals allow up to 4 fractional digits and use a fixed precision.
// We normalize the fraction to 4 digits and compare the scaled integer value
// against Cedar's documented decimal range.
const isValidDecimalLiteral = (value: string) => {
  const match = value.match(/^(?<sign>-?)(?<whole>\d+)\.(?<fraction>\d{1,4})$/);

  if (!match?.groups) {
    return false;
  }

  const sign = match.groups.sign;
  const whole = match.groups.whole;
  const fraction = match.groups.fraction;

  if (sign === undefined || whole === undefined || fraction === undefined) {
    return false;
  }

  const scaledValue = BigInt(`${sign}${whole}${fraction.padEnd(4, "0")}`);

  return scaledValue >= decimalLowerBound && scaledValue <= decimalUpperBound;
};

const durationUnitOrder = {
  d: 0,
  h: 1,
  m: 2,
  s: 3,
  ms: 4,
} as const;

const durationUnitMultiplier = {
  d: 86400000n,
  h: 3600000n,
  m: 60000n,
  s: 1000n,
  ms: 1n,
} as const;

type DurationUnit = keyof typeof durationUnitOrder;

const isValidDurationLiteral = (value: string) => {
  const sign = value.startsWith("-") ? -1n : 1n;
  const unsignedValue = value.startsWith("-") ? value.slice(1) : value;

  if (unsignedValue.length === 0) {
    return false;
  }

  const segmentPattern = /(\d+)(ms|d|h|m|s)/g;
  const seenUnits = new Set<DurationUnit>();
  let lastUnitOrder = -1;
  let totalMilliseconds = 0n;
  let consumedLength = 0;

  for (const match of unsignedValue.matchAll(segmentPattern)) {
    const [segment, amountValue, unitValue] = match;

    if (amountValue === undefined || unitValue === undefined) {
      return false;
    }

    const unit = unitValue as DurationUnit;
    const unitOrder = durationUnitOrder[unit];

    if (match.index !== consumedLength || seenUnits.has(unit)) {
      return false;
    }

    if (unitOrder <= lastUnitOrder) {
      return false;
    }

    seenUnits.add(unit);
    lastUnitOrder = unitOrder;
    consumedLength += segment.length;

    const amount = BigInt(amountValue);
    const multiplier = durationUnitMultiplier[unit];

    totalMilliseconds += amount * multiplier;
  }

  if (consumedLength !== unsignedValue.length) {
    return false;
  }

  const signedMilliseconds = totalMilliseconds * sign;

  return (
    signedMilliseconds >= longLowerBound && signedMilliseconds <= longUpperBound
  );
};

const decimalLiteralArgumentSchema = literalStringValueSchema.refine(
  ({ Value }) => isValidDecimalLiteral(Value),
  "decimal() requires a valid Cedar decimal literal"
);

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
