import {
  parseTemplateString,
  variablePattern,
} from "../primitives/Variable/utils";

import { KeyValueType } from "../../schema/primitives/keyValue";
import { MutationType } from "../../schema/procedures/mutation";
import { QueryType } from "../../schema/procedures/query";
import { useResolveVariableString } from "../primitives/Variable/utils/useResolveVariableString";
import { useTranslation } from "react-i18next";

// TODO: POC only, needs to be refactored
export const useGetUrl = () => {
  const resolveVariableStringFn = useResolveVariableString();
  const getKeyValueArrayFn = useResolveKeyValueArray();

  return (input: QueryType | MutationType) => {
    const { baseUrl, queryParams } = input;

    const queryParamsParsed = getKeyValueArrayFn(queryParams ?? []);

    const searchParams = new URLSearchParams();
    queryParamsParsed?.forEach(({ key, value }) => {
      searchParams.append(key, value);
    });

    const queryString = searchParams.toString();
    const url = queryString ? baseUrl.concat("?", queryString) : baseUrl;

    const templateFragments = url.split(variablePattern);

    // TODO: use getKeyValueArrayFn here as well
    return templateFragments
      .map((fragment, index) => {
        const isVariable = index % 2 === 1;
        if (isVariable) {
          const a = resolveVariableStringFn(fragment);
          if (!a.success) {
            throw new Error("🚀");
          }
          return a.data;
        }

        return fragment;
      })
      .join("");
  };
};

/**
 * A hook that returns a function that lets you analyze a list of KeyValue
 * primitives for variable strings and replace them with their resolved values.
 *
 * Example:
 *
 * const payload = [{
 *   key: "id",
 *   value: "company-id-{{query.company-list.data.0.id}}",
 * }]
 *
 * const resolveKeyValueArrayFn = useResolveKeyValueArray();
 *
 * console.log(resolveKeyValueArrayFn(payload)); // Output: [{ key: "id", value: "company-id-apple" }]
 *
 */
export const useResolveKeyValueArray = () => {
  const resolveVariableStringFn = useResolveVariableString();
  const { t } = useTranslation();
  return (input: KeyValueType[]): KeyValueType[] =>
    // iterate over every KeyValue primitive in the input array
    input.map(({ key, value }) => {
      /**
       * iterate over every variable placeholder in the
       * value field and replace it with its resolved value
       * from the React context
       */
      const parsedValue = parseTemplateString(value, (match) => {
        const result = resolveVariableStringFn(match);
        if (!result.success) {
          throw new Error(
            t(`direktivPage.error.templateString.${result.error}`)
          );
        }
        return result.data;
      }).join("");
      return { key, value: parsedValue };
    });
};
