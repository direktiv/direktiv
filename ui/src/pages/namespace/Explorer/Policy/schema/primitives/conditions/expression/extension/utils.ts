import { z } from "zod";

export const cedarIpLiteralSchema = z.union([
  z.string().ip({ version: "v4" }),
  z.string().ip({ version: "v6" }),
  z.string().cidr({ version: "v4" }),
  z.string().cidr({ version: "v6" }),
]);

export const cedarDatetimeLiteralSchema = z.union([
  z.string().date(),
  z.string().datetime({ offset: true }),
]);

const longLowerBound = BigInt("-9223372036854775808");
const longUpperBound = BigInt("9223372036854775807");

// Good-enough decimal validation for this schema layer: require a Cedar-like
// decimal string shape, but leave overflow and exact bound checks out.
export const isValidDecimalLiteral = (value: string) =>
  /^-?\d+\.\d{1,4}$/.test(value);

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

export const isValidDurationLiteral = (value: string) => {
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
