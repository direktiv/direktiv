import { JSXValueSchema, getValueFromJsonPath } from "./utils";

import { Error } from "./Error";
import { VariableObjectValidated } from "../../../../../schema/primitives/variable";
import { twMergeClsx } from "~/util/helpers";
import { useMode } from "../../../../context/pageCompilerContext";
import { useQueryClient } from "@tanstack/react-query";

type TemplateStringProps = {
  variable: VariableObjectValidated;
};

export const QueryVariable = ({ variable }: TemplateStringProps) => {
  const { src, id, pointer } = variable;

  const mode = useMode();
  const cacheKey = [id];
  const queryClient = useQueryClient();
  const queryState = queryClient.getQueryState(cacheKey);

  if (queryState === undefined)
    return <Error value={src}>queryIdNotFound</Error>;

  const cachedData = queryClient.getQueryData(cacheKey);
  const [data, error] = getValueFromJsonPath(cachedData, pointer);
  if (error) {
    return <Error value={src}>{error}</Error>;
  }

  const dataParsed = JSXValueSchema.safeParse(data);
  if (!dataParsed.success) {
    return <Error value={src}>couldNotStringify</Error>;
  }

  return (
    <span
      className={twMergeClsx(
        mode === "preview" &&
          "border border-gray-9 bg-gray-4 dark:bg-gray-dark-4 dark:border-gray-dark-9"
      )}
    >
      {dataParsed.data}
    </span>
  );
};
