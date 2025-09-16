import { KeyValueResolverFunction } from "./utils";
import { KeyValueType } from "../../../schema/primitives/keyValue";
import { useStringInterpolation } from "../Variable/utils/useStringInterpolation";

/**
 * Hook that returns a function to resolve an array of key-value pairs by interpolating
 * template variables in each value using provided local and context variables.
 *
 * keyValueArray:
 * [
 *   { key: "name", value: "{{this.firstName}} {{this.lastName}}" },
 *   { key: "created-at", value: "{{query.user.data.createdAt}}" }
 * ]
 *
 * localVariables:
 * { this: { form: { firstName: "John", lastName: "Doe" } } }
 *
 * return value:
 * [
 *   { key: "name", value: "John Doe" },
 *   { key: "created-at", value: "2023-06-15T09:30:00Z" }
 * ]
 */
export const useKeyValueArrayResolver = (): KeyValueResolverFunction<
  KeyValueType[]
> => {
  const interpolateString = useStringInterpolation();
  return (keyValueArray, localVariables) =>
    keyValueArray.map(({ key, value }) => {
      const parsedValue = interpolateString(value, localVariables);
      return { key, value: parsedValue };
    });
};
