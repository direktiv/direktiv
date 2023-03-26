import * as Dialog from "@radix-ui/react-dialog";

import { Folder, PlusCircle } from "lucide-react";
import { SubmitHandler, useForm } from "react-hook-form";

import Button from "../../../componentsNext/Button";
import clsx from "clsx";
import { useCreateDirectory } from "../../../api/tree/mutate/createDirectory";

// TODO: this must be infered from the api input type
type FormInput = {
  name: string;
};

const NewDirectory = ({ path }: { path?: string }) => {
  const {
    register,
    handleSubmit,
    formState: { isDirty },
  } = useForm<FormInput>({
    // resolver: zodResolver(someSchame),resolver: zodResolver(), // TODO: may add zod resolver
    defaultValues: {},
  });

  const { mutate, isLoading } = useCreateDirectory();

  const onSubmit: SubmitHandler<FormInput> = ({ name }) => {
    mutate({ path, directory: name });
  };

  return (
    <Dialog.Content
      className={clsx(
        "fixed z-50 grid w-full gap-2 rounded-b-lg bg-base-100 p-6 shadow-md animate-in data-[state=open]:fade-in-90 data-[state=open]:slide-in-from-bottom-10 sm:max-w-[425px] sm:rounded-lg sm:zoom-in-90 data-[state=open]:sm:slide-in-from-bottom-0"
      )}
    >
      <form onSubmit={handleSubmit(onSubmit)}>
        <div className="text-mauve12 m-0 flex gap-2 text-[17px] font-medium">
          <Folder /> Create a new Folder
        </div>
        <div className="text-mauve11 mt-[10px] mb-5 text-[15px] leading-normal">
          Please enter the name of the new folder.
        </div>
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
            defaultValue="Folder Name"
            {...register("name")}
          />
        </fieldset>
        <div className="flex justify-end gap-2">
          <Dialog.Close asChild>
            <Button variant="ghost">Cancel</Button>
          </Dialog.Close>
          <Button type="submit" disabled={!isDirty} loading={isLoading}>
            {!isLoading && <PlusCircle />}
            Create
          </Button>
        </div>
      </form>
    </Dialog.Content>
  );
};

export default NewDirectory;
