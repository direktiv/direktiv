import { FC, useState } from "react";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { JSONSchemaForm } from "~/design/JSONschemaForm";
import { RJSFSchema } from "@rjsf/utils";
import { Save } from "lucide-react";
import { ScrollArea } from "~/design/ScrollArea";
import { endpointBaseFormSchema } from "./utils";
import { stringify } from "json-to-pretty-yaml";
import { useNodeContent } from "~/api/tree/query/node";
import { usePlugins } from "~/api/gateway/query/getPlugins";
import { useTranslation } from "react-i18next";
import { useUpdateWorkflow } from "~/api/tree/mutate/updateWorkflow";
import yamljs from "js-yaml";

export type NodeContentType = ReturnType<typeof useNodeContent>["data"];

const EndpointEditor: FC<{
  data: NonNullable<NodeContentType>;
  path: string;
}> = ({ data, path }) => {
  const { t } = useTranslation();
  const workflowData = atob(data.revision?.source ?? "");
  const { data: plugins } = usePlugins();
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [error, setError] = useState<string | undefined>();
  const [endpointConfigJson, setEndpointConfigJson] = useState(() => {
    let json;
    try {
      json = yamljs.load(workflowData);
    } catch (e) {
      json = null;
    }

    return json as Record<string, unknown>;
  });

  const { mutate: updateEndpoint, isLoading } = useUpdateWorkflow({
    onError: (error) => {
      error && setError(error);
    },
    onSuccess: () => {
      setHasUnsavedChanges(false);
    },
  });

  if (!plugins) return null;

  const onSaveClicked = () => {
    const toSave = stringify(endpointConfigJson);
    if (toSave) {
      setError(undefined);
      updateEndpoint({
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
        <hr />
        <pre>{stringify(endpointConfigJson)}</pre>
      </Card>
      <Card className="flex grow flex-col p-4">
        <ScrollArea className="h-full p-4">
          <JSONSchemaForm
            formData={endpointConfigJson}
            onChange={(e) => {
              if (e.formData) {
                console.log("🚀", e.formData);
                setHasUnsavedChanges(true);
                setEndpointConfigJson(e.formData);
                // const formDataWithHeader = addServiceHeader(e.formData);
                // setServiceConfigJson(formDataWithHeader);
                // setValue("fileContent", stringify(formDataWithHeader));
              }
            }}
            schema={endpointBaseFormSchema}
          />
        </ScrollArea>
      </Card>
      <div className="flex flex-col justify-end gap-4 sm:flex-row sm:items-center">
        <Button variant="outline" disabled={isLoading} onClick={onSaveClicked}>
          <Save />
          {t("pages.explorer.gateway.editor.saveBtn")}
        </Button>
      </div>
    </div>
  );
};

export default EndpointEditor;
