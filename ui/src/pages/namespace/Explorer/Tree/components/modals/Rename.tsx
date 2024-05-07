import { BaseFileSchemaType, FileNameSchema } from "~/api/files/schema";
import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { SubmitHandler, useForm } from "react-hook-form";
import { getFilenameFromPath, getParentFromPath } from "~/api/files/utils";

import Button from "~/design/Button";
import FormErrors from "~/components/FormErrors";
import Input from "~/design/Input";
import { TextCursorInput } from "lucide-react";
import { addYamlFileExtension } from "../../utils";
import { useRenameFile } from "~/api/files/mutate/renameFile";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type FormInput = {
  name: string;
};

const Rename = ({
  file,
  close,
  unallowedNames,
}: {
  file: BaseFileSchemaType;
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
        name: FileNameSchema.transform((enteredName) => {
          if (file.type !== "directory" && file.type !== "file") {
            return addYamlFileExtension(enteredName);
          }
          return enteredName;
        }).refine(
          (nameWithExtension) =>
            !unallowedNames.some(
              (unallowedName) => unallowedName === nameWithExtension
            ),
          {
            message: t("pages.explorer.tree.rename.nameAlreadyExists"),
          }
        ),
      })
    ),
    defaultValues: {
      name: getFilenameFromPath(file.path),
    },
  });

  const { mutate: rename, isPending } = useRenameFile({
    onSuccess: () => {
      close();
    },
  });

  const onSubmit: SubmitHandler<FormInput> = ({ name }) => {
    rename({
      file,
      payload: {
        path: `${getParentFromPath(file.path)}/${name}`,
      },
    });
  };

  // you can not submit if the form has not changed or if there are any errors and
  // you have already submitted the form (errors will first show up after submit)
  const disableSubmit = !isDirty || (isSubmitted && !isValid);

  const formId = `new-dir-${file.path}`;

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <TextCursorInput />
          {t("pages.explorer.tree.rename.title")}
        </DialogTitle>
      </DialogHeader>
      <div className="my-3">
        <FormErrors errors={errors} className="mb-5" />
        <form id={formId} onSubmit={handleSubmit(onSubmit)}>
          <Input {...register("name")} data-testid="node-rename-input" />
        </form>
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.explorer.tree.rename.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
          data-testid="node-rename-submit"
          type="submit"
          disabled={disableSubmit}
          loading={isPending}
          form={formId}
        >
          {!isPending && <TextCursorInput />}
          {t("pages.explorer.tree.rename.renameBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default Rename;
