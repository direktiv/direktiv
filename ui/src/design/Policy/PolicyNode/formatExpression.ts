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

type BinaryPayload = { left: ExpressionType; right: ExpressionType };

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

const formatUnknownValue = (object: { value: unknown }) => {
  const value = object.value;

  if (isEntityValue(value)) return formatEntityValue(value);
  if (
    typeof value === "number" ||
    typeof value === "boolean" ||
    value === null
  ) {
    return String(value);
  }
  if (typeof value === "string") return JSON.stringify(value);
  return JSON.stringify(value) ?? String(value);
};

const formatUnknownResult = (unknownExpression: { Unknown: unknown }) => {
  const firstEntry = Object.entries({ Unknown: unknownExpression.Unknown })[0];
  if (firstEntry === undefined) return "unknown()";
  return `unknown(${JSON.stringify(firstEntry[1])})`;
};

const formatLikePattern = (pattern: PatternElement[]) =>
  pattern
    .map((patternElement) =>
      patternElement === "Wildcard" ? "*" : patternElement.Literal
    )
    .join("");

const formatRecordExpression = (record: {
  Record: Record<string, ExpressionType>;
}) =>
  `{${Object.entries(record.Record)
    .map(([key, value]) => `${JSON.stringify(key)}: ${formatExpression(value)}`)
    .join(", ")}}`;

const formatIfThenElseExpression = (expression: {
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

const formatSetExpression = (expression: { Set: ExpressionType[] }) =>
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
  if ("Value" in operand && typeof operand.Value !== "object") {
    return formatUnknownValue({ value: operand.Value });
  }
  if ("Slot" in operand) return operand.Slot;
  return formatExpression(operand);
};

const formatAttributeExpression = (
  data:
    | {
        ".": { left: ExpressionType; attr: string };
      }
    | {
        has: { left: ExpressionType; attr: string };
      }
) => {
  if ("." in data) {
    return `${formatExpression(data["."].left)}.${data["."].attr}`;
  }
  if ("has" in data) {
    return `${formatExpression(data.has.left)} has ${data.has.attr}`;
  }

  return "unknown attribute";
};

const formatUnaryExpression = (expression: {
  [key: string]: { arg: ExpressionType };
}) => {
  const keys = Object.keys(expression);
  const unary = keys[0] ?? "isEmpty";
  const operand = expression[unary]?.arg ?? { Value: "<missing-arg>" };
  return `${unary}(${formatExpression(operand)})`;
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

const formatLikeExpression = (expression: {
  like: { left: ExpressionType; pattern: PatternElement[] };
}) => {
  const {
    like: { left, pattern },
  } = expression;
  return `${formatExpression(left)} like "${formatLikePattern(pattern)}"`;
};

const formatVarExpression = (varExpression: { Var: string }) =>
  varExpression.Var;

const formatSlotExpression = (slotExpression: { Slot: string }) =>
  slotExpression.Slot;

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
    return formatVarExpression(varResult.data);
  }

  const valueResult = ValueExpressionSchema.safeParse(expression);

  if (valueResult.success) {
    return formatUnknownValue({ value: valueResult.data.Value });
  }

  const slotResult = SlotExpressionSchema.safeParse(expression);

  if (slotResult.success) {
    return formatSlotExpression(slotResult.data);
  }

  const unknownResult = UnknownExpressionSchema.safeParse(expression);

  if (unknownResult.success) {
    return formatUnknownResult(unknownResult.data);
  }

  const attributeResult =
    AttributeExpressionSchema(ExpressionSchema).safeParse(expression);

  if (attributeResult.success) {
    return formatAttributeExpression(attributeResult.data);
  }

  const likeResult =
    LikeExpressionSchema(ExpressionSchema).safeParse(expression);

  if (likeResult.success) {
    return formatLikeExpression(likeResult.data);
  }

  const ifThenElseResult =
    IfThenElseExpressionSchema(ExpressionSchema).safeParse(expression);
  if (ifThenElseResult.success) {
    return formatIfThenElseExpression(ifThenElseResult.data);
  }

  const setResult = SetExpressionSchema(ExpressionSchema).safeParse(expression);
  if (setResult.success) {
    return formatSetExpression(setResult.data);
  }

  const recordResult =
    RecordExpressionSchema(ExpressionSchema).safeParse(expression);
  if (recordResult.success) {
    return formatRecordExpression(recordResult.data);
  }

  const unaryResult =
    UnaryExpressionSchema(ExpressionSchema).safeParse(expression);
  if (unaryResult.success) {
    return formatUnaryExpression(
      unaryResult.data as {
        [key: string]: { arg: ExpressionType };
      }
    );
  }

  return JSON.stringify(expression) ?? String(expression);
};
