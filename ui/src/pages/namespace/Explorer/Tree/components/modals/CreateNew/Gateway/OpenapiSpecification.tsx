import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Network, PlusCircle } from "lucide-react";
import { SubmitHandler, useForm } from "react-hook-form";

import Button from "~/design/Button";
import { FileNameSchema } from "~/api/files/schema";
import FormErrors from "~/components/FormErrors";
import InfoTooltip from "~/components/NamespaceEdit/InfoTooltip";
import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";
import { addYamlFileExtension } from "../../../../utils";
import { encode } from "js-base64";
import { jsonToYaml } from "~/pages/namespace/Explorer/utils";
import { useCreateFile } from "~/api/files/mutate/createFile";
import { useNamespace } from "~/util/store/namespace";
import { useNavigate } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type FormInput = {
  name: string;
  fileContent: string;
};

const NewOpenapiSpecification = ({
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

  // TODO: This is the minimal OpenAPI object that is used to create the base file.
  const baseOpenApiObject = {
    openapi: "3.0.0",
    info: {
      title: namespace,
      version: "1.0.0",
      description: "Minimal OpenAPI Base Specification",
    },
  };

  const defaultMinimalOpenApiYaml = jsonToYaml(baseOpenApiObject);

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
      fileContent: defaultMinimalOpenApiYaml,
    },
  });

  const { mutate: createFile, isPending } = useCreateFile({
    onSuccess: (data) => {
      namespace &&
        navigate({
          to: "/n/$namespace/explorer/openapiSpecification/$",
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
        type: "gateway",
        mimeType: "application/yaml",
        data: encode(fileContent),
      },
    });
  };

  const disableSubmit = !isDirty || (isSubmitted && !isValid);

  const formId = `new-openapiBaseSpec-${path}`;
  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Network />
          New OpenAPI Specification
        </DialogTitle>
      </DialogHeader>

      <div className="my-3 flex flex-col gap-5">
        <FormErrors errors={errors} className="mb-5" />
        <form
          id={formId}
          onSubmit={handleSubmit(onSubmit)}
          className="flex flex-col gap-5"
        >
          <div>
            <fieldset className="flex items-center gap-5">
              <label className="w-[90px] text-right text-[14px]" htmlFor="name">
                {t("pages.explorer.tree.newRoute.nameLabel")}
              </label>
              <InputWithButton>
                <Input
                  id="name"
                  placeholder={t(
                    "pages.explorer.tree.newRoute.namePlaceholder"
                  )}
                  {...register("name")}
                />
                <InfoTooltip>
                  Every namespace should only have <strong>one</strong> OpenAPI
                  Specification
                </InfoTooltip>
              </InputWithButton>
            </fieldset>
          </div>
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

export default NewOpenapiSpecification;
