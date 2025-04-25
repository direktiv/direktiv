import { VariableObject } from "../../../../../schema/primitives/variable";
// import { useQueryClient } from "@tanstack/react-query";

type TemplateStringProps = Omit<VariableObject, "namespace">;

export const QueryVariable = ({ id, pointer }: TemplateStringProps) => (
  // const client = useQueryClient();
  // const cachedData = client.getQueryData([id]);

  <>
    {id} {pointer}
  </>
);
