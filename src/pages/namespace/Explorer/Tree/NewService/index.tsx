import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Play, PlusCircle } from "lucide-react";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { JSONSchemaForm } from "~/design/JSONschemaForm";
import { RJSFSchema } from "@rjsf/utils";
import { ScrollArea } from "~/design/ScrollArea";
import prettyYAML from "json-to-pretty-yaml";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const serviceSchema: RJSFSchema = {
  properties: {
    image: {
      title: "Image",
      type: "string",
    },
    scale: {
      title: "Scale",
      type: "integer",
    },
    size: {
      title: "size",
      type: "integer",
    },
    cmd: {
      title: "Cmd",
      type: "string",
    },
  },
  required: ["name"],
  type: "object",
};

// type FormInput = {
//   name: string;
//   fileContent: string;
// };

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
  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Play />
          {t("pages.explorer.tree.newService.title")}
        </DialogTitle>
      </DialogHeader>

      <Card className="h-96 w-full p-4 sm:h-[500px]">
        <code className="block">{JSON.stringify(serviceConfig)}</code>

        <code className="block">{prettyYAML.stringify(serviceConfig)}</code>
        <ScrollArea className="h-full">
          <JSONSchemaForm
            formData={{
              image: "coo",
            }}
            onChange={(e) => {
              console.log("🚀", JSON.stringify(e.formData));
              console.log(prettyYAML.stringify(e.formData));
              setServiceConfig(e.formData);
              // setServiceConfig(prettyYAML.stringify(e.formData));
            }}
            schema={serviceSchema}
          />
        </ScrollArea>
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
