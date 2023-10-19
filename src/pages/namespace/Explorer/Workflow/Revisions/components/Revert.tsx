import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { ToastAction, useToast } from "~/design/Toast";
import { Trans, useTranslation } from "react-i18next";

import Button from "~/design/Button";
import { TrimmedRevisionSchemaType } from "~/api/tree/schema/node";
import { Undo } from "lucide-react";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useNavigate } from "react-router-dom";
import { useNodeContent } from "~/api/tree/query/node";
import { useUpdateWorkflow } from "~/api/tree/mutate/updateWorkflow";

const Revert = ({
  path,
  revision,
  close,
}: {
  path: string;
  revision: TrimmedRevisionSchemaType;
  close: () => void;
}) => {
  const { t } = useTranslation();
  const { toast } = useToast();
  const navigate = useNavigate();
  const namespace = useNamespace();
  const { data, isSuccess } = useNodeContent({ path, revision: revision.name });
  const { mutate: updateWorkflow, isLoading: updateIsLoading } =
    useUpdateWorkflow({
      onSuccess: () => {
        toast({
          title: t(
            "pages.explorer.tree.workflow.revisions.overview.list.revert.success.title"
          ),
          description: t(
            "pages.explorer.tree.workflow.revisions.overview.list.revert.success.description",
            { name: revision.name }
          ),
          action: (
            <ToastAction
              altText={t(
                "pages.explorer.tree.workflow.revisions.overview.list.revert.success.action"
              )}
              onClick={() => {
                if (!namespace) {
                  return;
                }
                navigate(
                  pages.explorer.createHref({
                    namespace,
                    path,
                    subpage: "workflow",
                  })
                );
              }}
              data-testid="workflow-revert-revision-toast-action"
            >
              {t(
                "pages.explorer.tree.workflow.revisions.overview.list.revert.success.action"
              )}
            </ToastAction>
          ),
          variant: "success",
        });
        close();
      },
      onError: () => {
        toast({
          title: t("api.generic.error"),
          description: t(
            "pages.explorer.tree.workflow.revisions.overview.list.revert.error.description"
          ),
          variant: "error",
        });
      },
    });

  const workflowData = atob(data?.revision?.source ?? "");
  const isLoading = !isSuccess || updateIsLoading;

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Undo />
          {t(
            "pages.explorer.tree.workflow.revisions.overview.list.revert.title"
          )}
        </DialogTitle>
      </DialogHeader>
      <div className="my-3">
        <Trans
          i18nKey="pages.explorer.tree.workflow.revisions.overview.list.revert.description"
          values={{ name: revision.name }}
        />
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t(
              "pages.explorer.tree.workflow.revisions.overview.list.revert.cancelBtn"
            )}
          </Button>
        </DialogClose>
        <Button
          onClick={() => {
            if (workflowData) {
              updateWorkflow({ path, fileContent: workflowData });
            }
          }}
          loading={isLoading}
          data-testid="dialog-revert-revision-btn-submit"
        >
          {!isLoading && <Undo />}
          {t(
            "pages.explorer.tree.workflow.revisions.overview.list.revert.revertBtn"
          )}
        </Button>
      </DialogFooter>
    </>
  );
};

export default Revert;
