import { ExtendedKeyValueType } from "../../../schema/primitives/extendedKeyValue";
import { KeyValueResolverFunction } from "./utils";
import { useStringInterpolation } from "../Variable/utils/useStringInterpolation";
import { useTranslation } from "react-i18next";
import { useVariableResolver } from "../Variable/utils/useVariableResolver";

/**
 * Hook that returns a function to resolve an array of extended key-value pairs.
 * Transforms them into a regular key-value array by interpolating template variables
 * and resolving different value types using provided local and context variables.
 *
 * keyValueArray input:
 * [
 *   {
 *     key: "name",
 *     value: {
 *       type: "string",
 *       value: "{{this.form.firstName}} {{this.form.lastName}}",
 *     },
 *   },
 *   {
 *     key: "isActive",
 *     value: { type: "variable", value: "query.user.data.isActive" },
 *   },
 * ];
 *
 * localVariables:
 * { this: { form: { firstName: "John", lastName: "Doe" } } }
 *
 * output: [
 *   { key: "name", value: "John Doe" },
 *   { key: "isActive", value: true }
 * ]
 */
export const useExtendedKeyValueArrayResolver = (): KeyValueResolverFunction<
  ExtendedKeyValueType[]
> => {
  const { t } = useTranslation();
  const interpolateString = useStringInterpolation();
  const resolveVariable = useVariableResolver();
  return (extendedKeyValueArray, localVariables) =>
    extendedKeyValueArray.map(({ key, value: valueType }) => {
      switch (valueType.type) {
        case "string": {
          return {
            key,
            value: interpolateString(valueType.value, localVariables),
          };
        }
        case "variable": {
          const resolvedVariable = resolveVariable(
            valueType.value,
            localVariables
          );
          if (!resolvedVariable.success) {
            throw new Error(
              t(`direktivPage.error.templateString.${resolvedVariable.error}`, {
                variable: valueType.value,
              })
            );
          }
          return { key, value: resolvedVariable.data };
        }
        case "boolean":
        case "number":
          {
            return { key, value: valueType.value };
          }

          throw new Error(
            `${valueType.type} is not implemented for extended key value`
          );
      }
    });
};
