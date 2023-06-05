import { ArrowLeft, GitMerge, Tag, Undo } from "lucide-react";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import CopyButton from "~/design/CopyButton";
import Editor from "~/design/Editor";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useNavigate } from "react-router-dom";
import { useNodeContent } from "~/api/tree/query/node";
import { useNodeTags } from "~/api/tree/query/tags";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";

const WorkflowRevisionsPage = () => {
  const namespace = useNamespace();

  const { t } = useTranslation();
  const navigate = useNavigate();
  const { revision: selectedRevision, path } = pages.explorer.useParams();
  const theme = useTheme();
  const { data } = useNodeContent({
    path,
    revision: selectedRevision,
  });
  const { data: tags } = useNodeTags({ path });
  const workflowData = data?.revision?.source && atob(data?.revision?.source);
  const isTag =
    tags?.results?.some((tag) => tag.name === selectedRevision) ?? false;

  const Icon = isTag ? Tag : GitMerge;

  if (!namespace) return null;
  if (!selectedRevision) return null;
  if (!workflowData) return null;

  const backLink = pages.explorer.createHref({
    namespace,
    path,
    subpage: "workflow-revisions",
  });

  return (
    <div className="flex grow flex-col space-y-4">
      <div className="flex gap-x-4">
        <h3 className="group flex grow items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
          <Icon aria-hidden="true" className="h-5" />
          {selectedRevision}
          <CopyButton
            value={selectedRevision}
            buttonProps={{
              variant: "outline",
              className: "hidden group-hover:inline-flex",
              size: "sm",
            }}
          >
            {(copied) =>
              copied
                ? t(
                    "pages.explorer.tree.workflow.revisions.overview.detail.copied"
                  )
                : t(
                    "pages.explorer.tree.workflow.revisions.overview.detail.copy"
                  )
            }
          </CopyButton>
        </h3>
        {/* TODO: change to a Link as soon out Button component support asChild prop (DIR-597) */}
        <Button
          variant="outline"
          onClick={() => {
            navigate(backLink);
          }}
        >
          <ArrowLeft />
          {t("pages.explorer.tree.workflow.revisions.overview.detail.backBtn")}
        </Button>
        <Button variant="outline">
          <Undo />
          {t(
            "pages.explorer.tree.workflow.revisions.overview.detail.revertBtn"
          )}
        </Button>
      </div>
      <Card className="grow p-4">
        <Editor
          value={workflowData}
          theme={theme ?? undefined}
          options={{ readOnly: true }}
        />
      </Card>
    </div>
  );
};

export default WorkflowRevisionsPage;
