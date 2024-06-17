import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import {
  FileNameSchema,
  WorkflowType,
  workflowTypes,
} from "~/api/files/schema";
import { Play, PlusCircle } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";
import { SubmitHandler, useForm } from "react-hook-form";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import FormErrors from "~/components/FormErrors";
import Input from "~/design/Input";
import { Textarea } from "~/design/TextArea";
import { addYamlFileExtension } from "../../../../utils";
import { encode } from "js-base64";
import { useCreateFile } from "~/api/files/mutate/createFile";
import { useNamespace } from "~/util/store/namespace";
import { useNavigate } from "react-router-dom";
import { useNotifications } from "~/api/notifications/query/get";
import { usePages } from "~/util/router/pages";
import { useState } from "react";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { workflowTemplates } from "./templates";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

const defaultType: WorkflowType = "yaml";

type FormInput = {
  name: string;
  selectedTemplateName: string;
  workflowType: WorkflowType;
  fileContent: string;
};

const getFirstTemplateFor = (type: WorkflowType) => {
  const entry = Object.entries(workflowTemplates[type])[0];
  if (!entry) {
    throw Error("no!");
  }
  return entry[1];
};

const NewWorkflow = ({
  path,
  close,
  unallowedNames,
}: {
  path?: string;
  close: () => void;
  unallowedNames?: string[];
}) => {
  const pages = usePages();
  const { t } = useTranslation();
  const namespace = useNamespace();
  const navigate = useNavigate();
  const { refetch: updateNotificationBell } = useNotifications();

  const theme = useTheme();
  const [workflowData, setWorkflowData] = useState<string>(
    getFirstTemplateFor(defaultType).data
  );

  const resolver = zodResolver(
    z.object({
      name: FileNameSchema.transform((enteredName) =>
        workflowType === "typescript"
          ? enteredName // todo: add automatic extension like below
          : addYamlFileExtension(enteredName)
      ).refine(
        (nameWithExtension) =>
          !(unallowedNames ?? []).some(
            (unallowedName) => unallowedName === nameWithExtension
          ),
        {
          message: t("pages.explorer.tree.newWorkflow.nameAlreadyExists"),
        }
      ),
      fileContent: z.string(),
    })
  );

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { isDirty, errors, isValid, isSubmitted },
  } = useForm<FormInput>({
    resolver,
    defaultValues: {
      selectedTemplateName: getFirstTemplateFor(defaultType).name,
      workflowType: defaultType,
      fileContent: getFirstTemplateFor(defaultType).data,
    },
  });

  const { mutate: createFile, isPending } = useCreateFile({
    onSuccess: (data) => {
      /**
       * creating a new workflow might introduce an uninitialized secret.
       * We need to update the notification bell, to see potential new messages.
       */
      updateNotificationBell();
      namespace &&
        navigate(
          pages.explorer.createHref({
            namespace,
            path: data.data.path,
            subpage: "workflow",
          })
        );
      close();
    },
  });

  const onSubmit: SubmitHandler<FormInput> = ({ name, fileContent }) => {
    createFile({
      path,
      payload: {
        name,
        data: encode(fileContent),
        type: "workflow",
        mimeType: workflowMimeType,
      },
    });
  };

  // you can not submit if the form has not changed or if there are any errors and
  // you have already submitted the form (errors will first show up after submit)
  const disableSubmit = !isDirty || (isSubmitted && !isValid);

  const formId = `new-worfklow-${path}`;

  const workflowType: (typeof workflowTypes)[number] = watch("workflowType");

  const selectedTemplateName: string = watch("selectedTemplateName");

  const workflowMimeType =
    workflowType === "typescript"
      ? "application/x-typescript"
      : "application/yaml";

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Play />
          {t("pages.explorer.tree.newWorkflow.title")}
        </DialogTitle>
      </DialogHeader>
      <div className="my-3">
        <FormErrors errors={errors} className="mb-5" />
        <form
          id={formId}
          onSubmit={handleSubmit(onSubmit)}
          className="flex flex-col gap-y-5"
        >
          <fieldset className="flex items-center gap-5">
            <label className="w-[100px] text-right text-[14px]" htmlFor="name">
              {t("pages.explorer.tree.newWorkflow.nameLabel")}
            </label>
            <Input
              data-testid="new-workflow-name"
              id="name"
              placeholder={t("pages.explorer.tree.newWorkflow.namePlaceholder")}
              {...register("name")}
            />
          </fieldset>
          <fieldset className="flex items-center gap-5">
            <label className="w-[100px] text-right text-[14px]" htmlFor="type">
              {t("pages.explorer.tree.newWorkflow.type.label")}
            </label>
            <Select
              value={workflowType}
              onValueChange={(value: (typeof workflowTypes)[number]) => {
                setValue("workflowType", value);

                const match = getFirstTemplateFor(value);
                if (match) {
                  setValue("selectedTemplateName", match.name);
                  setValue("fileContent", match.data);
                  setWorkflowData(match.data);
                }
              }}
            >
              <SelectTrigger id="type" variant="outline" block>
                <SelectValue
                  placeholder={t(
                    "pages.explorer.tree.newWorkflow.type.placeholder"
                  )}
                  defaultValue={defaultType}
                />
              </SelectTrigger>
              <SelectContent>
                {workflowTypes.map((type) => (
                  <SelectItem value={type} key={type}>
                    {t(`pages.explorer.tree.newWorkflow.type.${type}`)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </fieldset>
          <fieldset className="flex items-center gap-5">
            <label
              className="w-[100px] text-right text-[14px]"
              htmlFor="template"
            >
              {t("pages.explorer.tree.newWorkflow.template.label")}
            </label>
            <Select
              value={selectedTemplateName}
              onValueChange={(value) => {
                const match = workflowTemplates[workflowType][value];
                if (match) {
                  setValue("selectedTemplateName", value);
                  setWorkflowData(match.data);
                }
              }}
            >
              <SelectTrigger id="template" variant="outline" block>
                <SelectValue
                  placeholder={t(
                    "pages.explorer.tree.newWorkflow.template.placeholder"
                  )}
                />
              </SelectTrigger>
              <SelectContent>
                {Object.keys(workflowTemplates[workflowType]).map((name) => (
                  <SelectItem value={name} key={name}>
                    {name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </fieldset>
          <fieldset className="flex items-start gap-5">
            <Textarea className="hidden" {...register("fileContent")} />
            <Card className="h-96 w-full p-4" noShadow background="weight-1">
              <Editor
                value={workflowData}
                onChange={(newData) => {
                  if (newData) {
                    setWorkflowData(newData);
                    setValue("fileContent", newData);
                  }
                }}
                theme={theme ?? undefined}
              />
            </Card>
          </fieldset>
        </form>
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.explorer.tree.newWorkflow.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
          data-testid="new-workflow-submit"
          type="submit"
          disabled={disableSubmit}
          loading={isPending}
          form={formId}
        >
          {!isPending && <PlusCircle />}
          {t("pages.explorer.tree.newWorkflow.createBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default NewWorkflow;
