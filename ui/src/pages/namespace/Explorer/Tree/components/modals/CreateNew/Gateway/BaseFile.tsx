import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Network, PlusCircle } from "lucide-react";
import { SubmitHandler, useForm } from "react-hook-form";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { Editor } from "@monaco-editor/react";
import { FileNameSchema } from "~/api/files/schema";
import FormErrors from "~/components/FormErrors";
import InfoTooltip from "~/components/NamespaceEdit/InfoTooltip";
import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";
import { Textarea } from "~/design/TextArea";
import { addYamlFileExtension } from "../../../../utils";
import { defaultEndpointFileYaml } from "~/pages/namespace/Explorer/Endpoint/EndpointEditor/utils";
import { encode } from "js-base64";
import { useCreateFile } from "~/api/files/mutate/createFile";
import { useNamespace } from "~/util/store/namespace";
import { useNavigate } from "react-router-dom";
import { usePages } from "~/util/router/pages";
import { useState } from "react";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type FormInput = {
  name: string;
  fileContent: string;
};

const NewapiBaseSpec = ({
  path,
  close,
  unallowedNames,
}: {
  path?: string;
  close: () => void;
  unallowedNames?: string[];
}) => {
  const theme = useTheme();
  const [baseFileData, setBaseFileData] = useState<string>(
    defaultEndpointFileYaml
  );
  const pages = usePages();
  const { t } = useTranslation();
  const namespace = useNamespace();
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
          message: t("pages.explorer.tree.newRoute.nameAlreadyExists"), // Fix this
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
            subpage: "baseFile",
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
        type: "gateway",
        mimeType: "application/yaml",
        data: encode(fileContent),
      },
    });
  };

  // you can not submit if the form has not changed or if there are any errors and
  // you have already submitted the form (errors will first show up after submit)
  const disableSubmit = !isDirty || (isSubmitted && !isValid);

  const formId = `new-openapiBaseSpec-${path}`;
  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Network />
          NEW BaseFILE
          {/* {t("pages.explorer.tree.newRoute.title")} */}
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
                  Base Specification
                </InfoTooltip>
              </InputWithButton>
            </fieldset>
          </div>
          <div>
            <fieldset className="flex items-start gap-5">
              <Textarea className="hidden" {...register("fileContent")} />
              <Card className="h-96 w-full p-4" noShadow background="weight-1">
                <Editor
                  value={baseFileData}
                  onChange={(newData) => {
                    if (newData) {
                      setBaseFileData(newData);
                      setValue("fileContent", newData);
                    }
                  }}
                  theme={theme ?? undefined}
                />
              </Card>
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

export default NewapiBaseSpec;
