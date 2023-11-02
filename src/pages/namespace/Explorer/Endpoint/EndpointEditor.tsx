import { FC, useState } from "react";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { Save } from "lucide-react";
import { useNodeContent } from "~/api/tree/query/node";
import { usePlugins } from "~/api/gateway/query/getPlugins";
import { useTranslation } from "react-i18next";
import { useUpdateWorkflow } from "~/api/tree/mutate/updateWorkflow";

export type NodeContentType = ReturnType<typeof useNodeContent>["data"];

const EndpointEditor: FC<{
  data: NonNullable<NodeContentType>;
  path: string;
}> = ({ data, path }) => {
  const { t } = useTranslation();
  const { data: plugins } = usePlugins();
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [error, setError] = useState<string | undefined>();

  const { mutate: updateService, isLoading } = useUpdateWorkflow({
    onError: (error) => {
      error && setError(error);
    },
    onSuccess: () => {
      setHasUnsavedChanges(false);
    },
  });

  const workflowData = atob(data.revision?.source ?? "");

  const onSave = (toSave: string | undefined) => {
    if (toSave) {
      setError(undefined);
      updateService({
        path,
        fileContent: toSave,
      });
    }
  };

  return (
    <div className="relative flex grow flex-col space-y-4 p-5">
      <Card className="flex flex-col p-4" noShadow>
        <div>Error: {error}</div>
        <div>hasUnsavedChanges: {hasUnsavedChanges ? "yes" : "no"}</div>
        <pre>{workflowData}</pre>
      </Card>
      <Card className="flex grow flex-col p-4"></Card>
      <div className="flex flex-col justify-end gap-4 sm:flex-row sm:items-center">
        <Button variant="outline" disabled={isLoading}>
          <Save />
          {t("pages.explorer.gateway.editor.saveBtn")}
        </Button>
      </div>
    </div>
  );
};

export default EndpointEditor;
