import { KeyValueType } from "../../schema/primitives/keyValue";
import { MutationType } from "../../schema/procedures/mutation";
import { QueryType } from "../../schema/procedures/query";
import { parseTemplateString } from "../primitives/Variable/utils";
import { useResolveVariableString } from "../primitives/Variable/utils/useResolveVariableString";
import { useTranslation } from "react-i18next";

export const useGetUrl = () => {
  const getKeyValueArrayFn = useResolveKeyValueArray();
  const replaceVariablesFn = useReplaceVariables();

  return (input: QueryType | MutationType) => {
    const { baseUrl, queryParams } = input;
    const queryParamsParsed = getKeyValueArrayFn(queryParams ?? []);
    const searchParams = new URLSearchParams();
    queryParamsParsed?.forEach(({ key, value }) => {
      searchParams.append(key, value);
    });

    const queryString = searchParams.toString();
    const url = replaceVariablesFn(baseUrl);
    return queryString ? url.concat("?", queryString) : url;
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
  const replaceVariablesFn = useReplaceVariables();
  return (input: KeyValueType[]): KeyValueType[] =>
    // iterate over every KeyValue primitive in the input array
    input.map(({ key, value }) => {
      /**
       * iterate over every variable placeholder in the
       * value field and replace it with its resolved value
       * from the React context
       */
      const parsedValue = replaceVariablesFn(value);
      return { key, value: parsedValue };
    });
};

// TODO: rename
const useReplaceVariables = () => {
  const { t } = useTranslation();
  const resolveVariableStringFn = useResolveVariableString();
  return (input: string) =>
    parseTemplateString(input, (match) => {
      const result = resolveVariableStringFn(match);
      if (!result.success) {
        throw new Error(t(`direktivPage.error.templateString.${result.error}`));
      }
      return result.data;
    }).join("");
};
