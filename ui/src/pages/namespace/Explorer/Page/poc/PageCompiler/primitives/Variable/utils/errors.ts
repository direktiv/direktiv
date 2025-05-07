export type ValidateVariableError =
  | "namespaceUndefined"
  | "idUndefined"
  | "pointerUndefined";

export type JsonPathError = "invalidJson" | "invalidPath";

type VariableError = "NoStateForId";

type ArrayError = "notAnArray";

type JSXError = "couldNotStringify";

export type ResolveVariableError =
  | ValidateVariableError
  | JsonPathError
  | VariableError;

export type ResolveVariableArrayError = ResolveVariableError | ArrayError;

export type ResolveVariableJSXError = ResolveVariableError | JSXError;
