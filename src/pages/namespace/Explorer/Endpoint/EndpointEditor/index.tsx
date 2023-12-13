import { FC, useState } from "react";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import EndpointPreview from "../EndpointPreview";
import { RouteSchemeType } from "~/api/gateway/schema";
import { Save } from "lucide-react";
import { stringify } from "json-to-pretty-yaml";
import { useNodeContent } from "~/api/tree/query/node";
import { useTranslation } from "react-i18next";
import { useUpdateWorkflow } from "~/api/tree/mutate/updateWorkflow";
import yamljs from "js-yaml";

type NodeContentType = ReturnType<typeof useNodeContent>["data"];

type EndpointEditorProps = {
  path: string;
  data: NonNullable<NodeContentType>;
  route?: RouteSchemeType;
};

const EndpointEditor: FC<EndpointEditorProps> = ({ data, path }) => {
  const { t } = useTranslation();
  const workflowData = atob(data.revision?.source ?? "");

  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [routeConfigJson, setRouteConfigJson] = useState(() => {
    let json;
    try {
      json = yamljs.load(workflowData);
    } catch (e) {
      json = null;
    }

    return json as Record<string, unknown>;
  });

  const { mutate: updateRoute, isLoading } = useUpdateWorkflow({
    onSuccess: () => {
      setHasUnsavedChanges(false);
    },
  });

  const onSaveClicked = () => {
    const toSave = stringify(routeConfigJson);
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
        <div className="grow">FORM goes here</div>
        <div className="flex justify-end gap-2 pt-2 text-sm text-gray-8 dark:text-gray-dark-8">
          {hasUnsavedChanges && (
            <span className="text-center">
              {t("pages.explorer.workflow.editor.unsavedNote")}
            </span>
          )}
        </div>
      </Card>
      <div className="flex flex-col justify-end gap-4 sm:flex-row sm:items-center">
        <EndpointPreview fileContent={stringify(routeConfigJson)} />
        <Button variant="outline" disabled={isLoading} onClick={onSaveClicked}>
          <Save />
          {t("pages.explorer.endpoint.editor.saveBtn")}
        </Button>
      </div>
    </div>
  );
};

export default EndpointEditor;
