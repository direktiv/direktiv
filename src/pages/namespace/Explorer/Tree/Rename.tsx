import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../../../../design/Dialog";
import { NodeSchemaType, fileNameSchema } from "../../../../api/tree/schema";
import { SubmitHandler, useForm } from "react-hook-form";

import Alert from "../../../../design/Alert";
import Button from "../../../../design/Button";
import Input from "../../../../design/Input";
import { TextCursorInput } from "lucide-react";
import { useRenameNode } from "../../../../api/tree/mutate/renameNode";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type FormInput = {
  name: string;
};

const Rename = ({
  node,
  close,
  unallowedNames,
}: {
  node: NodeSchemaType;
  close: () => void;
  unallowedNames: string[];
}) => {
  const {
    register,
    handleSubmit,
    formState: { isDirty, errors, isValid, isSubmitted },
  } = useForm<FormInput>({
    resolver: zodResolver(
      z.object({
        name: fileNameSchema.and(
          z.string().refine((name) => !unallowedNames.some((n) => n === name), {
            message: "The name already exists",
          })
        ),
      })
    ),
    defaultValues: {
      name: node.name,
    },
  });

  const { mutate: rename, isLoading } = useRenameNode({
    onSuccess: () => {
      close();
    },
  });

  const onSubmit: SubmitHandler<FormInput> = ({ name }) => {
    rename({ node, newName: name });
  };

  // you can not submit if the form has not changed or if there are any errors and
  // you have already submitted the form (errors will first show up after submit)
  const disableSubmit = !isDirty || (isSubmitted && !isValid);

  const formId = `new-dir-${node.path}`;
  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <TextCursorInput /> Rename
        </DialogTitle>
      </DialogHeader>
      <div className="my-3">
        {!!errors.name && (
          <Alert variant="error" className="mb-5">
            <p>{errors.name.message}</p>
          </Alert>
        )}
        <form id={formId} onSubmit={handleSubmit(onSubmit)}>
          <Input {...register("name")} data-testid="node-rename-input" />
        </form>
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">Cancel</Button>
        </DialogClose>
        <Button
          data-testid="node-rename-submit"
          type="submit"
          disabled={disableSubmit}
          loading={isLoading}
          form={formId}
        >
          {!isLoading && <TextCursorInput />}
          Rename
        </Button>
      </DialogFooter>
    </>
  );
};

export default Rename;
