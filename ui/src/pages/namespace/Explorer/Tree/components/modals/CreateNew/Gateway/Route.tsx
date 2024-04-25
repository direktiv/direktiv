import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Network, PlusCircle } from "lucide-react";
import { SubmitHandler, useForm } from "react-hook-form";

import Button from "~/design/Button";
import FormErrors from "~/components/FormErrors";
import Input from "~/design/Input";
import { addYamlFileExtension } from "../../../../utils";
import { defaultEndpointFileYaml } from "~/pages/namespace/Explorer/Endpoint/EndpointEditor/utils";
import { encode } from "js-base64";
import { fileNameSchema } from "~/api/tree_obsolete/schema/node";
import { pages } from "~/util/router/pages";
import { useCreateFile } from "~/api/files/mutate/createFile";
import { useNamespace } from "~/util/store/namespace";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type FormInput = {
  name: string;
  fileContent: string;
};

const NewRoute = ({
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
            message: t("pages.explorer.tree.newRoute.nameAlreadyExists"),
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
      fileContent: defaultEndpointFileYaml,
    },
  });

  const { mutate: createFile, isPending } = useCreateFile({
    onSuccess: (data) => {
      namespace &&
        navigate(
          pages.explorer.createHref({
            namespace,
            path: data.data.path,
            subpage: "endpoint",
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
        type: "endpoint",
        mimeType: "application/yaml",
        data: encode(fileContent),
      },
    });
  };

  // you can not submit if the form has not changed or if there are any errors and
  // you have already submitted the form (errors will first show up after submit)
  const disableSubmit = !isDirty || (isSubmitted && !isValid);

  const formId = `new-route-${path}`;
  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Network /> {t("pages.explorer.tree.newRoute.title")}
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
              placeholder={t("pages.explorer.tree.newRoute.namePlaceholder")}
              {...register("name")}
            />
          </fieldset>
        </form>
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.explorer.tree.newRoute.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
          type="submit"
          disabled={disableSubmit}
          loading={isPending}
          form={formId}
        >
          {!isPending && <PlusCircle />}
          {t("pages.explorer.tree.newRoute.createBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default NewRoute;
