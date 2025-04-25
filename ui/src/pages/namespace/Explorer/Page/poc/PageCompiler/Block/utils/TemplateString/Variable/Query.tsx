import { TemplateStringSeparator } from "../../../../../schema/primitives/templateString";
import { useQueryClient } from "@tanstack/react-query";

type TemplateStringProps = {
  id: string;
  pointer: string;
};

const getObjectValueByPath = (
  obj: unknown,
  path: string
): string | undefined => {
  if (!obj || !path || typeof obj !== "object") {
    return undefined;
  }

  const pathParts = path.split(TemplateStringSeparator);
  let current = obj;

  for (const part of pathParts) {
    if (
      current &&
      typeof current === "object" &&
      current !== null &&
      part in current
    ) {
      current = (current as Record<string, unknown>)[part];
    } else {
      return undefined; // Path not found
    }
  }

  return current;
};

export const QueryVariable = ({ id, pointer }: TemplateStringProps) => {
  const client = useQueryClient();
  const cachedData = client.getQueryData([id]);
  return (
    <span className="bg-success-4 text-success-11 dark:bg-success-dark-4 dark:text-success-dark-11">
      {getObjectValueByPath(cachedData, pointer)}
    </span>
  );
};
