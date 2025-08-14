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

type ResolverFunctionWithoutError<TType> = (
  value: VariableType,
  options?: {
    formEvent: FormEvent<HTMLFormElement>;
  }
) => TType;

type ResolverFunctionWithError<TType, TError> = (
  value: VariableType,
  options?: {
    formEvent: FormEvent<HTMLFormElement>;
  }
) => Result<TType, TError>;

export type ResolverFunction<TType, TError = never> = [TError] extends [never]
  ? ResolverFunctionWithoutError<TType>
  : ResolverFunctionWithError<TType, TError>;
