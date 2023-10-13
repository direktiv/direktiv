import { Code, Columns, Play, PlusCircle } from "lucide-react";
import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { JSONSchemaForm } from "~/design/JSONschemaForm";
import { RJSFSchema } from "@rjsf/utils";
import { ScrollArea } from "~/design/ScrollArea";
import { Toggle } from "~/design/Toggle";
import { stringify } from "json-to-pretty-yaml";
import { twMergeClsx } from "~/util/helpers";
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

const serviceFormSchema: RJSFSchema = {
  properties: {
    image: {
      title: "Image",
      type: "string",
    },
    name: {
      title: "Name",
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
  required: ["image", "name"],
  type: "object",
};

const defaultService = {
  direktiv_api: "service/v1",
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
  const [splitView, setSplitView] = useState(true);

  const [serviceConfigJson, setServiceConfigJson] = useState(defaultService);
  const [serviceConfigYaml, setServiceConfigYaml] = useState(
    jsonToYaml(defaultService)
  );

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
      <Card className="flex flex-col gap-4 p-4" noShadow>
        <div
          className={twMergeClsx(
            "grid h-[600px] gap-5",
            splitView
              ? "grid-rows-2 md:grid-cols-2 md:grid-rows-none"
              : "grid-rows-1 md:grid-cols-1"
          )}
        >
          {splitView && (
            <Card background="weight-1">
              <ScrollArea className="h-full p-4">
                <JSONSchemaForm
                  formData={serviceConfigJson}
                  onChange={(e) => {
                    if (e.formData) {
                      setServiceConfigJson(e.formData);
                      setServiceConfigYaml(jsonToYaml(e.formData));
                    }
                  }}
                  schema={serviceFormSchema}
                />
              </ScrollArea>
            </Card>
          )}
          <Card className="flex p-4" background="weight-1">
            <Editor
              value={serviceConfigYaml}
              theme={theme ?? undefined}
              onChange={(newData) => {
                if (newData) {
                  setServiceConfigYaml(newData);
                  let json;
                  try {
                    json = yamljs.load(newData);
                  } catch (e) {
                    json = null;
                  }
                  if (typeof json === "object") {
                    setServiceConfigJson(json);
                  }
                }
              }}
              options={{
                readOnly: splitView,
              }}
            />
          </Card>
        </div>
        <ButtonBar className="self-end">
          <Toggle
            pressed={splitView}
            onClick={() => {
              setSplitView(true);
            }}
          >
            <Columns />
          </Toggle>
          <Toggle
            pressed={!splitView}
            onClick={() => {
              setSplitView(false);
            }}
          >
            <Code />
          </Toggle>
        </ButtonBar>
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
