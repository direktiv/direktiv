import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Play, PlusCircle } from "lucide-react";
import { SubmitHandler, useForm } from "react-hook-form";
import {
  defaultServiceYaml,
  serviceHeader,
  useServiceFormSchema,
} from "./config";

import Button from "~/design/Button";
import FormErrors from "~/components/FormErrors";
import Input from "~/design/Input";
import { addYamlFileExtension } from "../../../../utils";
import { fileNameSchema } from "~/api/tree/schema/node";
import { pages } from "~/util/router/pages";
import { useCreateWorkflow } from "~/api/tree/mutate/createWorkflow";
import { useNamespace } from "~/util/store/namespace";
import { useNavigate } from "react-router-dom";
import { useState } from "react";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
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
            message: t("pages.explorer.tree.newService.nameAlreadyExists"),
          }
        ),
      fileContent: z.string(),
    })
  );

  const {
    register,
    handleSubmit,
    setValue,
    getValues,
    formState: { isDirty, errors, isValid, isSubmitted },
  } = useForm<FormInput>({
    resolver,
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

      <div className="my-3">
        <FormErrors errors={errors} className="mb-5" />
        <form id={formId} onSubmit={handleSubmit(onSubmit)}>
          <fieldset className="flex items-center gap-5">
            <label className="w-[90px] text-right text-[14px]" htmlFor="name">
              {t("pages.explorer.tree.newRoute.nameLabel")}
            </label>
            <Input
              id="name"
              placeholder={t("pages.explorer.tree.newService.namePlaceholder")}
              {...register("name")}
            />
          </fieldset>
        </form>
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
