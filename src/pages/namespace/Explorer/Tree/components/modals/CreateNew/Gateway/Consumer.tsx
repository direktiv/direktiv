import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Network, PlusCircle } from "lucide-react";
import { SubmitHandler, useForm } from "react-hook-form";

import Button from "~/design/Button";
import FormErrors from "~/componentsNext/FormErrors";
import Input from "~/design/Input";
import { addYamlFileExtension } from "../../../../utils";
import { defaultConsumerFileYaml } from "~/pages/namespace/Explorer/Consumer/ConsumerEditor/utils";
import { fileNameSchema } from "~/api/tree/schema/node";
import { pages } from "~/util/router/pages";
import { useCreateWorkflow } from "~/api/tree/mutate/createWorkflow";
import { useNamespace } from "~/util/store/namespace";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type FormInput = {
  name: string;
  fileContent: string;
};

const NewConsumer = ({
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
  const {
    register,
    handleSubmit,
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
              message: t("pages.explorer.tree.newConsumer.nameAlreadyExists"),
            }
          ),
        fileContent: z.string(),
      })
    ),
    defaultValues: {
      fileContent: defaultConsumerFileYaml,
    },
  });

  const { mutate: createEndpoint, isLoading } = useCreateWorkflow({
    onSuccess: (data) => {
      namespace &&
        navigate(
          pages.explorer.createHref({
            namespace,
            path: data.node.path,
            subpage: "consumer",
          })
        );
      close();
    },
  });

  const onSubmit: SubmitHandler<FormInput> = ({ name, fileContent }) => {
    createEndpoint({ path, name, fileContent });
  };

  // you can not submit if the form has not changed or if there are any errors and
  // you have already submitted the form (errors will first show up after submit)
  const disableSubmit = !isDirty || (isSubmitted && !isValid);

  const formId = `new-consumer-${path}`;
  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Network /> {t("pages.explorer.tree.newConsumer.title")}
        </DialogTitle>
      </DialogHeader>

      <div className="my-3">
        <FormErrors errors={errors} className="mb-5" />
        <form id={formId} onSubmit={handleSubmit(onSubmit)}>
          <fieldset className="flex items-center gap-5">
            <label className="w-[90px] text-right text-[14px]" htmlFor="name">
              {t("pages.explorer.tree.newConsumer.nameLabel")}
            </label>
            <Input
              id="name"
              placeholder={t("pages.explorer.tree.newConsumer.namePlaceholder")}
              {...register("name")}
            />
          </fieldset>
        </form>
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.explorer.tree.newConsumer.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
          type="submit"
          disabled={disableSubmit}
          loading={isLoading}
          form={formId}
        >
          {!isLoading && <PlusCircle />}
          {t("pages.explorer.tree.newConsumer.createBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default NewConsumer;
