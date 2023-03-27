import * as Dialog from "@radix-ui/react-dialog";

import { PlusCircle, Trash } from "lucide-react";

import Button from "../../../componentsNext/Button";
import { NodeSchemaType } from "../../../api/tree/schema";
import { useCreateDirectory } from "../../../api/tree/mutate/createDirectory";
import { useNamespace } from "../../../util/store/namespace";

const Delete = ({
  file,
  close,
}: {
  file: NodeSchemaType;
  close: () => void;
}) => {
  const namespace = useNamespace();
  const { mutate, isLoading } = useCreateDirectory({
    onSuccess: () => {
      close();
    },
  });

  const onSubmit = () => {
    console.log("🚀 delete");
    // mutate({ path, directory: name });
  };

  return (
    <form onSubmit={onSubmit}>
      <div className="text-mauve12 m-0 flex items-center gap-2 text-[17px] font-medium">
        <Trash /> Delete
      </div>
      <div className="text-mauve11 mt-[10px] mb-5 text-[15px] leading-normal">
        Are you sure you want to delete <b>{file.name}</b>? This can not be
        undone.
      </div>
      {file.type === "directory" && (
        <div className="text-mauve11 mt-[10px] mb-5 text-[15px] leading-normal">
          All content of this directoy will be deleted as well
        </div>
      )}
      <div className="flex justify-end gap-2">
        <Dialog.Close asChild>
          <Button variant="ghost">Cancel</Button>
        </Dialog.Close>
        <Button type="submit" variant="destructive" loading={isLoading}>
          {!isLoading && <PlusCircle />}
          Delete
        </Button>
      </div>
    </form>
  );
};

export default Delete;
