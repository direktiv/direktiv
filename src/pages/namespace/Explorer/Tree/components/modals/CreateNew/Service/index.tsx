import { Code, Columns, Play, PlusCircle } from "lucide-react";
import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { SubmitHandler, useForm } from "react-hook-form";
import {
  addServiceHeader,
  defaultServiceYaml,
  serviceHeader,
  useServiceFormSchema,
} from "./config";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import FormErrors from "~/componentsNext/FormErrors";
import Input from "~/design/Input";
import { JSONSchemaForm } from "~/design/JSONschemaForm";
import { ScrollArea } from "~/design/ScrollArea";
import ServiceHelp from "~/pages/namespace/Explorer/Service/ServiceHelp";
import { Toggle } from "~/design/Toggle";
import { addYamlFileExtension } from "../../../../utils";
import { fileNameSchema } from "~/api/tree/schema/node";
import { pages } from "~/util/router/pages";
import { stringify } from "json-to-pretty-yaml";
import { twMergeClsx } from "~/util/helpers";
import { useCreateWorkflow } from "~/api/tree/mutate/createWorkflow";
import { useNamespace } from "~/util/store/namespace";
import { useNavigate } from "react-router-dom";
import { useState } from "react";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import yamljs from "js-yaml";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

export type FormInput = {
  name: string;
  fileContent: string;
};

const NewService = ({
  path,
  unallowedNames,
}: {
  path?: string;
  close: () => void;
  unallowedNames?: string[];
}) => {
  const { t } = useTranslation();
  const namespace = useNamespace();
  const navigate = useNavigate();
  const theme = useTheme();

  const [splitView, setSplitView] = useState(true);
  const [serviceConfigJson, setServiceConfigJson] = useState(serviceHeader);
  const serviceFormSchema = useServiceFormSchema();

  const {
    register,
    handleSubmit,
    setValue,
    getValues,
    formState: { isDirty, errors, isValid, isSubmitted },
  } = useForm<FormInput>({
    resolver: zodResolver(
      z.object({
        name: fileNameSchema
          .transform((enteredName) => addYamlFileExtension(enteredName))
          .refine(
            (nameWithExtension) =>
              !(unallowedNames ?? []).some(
                (unallowedName) => unallowedName === nameWithExtension
              ),
            {
              message: t("pages.explorer.tree.newService.nameAlreadyExists"),
            }
          ),
        fileContent: z.string(),
      })
    ),
    defaultValues: {
      fileContent: defaultServiceYaml,
    },
  });

  const { mutate: createService, isLoading } = useCreateWorkflow({
    onSuccess: (data) => {
      namespace &&
        navigate(
          pages.explorer.createHref({
            namespace,
            path: data.node.path,
            subpage: "service",
          })
        );
      close();
    },
  });

  const onSubmit: SubmitHandler<FormInput> = ({ name, fileContent }) => {
    createService({ path, name, fileContent });
  };

  // you can not submit if the form has not changed or if there are any errors and
  // you have already submitted the form (errors will first show up after submit)
  const disableSubmit = !isDirty || (isSubmitted && !isValid);
  const formId = `new-service-${path}`;

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Play />
          {t("pages.explorer.tree.newService.title")}
        </DialogTitle>
      </DialogHeader>
      <div className="my-5 flex flex-col gap-y-5">
        <FormErrors errors={errors} className="mb-5" />
        <form id={formId} onSubmit={handleSubmit(onSubmit)}>
          <fieldset className="flex items-center gap-5">
            <label className="w-[100px] text-right text-[14px]" htmlFor="name">
              {t("pages.explorer.tree.newService.nameLabel")}
            </label>
            <Input
              data-testid="new-workflow-name"
              id="name"
              placeholder={t("pages.explorer.tree.newService.namePlaceholder")}
              {...register("name")}
            />
          </fieldset>
        </form>
        <div className="flex flex-col gap-4">
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
                        const formDataWithHeader = addServiceHeader(e.formData);
                        setServiceConfigJson(formDataWithHeader);
                        setValue("fileContent", stringify(formDataWithHeader));
                      }
                    }}
                    schema={serviceFormSchema}
                  />
                </ScrollArea>
              </Card>
            )}
            <Card className="flex p-4" background="weight-1">
              <Editor
                value={getValues("fileContent")}
                theme={theme ?? undefined}
                onChange={(newData) => {
                  if (newData && !splitView) {
                    setValue("fileContent", newData);
                    let json;
                    try {
                      json = yamljs.load(newData);
                    } catch (e) {
                      json = null;
                    }
                    if (json && typeof json === "object") {
                      const formDataWithHeader = addServiceHeader(json);
                      setServiceConfigJson(formDataWithHeader);
                    }
                  }
                }}
                options={{
                  readOnly: splitView,
                }}
              />
            </Card>
          </div>
          <div className="flex justify-end gap-4">
            <ButtonBar>
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
            <ServiceHelp />
          </div>
        </div>
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
