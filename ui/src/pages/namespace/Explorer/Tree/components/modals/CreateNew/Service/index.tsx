import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Play, PlusCircle } from "lucide-react";
import { SubmitHandler, useForm } from "react-hook-form";

import Button from "~/design/Button";
import { FileNameSchema } from "~/api/files/schema";
import FormErrors from "~/components/FormErrors";
import Input from "~/design/Input";
import { addYamlFileExtension } from "../../../../utils";
import { defaultServiceYaml } from "~/pages/namespace/Explorer/Service/ServiceEditor/utils";
import { encode } from "js-base64";
import { useCreateFile } from "~/api/files/mutate/createFile";
import { useNavigate } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type FormInput = {
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
  const navigate = useNavigate();

  const resolver = zodResolver(
    z.object({
      name: FileNameSchema.transform((enteredName) =>
        addYamlFileExtension(enteredName)
      ).refine(
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
    formState: { isDirty, errors, isValid, isSubmitted },
  } = useForm<FormInput>({
    resolver,
    defaultValues: {
      fileContent: defaultServiceYaml,
    },
  });

  const { mutate: createFile, isPending } = useCreateFile({
    onSuccess: (data) => {
      navigate({
        to: "/n/$namespace/explorer/service/$",
        from: "/n/$namespace",
        params: { _splat: data.data.path },
      });
      close();
    },
  });

  const onSubmit: SubmitHandler<FormInput> = ({ name, fileContent }) => {
    createFile({
      path,
      payload: {
        name,
        mimeType: "application/yaml",
        type: "service",
        data: encode(fileContent),
      },
    });
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
          loading={isPending}
          form={formId}
        >
          {!isPending && <PlusCircle />}
          {t("pages.explorer.tree.newService.createBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default NewService;
