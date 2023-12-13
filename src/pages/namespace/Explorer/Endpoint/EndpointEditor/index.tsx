import { FC, useState } from "react";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import EndpointPreview from "../EndpointPreview";
import { Form } from "./Form";
import { RouteSchemeType } from "~/api/gateway/schema";
import { Save } from "lucide-react";
import { serializeEndpointFile } from "../utils";
import { stringify } from "json-to-pretty-yaml";
import { useNodeContent } from "~/api/tree/query/node";
import { useTranslation } from "react-i18next";
import { useUpdateWorkflow } from "~/api/tree/mutate/updateWorkflow";

type NodeContentType = ReturnType<typeof useNodeContent>["data"];

type EndpointEditorProps = {
  path: string;
  data: NonNullable<NodeContentType>;
  route?: RouteSchemeType;
};

const EndpointEditor: FC<EndpointEditorProps> = ({ data, path }) => {
  const { t } = useTranslation();
  const endpointFileContent = atob(data.revision?.source ?? "");

  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const endpointConfig = serializeEndpointFile(endpointFileContent);

  const { mutate: updateRoute, isLoading } = useUpdateWorkflow({
    onSuccess: () => {
      setHasUnsavedChanges(false);
    },
  });

  const onSaveClicked = () => {
    const toSave = stringify(endpointConfig);
    if (toSave) {
      updateRoute({
        path,
        fileContent: toSave,
      });
    }
  };

  return (
    <div className="relative flex grow flex-col space-y-4 p-5">
      <Card className="flex grow flex-col p-4">
        <div className="grid grow grid-cols-2">
          <div>
            <Form endpointConfig={endpointConfig} />
          </div>
          <Card className="grid grid-rows-2 p-5">
            <pre>{endpointFileContent}</pre>
            <pre>{JSON.stringify(endpointConfig, null, 2)}</pre>
          </Card>
        </div>
        <div className="flex justify-end gap-2 pt-2 text-sm text-gray-8 dark:text-gray-dark-8">
          {hasUnsavedChanges && (
            <span className="text-center">
              {t("pages.explorer.workflow.editor.unsavedNote")}
            </span>
          )}
        </div>
      </Card>
      <div className="flex flex-col justify-end gap-4 sm:flex-row sm:items-center">
        <EndpointPreview fileContent={stringify(endpointConfig)} />
        <Button variant="outline" disabled={isLoading} onClick={onSaveClicked}>
          <Save />
          {t("pages.explorer.endpoint.editor.saveBtn")}
        </Button>
      </div>
    </div>
  );
};

export default EndpointEditor;
