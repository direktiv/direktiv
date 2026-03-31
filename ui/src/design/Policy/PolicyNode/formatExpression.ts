import {
  ExpressionBinaryOperators,
  ExpressionUnaryOperators,
} from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/utils";

import type { ExpressionType } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/types";
import type { PatternElement } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/like";

type UnaryPayload = { arg: ExpressionType };
type BinaryPayload = { left: ExpressionType; right: ExpressionType };

const fallBackExpressionUnary: UnaryPayload = {
  arg: { Value: "<missing-arg>" },
};

const fallbackExpressionBinary: BinaryPayload = {
  left: { Value: "<missing-left>" },
  right: { Value: "<missing-right>" },
};

const isEntityValue = (
  value: unknown
): value is { __entity: { type: unknown; id: unknown } } => {
  if (
    value === null ||
    typeof value !== "object" ||
    !("__entity" in value) ||
    value.__entity === null ||
    typeof value.__entity !== "object"
  ) {
    return false;
  }

  const entity = value.__entity as { type?: unknown; id?: unknown };

  return "type" in entity && "id" in entity;
};

const formatEntityValue = (value: {
  __entity: { type: unknown; id: unknown };
}) => {
  const { type, id } = value.__entity;

  const formattedType = String(type).toLowerCase();
  const formattedId = String(id).toLowerCase();

  return `type: ${formattedType} id: ${formattedId}`;
};

const isPrimitive = (
  value: unknown
): value is string | number | boolean | null =>
  typeof value === "string" ||
  typeof value === "number" ||
  typeof value === "boolean" ||
  value === null;

const isRecordLikeValue = (
  value: unknown
): value is Record<string, string | number | boolean | null> => {
  if (value === null || typeof value !== "object" || Array.isArray(value)) {
    return false;
  }

  return Object.values(value).every(isPrimitive);
};

const formatUnknownValue = (value: unknown) => {
  if (isEntityValue(value)) return formatEntityValue(value);
  if (isRecordLikeValue(value)) {
    return Object.entries(value)
      .map(([key, v]) => `${key}: ${String(v)}`)
      .join(", ");
  }
  if (typeof value === "string") return JSON.stringify(value);
  if (
    typeof value === "number" ||
    typeof value === "boolean" ||
    value === null
  ) {
    return String(value);
  }
  return String(value);
};

const formatLikePattern = (pattern: PatternElement[]) =>
  pattern
    .map((patternElement) =>
      patternElement === "Wildcard" ? "*" : patternElement.Literal
    )
    .join("");

const formatRecord = (record: Record<string, ExpressionType>) =>
  `{${Object.entries(record)
    .map(([key, value]) => `${JSON.stringify(key)}: ${formatExpression(value)}`)
    .join(", ")}}`;

export const formatExpression = (expression: ExpressionType): string => {
  if ("Var" in expression) return expression.Var;
  if ("Value" in expression) return formatUnknownValue(expression.Value);
  if ("Slot" in expression) return expression.Slot;

  if ("Unknown" in expression) {
    const firstEntry = Object.entries(expression.Unknown)[0];
    if (firstEntry === undefined) return "unknown()";

    return `unknown(${JSON.stringify(firstEntry[1])})`;
  }

  if ("." in expression) {
    const { left, attr } = expression["."];
    return `${formatExpression(left)}.${attr}`;
  }

  if ("has" in expression) {
    const { left, attr } = expression.has;
    return `${formatExpression(left)} has ${attr}`;
  }

  if ("is" in expression) {
    const { left, entity_type: entityType, in: inExpr } = expression.is;
    if (inExpr === undefined)
      return `${formatExpression(left)} is ${entityType}`;

    return `${formatExpression(left)} is ${entityType} in ${formatExpression(inExpr)}`;
  }

  if ("like" in expression) {
    const { left, pattern } = expression.like;
    return `${formatExpression(left)} like ${formatLikePattern(pattern)}`;
  }

  if ("if-then-else" in expression) {
    const {
      if: ifExpr,
      then: thenExpr,
      else: elseExpr,
    } = expression["if-then-else"];
    return `if ${formatExpression(ifExpr)} then ${formatExpression(thenExpr)} else ${formatExpression(elseExpr)}`;
  }

  if ("Set" in expression) {
    return `[${expression.Set.map((item) => formatExpression(item)).join(", ")}]`;
  }

  if ("Record" in expression) return formatRecord(expression.Record);

  if ("getTag" in expression) {
    const { left, right } = expression.getTag;
    if ("Value" in right && typeof right.Value === "string") {
      return `${formatExpression(left)} getTag ${right.Value}`;
    }

    return `${formatExpression(left)} getTag ${formatExpression(right)}`;
  }

  for (const operator of ExpressionUnaryOperators) {
    if (!(operator in expression)) continue;

    const payload =
      (expression as Record<string, UnaryPayload>)[operator] ??
      fallBackExpressionUnary;
    if (operator === "!") return `!${formatExpression(payload.arg)}`;
    if (operator === "neg") return `-${formatExpression(payload.arg)}`;

    return `${operator}(${formatExpression(payload.arg)})`;
  }

  for (const operator of ExpressionBinaryOperators) {
    if (!(operator in expression)) continue;
    const payload =
      (
        expression as Partial<
          Record<(typeof ExpressionBinaryOperators)[number], BinaryPayload>
        >
      )[operator] ?? fallbackExpressionBinary;
    return `${formatExpression(payload.left)} ${operator} ${formatExpression(payload.right)}`;
  }

  return JSON.stringify(expression) ?? String(expression);
};
