import { NamespaceCreatedSchema } from "../schema";
import { apiFactory } from "../../utils";
import { useApiKey } from "../../../util/store/apiKey";
import { useMutation } from "@tanstack/react-query";
import { useToast } from "../../../design/Toast";

const createNamespace = apiFactory({
  pathFn: ({ name }: { name: string }) => `/api/namespaces/${name}`,
  method: "PUT",
  schema: NamespaceCreatedSchema,
});

type ResolvedCreateNamespace = Awaited<ReturnType<typeof createNamespace>>;

export const useCreateNamespace = ({
  onSuccess,
}: { onSuccess?: (data: ResolvedCreateNamespace) => void } = {}) => {
  const apiKey = useApiKey();
  const { toast } = useToast();

  return useMutation({
    mutationFn: ({ name }: { name: string }) =>
      createNamespace({
        apiKey: apiKey ?? undefined,
        params: undefined,
        pathParams: {
          name,
        },
      }),
    onSuccess(data, variables) {
      toast({
        title: "Namespace created",
        description: `Namespace ${variables.name} was created successfully.`,
        variant: "success",
      });
      onSuccess?.(data);
    },
    onError: () => {
      toast({
        title: "An error occurred",
        description: "could not create directory 😢",
        variant: "error",
      });
    },
  });
};
