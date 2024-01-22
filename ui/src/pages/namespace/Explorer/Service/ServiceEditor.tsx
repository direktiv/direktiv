import { FC, useState } from "react";

import Button from "~/design/Button";
import { CodeEditor } from "../Workflow/Active/CodeEditor";
import { Save } from "lucide-react";
import ServiceHelp from "./ServiceHelp";
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

  const serviceDataFromServer = atob(data?.source ?? "");

  const { mutate: updateService, isLoading } = useUpdateWorkflow({
    onError: (error) => {
      error && setError(error);
    },
    onSuccess: () => {
      setHasUnsavedChanges(false);
    },
  });

  const [editorContent, setEditorContent] = useState(serviceDataFromServer);

  const onEditorContentUpdate = (newData: string) => {
    setHasUnsavedChanges(serviceDataFromServer !== newData);
    setEditorContent(newData ?? "");
  };

  const onSave = (toSave: string | undefined) => {
    if (toSave) {
      setError(undefined);
      updateService({
        path,
        fileContent: toSave,
      });
    }
  };

  if (!namespace) return null;

  return (
    <div className="relative flex grow flex-col space-y-4 p-5">
      <CodeEditor
        value={serviceDataFromServer}
        onValueChange={onEditorContentUpdate}
        // TODO: may remove this
        createdAt={data.node.createdAt}
        error={error}
        hasUnsavedChanges={hasUnsavedChanges}
        onSave={onSave}
      />

      <div className="flex flex-col justify-end gap-4 sm:flex-row sm:items-center">
        <ServiceHelp />
        <Button
          variant="outline"
          disabled={isLoading}
          onClick={() => {
            onSave(editorContent);
          }}
        >
          <Save />
          {t("pages.explorer.service.editor.saveBtn")}
        </Button>
      </div>
    </div>
  );
};

export default ServiceEditor;
