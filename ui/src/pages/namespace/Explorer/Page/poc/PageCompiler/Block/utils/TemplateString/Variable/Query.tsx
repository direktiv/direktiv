import { JSXValueSchema, getValueFromJsonPath } from "./utils";

import { Error } from "./Error";
import { twMergeClsx } from "~/util/helpers";
import { useMode } from "../../../../context/pageCompilerContext";
import { useQueryClient } from "@tanstack/react-query";

type TemplateStringProps = {
  id: string;
  pointer: string;
};

export const QueryVariable = ({ id, pointer }: TemplateStringProps) => {
  const mode = useMode();
  const cacheKey = [id];

  const queryClient = useQueryClient();

  const queryState = queryClient.getQueryState(cacheKey);
  const cachedData = queryClient.getQueryData(cacheKey);
  const [data, error] = getValueFromJsonPath(cachedData, pointer);

  if (queryState === undefined)
    return (
      <Error value="queryIdNotFound">
        Could not find a query with id <code>{id}</code>.
      </Error>
    );

  if (error) {
    return (
      <Error value={error}>
        Error when trying to access <code>{pointer}</code> in query with id{" "}
        <code>{id}</code>.
      </Error>
    );
  }

  const dataParsed = JSXValueSchema.safeParse(data);

  if (!dataParsed.success) {
    return (
      <Error value="couldNotStringify">
        Error when trying to render <code>{pointer}</code> in query with id{" "}
        <code>{id}</code>. Make sure it is from one of the following types:{" "}
        <code>String</code>, <code>Number</code>, <code>Boolean</code>,{" "}
        <code>Null</code> or <code>Undefined</code>.
      </Error>
    );
  }

  return (
    <span
      className={twMergeClsx(
        mode === "inspect" &&
          "border border-gray-9 bg-gray-4 dark:bg-gray-dark-4 dark:border-gray-dark-9"
      )}
    >
      {dataParsed.data}
    </span>
  );
};
