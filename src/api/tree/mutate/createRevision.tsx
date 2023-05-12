import { ToastAction, useToast } from "../../../design/Toast";

import { WorkflowCreatedSchema } from "../schema";
import { apiFactory } from "../../utils";
import { forceLeadingSlash } from "../utils";
import { pages } from "../../../util/router/pages";
import { useApiKey } from "../../../util/store/apiKey";
import { useMutation } from "@tanstack/react-query";
import { useNamespace } from "../../../util/store/namespace";
import { useNavigate } from "react-router-dom";

const createRevision = apiFactory({
  pathFn: ({ namespace, path }: { namespace: string; path: string }) =>
    `/api/namespaces/${namespace}/tree${forceLeadingSlash(
      path
    )}?op=save-workflow&ref=latest`,
  method: "POST",
  schema: WorkflowCreatedSchema,
});

export const useCreateRevision = () => {
  const apiKey = useApiKey();
  const namespace = useNamespace();
  const { toast } = useToast();
  const navigate = useNavigate();

  if (!namespace) {
    throw new Error("namespace is undefined");
  }

  return useMutation({
    mutationFn: ({ path }: { path: string }) =>
      createRevision({
        apiKey: apiKey ?? undefined,
        payload: undefined,
        urlParams: {
          namespace: namespace,
          path,
        },
      }),
    onSuccess: (data, variables) => {
      toast({
        title: "Revision created",
        description: `Revision ${data.revision.name} was created`,
        variant: "success",
        action: (
          <ToastAction
            altText="Open Revision"
            onClick={() => {
              navigate(
                pages.explorer.createHref({
                  namespace,
                  path: variables.path,
                  subpage: "workflow-revisions",
                  revision: data.revision.name,
                })
              );
            }}
          >
            Open Revision
          </ToastAction>
        ),
      });
    },
    onError: () => {
      toast({
        title: "An error occurred",
        description: "could not create revision 😢",
        variant: "error",
      });
    },
  });
};
