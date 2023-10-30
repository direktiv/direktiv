import { FC, useState } from "react";
import { Save, Tag } from "lucide-react";

import Button from "~/design/Button";
import { CodeEditor } from "../Workflow/Active/CodeEditor";
import { useNamespace } from "~/util/store/namespace";
import { useNodeContent } from "~/api/tree/query/node";
import { useTranslation } from "react-i18next";
import { useUpdateWorkflow } from "~/api/tree/mutate/updateWorkflow";

export type NodeContentType = ReturnType<typeof useNodeContent>["data"];

const ServiceEditor: FC<{
  data: NonNullable<NodeContentType>;
  path: string;
}> = ({ data, path }) => {
  const { t } = useTranslation();
  const namespace = useNamespace();
  const [error, setError] = useState<string | undefined>();
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);

  const workflowDataFromServer = atob(data?.revision?.source ?? "");

  const { mutate: updateWorkflow, isLoading } = useUpdateWorkflow({
    onError: (error) => {
      error && setError(error);
    },
    onSuccess: () => {
      setHasUnsavedChanges(false);
    },
  });

  const [editorContent, setEditorContent] = useState(workflowDataFromServer);

  const onEditorContentUpdate = (newData: string) => {
    setHasUnsavedChanges(workflowDataFromServer !== newData);
    setEditorContent(newData ?? "");
  };

  const onSave = (toSave: string | undefined) => {
    if (toSave) {
      setError(undefined);
      updateWorkflow({
        path,
        fileContent: toSave,
      });
    }
  };

  if (!namespace) return null;

  return (
    <div className="relative flex grow flex-col space-y-4 p-5">
      <h3 className="flex items-center gap-x-2 font-bold">
        <Tag className="h-5" />
        {t("pages.explorer.workflow.headline")}
      </h3>
      <CodeEditor
        value={workflowDataFromServer}
        onValueChange={onEditorContentUpdate}
        createdAt={data.revision?.createdAt}
        error={error}
        hasUnsavedChanges={hasUnsavedChanges}
        onSave={onSave}
      />

      <div className="flex flex-col justify-end gap-4 sm:flex-row sm:items-center">
        <Button
          variant="outline"
          disabled={isLoading}
          onClick={() => {
            onSave(editorContent);
          }}
          data-testid="workflow-editor-btn-save"
        >
          <Save />
          {t("pages.explorer.workflow.editor.saveBtn")}
        </Button>
      </div>
    </div>
  );
};

export default ServiceEditor;
