import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../../../../../design/Dialog";
import { SubmitHandler, useForm } from "react-hook-form";
import { Trans, useTranslation } from "react-i18next";

import Alert from "../../../../../design/Alert";
import Button from "../../../../../design/Button";
import Input from "../../../../../design/Input";
import { Tag } from "lucide-react";
import { TrimedRevisionSchemaType } from "../../../../../api/tree/schema";
import { useCreateTag } from "../../../../../api/tree/mutate/createTag";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type FormInput = {
  name: string;
};

const CreateTag = ({
  path,
  revision,
  close,
  unallowedNames,
}: {
  path: string;
  revision: TrimedRevisionSchemaType;
  close: () => void;
  unallowedNames: string[];
}) => {
  const { t } = useTranslation();
  const {
    register,
    handleSubmit,
    formState: { isDirty, errors, isValid, isSubmitted },
  } = useForm<FormInput>({
    resolver: zodResolver(
      z.object({
        name: z
          .string()
          .refine((name) => !unallowedNames.some((n) => n === name), {
            message: t(
              "pages.explorer.tree.workflow.revisions.tag.alreadyExist"
            ),
          }),
      })
    ),
  });

  const { mutate: createTag, isLoading } = useCreateTag({
    onSuccess: () => {
      close();
    },
  });

  const onSubmit: SubmitHandler<FormInput> = ({ name }) => {
    createTag({ path, ref: revision.name, tag: name });
  };

  // you can not submit if the form has not changed or if there are any errors and
  // you have already submitted the form (errors will first show up after submit)
  const disableSubmit = !isDirty || (isSubmitted && !isValid);

  const formId = `new-dir-${revision.name}`;
  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Tag /> {t("pages.explorer.tree.workflow.revisions.tag.titleTag")}
        </DialogTitle>
      </DialogHeader>
      <div className="my-3 flex flex-col gap-y-5">
        {!!errors.name && (
          <Alert variant="error" className="mb-5">
            <p>{errors.name.message}</p>
          </Alert>
        )}
        <div>
          <Trans
            i18nKey="pages.explorer.tree.workflow.revisions.tag.description"
            values={{ name: revision.name }}
          />
        </div>
        <form id={formId} onSubmit={handleSubmit(onSubmit)}>
          <Input {...register("name")} />
        </form>
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.explorer.tree.workflow.revisions.tag.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
          type="submit"
          disabled={disableSubmit}
          loading={isLoading}
          form={formId}
        >
          {!isLoading && <Tag />}
          {t("pages.explorer.tree.workflow.revisions.tag.createBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default CreateTag;
