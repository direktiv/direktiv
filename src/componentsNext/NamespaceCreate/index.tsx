import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../../design/Dialog";
import { Home, PlusCircle } from "lucide-react";
import { SubmitHandler, useForm } from "react-hook-form";

import Alert from "../../design/Alert";
import Button from "../../design/Button";
import Input from "../../design/Input";
import { fileNameSchema } from "../../api/tree/schema";
import { pages } from "../../util/router/pages";
import { useCreateNamespace } from "../../api/namespaces/mutate/createNamespace";
import { useListNamespaces } from "../../api/namespaces/query/get";
import { useNamespaceActions } from "../../util/store/namespace";
import { useNavigate } from "react-router-dom";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type FormInput = {
  name: string;
};

const NamespaceCreate = ({ close }: { close: () => void }) => {
  const { data } = useListNamespaces();
  const { setNamespace } = useNamespaceActions();
  const existingNamespaces = data?.results.map((n) => n.name) || [];

  const navigate = useNavigate();
  const {
    register,
    handleSubmit,
    formState: { isDirty, errors, isValid, isSubmitted },
  } = useForm<FormInput>({
    resolver: zodResolver(
      z.object({
        name: fileNameSchema.and(
          z
            .string()
            .refine((name) => !existingNamespaces.some((n) => n === name), {
              message: "The name already exists",
            })
        ),
      })
    ),
  });

  const { mutate: createNamespace, isLoading } = useCreateNamespace({
    onSuccess: (data) => {
      setNamespace(data.namespace.name);
      navigate(
        pages.explorer.createHref({
          namespace: data.namespace.name,
        })
      );
      close();
    },
  });

  const onSubmit: SubmitHandler<FormInput> = ({ name }) => {
    createNamespace({ name });
  };

  // you can not submit if the form has not changed or if there are any errors and
  // you have already submitted the form (errors will first show up after submit)
  const disableSubmit = !isDirty || (isSubmitted && !isValid);

  const formId = `new-namespace`;
  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Home /> Create a new namespace
        </DialogTitle>
      </DialogHeader>

      <div className="my-3">
        {!!errors.name && (
          <Alert variant="error" className="mb-5">
            <p>{errors.name.message}</p>
          </Alert>
        )}
        <form id={formId} onSubmit={handleSubmit(onSubmit)}>
          <fieldset className="flex items-center gap-5">
            <label className="w-[90px] text-right text-[15px]" htmlFor="name">
              Namespace
            </label>
            <Input
              id="name"
              data-testid="new-namespace-name"
              placeholder="new-namespace-name"
              {...register("name")}
            />
          </fieldset>
        </form>
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">Cancel</Button>
        </DialogClose>
        <Button
          data-testid="new-namespace-submit"
          type="submit"
          disabled={disableSubmit}
          loading={isLoading}
          form={formId}
        >
          {!isLoading && <PlusCircle />}
          Create
        </Button>
      </DialogFooter>
    </>
  );
};

export default NamespaceCreate;
