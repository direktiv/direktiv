import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Play, PlusCircle } from "lucide-react";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { JSONSchemaForm } from "~/design/JSONschemaForm";
import { RJSFSchema } from "@rjsf/utils";
import { ScrollArea } from "~/design/ScrollArea";
import { stringify } from "json-to-pretty-yaml";
import { useState } from "react";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import yamljs from "js-yaml";

const jsonToYaml = (json: Record<string, unknown>) => {
  if (Object.keys(json).length === 0) {
    return "";
  }
  return stringify(json);
};

const serviceSchema: RJSFSchema = {
  properties: {
    image: {
      title: "Image",
      type: "string",
    },
    scale: {
      title: "Scale",
      type: "integer",
      enum: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9],
    },
    size: {
      title: "size",
      type: "integer",
      enum: ["large", "medium", "small"],
    },
    cmd: {
      title: "Cmd",
      type: "string",
    },
  },
  required: ["name"],
  type: "object",
};

const NewService = ({
  path,
}: {
  path?: string;
  close: () => void;
  unallowedNames?: string[];
}) => {
  const { t } = useTranslation();
  const disableSubmit = false;
  const isLoading = false;

  const [serviceConfig, setServiceConfig] = useState({});
  const formId = `new-service-${path}`;
  const theme = useTheme();

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Play />
          {t("pages.explorer.tree.newService.title")}
        </DialogTitle>
      </DialogHeader>

      <Card className="h-96 w-full p-4" noShadow background="weight-1">
        <ScrollArea className="h-full">
          <JSONSchemaForm
            formData={serviceConfig}
            onChange={(e) => {
              setServiceConfig(e.formData);
            }}
            schema={serviceSchema}
          />
        </ScrollArea>
      </Card>
      <Card className="h-96 w-full p-4" noShadow background="weight-1">
        <Editor
          value={jsonToYaml(serviceConfig)}
          onChange={(newData) => {
            if (newData) {
              const json = yamljs.load(newData);
              if (typeof json === "object") {
                // setServiceConfig(json);
              }
            }
          }}
          theme={theme ?? undefined}
        />
      </Card>
      <Card className="w-full p-4" noShadow background="weight-1">
        <code className="block">{JSON.stringify(serviceConfig)}</code>
        <code className="block">{jsonToYaml(serviceConfig)}</code>
      </Card>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.explorer.tree.newService.cancelBtn")}
          </Button>
        </DialogClose>

        <Button
          data-testid="new-workflow-submit"
          type="submit"
          disabled={disableSubmit}
          loading={isLoading}
          form={formId}
        >
          {!isLoading && <PlusCircle />}
          {t("pages.explorer.tree.newService.createBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default NewService;
