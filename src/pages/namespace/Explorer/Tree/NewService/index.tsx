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
import { useState } from "react";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";

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
  const [serviceConfig, setServiceConfig] = useState({
    direktiv_api: "service/v1",
  });
  const formId = `new-service-${path}`;
  const theme = useTheme();

  /**
   * TODO:
   * [x] disable editor
   * [x] style headlines (with icons?)
   * - JSONSchemaForm onChange,
   * - add a zod schema that will
   *   - validate the formdata
   *   - parse out empty strings
   *   - handles missing fields (or dies the schema do this?)
   * - add a toggle to use code only
   */

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Play />
          {t("pages.explorer.tree.newService.title")}
        </DialogTitle>
      </DialogHeader>
      <div className="flex flex-col gap-5">
        <Card className="h-auto w-full p-4" noShadow background="weight-1">
          <ScrollArea className="h-full">
            <JSONSchemaForm
              formData={serviceConfig}
              onChange={(e) => {
                if (e.formData) {
                  setServiceConfig(e.formData);
                }
              }}
              schema={serviceFormSchema}
            />
          </ScrollArea>
        </Card>
        <Card className="h-96 w-full p-4" noShadow background="weight-1">
          <Editor
            value={jsonToYaml(serviceConfig)}
            theme={theme ?? undefined}
            options={{
              readOnly: true,
            }}
          />
        </Card>
        <ButtonBar>
          <Toggle
            pressed={!splitView}
            onClick={() => {
              setSplitView(false);
            }}
          >
            <Code />
          </Toggle>
          <Toggle
            pressed={splitView}
            onClick={() => {
              setSplitView(true);
            }}
          >
            <Columns />
          </Toggle>
        </ButtonBar>
      </div>
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
