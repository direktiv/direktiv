export type ValidateVariableError =
  | "namespaceInvalid"
  | "idUndefined"
  | "pointerUndefined";

export type JsonPathError = "invalidJson" | "invalidPath";

type VariableError = "NoStateForId";

type ArrayError = "notAnArray";

type StringifyError = "couldNotStringify";

export type ResolveVariableError =
  | ValidateVariableError
  | JsonPathError
  | VariableError;

export type ResolveVariableArrayError = ResolveVariableError | ArrayError;

export type ResolveVariableStringError = ResolveVariableError | StringifyError;
