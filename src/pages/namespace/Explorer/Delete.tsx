import * as Dialog from "@radix-ui/react-dialog";

import { PlusCircle, Trash } from "lucide-react";

import Button from "../../../componentsNext/Button";
import { NodeSchemaType } from "../../../api/tree/schema";
import { useDeleteNode } from "../../../api/tree/mutate/deleteNode";

const Delete = ({
  node,
  close,
}: {
  node: NodeSchemaType;
  close: () => void;
}) => {
  const { mutate, isLoading } = useDeleteNode({
    onSuccess: () => {
      close();
    },
  });

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        mutate({ node });
      }}
    >
      <div className="text-mauve12 m-0 flex items-center gap-2 text-[17px] font-medium">
        <Trash /> Delete
      </div>
      <div className="text-mauve11 mt-[10px] mb-5 text-[15px] leading-normal">
        Are you sure you want to delete <b>{node.name}</b>? This can not be
        undone.
      </div>
      {node.type === "directory" && (
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
