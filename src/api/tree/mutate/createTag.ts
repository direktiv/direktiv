import { TagCreatedSchema } from "../schema";
import { apiFactory } from "../../utils";
import { forceLeadingSlash } from "../utils";
import { useApiKey } from "../../../util/store/apiKey";
import { useMutation } from "@tanstack/react-query";
import { useNamespace } from "../../../util/store/namespace";
import { useToast } from "../../../design/Toast";

const createTag = apiFactory({
  pathFn: ({
    namespace,
    path,
    ref,
  }: {
    namespace: string;
    path: string;
    ref: string;
  }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(
      path
    )}?op=tag&ref=${ref}`,
  method: "POST",
  schema: TagCreatedSchema,
});

export const useCreateTag = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutation({
    mutationFn: ({
      path,
      ref,
      tag,
    }: {
      path: string;
      ref: string;
      tag: string;
    }) =>
      createTag({
        apiKey: apiKey ?? undefined,
        params: { tag },
        pathParams: {
          namespace: namespace,
          path,
          ref,
        },
      }),
    onSuccess: (_, variables) => {
      toast({
        title: "Tag created",
        description: `Tag ${variables.tag} was created`,
        variant: "success",
      });
    },
    onError: () => {
      toast({
        title: "An error occurred",
        description: "could not create tag 😢",
        variant: "error",
      });
    },
  });
};
