export type ValidateVariableError =
  | "namespaceInvalid"
  | "idUndefined"
  | "pointerUndefined";

export type JsonPathError = "invalidJson" | "invalidPath";

type VariableError = "NoStateForId";

type ArrayError = "notAnArray";

type StringArrayError = "notAnArrayOfStrings";

type BooleanError = "notABoolean";

type NumberError = "notANumber";

type StringifyError = "couldNotStringify";

export type ResolveVariableError =
  | ValidateVariableError
  | JsonPathError
  | VariableError;

export type ResolveVariableArrayError = ResolveVariableError | ArrayError;

export type ResolveVariableStringArrayError =
  | ResolveVariableArrayError
  | StringArrayError;

export type ResolveVariableStringError = ResolveVariableError | StringifyError;

export type ResolveVariableBooleanError = ResolveVariableError | BooleanError;

export type ResolveVariableNumberError = ResolveVariableError | NumberError;
