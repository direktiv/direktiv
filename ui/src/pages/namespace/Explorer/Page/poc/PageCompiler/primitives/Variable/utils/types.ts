import { FormEvent } from "react";
import { VariableType } from "../../../../schema/primitives/variable";

export type ValidationResult<DataType, E> = Success<DataType> | Failure<E>;

type Success<DataType> = {
  success: true;
  data: DataType;
};

type Failure<E> = {
  success: false;
  error: E;
};

type ResolverFunctionWithoutError<DataType> = (
  value: VariableType,
  options?: {
    formEvent: FormEvent<HTMLFormElement>;
  }
) => DataType;

type ResolverFunctionWithError<DataType, Error> = (
  value: VariableType,
  options?: {
    formEvent: FormEvent<HTMLFormElement>;
  }
) => ValidationResult<DataType, Error>;

/**
 * Unified resolver function type that conditionally returns either a direct
 * value or a ValidationResult type based on whether an error type is provided
 */
export type ResolverFunction<DataType, Error = never> = [Error] extends [never]
  ? ResolverFunctionWithoutError<DataType>
  : ResolverFunctionWithError<DataType, Error>;
