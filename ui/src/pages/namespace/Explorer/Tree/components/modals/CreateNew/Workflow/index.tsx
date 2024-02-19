import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
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
import { fileNameSchema } from "~/api/tree/schema/node";
import { pages } from "~/util/router/pages";
import { useCreateNode } from "~/api/filesTree/mutate/createFile";
import { useNamespace } from "~/util/store/namespace";
import { useNamespaceLinting } from "~/api/namespaceLinting/query/useNamespaceLinting";
import { useNavigate } from "react-router-dom";
import { useState } from "react";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import workflowTemplates from "./templates";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type FormInput = {
  name: string;
  fileContent: string;
};

const defaultWorkflowTemplate = workflowTemplates[0];

const NewWorkflow = ({
  path,
  close,
  unallowedNames,
}: {
  path?: string;
  close: () => void;
  unallowedNames?: string[];
}) => {
  const { t } = useTranslation();
  const namespace = useNamespace();
  const navigate = useNavigate();
  const { refetch: updateNotificationBell } = useNamespaceLinting();

  const theme = useTheme();
  const [workflowData, setWorkflowData] = useState<string>(
    defaultWorkflowTemplate.data
  );

  const resolver = zodResolver(
    z.object({
      name: fileNameSchema
        .transform((enteredName) => addYamlFileExtension(enteredName))
        .refine(
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
    formState: { isDirty, errors, isValid, isSubmitted },
  } = useForm<FormInput>({
    resolver,
    defaultValues: {
      fileContent: defaultWorkflowTemplate.data,
    },
  });

  const { mutate, isLoading } = useCreateNode({
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
    mutate({
      path,
      file: {
        name,
        data: encode(fileContent),
        type: "workflow",
        mimeType: "application/direktiv",
      },
    });
  };

  // you can not submit if the form has not changed or if there are any errors and
  // you have already submitted the form (errors will first show up after submit)
  const disableSubmit = !isDirty || (isSubmitted && !isValid);

  const formId = `new-worfklow-${path}`;
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
            <label
              className="w-[100px] text-right text-[14px]"
              htmlFor="template"
            >
              {t("pages.explorer.tree.newWorkflow.templateLabel")}
            </label>
            <Select
              onValueChange={(value) => {
                const matchingWf = workflowTemplates.find(
                  (t) => t.name === value
                );
                if (matchingWf) {
                  setValue("fileContent", matchingWf.data);
                  setWorkflowData(matchingWf.data);
                }
              }}
            >
              <SelectTrigger id="template" variant="outline" block>
                <SelectValue
                  placeholder={defaultWorkflowTemplate.name}
                  defaultValue={defaultWorkflowTemplate.data}
                />
              </SelectTrigger>
              <SelectContent>
                {workflowTemplates.map((t) => (
                  <SelectItem value={t.name} key={t.name}>
                    {t.name}
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
          loading={isLoading}
          form={formId}
        >
          {!isLoading && <PlusCircle />}
          {t("pages.explorer.tree.newWorkflow.createBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default NewWorkflow;
