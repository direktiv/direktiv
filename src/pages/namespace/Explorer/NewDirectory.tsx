import * as Dialog from "@radix-ui/react-dialog";

import { Folder, PlusCircle } from "lucide-react";
import { SubmitHandler, useForm } from "react-hook-form";

import Alert from "../../../componentsNext/Alert";
import Button from "../../../componentsNext/Button";
import { fileNameSchema } from "../../../api/tree/schema";
import { pages } from "../../../util/router/pages";
import { useCreateDirectory } from "../../../api/tree/mutate/createDirectory";
import { useNamespace } from "../../../util/store/namespace";
import { useNavigate } from "react-router-dom";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type FormInput = {
  name: string;
};

const NewDirectory = ({
  path,
  close,
  unallowedNames,
}: {
  path?: string;
  close: () => void;
  unallowedNames: string[];
}) => {
  const namespace = useNamespace();
  const navigate = useNavigate();
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
  });

  const { mutate, isLoading } = useCreateDirectory({
    onSuccess: (data) => {
      namespace &&
        navigate(
          pages.explorer.createHref({ namespace, path: data.node.path })
        );
      close();
    },
  });

  const onSubmit: SubmitHandler<FormInput> = ({ name }) => {
    mutate({ path, directory: name });
  };

  // you can not submit if the form has not changed or if there are any errors and
  // you have already submitted the form (errors will first show up after submit)
  const disableSubmit = !isDirty || (isSubmitted && !isValid);

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <div className="text-mauve12 m-0 flex items-center gap-2 text-[17px] font-medium ">
        <Folder /> Create a new Folder
      </div>
      <div className="text-mauve11 mt-[10px] mb-5 text-[15px] leading-normal">
        Please enter the name of the new folder.
      </div>

      {!!errors.name && (
        <Alert variant="error" className="mb-5">
          <p>{errors.name.message}</p>
        </Alert>
      )}

      <fieldset className="mb-[15px] flex items-center gap-5">
        <label
          className="text-violet11 w-[90px] text-right text-[15px]"
          htmlFor="name"
        >
          Name
        </label>
        <input
          className="text-violet11 shadow-violet7 focus:shadow-violet8 inline-flex h-[35px] w-full flex-1 items-center justify-center rounded-[4px] px-[10px] text-[15px] leading-none shadow-[0_0_0_1px] outline-none focus:shadow-[0_0_0_2px]"
          id="name"
          placeholder="folder-name"
          {...register("name")}
        />
      </fieldset>
      <div className="flex justify-end gap-2">
        <Dialog.Close asChild>
          <Button variant="ghost">Cancel</Button>
        </Dialog.Close>
        <Button type="submit" disabled={disableSubmit} loading={isLoading}>
          {!isLoading && <PlusCircle />}
          Create
        </Button>
      </div>
    </form>
  );
};

export default NewDirectory;
