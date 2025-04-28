import { getValueFromJsonPath } from "./utils";
import { useQueryClient } from "@tanstack/react-query";

type TemplateStringProps = {
  id: string;
  pointer: string;
};

export const QueryVariable = ({ id, pointer }: TemplateStringProps) => {
  const client = useQueryClient();
  const cachedData = client.getQueryData([id]);
  const data = getValueFromJsonPath(cachedData, pointer);
  return (
    <span className="bg-success-4 text-success-11 dark:bg-success-dark-4 dark:text-success-dark-11">
      {data}
    </span>
  );
};
