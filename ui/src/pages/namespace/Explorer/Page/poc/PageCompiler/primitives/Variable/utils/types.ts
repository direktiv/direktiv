import { FormEvent } from "react";
import { VariableType } from "../../../../schema/primitives/variable";

export type Result<T, E> = Success<T> | Failure<E>;

type Success<T> = {
  success: true;
  data: T;
};

type Failure<E> = {
  success: false;
  error: E;
};

export type ResolverFunction<TType> = (
  value: VariableType,
  options?: {
    formEvent: FormEvent<HTMLFormElement>;
  }
) => TType;

export type ResolverFunctionWithError<TType, TError> = (
  value: VariableType,
  options?: {
    formEvent: FormEvent<HTMLFormElement>;
  }
) => Result<TType, TError>;
