import { AttributeExpressionSchema } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/attribute";
import { BinaryExpressionSchema } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/binary";
import { ExpressionSchema } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression";
import type { ExpressionType } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/types";
import { IfThenElseExpressionSchema } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/ifThenElse";
import { IsExpressionSchema } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/is";
import { LikeExpressionSchema } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/like";
import type { PatternElement } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/like";
import { RecordExpressionSchema } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/record";
import { SetExpressionSchema } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/set";
import { SlotExpressionSchema } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/slot";
import { UnaryExpressionSchema } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/unary";
import { UnknownExpressionSchema } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/unknown";
import { ValueExpressionSchema } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/value";
import { VarExpressionSchema } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/var";

type UnaryPayload = { arg: ExpressionType };
type BinaryPayload = { left: ExpressionType; right: ExpressionType };

const fallBackExpressionUnary: UnaryPayload = {
  arg: { Value: "<missing-arg>" },
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

  const formattedType = String(type);
  const formattedId = JSON.stringify(String(id));

  return `${formattedType}::${formattedId}`;
};

const formatUnknownValue = (value: unknown) => {
  if (isEntityValue(value)) return formatEntityValue(value);
  if (typeof value === "string") return JSON.stringify(value);
  if (
    typeof value === "number" ||
    typeof value === "boolean" ||
    value === null
  ) {
    return String(value);
  }
  return JSON.stringify(value) ?? String(value);
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

const formatIfThenElse = (expression: {
  "if-then-else": {
    if: ExpressionType;
    then: ExpressionType;
    else: ExpressionType;
  };
}) => {
  const {
    "if-then-else": { if: ifExpr, then: thenExpr, else: elseExpr },
  } = expression;
  return `if ${formatExpression(ifExpr)} then ${formatExpression(
    thenExpr
  )} else ${formatExpression(elseExpr)}`;
};

const formatSet = (expression: { Set: ExpressionType[] }) =>
  `[${expression.Set.map((item) => formatExpression(item)).join(", ")}]`;

const getSingleObjectKey = <T extends Record<string, unknown>>(
  obj: T,
  errorMessage: string
): keyof T => {
  const keys = Object.keys(obj);
  if (keys.length !== 1) {
    throw new Error(errorMessage);
  }
  return keys[0] as keyof T;
};

const formatBinaryOperand = (operand: ExpressionType): string => {
  if ("Var" in operand) return operand.Var;
  if ("Value" in operand) return formatUnknownValue(operand.Value);
  if ("Slot" in operand) return operand.Slot;
  return formatExpression(operand);
};

const formatBinaryExpression = (expression: {
  [key: string]: BinaryPayload;
}) => {
  const operator = getSingleObjectKey(
    expression,
    "Binary expression is missing its operator"
  );
  const payload = expression[operator]!;

  switch (operator) {
    case "contains":
      return `${formatExpression(payload.left)}.contains(${formatExpression(payload.right)})`;
    case "containsAll":
      return `${formatExpression(payload.left)}.containsAll(${formatExpression(payload.right)})`;
    case "containsAny":
      return `${formatExpression(payload.left)}.containsAny(${formatExpression(payload.right)})`;
    case "hasTag":
      return `${formatExpression(payload.left)}.hasTag(${formatExpression(payload.right)})`;
    case "getTag":
      return `${formatExpression(payload.left)}.getTag(${formatExpression(payload.right)})`;
    case "==":
    case "!=":
    case "in":
    case "<":
    case "<=":
    case ">":
    case ">=":
    case "&&":
    case "||":
    case "+":
    case "-":
    case "*":
      return `${formatBinaryOperand(payload.left)} ${operator} ${formatBinaryOperand(payload.right)}`;
    default:
      throw new Error(`Unsupported binary operator: ${String(operator)}`);
  }
};

const formatIsExpression = (expression: {
  is: { left: ExpressionType; entity_type: string; in?: ExpressionType };
}) => {
  const {
    is: { left, entity_type, in: inExpression },
  } = expression;
  const leftText = formatExpression(left);

  if (inExpression === undefined) {
    return `${leftText} is ${entity_type}`;
  }

  return `${leftText} is ${entity_type} in ${formatExpression(inExpression)}`;
};

export const formatExpression = (expression: ExpressionType): string => {
  const binaryExpression =
    BinaryExpressionSchema(ExpressionSchema).safeParse(expression);

  if (binaryExpression.success) {
    return formatBinaryExpression(binaryExpression.data);
  }

  const isExpression =
    IsExpressionSchema(ExpressionSchema).safeParse(expression);

  if (isExpression.success) {
    return formatIsExpression(isExpression.data);
  }

  const varResult = VarExpressionSchema.safeParse(expression);
  if (varResult.success) {
    return varResult.data.Var;
  }

  const valueResult = ValueExpressionSchema.safeParse(expression);
  if (valueResult.success) {
    return formatUnknownValue(valueResult.data.Value);
  }

  const slotResult = SlotExpressionSchema.safeParse(expression);
  if (slotResult.success) {
    return slotResult.data.Slot;
  }

  const unknownResult = UnknownExpressionSchema.safeParse(expression);
  if (unknownResult.success) {
    const firstEntry = Object.entries(unknownResult.data.Unknown)[0];
    if (firstEntry === undefined) return "unknown()";
    return `unknown(${JSON.stringify(firstEntry[1])})`;
  }

  const attributeResult =
    AttributeExpressionSchema(ExpressionSchema).safeParse(expression);
  if (attributeResult.success) {
    const data = attributeResult.data;
    if ("." in data) {
      return `${formatExpression(data["."].left)}.${data["."].attr}`;
    }
    if ("has" in data) {
      return `${formatExpression(data.has.left)} has ${data.has.attr}`;
    }
  }

  const likeResult =
    LikeExpressionSchema(ExpressionSchema).safeParse(expression);
  if (likeResult.success) {
    const { left, pattern } = likeResult.data.like;
    return `${formatExpression(left)} like "${formatLikePattern(pattern)}"`;
  }

  const ifThenElseResult =
    IfThenElseExpressionSchema(ExpressionSchema).safeParse(expression);
  if (ifThenElseResult.success) {
    return formatIfThenElse(ifThenElseResult.data);
  }

  const setResult = SetExpressionSchema(ExpressionSchema).safeParse(expression);
  if (setResult.success) {
    return formatSet(setResult.data);
  }

  const recordResult =
    RecordExpressionSchema(ExpressionSchema).safeParse(expression);
  if (recordResult.success) {
    return formatRecord(recordResult.data.Record);
  }

  const unaryResult =
    UnaryExpressionSchema(ExpressionSchema).safeParse(expression);
  if (unaryResult.success) {
    const data = unaryResult.data;
    const unaryOperator = Object.keys(data)[0] as keyof typeof data;
    const payload =
      (data[unaryOperator] as UnaryPayload) ?? fallBackExpressionUnary;
    if (unaryOperator === "!") return `!${formatExpression(payload.arg)}`;
    if (unaryOperator === "neg") return `-${formatExpression(payload.arg)}`;
    return `${unaryOperator}(${formatExpression(payload.arg)})`;
  }

  return JSON.stringify(expression) ?? String(expression);
};
